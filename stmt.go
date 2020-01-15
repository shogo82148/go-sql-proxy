package proxy

import (
	"context"
	"database/sql/driver"
)

// Stmt adds hook points into "database/sql/driver".Stmt.
type Stmt struct {
	// Stmt is the original statement.
	// It may be nil because some sql drivers support skipping Prepare.
	Stmt driver.Stmt

	QueryString string
	Proxy       *Proxy
	Conn        *Conn
}

// Close closes the statement.
// It just calls the original Close method.
func (stmt *Stmt) Close() error {
	return stmt.Stmt.Close()
}

// NumInput returns the number of placeholder parameters.
// It just calls the original NumInput method.
func (stmt *Stmt) NumInput() int {
	return stmt.Stmt.NumInput()
}

// Exec executes a query that doesn't return rows.
// It will trigger PreExec, Exec, PostExec hooks.
// NOT SUPPORTED: use ExecContext instead
func (stmt *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	panic("not supported")
}

// ExecContext executes a query that doesn't return rows.
// It will trigger PreExec, Exec, PostExec hooks.
func (stmt *Stmt) ExecContext(c context.Context, args []driver.NamedValue) (driver.Result, error) {
	var ctx interface{}
	var err error
	var result driver.Result
	hooks := stmt.Proxy.getHooks(c)
	if hooks != nil {
		defer func() { hooks.postExec(c, ctx, stmt, args, result, err) }()
		if ctx, err = hooks.preExec(c, stmt, args); err != nil {
			return nil, err
		}
	}

	if execerContext, ok := stmt.Stmt.(driver.StmtExecContext); ok {
		result, err = execerContext.ExecContext(c, args)
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
		result, err = stmt.Stmt.Exec(dargs)
	}
	if err != nil {
		return result, err
	}

	if hooks != nil {
		if err = hooks.exec(c, ctx, stmt, args, result); err != nil {
			return result, err
		}
	}

	return result, nil
}

// Query executes a query that may return rows.
// It wil trigger PreQuery, Query, PostQuery hooks.
// NOT SUPPORTED: use QueryContext instead
func (stmt *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("not supported")
}

// QueryContext executes a query that may return rows.
// It wil trigger PreQuery, Query, PostQuery hooks.
func (stmt *Stmt) QueryContext(c context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var ctx interface{}
	var err error
	var rows driver.Rows
	hooks := stmt.Proxy.getHooks(c)
	if hooks != nil {
		defer func() { hooks.postQuery(c, ctx, stmt, args, rows, err) }()
		if ctx, err = hooks.preQuery(c, stmt, args); err != nil {
			return nil, err
		}
	}

	if queryCtx, ok := stmt.Stmt.(driver.StmtQueryContext); ok {
		rows, err = queryCtx.QueryContext(c, args)
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
		rows, err = stmt.Stmt.Query(dargs)
	}
	if err != nil {
		return nil, err
	}

	if hooks != nil {
		if err = hooks.query(c, ctx, stmt, args, rows); err != nil {
			return nil, err
		}
	}

	return rows, nil
}

// ColumnConverter returns a ValueConverter for the provided column index.
// If the original statement does not satisfy ColumnConverter,
// it returns driver.DefaultParameterConverter.
func (stmt *Stmt) ColumnConverter(idx int) driver.ValueConverter {
	if conv, ok := stmt.Stmt.(driver.ColumnConverter); ok {
		return conv.ColumnConverter(idx)
	}
	return driver.DefaultParameterConverter
}

// CheckNamedValue for implementing NamedValueChecker
func (stmt *Stmt) CheckNamedValue(nv *driver.NamedValue) (err error) {
	if nvc, ok := stmt.Stmt.(namedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// When converting data in sql/driver/convert.go, it is checked first whether the `stmt`
	// implements `NamedValueChecker`, and then checks if `conn` implements NamedValueChecker.
	// In the case of "go-sql-proxy", the `proxy.Stmt` "implements" `CheckNamedValue` here,
	// so we also check both `stmt` and `conn` inside here.
	if nvc, ok := stmt.Conn.Conn.(namedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// fallback to default
	return defaultCheckNamedValue(nv)
}
