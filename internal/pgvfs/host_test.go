package pgvfs

import (
	"bytes"
	"os"
	"testing"

	"github.com/yashikota/minipsql/internal/memfs"
)

func TestHostFileDescriptorsAndStreams(t *testing.T) {
	filesystem := memfs.New()
	if err := filesystem.Mkdir("pgdata", 0o700); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	host := New(Options{FS: filesystem, Stdout: &stdout, Args: []string{"postgres", "--single"}})

	fd := host.Open("pgdata/relation", 0o100|2, 0o600)
	if fd < 0 {
		t.Fatalf("open = %d", fd)
	}
	if got := host.Write(int32(fd), []byte("tuple")); got != 5 {
		t.Fatalf("write = %d", got)
	}
	if got := host.PWrite(int32(fd), []byte("T"), 0); got != 1 {
		t.Fatalf("pwrite = %d", got)
	}
	if got := host.Seek(int32(fd), 0, 0); got != 0 {
		t.Fatalf("seek = %d", got)
	}
	buf := make([]byte, 5)
	if got := host.Read(int32(fd), buf); got != 5 || string(buf) != "Tuple" {
		t.Fatalf("read = %d, %q", got, buf)
	}
	if got := host.Write(1, []byte("ok")); got != 2 || stdout.String() != "ok" {
		t.Fatalf("stdout = %d, %q", got, stdout.String())
	}
	if got := host.Close(int32(fd)); got != 0 {
		t.Fatalf("close = %d", got)
	}
	if got := host.Read(int32(fd), buf); got != -errBadFD {
		t.Fatalf("closed read = %d", got)
	}
	if host.Argc() != 2 {
		t.Fatalf("argc = %d", host.Argc())
	}
	if arg, ok := host.Arg(1); !ok || arg != "--single" {
		t.Fatalf("arg = %q, %v", arg, ok)
	}
}

func TestHostDirectoryAndErrno(t *testing.T) {
	host := New(Options{})
	if got := host.Mkdir("base", 0o700); got != 0 {
		t.Fatalf("mkdir = %d", got)
	}
	file := host.Open("base/1", int32(0o100|os.O_WRONLY), 0o600)
	if file < 0 {
		t.Fatalf("open = %d", file)
	}
	_ = host.Close(int32(file))
	dir := host.OpenDir("base")
	if dir < 0 {
		t.Fatalf("opendir = %d", dir)
	}
	name, size := host.ReadDir(int32(dir))
	if name != "1" || size != 1 {
		t.Fatalf("readdir = %q, %d", name, size)
	}
	if _, size := host.ReadDir(int32(dir)); size != 0 {
		t.Fatalf("readdir eof = %d", size)
	}
	if got := host.Rmdir("base"); got != -errNotEmpty {
		t.Fatalf("non-empty rmdir = %d", got)
	}
	if got := host.Access("missing"); got != -errNoEntry {
		t.Fatalf("missing access = %d", got)
	}
}
