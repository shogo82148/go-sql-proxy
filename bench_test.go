// +build go1.8

package proxy

import (
	"context"
	"database/sql/driver"
	"testing"
)

type nullConn struct{}

func (nullConn) Prepare(query string) (driver.Stmt, error)                     { return nil, nil }
func (nullConn) Close() error                                                  { return nil }
func (nullConn) Begin() (driver.Tx, error)                                     { return nil, nil }
func (nullConn) Exec(query string, args []driver.Value) (driver.Result, error) { return nil, nil }

type nullConnCtx struct{}

func (nullConnCtx) Prepare(query string) (driver.Stmt, error)                     { return nil, nil }
func (nullConnCtx) Close() error                                                  { return nil }
func (nullConnCtx) Begin() (driver.Tx, error)                                     { return nil, nil }
func (nullConnCtx) Exec(query string, args []driver.Value) (driver.Result, error) { return nil, nil }
func (nullConnCtx) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

type nullLogger struct{}

func (nullLogger) Output(calldepth int, s string) error { return nil }

func BenchmarkDirectly(b *testing.B) {
	conn := nullConn{}
	args := []driver.Value{int64(123456789)}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", args)
	}
}

func BenchmarkNilHook(b *testing.B) {
	ctx := context.Background()
	conn := &Conn{
		Conn:  nullConn{},
		Proxy: &Proxy{},
	}
	args := []driver.NamedValue{
		{
			Ordinal: 1,
			Value:   int64(123456789),
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ExecContext(ctx, "CREATE TABLE t1 (id INTEGER PRIMARY KEY)", args)
	}
}

func BenchmarkNilHookCtx(b *testing.B) {
	ctx := context.Background()
	conn := &Conn{
		Conn:  nullConnCtx{},
		Proxy: &Proxy{},
	}
	args := []driver.NamedValue{
		{
			Ordinal: 1,
			Value:   int64(123456789),
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ExecContext(ctx, "CREATE TABLE t1 (id INTEGER PRIMARY KEY)", args)
	}
}

func BenchmarkTracer(b *testing.B) {
	ctx := context.Background()
	conn := &Conn{
		Conn: nullConnCtx{},
		Proxy: &Proxy{
			hooks: NewTraceHooks(TracerOptions{
				Outputter: nullLogger{},
			}),
		},
	}
	args := []driver.NamedValue{
		{
			Ordinal: 1,
			Value:   int64(123456789),
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ExecContext(ctx, "CREATE TABLE t1 (id INTEGER PRIMARY KEY)", args)
	}
}
