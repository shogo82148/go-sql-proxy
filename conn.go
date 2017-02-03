// +build go1.8

package proxy

import (
	"context"
	"database/sql/driver"
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
	stmt, err := conn.Conn.Prepare(query) // TODO: call PrepareContext if conn.Conn satisfies ConnPrepareContext
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
	var err error
	var ctx interface{}

	var tx driver.Tx
	defer func() { conn.Proxy.Hooks.postBegin(c, ctx, conn, err) }()

	if ctx, err = conn.Proxy.Hooks.preBegin(c, conn); err != nil {
		return nil, err
	}

	tx, err = conn.Conn.Begin() // TODO: call BeginContext if conn.Conn satisfies ConnBeginContext
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

	result, err = execer.Exec(query, namedValuesToValues(args)) // TODO: call ExecContext if conn.Conn satisfies ConnExecContext
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
