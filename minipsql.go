// Package minipsql provides isolated, in-memory PostgreSQL clusters for tests.
package minipsql

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/yashikota/minipsql/internal/engine"
)

// Options configures a new isolated PostgreSQL cluster.
type Options struct {
	Database string
	User     string
}

func (o Options) normalized() Options {
	if o.Database == "" {
		o.Database = "postgres"
	}
	if o.User == "" {
		o.User = "postgres"
	}
	return o
}

// Instance owns one isolated in-memory cluster and its database/sql pool.
// Close must be called when the test is finished.
type Instance struct {
	cluster engine.Cluster
	db      *sql.DB
	once    sync.Once
	err     error
}

// New creates and initializes an isolated in-memory PostgreSQL cluster.
func New(ctx context.Context, opts Options) (*Instance, error) {
	opts = opts.normalized()
	cluster, err := engine.New(ctx, engine.Options{Database: opts.Database, User: opts.User})
	if err != nil {
		if errors.Is(err, engine.ErrUnavailable) {
			return nil, ErrEngineUnavailable
		}
		return nil, err
	}

	connector := newConnector(cluster, opts)
	db := sql.OpenDB(connector)
	// The embedded single-user backend owns one transaction stream per
	// connection. Keep database/sql on that stream so BEGIN/COMMIT state and
	// temporary objects remain coherent.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	instance := &Instance{cluster: cluster, db: db}
	if err := db.PingContext(ctx); err != nil {
		_ = instance.Close()
		return nil, err
	}
	return instance, nil
}

// DB returns the database/sql pool connected to the instance.
func (i *Instance) DB() *sql.DB {
	if i == nil {
		return nil
	}
	return i.db
}

// Close closes all SQL connections and releases the cluster's memory.
func (i *Instance) Close() error {
	if i == nil {
		return nil
	}
	i.once.Do(func() {
		if i.db != nil {
			i.err = i.db.Close()
		}
		if err := i.cluster.Close(); i.err == nil {
			i.err = err
		}
	})
	return i.err
}
