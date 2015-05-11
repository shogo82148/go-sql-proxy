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
	result, err := stmt.Stmt.Exec(args)
	if err != nil {
		return nil, err
	}

	if hook := stmt.Proxy.Hooks.Exec; hook != nil {
		if err := hook(stmt, args, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (stmt *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	rows, err := stmt.Stmt.Query(args)
	if err != nil {
		return nil, err
	}

	if hook := stmt.Proxy.Hooks.Query; hook != nil {
		if err := hook(stmt, args, rows); err != nil {
			return nil, err
		}
	}

	return rows, nil
}
