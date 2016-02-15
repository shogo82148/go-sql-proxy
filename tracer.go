package proxy

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

type logger struct{}

func (_ logger) Output(calldepth int, s string) error {
	return log.Output(calldepth, s)
}

// Outputter is what is used by the tracing proxy created via `NewTraceProxy`.
// Anything that implements a `log.Logger` style `Output` method will satisfy
// this interface.
type Outputter interface {
	Output(calldepth int, s string) error
}

type Filter interface {
	DoOutput(packageName string) bool
}

type PackageFilter map[string]struct{}

func (f PackageFilter) DoOutput(packageName string) bool {
	_, ok := f[packageName]
	return !ok
}

func (f PackageFilter) Ignore(packageName string) {
	f[packageName] = struct{}{}
}

func findCaller(f Filter) int {
	// i starts 4. 0: findCaller, 1: hooks, 2: proxy-funcs, 3: database/sql, and equals or greater than 4: user-funcs
	for i := 4; ; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
		parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		pl := len(parts)
		packageName := ""
		if parts[pl-2][0] == '(' {
			packageName = strings.Join(parts[0:pl-2], ".")
		} else {
			packageName = strings.Join(parts[0:pl-1], ".")
		}

		if f.DoOutput(packageName) {
			return i
		}
	}
	return 0
}

// NewTraceProxy generates a proxy that logs queries.
func NewTraceProxy(d driver.Driver, o Outputter) *Proxy {
	return NewTraceProxyWithFilter(d, o, nil)
}

// NewTraceProxyWithFilter generates a proxy that logs queries.
func NewTraceProxyWithFilter(d driver.Driver, o Outputter, f Filter) *Proxy {
	if f == nil {
		f = PackageFilter{
			"database/sql":                    struct{}{},
			"github.com/shogo82148/txmanager": struct{}{},
		}
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

// RegisterTracer creates proxies that logs queries from the sql drivers already registered,
// and registers the proxies as sql driver.
// The proxies' names have suffix ":trace".
func RegisterTracer() {
	for _, driver := range sql.Drivers() {
		if strings.HasSuffix(driver, ":trace") {
			continue
		}
		db, _ := sql.Open(driver, "")
		defer db.Close()
		sql.Register(driver+":trace", NewTraceProxy(db.Driver(), logger{}))
	}
}
