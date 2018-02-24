package proxy

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
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
	hooks := conn.Proxy.getHooks(c)

	if hooks != nil {
		defer func() { hooks.postPing(c, ctx, conn, err) }()
		if ctx, err = hooks.prePing(c, conn); err != nil {
			return err
		}
	}

	if p, ok := conn.Conn.(driver.Pinger); ok {
		err = p.Ping(c)
		if err != nil {
			return err
		}
	}

	if hooks != nil {
		err = hooks.ping(c, ctx, conn)
	}
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
	hooks := conn.Proxy.getHooks(c)
	if hooks != nil {
		defer func() { hooks.postBegin(c, ctx, conn, err) }()
		if ctx, err = hooks.preBegin(c, conn); err != nil {
			return nil, err
		}
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
				err = c.Err()
			}
		}
	}
	if err != nil {
		return nil, err
	}

	if hooks != nil {
		if err = hooks.begin(c, ctx, conn); err != nil {
			tx.Rollback()
			return nil, err
		}
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

	// set the hooks.
	var stmt *Stmt
	var ctx interface{}
	var err error
	var result driver.Result
	hooks := conn.Proxy.getHooks(c)
	if hooks != nil {
		stmt = &Stmt{
			QueryString: query,
			Proxy:       conn.Proxy,
		}
		defer func() { hooks.postExec(c, ctx, stmt, args, result, err) }()
		if ctx, err = hooks.preExec(c, stmt, args); err != nil {
			return nil, err
		}
	}

	// call the original method.
	if execerCtx, ok := execer.(driver.ExecerContext); ok {
		result, err = execerCtx.ExecContext(c, query, args)
	} else {
		select {
		default:
		case <-c.Done():
			return nil, c.Err()
		}
		dargs, err0 := namedValuesToValues(args)
		if err0 != nil {
			return nil, err0
		}
		result, err = execer.Exec(query, dargs)
	}
	if err != nil {
		return nil, err
	}

	if hooks != nil {
		if err = hooks.exec(c, ctx, stmt, args, result); err != nil {
			return nil, err
		}
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

	var stmt *Stmt
	var ctx interface{}
	var err error
	var rows driver.Rows
	hooks := conn.Proxy.getHooks(c)
	if hooks != nil {
		stmt := &Stmt{
			QueryString: query,
			Proxy:       conn.Proxy,
		}
		defer func() { hooks.postQuery(c, ctx, stmt, args, rows, err) }()
		if ctx, err = hooks.preQuery(c, stmt, args); err != nil {
			return nil, err
		}
	}

	// call the original method.
	if queryerCtx, ok := conn.Conn.(driver.QueryerContext); ok {
		rows, err = queryerCtx.QueryContext(c, query, args)
	} else {
		select {
		default:
		case <-c.Done():
			return nil, c.Err()
		}
		dargs, err0 := namedValuesToValues(args)
		if err0 != nil {
			return nil, err0
		}
		rows, err = queryer.Query(query, dargs)
	}
	if err != nil {
		return nil, err
	}

	if hooks != nil {
		if err = hooks.query(c, ctx, stmt, args, rows); err != nil {
			rows.Close()
			return nil, err
		}
	}

	return rows, nil
}
