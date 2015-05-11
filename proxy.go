// a proxy package is a proxy driver for dabase/sql.

package proxy

import "database/sql/driver"

type Proxy struct {
	Driver driver.Driver
	Hooks  *Hooks
}

type Hooks struct {
	Open     func(conn *Conn) error
	Exec     func(stmt *Stmt, args []driver.Value, result driver.Result) error
	Query    func(stmt *Stmt, args []driver.Value, rows driver.Rows) error
	Begin    func(conn *Conn) error
	Commit   func(tx *Tx) error
	Rollback func(tx *Tx) error
}

func NewProxy(driver driver.Driver, hooks *Hooks) *Proxy {
	if hooks == nil {
		hooks = &Hooks{}
	}
	return &Proxy{
		Driver: driver,
		Hooks:  hooks,
	}
}

func (p *Proxy) Open(name string) (driver.Conn, error) {
	conn, err := p.Driver.Open(name)
	if err != nil {
		return nil, err
	}

	proxyConn := &Conn{
		Conn:  conn,
		Proxy: p,
	}
	if hook := p.Hooks.Open; hook != nil {
		if err := hook(proxyConn); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return proxyConn, nil
}
