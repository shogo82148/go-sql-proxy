package proxy

import (
	"database/sql/driver"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Outputter is what is used by the tracing proxy created via `NewTraceProxy`.
// Anything that implements a `log.Logger` style `Output` method will satisfy
// this interface.
type Outputter interface {
	Output(calldepth int, s string) error
}

type Filter interface {
	DoOutput(file string) bool
}

type defaultFilter struct{}

func (_ defaultFilter) DoOutput(file string) bool {
	// skip database/sql
	switch {
	case strings.HasPrefix(file, "database/sql/"):
	case strings.HasPrefix(file, "github.com/shogo82148/txmanager/"):
	default:
		return true
	}
	return false
}

func findCaller(f Filter) int {
	// i starts 4 because 0: findCaller, 1: hooks, 2: proxy-funcs, 3: database/sql, and equals or greater than 4: user-funcs
	for i := 4; ; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok {
			break
		}

		srcIndex := strings.LastIndex(file, "/src/")
		if srcIndex < 0 {
			return i
		}
		if f.DoOutput(file[srcIndex+5:]) {
			return i
		}
	}
	return 0
}

// NewTraceProxy generates a proxy that logs queries.
func NewTraceProxy(d driver.Driver, o Outputter, f Filter) *Proxy {
	if f == nil {
		f = defaultFilter{}
	}

	return &Proxy{
		Driver: d,
		Hooks: &Hooks{
			PreOpen: func(_ string) (interface{}, error) {
				return time.Now(), nil
			},
			PostOpen: func(ctx interface{}, _ driver.Conn) error {
				o.Output(
					findCaller(f),
					fmt.Sprintf(
						"Open (%s)",
						time.Since(ctx.(time.Time)),
					),
				)
				return nil
			},
			PreExec: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
				return time.Now(), nil
			},
			PostExec: func(ctx interface{}, stmt *Stmt, args []driver.Value, _ driver.Result) error {
				o.Output(
					findCaller(f),
					fmt.Sprintf(
						"Exec: %s; args = %v (%s)",
						stmt.QueryString,
						args,
						time.Since(ctx.(time.Time)),
					),
				)
				return nil
			},
			PreQuery: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
				return time.Now(), nil
			},
			PostQuery: func(ctx interface{}, stmt *Stmt, args []driver.Value, _ driver.Rows) error {
				o.Output(
					findCaller(f),
					fmt.Sprintf(
						"Query: %s; args = %v (%s)",
						stmt.QueryString,
						args,
						time.Since(ctx.(time.Time)),
					),
				)
				return nil
			},
			PreBegin: func(_ *Conn) (interface{}, error) {
				return time.Now(), nil
			},
			PostBegin: func(ctx interface{}, _ *Conn) error {
				o.Output(
					findCaller(f),
					fmt.Sprintf("Begin (%s)", time.Since(ctx.(time.Time))),
				)
				return nil
			},
			PreCommit: func(_ *Tx) (interface{}, error) {
				return time.Now(), nil
			},
			PostCommit: func(ctx interface{}, _ *Tx) error {
				o.Output(
					findCaller(f),
					fmt.Sprintf("Commit (%s)", time.Since(ctx.(time.Time))),
				)
				return nil
			},
			PreRollback: func(_ *Tx) (interface{}, error) {
				return time.Now(), nil
			},
			PostRollback: func(ctx interface{}, _ *Tx) error {
				o.Output(
					findCaller(f),
					fmt.Sprintf("Rollback (%s)", time.Since(ctx.(time.Time))),
				)
				return nil
			},
		},
	}
}
