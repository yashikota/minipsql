// Command fixturepack creates a deterministic zip from an initialized cluster.
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	source := flag.String("source", "", "source directory")
	output := flag.String("output", "", "output zip")
	flag.Parse()
	if *source == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "-source and -output are required")
		os.Exit(2)
	}
	var names []string
	check(filepath.WalkDir(*source, func(name string, entry fs.DirEntry, err error) error {
		if err != nil { return err }
		if name != *source { names = append(names, name) }
		return nil
	}))
	sort.Strings(names)
	out, err := os.Create(*output); check(err)
	zw := zip.NewWriter(out)
	epoch := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, name := range names {
		info, err := os.Lstat(name); check(err)
		header, err := zip.FileInfoHeader(info); check(err)
		rel, err := filepath.Rel(*source, name); check(err)
		header.Name = filepath.ToSlash(rel)
		header.Modified = epoch
		header.SetMode(info.Mode())
		if info.IsDir() {
			header.Name = strings.TrimSuffix(header.Name, "/") + "/"
			header.Method = zip.Store
			_, err = zw.CreateHeader(header); check(err)
			continue
		}
		header.Method = zip.Deflate
		writer, err := zw.CreateHeader(header); check(err)
		input, err := os.Open(name); check(err)
		_, err = io.Copy(writer, input)
		closeErr := input.Close()
		check(err); check(closeErr)
	}
	check(zw.Close())
	check(out.Close())
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
