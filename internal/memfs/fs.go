// Package memfs implements the instance-owned filesystem used by pgrust.
package memfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

// FS is a concurrent in-memory filesystem. Paths are slash-separated and
// relative to the filesystem root.
type FS struct {
	mu    sync.RWMutex
	nodes map[string]*node
	now   func() time.Time
}

type node struct {
	name    string
	mode    os.FileMode
	modTime time.Time
	data    []byte
}

func (n *node) isDir() bool { return n.mode.IsDir() }

// New returns an empty filesystem containing only its root directory.
func New() *FS {
	now := time.Now
	return &FS{
		nodes: map[string]*node{"": {mode: os.ModeDir | 0o700, modTime: now()}},
		now:   now,
	}
}

// Load copies a read-only filesystem tree into this filesystem. The copy is
// independent: later writes never mutate the source fixture.
func (m *FS) Load(source fs.FS) error {
	return fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == "." {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return m.Mkdir(name, info.Mode().Perm())
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		file, err := m.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := file.Write(data); err != nil {
			_ = file.Close()
			return err
		}
		return file.Close()
	})
}

func clean(name string) (string, error) {
	if name == "" || name == "." {
		return "", nil
	}
	if strings.HasPrefix(name, "/") {
		return "", fs.ErrPermission
	}
	name = path.Clean(name)
	if name == ".." || strings.HasPrefix(name, "../") {
		return "", fs.ErrPermission
	}
	return name, nil
}

func parent(name string) string {
	if index := strings.LastIndexByte(name, '/'); index >= 0 {
		return name[:index]
	}
	return ""
}

