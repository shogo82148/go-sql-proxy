package proxy

import (
	"database/sql/driver"
)

// Conn adds hook points into "database/sql/driver".Conn.
type Conn struct {
	Conn  driver.Conn
	Proxy *Proxy
}

// Prepare returns a prepared statement which is wrapped by Stmt.
func (conn *Conn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := conn.Conn.Prepare(query)
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
	var err error
	var ctx interface{}

	var tx driver.Tx
	if h := conn.Proxy.Hooks.PostBegin; h != nil {
		defer func() { h(ctx, conn) }()
	}

	if h := conn.Proxy.Hooks.PreBegin; h != nil {
		if ctx, err = h(conn); err != nil {
			return nil, err
		}
	}

	tx, err = conn.Conn.Begin()
	if err != nil {
		return nil, err
	}

	if hook := conn.Proxy.Hooks.Begin; hook != nil {
		if err = hook(ctx, conn); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return &Tx{
		Tx:    tx,
		Proxy: conn.Proxy,
	}, nil
}

// Exec calls the original Exec method of the connection.
// It will trigger PreExec, Exec, PostExec hooks.
//
// If the original connection does not satisfy "database/sql/driver".Execer, it return ErrSkip error.
func (conn *Conn) Exec(query string, args []driver.Value) (driver.Result, error) {
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

	if h := stmt.Proxy.Hooks.PostExec; h != nil {
		defer func() { h(ctx, stmt, args, result) }()
	}
	if h := stmt.Proxy.Hooks.PreExec; h != nil {
		if ctx, err = h(stmt, args); err != nil {
			return nil, err
		}
	}

	result, err = execer.Exec(query, args)
	if err != nil {
		return nil, err
	}

	if h := stmt.Proxy.Hooks.Exec; h != nil {
		if err := h(ctx, stmt, args, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Query executes a query that may return rows.
// It wil trigger PreQuery, Query, PostQuery hooks.
//
// If the orginal connection does not satisfy "database/sql/driver".Queryer, it return ErrSkip error.
func (conn *Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
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

	if h := stmt.Proxy.Hooks.PostQuery; h != nil {
		defer func() { h(ctx, stmt, args, rows) }()
	}

	if h := stmt.Proxy.Hooks.PreQuery; h != nil {
		if ctx, err = h(stmt, args); err != nil {
			return nil, err
		}
	}

	rows, err = queryer.Query(query, args)
	if err != nil {
		return nil, err
	}

	if h := stmt.Proxy.Hooks.Query; h != nil {
		if err := h(ctx, stmt, args, rows); err != nil {
			return nil, err
		}
	}

	return rows, nil
}
