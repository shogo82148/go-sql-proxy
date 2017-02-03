// +build go1.8

package proxy

import (
	"context"
	"database/sql/driver"
	"errors"

	"github.com/golang/go/src/database/sql"
)

// Conn adds hook points into "database/sql/driver".Conn.
type Conn struct {
	Conn  driver.Conn
	Proxy *Proxy
}

// Ping verifies a connection to the database is still alive.
// It will trigger PrePing, Ping, PostPing hooks.
//
// If the original connection does not satisfy "database/sql/driver".Pinger, it does nothing.
func (conn *Conn) Ping(c context.Context) error {
	var err error
	var ctx interface{}
	defer func() { conn.Proxy.Hooks.postPing(c, ctx, conn, err) }()
	if ctx, err = conn.Proxy.Hooks.prePing(c, conn); err != nil {
		return err
	}

	if p, ok := conn.Conn.(driver.Pinger); ok {
		err = p.Ping(c)
		if err != nil {
			return err
		}
	}

	err = conn.Proxy.Hooks.ping(c, ctx, conn)
	return err
}

// Prepare returns a prepared statement which is wrapped by Stmt.
// NOT SUPPORTED: use PrepareContext instead
func (conn *Conn) Prepare(query string) (driver.Stmt, error) {
	panic("not supported")
}

// PrepareContext returns a prepared statement which is wrapped by Stmt.
func (conn *Conn) PrepareContext(c context.Context, query string) (driver.Stmt, error) {
	var stmt driver.Stmt
	var err error
	if connCtx, ok := conn.Conn.(driver.ConnPrepareContext); ok {
		stmt, err = connCtx.PrepareContext(c, query)
	} else {
		stmt, err = conn.Conn.Prepare(query)
		if err == nil {
			select {
			default:
			case <-c.Done():
				stmt.Close()
				return nil, c.Err()
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return &Stmt{
		Stmt:        stmt,
		QueryString: query,
		Proxy:       conn.Proxy,
	}, nil
}

// Close calls the original Close method.
func (conn *Conn) Close() error {
	return conn.Conn.Close()
}

// Begin starts and returns a new transaction which is wrapped by Tx.
// It will trigger PreBegin, Begin, PostBegin hooks.
// NOT SUPPORTED: use BeginContext instead
func (conn *Conn) Begin() (driver.Tx, error) {
	panic("not supported")
}

// BeginContext starts and returns a new transaction which is wrapped by Tx.
// It will trigger PreBegin, Begin, PostBegin hooks.
func (conn *Conn) BeginTx(c context.Context, opts driver.TxOptions) (driver.Tx, error) {
	// set the hooks.
	var err error
	var ctx interface{}
	var tx driver.Tx
	defer func() { conn.Proxy.Hooks.postBegin(c, ctx, conn, err) }()
	if ctx, err = conn.Proxy.Hooks.preBegin(c, conn); err != nil {
		return nil, err
	}

	// call the original method.
	if connCtx, ok := conn.Conn.(driver.ConnBeginTx); ok {
		tx, err = connCtx.BeginTx(c, opts)
	} else {
		if c.Done() != context.Background().Done() {
			// the original driver does not support non-default transaction options.
			// so return error if non-default transaction is requested.
			if opts.Isolation != driver.IsolationLevel(sql.LevelDefault) {
				return nil, errors.New("proxy: driver does not support non-default isolation level")
			}
			if opts.ReadOnly {
				return nil, errors.New("proxy: driver does not support read-only transactions")
			}
		}
		tx, err = conn.Conn.Begin()
		if err == nil {
			// check the context is already done.
			select {
			default:
			case <-c.Done():
				tx.Rollback()
				return nil, c.Err()
			}
		}
	}
	if err != nil {
		return nil, err
	}

	if err = conn.Proxy.Hooks.begin(c, ctx, conn); err != nil {
		tx.Rollback()
		return nil, err
	}

	return &Tx{
		Tx:    tx,
		Proxy: conn.Proxy,
		ctx:   c,
	}, nil
}

// Exec calls the original Exec method of the connection.
// It will trigger PreExec, Exec, PostExec hooks.
//
// If the original connection does not satisfy "database/sql/driver".Execer, it return ErrSkip error.
// NOT SUPPORTED: use ExecContext instead
func (conn *Conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	panic("not supported")
}

// ExecContext calls the original Exec method of the connection.
// It will trigger PreExec, Exec, PostExec hooks.
//
// If the original connection does not satisfy "database/sql/driver".Execer, it return ErrSkip error.
func (conn *Conn) ExecContext(c context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := conn.Conn.(driver.Execer)
	if !ok {
		return nil, driver.ErrSkip
	}

	stmt := &Stmt{
		QueryString: query,
		Proxy:       conn.Proxy,
	}

	var ctx interface{}
	var err error
	var result driver.Result

	defer func() { stmt.Proxy.Hooks.postExec(c, ctx, stmt, args, result, err) }()
	if ctx, err = stmt.Proxy.Hooks.preExec(c, stmt, args); err != nil {
		return nil, err
	}

	if execerCtx, ok := execer.(driver.ExecerContext); ok {
		result, err = execerCtx.ExecContext(c, query, args)
	} else {
		result, err = execer.Exec(query, namedValuesToValues(args))
		if err == nil {
			select {
			default:
			case <-c.Done():
				return result, c.Err()
			}
		}
	}
	if err != nil {
		return nil, err
	}

	if err = stmt.Proxy.Hooks.exec(c, ctx, stmt, args, result); err != nil {
		return nil, err
	}

	return result, nil
}

// Query executes a query that may return rows.
// It wil trigger PreQuery, Query, PostQuery hooks.
//
// If the original connection does not satisfy "database/sql/driver".Queryer, it return ErrSkip error.
// NOT SUPPORTED: use QueryContext instead
func (conn *Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	panic("not supported")
}

// Query executes a query that may return rows.
// It wil trigger PreQuery, Query, PostQuery hooks.
//
// If the original connection does not satisfy "database/sql/driver".Queryer, it return ErrSkip error.
func (conn *Conn) QueryContext(c context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := conn.Conn.(driver.Queryer)
	if !ok {
		return nil, driver.ErrSkip
	}

	stmt := &Stmt{
		QueryString: query,
		Proxy:       conn.Proxy,
	}

	var ctx interface{}
	var err error
	var rows driver.Rows

	defer func() { stmt.Proxy.Hooks.postQuery(c, ctx, stmt, args, rows, err) }()
	if ctx, err = stmt.Proxy.Hooks.preQuery(c, stmt, args); err != nil {
		return nil, err
	}

	rows, err = queryer.Query(query, namedValuesToValues(args)) // TODO: call QueryContext if conn.Conn satisfies ConnQueryContext
	if err != nil {
		return nil, err
	}

	if err = stmt.Proxy.Hooks.query(c, ctx, stmt, args, rows); err != nil {
		return nil, err
	}

	return rows, nil
}
