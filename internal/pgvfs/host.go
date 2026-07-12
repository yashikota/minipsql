// Package pgvfs implements pgrust's custom host-import filesystem ABI.
package pgvfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/yashikota/minipsql/internal/memfs"
)

const (
	errPermission = 1
	errNoEntry    = 2
	errIO         = 5
	errBadFD      = 9
	errAccess     = 13
	errExists     = 17
	errNotDir     = 20
	errIsDir      = 21
	errInvalid    = 22
	errNotEmpty   = 39
)

// Host owns the descriptors and process-like streams for one guest backend.
// Multiple hosts may share one FS while retaining independent descriptor tables.
type Host struct {
	mu     sync.Mutex
	fs     *memfs.FS
	fds    map[int32]*memfs.File
	nextFD int32
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	args   []string
	now    func() time.Time
}

type Options struct {
	FS     *memfs.FS
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Args   []string
}

func New(opts Options) *Host {
	if opts.FS == nil {
		opts.FS = memfs.New()
	}
	if opts.Stdin == nil {
		opts.Stdin = emptyReader{}
	}
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	return &Host{
		fs: opts.FS, fds: make(map[int32]*memfs.File), nextFD: 3,
		stdin: opts.Stdin, stdout: opts.Stdout, stderr: opts.Stderr,
		args: append([]string(nil), opts.Args...), now: time.Now,
	}
}

type emptyReader struct{}

func (emptyReader) Read([]byte) (int, error) { return 0, io.EOF }

func (h *Host) FS() *memfs.FS { return h.fs }

