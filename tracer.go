package proxy

import (
	"database/sql/driver"
	"fmt"
	"log"
	"time"
)

// NewTraceProxy generates a proxy that logs queries.
func NewTraceProxy(d driver.Driver, logger *log.Logger) *Proxy {
	return &Proxy{
		Driver: d,
		Hooks: &Hooks{
			PreOpen: func(_ string) (interface{}, error) {
				return time.Now(), nil
			},
			PostOpen: func(ctx interface{}, _ driver.Conn) error {
				logger.Output(
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
				logger.Output(
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
				logger.Output(
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
				logger.Output(6, "Begin")
				return nil
			},
			Commit: func(tx *Tx) error {
				logger.Output(6, "Commit")
				return nil
			},
			Rollback: func(tx *Tx) error {
				logger.Output(8, "Rollback")
				return nil
			},
		},
	}
}
