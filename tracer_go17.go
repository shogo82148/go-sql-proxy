// +build !go1.8

package proxy

import (
	"database/sql/driver"
	"fmt"
	"runtime"
	"time"
)

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

// fallback version of findCaller.
// It uses runtime.FuncForPC which will be deprecated in Go1.9.
// See #16 for more detail.
func findCaller(f Filter) int {
	// skip starts 4. 0: Callers, 1: findCaller, 2: hooks, 3: proxy-funcs, 4: database/sql, and equals or greater than 5: user-funcs
	skip := 5
	for {
		var rpc [8]uintptr
		n := runtime.Callers(skip, rpc[:])

		for i, pc := range rpc[:n] {
			// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
			name := runtime.FuncForPC(pc).Name()
			dotIdx := 0
			for j := len(name) - 1; j >= 0; j-- {
				if name[j] == '.' {
					dotIdx = j
				} else if name[j] == '/' {
					break
				}
			}
			packageName := name[:dotIdx]
			if f.DoOutput(packageName) {
				// -1 because the meaning of skip differs between Caller and Callers.
				return skip + i - 1
			}
		}
		if n < len(rpc) {
			break
		}
		skip += n
	}
	return 0
}
