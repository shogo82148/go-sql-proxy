// +build !go1.8

package proxy

import (
	"database/sql/driver"
	"fmt"
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
