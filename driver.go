package minipsql

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/yashikota/minipsql/internal/engine"
)

type connector struct {
	cluster engine.Cluster
	opts    engine.Options
}

func newConnector(cluster engine.Cluster, opts Options) driver.Connector {
	return &connector{
		cluster: cluster,
		opts:    engine.Options{Database: opts.Database, User: opts.User},
	}
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	session, err := c.cluster.OpenSession(ctx, c.opts)
	if err != nil {
		return nil, err
	}
	return &conn{session: session}, nil
}

func (c *connector) Driver() driver.Driver { return connectorDriver{} }

type connectorDriver struct{}

func (connectorDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("minipsql: use minipsql.New instead of sql.Open")
}

type conn struct {
	session engine.Session
}

var (
	_ driver.Conn               = (*conn)(nil)
	_ driver.ConnBeginTx        = (*conn)(nil)
	_ driver.ConnPrepareContext = (*conn)(nil)
	_ driver.ExecerContext      = (*conn)(nil)
	_ driver.Pinger             = (*conn)(nil)
	_ driver.QueryerContext     = (*conn)(nil)
)

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *conn) PrepareContext(_ context.Context, query string) (driver.Stmt, error) {
	return &stmt{conn: c, query: query}, nil
}

func (c *conn) Close() error { return c.session.Close() }

func (c *conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	query := "BEGIN"
	if opts.Isolation != driver.IsolationLevel(0) {
		level, err := isolationSQL(opts.Isolation)
		if err != nil {
			return nil, err
		}
		query += " ISOLATION LEVEL " + level
	}
	if opts.ReadOnly {
		query += " READ ONLY"
	}
	if _, err := c.session.Execute(ctx, query, nil); err != nil {
		return nil, err
	}
	return &tx{conn: c}, nil
}

func isolationSQL(level driver.IsolationLevel) (string, error) {
	switch level {
	case driver.IsolationLevel(1):
		return "READ UNCOMMITTED", nil
	case driver.IsolationLevel(2):
		return "READ COMMITTED", nil
	case driver.IsolationLevel(4):
		return "REPEATABLE READ", nil
	case driver.IsolationLevel(6):
		return "SERIALIZABLE", nil
	default:
		return "", fmt.Errorf("minipsql: unsupported isolation level %d", level)
	}
}

func (c *conn) Ping(context.Context) error { return nil }

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	result, err := c.session.Execute(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return execResult{rowsAffected: result.RowsAffected}, nil
}

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	result, err := c.session.Execute(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return &rows{columns: result.Columns, values: result.Rows}, nil
}

type stmt struct {
	conn  *conn
	query string
}

var (
	_ driver.Stmt             = (*stmt)(nil)
	_ driver.StmtExecContext  = (*stmt)(nil)
	_ driver.StmtQueryContext = (*stmt)(nil)
)

func (*stmt) Close() error  { return nil }
func (*stmt) NumInput() int { return -1 }

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), namedValues(args))
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.QueryContext(context.Background(), namedValues(args))
}

func (s *stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.conn.ExecContext(ctx, s.query, args)
}

func (s *stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.conn.QueryContext(ctx, s.query, args)
}

func namedValues(values []driver.Value) []driver.NamedValue {
	args := make([]driver.NamedValue, len(values))
	for index, value := range values {
		args[index] = driver.NamedValue{Ordinal: index + 1, Value: value}
	}
	return args
}

type tx struct {
	conn *conn
	done bool
}

func (t *tx) Commit() error   { return t.finish("COMMIT") }
func (t *tx) Rollback() error { return t.finish("ROLLBACK") }

func (t *tx) finish(query string) error {
	if t.done {
		return errors.New("minipsql: transaction already closed")
	}
	t.done = true
	_, err := t.conn.session.Execute(context.Background(), query, nil)
	return err
}

type execResult struct{ rowsAffected int64 }

func (execResult) LastInsertId() (int64, error) {
	return 0, errors.New("minipsql: PostgreSQL does not support LastInsertId; use RETURNING")
}

func (r execResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type rows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *rows) Columns() []string { return append([]string(nil), r.columns...) }
func (*rows) Close() error        { return nil }

func (r *rows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	row := r.values[r.index]
	if len(row) != len(dest) {
		return fmt.Errorf("minipsql: engine returned %d values for %d columns", len(row), len(dest))
	}
	copy(dest, row)
	r.index++
	return nil
}
