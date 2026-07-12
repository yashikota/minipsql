//go:build !minipsql_stub

package engine

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yashikota/minipsql/internal/fixture"
	pgrust "github.com/yashikota/minipsql/internal/generated/pgrust"
	guesthost "github.com/yashikota/minipsql/internal/generated/pgrust/host"
	"github.com/yashikota/minipsql/internal/memfs"
	"github.com/yashikota/minipsql/internal/pgvfs"
)

type cluster struct {
	mu     sync.Mutex
	fs     *memfs.FS
	closed bool
}

func New(_ context.Context, _ Options) (Cluster, error) {
	fsys, err := fixture.NewFS()
	if err != nil {
		return nil, err
	}
	return &cluster{fs: fsys}, nil
}

func (c *cluster) OpenSession(ctx context.Context, opts Options) (Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, errors.New("minipsql: cluster is closed")
	}
	return startSession(ctx, c.fs, opts)
}

func (c *cluster) Close() error {
	c.mu.Lock()
	c.closed = true
	c.fs = nil
	c.mu.Unlock()
	return nil
}

type session struct {
	mu     sync.Mutex
	input  *io.PipeWriter
	frames <-chan []byte
	done   <-chan error
	stderr *lockedBuffer
	closed bool
}

func startSession(ctx context.Context, fsys *memfs.FS, opts Options) (*session, error) {
	inputReader, inputWriter := io.Pipe()
	output := newFrameWriter([]byte("backend> "))
	stderr := &lockedBuffer{}
	done := make(chan error, 1)
	host := pgvfs.New(pgvfs.Options{
		FS: fsys, Stdin: inputReader, Stdout: output, Stderr: stderr,
		Args: []string{"postgres", "--single", "-D", "/data", opts.Database},
	})
	imports := guesthost.Pgvfs{Host: host, Root: "/data"}
	module := pgrust.New(guesthost.Env{}, imports)
	go func() {
		var runErr error
		defer func() {
			if recovered := recover(); recovered != nil {
				if exit, ok := recovered.(pgvfs.ExitError); !ok || exit.Code != 0 {
					runErr = fmt.Errorf("minipsql: pgrust panic: %v", recovered)
				}
			}
			output.Close()
			done <- runErr
			close(done)
		}()
		if code := pgrust.Main(module, 0, 0); code != 0 {
			runErr = fmt.Errorf("minipsql: pgrust exited with code %d", code)
		}
	}()

	s := &session{input: inputWriter, frames: output.Frames(), done: done, stderr: stderr}
	select {
	case <-output.Frames():
		return s, nil
	case err := <-done:
		if err == nil {
			err = errors.New("minipsql: pgrust exited during startup")
		}
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	case <-ctx.Done():
		_ = inputWriter.Close()
		return nil, ctx.Err()
	}
}

