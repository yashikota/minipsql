package minipsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"
	"sync"
	"testing"

	"github.com/yashikota/minipsql/internal/engine"
)

func TestNewRunsPostgresInMemory(t *testing.T) {
	instance, err := New(context.Background(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = instance.Close() })
	var got int
	if err := instance.DB().QueryRowContext(context.Background(), "SELECT 1").Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("SELECT 1 = %d", got)
	}
	if _, err := instance.DB().ExecContext(context.Background(), "CREATE TABLE minipsql_tx (v integer)"); err != nil {
		t.Fatal(err)
	}
	tx, err := instance.DB().BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(context.Background(), "INSERT INTO minipsql_tx VALUES ($1)", int64(42)); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := instance.DB().QueryRowContext(context.Background(), "SELECT v FROM minipsql_tx").Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Fatalf("transaction value = %d", got)
	}
	if _, err := instance.DB().ExecContext(context.Background(), "SELECT * FROM minipsql_missing_relation"); err == nil {
		t.Fatal("missing relation query succeeded")
	}
}

func TestInstancesAreIsolated(t *testing.T) {
	ctx := context.Background()
	left, err := New(ctx, Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = left.Close() })
	right, err := New(ctx, Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = right.Close() })
	if _, err := left.DB().ExecContext(ctx, "CREATE TABLE only_in_left (v integer)"); err != nil {
		t.Fatal(err)
	}
	if _, err := right.DB().ExecContext(ctx, "SELECT * FROM only_in_left"); err == nil {
		t.Fatal("table leaked between isolated instances")
	}
}

func TestDatabaseSQLConnector(t *testing.T) {
	cluster := &fakeCluster{}
	db := sql.OpenDB(newConnector(cluster, Options{Database: "testdb", User: "tester"}))
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), "CREATE TABLE t (v int)"); err != nil {
		t.Fatal(err)
	}
	var got int64
	if err := db.QueryRowContext(context.Background(), "SELECT $1", int64(42)).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Fatalf("value = %d, want 42", got)
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	queries := cluster.queries()
	want := []string{
		"CREATE TABLE t (v int)",
		"SELECT $1",
		"BEGIN ISOLATION LEVEL SERIALIZABLE READ ONLY",
		"ROLLBACK",
	}
	if !reflect.DeepEqual(queries, want) {
		t.Fatalf("queries = %#v, want %#v", queries, want)
	}
	if cluster.opts != (engine.Options{Database: "testdb", User: "tester"}) {
		t.Fatalf("session options = %#v", cluster.opts)
	}
}

type fakeCluster struct {
	mu      sync.Mutex
	opts    engine.Options
	execSQL []string
}

func (f *fakeCluster) OpenSession(_ context.Context, opts engine.Options) (engine.Session, error) {
	f.mu.Lock()
	f.opts = opts
	f.mu.Unlock()
	return &fakeSession{cluster: f}, nil
}

func (*fakeCluster) Close() error { return nil }

func (f *fakeCluster) queries() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.execSQL...)
}

type fakeSession struct{ cluster *fakeCluster }

func (f *fakeSession) Execute(_ context.Context, query string, args []driver.NamedValue) (engine.Result, error) {
	f.cluster.mu.Lock()
	f.cluster.execSQL = append(f.cluster.execSQL, query)
	f.cluster.mu.Unlock()
	if query == "SELECT $1" {
		return engine.Result{Columns: []string{"?column?"}, Rows: [][]driver.Value{{args[0].Value}}}, nil
	}
	return engine.Result{}, nil
}

func (*fakeSession) Close() error { return nil }