// Open implements pgvfs.host_open and returns an fd or a negative Linux errno.
func (h *Host) Open(name string, flags int32, mode uint32) int64 {
	file, err := h.fs.OpenFile(name, openFlags(flags), os.FileMode(mode&0o777))
	if err != nil {
		return errno(err)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	fd := h.nextFD
	h.nextFD++
	h.fds[fd] = file
	return int64(fd)
}

func openFlags(flags int32) int {
	var out int
	switch flags & 3 {
	case 1:
		out = os.O_WRONLY
	case 2:
		out = os.O_RDWR
	default:
		out = os.O_RDONLY
	}
	if flags&0o100 != 0 {
		out |= os.O_CREATE
	}
	if flags&0o200 != 0 {
		out |= os.O_EXCL
	}
	if flags&0o1000 != 0 {
		out |= os.O_TRUNC
	}
	if flags&0o2000 != 0 {
		out |= os.O_APPEND
	}
	return out
}

func (h *Host) file(fd int32) (*memfs.File, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	f, ok := h.fds[fd]
	return f, ok
}

func (h *Host) Close(fd int32) int64 {
	h.mu.Lock()
	f, ok := h.fds[fd]
	if ok {
		delete(h.fds, fd)
	}
	h.mu.Unlock()
	if !ok {
		return -errBadFD
	}
	if err := f.Close(); err != nil {
		return errno(err)
	}
	return 0
}

func (h *Host) Read(fd int32, buffer []byte) int64 {
	if fd == 0 {
		n, err := h.stdin.Read(buffer)
		return ioResult(n, err)
	}
	f, ok := h.file(fd)
	if !ok {
		return -errBadFD
	}
	n, err := f.Read(buffer)
	return ioResult(n, err)
}

func (h *Host) Write(fd int32, buffer []byte) int64 {
	if fd == 1 {
		n, err := h.stdout.Write(buffer)
		return ioResult(n, err)
	}
	if fd == 2 {
		n, err := h.stderr.Write(buffer)
		return ioResult(n, err)
	}
	f, ok := h.file(fd)
	if !ok {
		return -errBadFD
	}
	n, err := f.Write(buffer)
	return ioResult(n, err)
}

func (h *Host) PRead(fd int32, buffer []byte, offset int64) int64 {
	f, ok := h.file(fd)
	if !ok {
		return -errBadFD
	}
	n, err := f.ReadAt(buffer, offset)
	return ioResult(n, err)
}

func (h *Host) PWrite(fd int32, buffer []byte, offset int64) int64 {
	f, ok := h.file(fd)
	if !ok {
		return -errBadFD
	}
	n, err := f.WriteAt(buffer, offset)
	return ioResult(n, err)
}

func (h *Host) Seek(fd int32, offset int64, whence int32) int64 {
	f, ok := h.file(fd)
	if !ok {
		return -errBadFD
	}
	position, err := f.Seek(offset, int(whence))
	if err != nil {
		return errno(err)
	}
	return position
}

func (h *Host) Sync(fd int32) int64 {
	f, ok := h.file(fd)
	if !ok {
		return -errBadFD
	}
	if err := f.Sync(); err != nil {
		return errno(err)
	}
	return 0
}

func (h *Host) Truncate(fd int32, size int64) int64 {
	f, ok := h.file(fd)
	if !ok {
		return -errBadFD
	}
	if err := f.Truncate(size); err != nil {
		return errno(err)
	}
	return 0
}

func (h *Host) Stat(name string) (os.FileInfo, int64) {
	info, err := h.fs.Stat(name)
	if err != nil {
		return nil, errno(err)
	}
	return info, 0
}

func (h *Host) FStat(fd int32) (os.FileInfo, int64) {
	f, ok := h.file(fd)
	if !ok {
		return nil, -errBadFD
	}
	info, err := f.Stat()
	if err != nil {
		return nil, errno(err)
	}
	return info, 0
}

func (h *Host) Unlink(name string) int64 { return result(h.fs.Remove(name)) }
func (h *Host) Mkdir(name string, mode uint32) int64 {
	return result(h.fs.Mkdir(name, os.FileMode(mode&0o777)))
}

func (h *Host) Rmdir(name string) int64 {
	info, err := h.fs.Stat(name)
	if err != nil {
		return errno(err)
	}
	if !info.IsDir() {
		return -errNotDir
	}
	return result(h.fs.Remove(name))
}

func (h *Host) Rename(oldName, newName string) int64 {
	return result(h.fs.Rename(oldName, newName))
}

func (h *Host) Access(name string) int64 {
	_, err := h.fs.Stat(name)
	return result(err)
}

func (h *Host) OpenDir(name string) int64 {
	info, err := h.fs.Stat(name)
	if err != nil {
		return errno(err)
	}
	if !info.IsDir() {
		return -errNotDir
	}
	return h.Open(name, 0, 0)
}

func (h *Host) ReadDir(fd int32) (string, int64) {
	f, ok := h.file(fd)
	if !ok {
		return "", -errBadFD
	}
	entries, err := f.ReadDir(1)
	if errors.Is(err, io.EOF) {
		return "", 0
	}
	if err != nil {
		return "", errno(err)
	}
	return entries[0].Name(), int64(len(entries[0].Name()))
}

func (h *Host) Arg(index int32) (string, bool) {
	if index < 0 || int(index) >= len(h.args) {
		return "", false
	}
	return h.args[index], true
}

func (h *Host) Argc() int64 { return int64(len(h.args)) }

func (h *Host) NowUnixNanos() int64 { return h.now().UnixNano() }

func ioResult(n int, err error) int64 {
	if err == nil || errors.Is(err, io.EOF) && n > 0 {
		return int64(n)
	}
	if errors.Is(err, io.EOF) {
		return 0
	}
	return errno(err)
}

func result(err error) int64 {
	if err == nil {
		return 0
	}
	return errno(err)
}

func errno(err error) int64 {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, fs.ErrNotExist):
		return -errNoEntry
	case errors.Is(err, fs.ErrExist):
		return -errExists
	case errors.Is(err, fs.ErrPermission):
		return -errAccess
	case errors.Is(err, fs.ErrClosed):
		return -errBadFD
	case stringsContains(err.Error(), "not a directory"):
		return -errNotDir
	case stringsContains(err.Error(), "is a directory"):
		return -errIsDir
	case stringsContains(err.Error(), "not empty"):
		return -errNotEmpty
	case stringsContains(err.Error(), "invalid"), stringsContains(err.Error(), "negative"):
		return -errInvalid
	default:
		return -errIO
	}
}

func stringsContains(value, fragment string) bool {
	for index := 0; index+len(fragment) <= len(value); index++ {
		if value[index:index+len(fragment)] == fragment {
			return true
		}
	}
	return false
}

// ExitError is raised by the generated adapter when pgvfs.host_proc_exit is
// invoked. It is recovered at the session boundary, never at an arbitrary SQL
// frame.
type ExitError struct{ Code int32 }

func (e ExitError) Error() string { return "pgrust guest exited" }

var _ error = ExitError{}

// Keep errPermission documented even though pgrust currently maps access
// denials to EACCES rather than EPERM.
var _ = errPermission
