package proxy

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Outputter is what is used by the tracing proxy created via `NewTraceProxy`.
// Anything that implements a `log.Logger` style `Output` method will satisfy
// this interface.
type Outputter interface{
	Output(calldepth int, s string) error
}

// NewTraceProxy generates a proxy that logs queries.
func NewTraceProxy(d driver.Driver, o Outputter) *Proxy {
	return &Proxy{
		Driver: d,
		Hooks: &Hooks{
			PreOpen: func(_ string) (interface{}, error) {
				return time.Now(), nil
			},
			PostOpen: func(ctx interface{}, _ driver.Conn) error {
				o.Output(
					7,
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
					7,
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
					9,
					fmt.Sprintf(
						"Query: %s; args = %v (%s)",
						stmt.QueryString,
						args,
						time.Since(ctx.(time.Time)),
					),
				)
				return nil
			},
			Begin: func(conn *Conn) error {
				o.Output(6, "Begin")
				return nil
			},
			Commit: func(tx *Tx) error {
				o.Output(6, "Commit")
				return nil
			},
			Rollback: func(tx *Tx) error {
				o.Output(8, "Rollback")
				return nil
			},
		},
	}
}
