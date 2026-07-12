package engine

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

func TestParseSingleUserResult(t *testing.T) {
	frame := []byte("\t 1: a\t(typeid = 23, len = 4, typmod = -1, byval = t)\n" +
		"\t 2: b\t(typeid = 25, len = -1, typmod = -1, byval = f)\n" +
		"\t----\n" +
		"\t 1:  = \"1\"\t(typeid = 23, len = 4, typmod = -1, byval = t)\n" +
		"\t 2:  = \"x\"\t(typeid = 25, len = -1, typmod = -1, byval = f)\n" +
		"\t----\n")
	got, err := parseResult(frame)
	if err != nil {
		t.Fatal(err)
	}
	want := Result{Columns: []string{"a", "b"}, Rows: [][]driver.Value{{"1", "x"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
}

func TestParseNullRow(t *testing.T) {
	frame := []byte("\t 1: n\t(typeid = 25, len = -1, typmod = -1, byval = f)\n\t----\n\t----\n")
	got, err := parseResult(frame)
	if err != nil {
		t.Fatal(err)
	}
	want := Result{Columns: []string{"n"}, Rows: [][]driver.Value{{nil}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
}

func TestBindPostgresLiterals(t *testing.T) {
	got, err := bind("SELECT $1, $2, $3", []driver.NamedValue{
		{Ordinal: 1, Value: int64(42)},
		{Ordinal: 2, Value: "it's"},
		{Ordinal: 3, Value: nil},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "SELECT 42, 'it''s', NULL"; got != want {
		t.Fatalf("query = %q, want %q", got, want)
	}
}
