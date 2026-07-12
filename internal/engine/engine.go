package engine

import (
	"context"
	"database/sql/driver"
	"errors"
)

var ErrUnavailable = errors.New("generated engine unavailable")

type Options struct {
	Database string
	User     string
}

type Result struct {
	Columns      []string
	Rows         [][]driver.Value
	RowsAffected int64
}

type Cluster interface {
	OpenSession(context.Context, Options) (Session, error)
	Close() error
}

type Session interface {
	Execute(context.Context, string, []driver.NamedValue) (Result, error)
	Close() error
}
