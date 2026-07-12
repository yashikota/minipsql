// Package host adapts pgrust's generated import ABI to minipsql's in-memory VFS.
package host

import (
	"encoding/binary"
	"os"
	"strings"

	"github.com/yashikota/minipsql/internal/generated/pgrust/base"
	"github.com/yashikota/minipsql/internal/pgvfs"
)

const errFault int64 = -14

type Pgvfs struct {
	Host *pgvfs.Host
	// Root is the absolute guest data directory mapped to the memfs root.
	Root string
}

func memory(m *base.Module, ptr, size int64) ([]byte, bool) {
	if ptr < 0 || size < 0 || uint64(ptr) > uint64(len(m.Memory)) || uint64(size) > uint64(len(m.Memory))-uint64(ptr) {
		return nil, false
	}
	return m.Memory[ptr : ptr+size], true
}

func guestString(m *base.Module, ptr, size int64) (string, bool) {
	b, ok := memory(m, ptr, size)
	return string(b), ok
}

func (p Pgvfs) guestPath(m *base.Module, ptr, size int64) (string, bool) {
	name, ok := guestString(m, ptr, size)
	if !ok {
		return "", false
	}
	root := strings.TrimSuffix(p.Root, "/")
	if root != "" && name == root {
		return ".", true
	}
	if root != "" && strings.HasPrefix(name, root+"/") {
		return strings.TrimPrefix(name, root+"/"), true
	}
	return strings.TrimPrefix(name, "/"), true
}

func (p Pgvfs) Host_now_ns(*base.Module) int64            { return p.Host.NowUnixNanos() }
func (p Pgvfs) Host_close(_ *base.Module, fd int32) int64 { return p.Host.Close(fd) }
func (p Pgvfs) Host_fsync(_ *base.Module, fd int32) int64 { return p.Host.Sync(fd) }

func (p Pgvfs) write(m *base.Module, fd int32, ptr, size int64) int64 {
	b, ok := memory(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.Write(fd, b)
}
func (p Pgvfs) read(m *base.Module, fd int32, ptr, size int64) int64 {
	b, ok := memory(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.Read(fd, b)
}
func (p Pgvfs) Host_stdout(m *base.Module, ptr, size int64) int64 { return p.write(m, 1, ptr, size) }
func (p Pgvfs) Host_stderr(m *base.Module, ptr, size int64) int64 { return p.write(m, 2, ptr, size) }
func (p Pgvfs) Host_stdin(m *base.Module, ptr, size int64) int64  { return p.read(m, 0, ptr, size) }
func (p Pgvfs) Host_read(m *base.Module, fd int32, ptr, size int64) int64 {
	return p.read(m, fd, ptr, size)
}
func (p Pgvfs) Host_write(m *base.Module, fd int32, ptr, size int64) int64 {
	return p.write(m, fd, ptr, size)
}

func (p Pgvfs) Host_pread(m *base.Module, fd int32, ptr, size, off int64) int64 {
	b, ok := memory(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.PRead(fd, b, off)
}
func (p Pgvfs) Host_pwrite(m *base.Module, fd int32, ptr, size, off int64) int64 {
	b, ok := memory(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.PWrite(fd, b, off)
}
func (p Pgvfs) Host_lseek(_ *base.Module, fd int32, off int64, whence int32) int64 {
	return p.Host.Seek(fd, off, whence)
}
func (p Pgvfs) Host_ftruncate(_ *base.Module, fd int32, size int64) int64 {
	return p.Host.Truncate(fd, size)
}

func (p Pgvfs) Host_open(m *base.Module, ptr, size int64, flags, mode int32) int64 {
	name, ok := p.guestPath(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.Open(name, flags, uint32(mode))
}
func (p Pgvfs) Host_unlink(m *base.Module, ptr, size int64) int64 {
	name, ok := p.guestPath(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.Unlink(name)
}
func (p Pgvfs) Host_mkdir(m *base.Module, ptr, size int64, mode int32) int64 {
	name, ok := p.guestPath(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.Mkdir(name, uint32(mode))
}
func (p Pgvfs) Host_rmdir(m *base.Module, ptr, size int64) int64 {
	name, ok := p.guestPath(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.Rmdir(name)
}
func (p Pgvfs) Host_access(m *base.Module, ptr, size int64, _ int32) int64 {
	name, ok := p.guestPath(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.Access(name)
}
func (p Pgvfs) Host_rename(m *base.Module, fromPtr, fromSize, toPtr, toSize int64) int64 {
	from, ok := p.guestPath(m, fromPtr, fromSize)
	if !ok {
		return errFault
	}
	to, ok := p.guestPath(m, toPtr, toSize)
	if !ok {
		return errFault
	}
	return p.Host.Rename(from, to)
}
func (p Pgvfs) Host_readlink(*base.Module, int64, int64, int64, int64) int64 { return -22 }

func writeStat(m *base.Module, ptr int64, info os.FileInfo) int64 {
	b, ok := memory(m, ptr, 64)
	if !ok {
		return errFault
	}
	mode := uint64(info.Mode().Perm()) | 0o100000
	if info.IsDir() {
		mode = uint64(info.Mode().Perm()) | 0o040000
	}
	words := [8]int64{int64(mode), 0, info.Size(), info.ModTime().Unix(), 1, 1, 1, (info.Size() + 511) / 512}
	for i, word := range words {
		binary.LittleEndian.PutUint64(b[i*8:], uint64(word))
	}
	return 0
}
func (p Pgvfs) Host_stat(m *base.Module, pathPtr, pathSize int64, _ int32, out int64) int64 {
	name, ok := p.guestPath(m, pathPtr, pathSize)
	if !ok {
		return errFault
	}
	info, result := p.Host.Stat(name)
	if result < 0 {
		return result
	}
	return writeStat(m, out, info)
}
func (p Pgvfs) Host_fstat(m *base.Module, fd int32, out int64) int64 {
	info, result := p.Host.FStat(fd)
	if result < 0 {
		return result
	}
	return writeStat(m, out, info)
}

func (p Pgvfs) Host_opendir(m *base.Module, ptr, size int64) int64 {
	name, ok := p.guestPath(m, ptr, size)
	if !ok {
		return errFault
	}
	return p.Host.OpenDir(name)
}
func (p Pgvfs) Host_readdir(m *base.Module, fd int32, ptr, size int64) int64 {
	name, result := p.Host.ReadDir(fd)
	if result <= 0 {
		return result
	}
	b, ok := memory(m, ptr, size)
	if !ok {
		return errFault
	}
	copy(b, name)
	if int64(len(name)) > size {
		return size
	}
	return int64(len(name))
}
func (p Pgvfs) Host_closedir(_ *base.Module, fd int32) int64 { return p.Host.Close(fd) }

func (p Pgvfs) Host_argc(*base.Module) int64 { return p.Host.Argc() }
func (p Pgvfs) Host_argv(m *base.Module, index int32, ptr, size int64) int64 {
	arg, ok := p.Host.Arg(index)
	if !ok {
		return -22
	}
	b, ok := memory(m, ptr, size)
	if !ok {
		return errFault
	}
	copy(b, arg)
	return int64(len(arg))
}
func (p Pgvfs) Host_proc_exit(_ *base.Module, code int32) { panic(pgvfs.ExitError{Code: code}) }
