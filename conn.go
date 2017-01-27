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

// Prepare returns a prepared statement which is wrapped by Stmt.
func (conn *Conn) Prepare(query string) (driver.Stmt, error) {
	return conn.PrepareContext(context.Background(), query)
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
func (conn *Conn) Begin() (driver.Tx, error) {
	return conn.BeginContext(context.Background())
}

// BeginContext starts and returns a new transaction which is wrapped by Tx.
// It will trigger PreBegin, Begin, PostBegin hooks.
func (conn *Conn) BeginContext(c context.Context) (driver.Tx, error) {
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
func (conn *Conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return conn.ExecContext(context.Background(), query, args)
}

// ExecContext calls the original Exec method of the connection.
// It will trigger PreExec, Exec, PostExec hooks.
//
// If the original connection does not satisfy "database/sql/driver".Execer, it return ErrSkip error.
func (conn *Conn) ExecContext(c context.Context, query string, args []driver.Value) (driver.Result, error) {
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

	result, err = execer.Exec(query, args) // TODO: call ExecContext if conn.Conn satisfies ConnExecContext
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
func (conn *Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return conn.QueryContext(context.Background(), query, args)
}

// Query executes a query that may return rows.
// It wil trigger PreQuery, Query, PostQuery hooks.
//
// If the original connection does not satisfy "database/sql/driver".Queryer, it return ErrSkip error.
func (conn *Conn) QueryContext(c context.Context, query string, args []driver.Value) (driver.Rows, error) {
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

	rows, err = queryer.Query(query, args) // TODO: call QueryContext if conn.Conn satisfies ConnQueryContext
	if err != nil {
		return nil, err
	}

	if err = stmt.Proxy.Hooks.query(c, ctx, stmt, args, rows); err != nil {
		return nil, err
	}

	return rows, nil
}
