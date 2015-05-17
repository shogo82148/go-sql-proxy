package proxy

import "database/sql/driver"

type Stmt struct {
	Stmt        driver.Stmt
	QueryString string
	Proxy       *Proxy
}

func (stmt *Stmt) Close() error {
	return stmt.Stmt.Close()
}

func (stmt *Stmt) NumInput() int {
	return stmt.Stmt.NumInput()
}

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