// Mkdir creates one directory. Its parent must already exist.
func (m *FS) Mkdir(name string, perm os.FileMode) error {
	name, err := clean(name)
	if err != nil {
		return err
	}
	if name == "" {
		return fs.ErrExist
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.nodes[name]; ok {
		return fs.ErrExist
	}
	p, ok := m.nodes[parent(name)]
	if !ok {
		return fs.ErrNotExist
	}
	if !p.isDir() {
		return errors.New("memfs: parent is not a directory")
	}
	m.nodes[name] = &node{name: path.Base(name), mode: os.ModeDir | perm.Perm(), modTime: m.now()}
	return nil
}

// OpenFile opens or creates a regular file using os.OpenFile flags.
func (m *FS) OpenFile(name string, flag int, perm os.FileMode) (*File, error) {
	name, err := clean(name)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[name]
	if !ok {
		if flag&os.O_CREATE == 0 {
			return nil, fs.ErrNotExist
		}
		p, exists := m.nodes[parent(name)]
		if !exists {
			return nil, fs.ErrNotExist
		}
		if !p.isDir() {
			return nil, errors.New("memfs: parent is not a directory")
		}
		n = &node{name: path.Base(name), mode: perm.Perm(), modTime: m.now()}
		m.nodes[name] = n
	} else if flag&(os.O_CREATE|os.O_EXCL) == os.O_CREATE|os.O_EXCL {
		return nil, fs.ErrExist
	}

	writable := flag&(os.O_WRONLY|os.O_RDWR) != 0
	readable := flag&os.O_WRONLY == 0
	if n.isDir() && writable {
		return nil, errors.New("memfs: is a directory")
	}
	if flag&os.O_TRUNC != 0 {
		if !writable {
			return nil, fs.ErrPermission
		}
		n.data = nil
		n.modTime = m.now()
	}
	offset := int64(0)
	if flag&os.O_APPEND != 0 {
		offset = int64(len(n.data))
	}
	return &File{fsys: m, path: name, node: n, offset: offset, readable: readable, writable: writable, append: flag&os.O_APPEND != 0}, nil
}

// Stat returns metadata for a path.
func (m *FS) Stat(name string) (os.FileInfo, error) {
	name, err := clean(name)
	if err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.nodes[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return infoOf(n), nil
}

// Remove deletes an empty directory or one regular file.
func (m *FS) Remove(name string) error {
	name, err := clean(name)
	if err != nil {
		return err
	}
	if name == "" {
		return fs.ErrPermission
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[name]
	if !ok {
		return fs.ErrNotExist
	}
	if n.isDir() {
		prefix := name + "/"
		for candidate := range m.nodes {
			if strings.HasPrefix(candidate, prefix) {
				return errors.New("memfs: directory not empty")
			}
		}
	}
	delete(m.nodes, name)
	return nil
}

// Rename atomically moves a file or directory tree.
func (m *FS) Rename(oldName, newName string) error {
	oldName, err := clean(oldName)
	if err != nil {
		return err
	}
	newName, err = clean(newName)
	if err != nil {
		return err
	}
	if oldName == "" || newName == "" || strings.HasPrefix(newName+"/", oldName+"/") {
		return fs.ErrPermission
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[oldName]
	if !ok {
		return fs.ErrNotExist
	}
	p, ok := m.nodes[parent(newName)]
	if !ok {
		return fs.ErrNotExist
	}
	if !p.isDir() {
		return errors.New("memfs: destination parent is not a directory")
	}
	if existing, ok := m.nodes[newName]; ok {
		if existing.isDir() != n.isDir() {
			return errors.New("memfs: incompatible destination")
		}
		if existing.isDir() {
			prefix := newName + "/"
			for candidate := range m.nodes {
				if strings.HasPrefix(candidate, prefix) {
					return errors.New("memfs: destination directory not empty")
				}
			}
		}
		delete(m.nodes, newName)
	}

	moves := make(map[string]*node)
	for candidate, child := range m.nodes {
		if candidate == oldName || strings.HasPrefix(candidate, oldName+"/") {
			suffix := strings.TrimPrefix(candidate, oldName)
			moves[newName+suffix] = child
			delete(m.nodes, candidate)
		}
	}
	n.name = path.Base(newName)
	for candidate, child := range moves {
		m.nodes[candidate] = child
	}
	return nil
}

// Clone returns a deep, independently writable copy of the filesystem.
func (m *FS) Clone() *FS {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clone := &FS{nodes: make(map[string]*node, len(m.nodes)), now: m.now}
	for name, n := range m.nodes {
		copyNode := *n
		copyNode.data = append([]byte(nil), n.data...)
		clone.nodes[name] = &copyNode
	}
	return clone
}

// File is an open file description.
type File struct {
	mu       sync.Mutex
	fsys     *FS
	path     string
	node     *node
	offset   int64
	dirIndex int
	readable bool
	writable bool
	append   bool
	closed   bool
}

func (f *File) check() error {
	if f.closed {
		return fs.ErrClosed
	}
	return nil
}

func (f *File) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.check(); err != nil {
		return 0, err
	}
	if !f.readable || f.node.isDir() {
		return 0, fs.ErrPermission
	}
	f.fsys.mu.RLock()
	defer f.fsys.mu.RUnlock()
	if f.offset >= int64(len(f.node.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.node.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *File) ReadAt(p []byte, offset int64) (int, error) {
	if offset < 0 {
		return 0, errors.New("memfs: negative offset")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fsys.mu.RLock()
	defer f.fsys.mu.RUnlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	if !f.readable || f.node.isDir() {
		return 0, fs.ErrPermission
	}
	if offset >= int64(len(f.node.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.node.data[offset:])
	if n != len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *File) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.check(); err != nil {
		return 0, err
	}
	if !f.writable || f.node.isDir() {
		return 0, fs.ErrPermission
	}
	f.fsys.mu.Lock()
	defer f.fsys.mu.Unlock()
	if f.append {
		f.offset = int64(len(f.node.data))
	}
	n := writeAt(f.node, p, f.offset)
	f.offset += int64(n)
	f.node.modTime = f.fsys.now()
	return n, nil
}

func (f *File) WriteAt(p []byte, offset int64) (int, error) {
	if offset < 0 {
		return 0, errors.New("memfs: negative offset")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fsys.mu.Lock()
	defer f.fsys.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	if !f.writable || f.node.isDir() {
		return 0, fs.ErrPermission
	}
	n := writeAt(f.node, p, offset)
	f.node.modTime = f.fsys.now()
	return n, nil
}

func writeAt(n *node, p []byte, offset int64) int {
	end := offset + int64(len(p))
	if end > int64(len(n.data)) {
		n.data = append(n.data, make([]byte, end-int64(len(n.data)))...)
	}
	return copy(n.data[offset:end], p)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.check(); err != nil {
		return 0, err
	}
	var next int64
	switch whence {
	case io.SeekStart:
		next = offset
	case io.SeekCurrent:
		next = f.offset + offset
	case io.SeekEnd:
		f.fsys.mu.RLock()
		next = int64(len(f.node.data)) + offset
		f.fsys.mu.RUnlock()
	default:
		return 0, errors.New("memfs: invalid whence")
	}
	if next < 0 {
		return 0, errors.New("memfs: negative position")
	}
	f.offset = next
	return next, nil
}

func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return fs.ErrClosed
	}
	f.closed = true
	return nil
}

func (f *File) Stat() (os.FileInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fsys.mu.RLock()
	defer f.fsys.mu.RUnlock()
	if f.closed {
		return nil, fs.ErrClosed
	}
	return infoOf(f.node), nil
}

func (f *File) Sync() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.check()
}

func (f *File) Truncate(size int64) error {
	if size < 0 {
		return errors.New("memfs: negative size")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fsys.mu.Lock()
	defer f.fsys.mu.Unlock()
	if f.closed {
		return fs.ErrClosed
	}
	if !f.writable || f.node.isDir() {
		return fs.ErrPermission
	}
	if size <= int64(len(f.node.data)) {
		f.node.data = f.node.data[:size]
	} else {
		f.node.data = append(f.node.data, make([]byte, size-int64(len(f.node.data)))...)
	}
	f.node.modTime = f.fsys.now()
	return nil
}

func (f *File) Name() string { return f.path }

func (f *File) ReadDir(count int) ([]os.DirEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.check(); err != nil {
		return nil, err
	}
	if !f.node.isDir() {
		return nil, errors.New("memfs: not a directory")
	}
	f.fsys.mu.RLock()
	defer f.fsys.mu.RUnlock()
	var entries []os.DirEntry
	for candidate, n := range f.fsys.nodes {
		if candidate != "" && parent(candidate) == f.path {
			entries = append(entries, dirEntry{infoOf(n)})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	if f.dirIndex >= len(entries) {
		if count > 0 {
			return nil, io.EOF
		}
		return []os.DirEntry{}, nil
	}
	end := len(entries)
	if count > 0 && f.dirIndex+count < end {
		end = f.dirIndex + count
	}
	out := append([]os.DirEntry(nil), entries[f.dirIndex:end]...)
	f.dirIndex = end
	return out, nil
}

type fileInfo struct {
	name    string
	mode    os.FileMode
	size    int64
	modTime time.Time
}

func infoOf(n *node) fileInfo {
	return fileInfo{name: n.name, mode: n.mode, size: int64(len(n.data)), modTime: n.modTime}
}

func (i fileInfo) Name() string       { return i.name }
func (i fileInfo) Size() int64        { return i.size }
func (i fileInfo) Mode() os.FileMode  { return i.mode }
func (i fileInfo) ModTime() time.Time { return i.modTime }
func (i fileInfo) IsDir() bool        { return i.mode.IsDir() }
func (fileInfo) Sys() any             { return nil }

type dirEntry struct{ fileInfo }

func (d dirEntry) Type() os.FileMode          { return d.Mode().Type() }
func (d dirEntry) Info() (os.FileInfo, error) { return d.fileInfo, nil }
