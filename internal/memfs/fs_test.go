package memfs

import (
	"io"
	"os"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestFileLifecycleAndCloneIsolation(t *testing.T) {
	m := New()
	if err := m.Mkdir("base", 0o700); err != nil {
		t.Fatal(err)
	}
	f, err := m.OpenFile("base/1", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("postgres")); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 8)
	if _, err := io.ReadFull(f, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "postgres" {
		t.Fatalf("content = %q", buf)
	}
	if err := f.Truncate(4); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	clone := m.Clone()
	cf, err := clone.OpenFile("base/1", os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cf.Write([]byte("copy")); err != nil {
		t.Fatal(err)
	}
	_ = cf.Close()

	original, err := m.OpenFile("base/1", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	originalBytes, err := io.ReadAll(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(originalBytes) != "post" {
		t.Fatalf("original changed through clone: %q", originalBytes)
	}
}

func TestLoadCopiesSourceTree(t *testing.T) {
	source := fstest.MapFS{
		"base/1": &fstest.MapFile{Data: []byte("catalog"), Mode: 0o600},
	}
	m := New()
	if err := m.Load(source); err != nil {
		t.Fatal(err)
	}
	file, err := m.OpenFile("base/1", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "catalog" {
		t.Fatalf("content = %q", got)
	}
	if _, err := file.WriteAt([]byte("C"), 0); err != nil {
		t.Fatal(err)
	}
	if string(source["base/1"].Data) != "catalog" {
		t.Fatal("Load retained source storage")
	}
}

func TestDirectoryRenameAndIteration(t *testing.T) {
	m := New()
	for _, dir := range []string{"pgdata", "pgdata/base"} {
		if err := m.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"a", "b"} {
		f, err := m.OpenFile("pgdata/base/"+name, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}
	if err := m.Rename("pgdata/base", "pgdata/base.old"); err != nil {
		t.Fatal(err)
	}
	dir, err := m.OpenFile("pgdata/base.old", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := dir.ReadDir(-1)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{entries[0].Name(), entries[1].Name()}
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("entries = %v", got)
	}
	if _, err := m.Stat("pgdata/base/a"); !os.IsNotExist(err) {
		t.Fatalf("old child still exists: %v", err)
	}
}

func TestRejectsEscapes(t *testing.T) {
	m := New()
	for _, name := range []string{"/tmp/file", "../file", "a/../../file"} {
		if _, err := m.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			t.Fatalf("OpenFile(%q) succeeded", name)
		}
	}
}
