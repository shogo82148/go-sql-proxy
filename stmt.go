package proxy

import "database/sql/driver"

// Stmt adds hook points into "database/sql/driver".Stmt.
type Stmt struct {
	// Stmt is the original statement.
	// It may be nil because some sql drivers support skipping Prepare.
	Stmt driver.Stmt

	QueryString string
	Proxy       *Proxy
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
func (stmt *Stmt) Exec(args []driver.Value) (driver.Result, error) {
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

	result, err = stmt.Stmt.Exec(args)
	if err != nil {
		return nil, err
	}

	if hook := stmt.Proxy.Hooks.Exec; hook != nil {
		if err := hook(ctx, stmt, args, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Query executes a query that may return rows.
// It wil trigger PreQuery, Query, PostQuery hooks.
func (stmt *Stmt) Query(args []driver.Value) (driver.Rows, error) {
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

	rows, err = stmt.Stmt.Query(args)
	if err != nil {
		return nil, err
	}

	if hook := stmt.Proxy.Hooks.Query; hook != nil {
		if err := hook(ctx, stmt, args, rows); err != nil {
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