func (s *session) Execute(ctx context.Context, query string, args []driver.NamedValue) (Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return Result{}, errors.New("minipsql: session is closed")
	}
	query, err := bind(query, args)
	if err != nil {
		return Result{}, err
	}
	query = strings.TrimSpace(query)
	if !strings.HasSuffix(query, ";") {
		query += ";"
	}
	mark := s.stderr.Len()
	if _, err := io.WriteString(s.input, query+"\n"); err != nil {
		return Result{}, fmt.Errorf("minipsql: write query: %w", err)
	}
	select {
	case frame, ok := <-s.frames:
		if !ok {
			return Result{}, s.exitError()
		}
		if message := postgresError(s.stderr.Since(mark)); message != "" {
			return Result{}, errors.New(message)
		}
		return parseResult(frame)
	case <-s.done:
		return Result{}, s.exitError()
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

func (s *session) exitError() error {
	message := strings.TrimSpace(s.stderr.String())
	if message == "" {
		message = "pgrust backend exited"
	}
	return errors.New("minipsql: " + message)
}

func (s *session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	err := s.input.Close()
	s.mu.Unlock()
	<-s.done
	return err
}

type frameWriter struct {
	mu     sync.Mutex
	delim  []byte
	buffer []byte
	frames chan []byte
	closed bool
}

func newFrameWriter(delim []byte) *frameWriter {
	return &frameWriter{delim: delim, frames: make(chan []byte, 2)}
}
func (w *frameWriter) Frames() <-chan []byte { return w.frames }
func (w *frameWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	w.buffer = append(w.buffer, p...)
	for {
		index := bytes.Index(w.buffer, w.delim)
		if index < 0 {
			break
		}
		w.frames <- append([]byte(nil), w.buffer[:index]...)
		w.buffer = append(w.buffer[:0], w.buffer[index+len(w.delim):]...)
	}
	return len(p), nil
}
func (w *frameWriter) Close() {
	w.mu.Lock()
	if !w.closed {
		w.closed = true
		close(w.frames)
	}
	w.mu.Unlock()
}

type lockedBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	b.data = append(b.data, p...)
	b.mu.Unlock()
	return len(p), nil
}
func (b *lockedBuffer) Len() int       { b.mu.Lock(); defer b.mu.Unlock(); return len(b.data) }
func (b *lockedBuffer) String() string { b.mu.Lock(); defer b.mu.Unlock(); return string(b.data) }
func (b *lockedBuffer) Since(mark int) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if mark > len(b.data) {
		mark = len(b.data)
	}
	return string(b.data[mark:])
}

func parseResult(frame []byte) (Result, error) {
	var result Result
	var row []driver.Value
	inRows := false
	for _, line := range strings.Split(string(frame), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "----" {
			if !inRows {
				inRows = true
				row = make([]driver.Value, len(result.Columns))
			} else {
				result.Rows = append(result.Rows, row)
				row = make([]driver.Value, len(result.Columns))
			}
			continue
		}
		colon := strings.Index(line, ": ")
		typeInfo := strings.Index(line, "\t(typeid =")
		if colon < 0 || typeInfo < 0 {
			continue
		}
		payload := line[colon+2 : typeInfo]
		if !inRows {
			result.Columns = append(result.Columns, payload)
			continue
		}
		equals := strings.Index(payload, " = ")
		if equals < 0 {
			continue
		}
		raw := payload[equals+3:]
		value, err := strconv.Unquote(raw)
		if err != nil {
			value = strings.Trim(raw, "\"")
		}
		column, err := strconv.Atoi(strings.TrimSpace(line[:colon]))
		if err == nil && column > 0 && column <= len(row) {
			row[column-1] = value
		}
	}
	return result, nil
}

func postgresError(stderr string) string {
	for _, severity := range []string{" FATAL:  ", " ERROR:  ", " PANIC:  "} {
		if index := strings.Index(stderr, severity); index >= 0 {
			return "minipsql: postgres:" + stderr[index+len(severity)-2:]
		}
	}
	return ""
}

func bind(query string, args []driver.NamedValue) (string, error) {
	for index := len(args) - 1; index >= 0; index-- {
		arg := args[index]
		ordinal := arg.Ordinal
		if ordinal == 0 {
			ordinal = index + 1
		}
		literal, err := sqlLiteral(arg.Value)
		if err != nil {
			return "", err
		}
		query = strings.ReplaceAll(query, "$"+strconv.Itoa(ordinal), literal)
	}
	return query, nil
}

func sqlLiteral(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "NULL", nil
	case bool:
		if v {
			return "TRUE", nil
		}
		return "FALSE", nil
	case int64, float64:
		return fmt.Sprint(v), nil
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'", nil
	case []byte:
		return "decode('" + hex.EncodeToString(v) + "', 'hex')", nil
	case time.Time:
		return "'" + v.Format(time.RFC3339Nano) + "'", nil
	default:
		return "", fmt.Errorf("minipsql: unsupported argument type %T", value)
	}
}
