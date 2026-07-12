package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/yashikota/minipsql/internal/fixture"
	pgrust "github.com/yashikota/minipsql/internal/generated/pgrust"
	guesthost "github.com/yashikota/minipsql/internal/generated/pgrust/host"
	"github.com/yashikota/minipsql/internal/pgvfs"
)

func main() {
	var stdout, stderr bytes.Buffer
	query := os.Getenv("SQL")
	if query == "" {
		query = "select 1;"
	}
	fsys, err := fixture.NewFS()
	if err != nil {
		panic(err)
	}
	h := pgvfs.New(pgvfs.Options{
		FS:     fsys,
		Stdin:  strings.NewReader(query + "\n"),
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"postgres", "--single", "-D", "/data", "postgres"},
	})
	imports := guesthost.Pgvfs{Host: h, Root: "/data"}
	m := pgrust.New(guesthost.Env{}, imports)
	code := int32(-1)
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				if exit, ok := recovered.(pgvfs.ExitError); ok {
					code = exit.Code
					return
				}
				panic(recovered)
			}
		}()
		code = pgrust.Main(m, 0, 0)
	}()
	fmt.Printf("exit=%d\nstdout:\n%s\nstderr:\n%s\n", code, stdout.Bytes(), stderr.Bytes())
	if code != 0 {
		os.Exit(int(code))
	}
}
