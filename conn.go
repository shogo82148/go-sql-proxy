package proxy

import (
	"database/sql/driver"
)

type Conn struct {
	Conn  driver.Conn
	Proxy *Proxy
}

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

func (conn *Conn) Close() error {
	return conn.Conn.Close()
}

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
