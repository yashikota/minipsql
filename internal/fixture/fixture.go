// Package fixture owns the immutable PostgreSQL cluster template.
package fixture

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"

	"github.com/yashikota/minipsql/internal/memfs"
)

//go:embed pgdata.zip
var data []byte

// NewFS returns an independent in-memory copy of the initialized cluster.
func NewFS() (*memfs.FS, error) {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("minipsql: open cluster fixture: %w", err)
	}
	fsys := memfs.New()
	if err := fsys.Load(archive); err != nil {
		return nil, fmt.Errorf("minipsql: load cluster fixture: %w", err)
	}
	return fsys, nil
}
