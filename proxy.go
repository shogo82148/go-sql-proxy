// a proxy package is a proxy driver for dabase/sql.

package proxy

import "database/sql/driver"

type Proxy struct {
	Driver driver.Driver
	Hooks  *Hooks
}

type Hooks struct {
	// PreOpen is a callback that gets called before any
	// attempt to open the sql connection is made, and is ALWAYS
	// called. If this callback returns an error, the underlying
	// driver's Open() method and Hooks.Open() are not called.
	// The first return value is passed to both Open() and PostOpen() hooks
	PreOpen  func(name string) (interface{}, error)

	// Open is only called if PreOpen() returns no error, and
	// the underlying driver's Open() returns no error.
	// `ctx` is the return value supplied from the PreOpen() hook
	Open     func(ctx interface{}, conn driver.Conn) error

	// PostOpen is a callback that gets called at the end of
	// the call to Proxy.Open(). It is ALWAYS called.
	// `ctx` is the return value supplied from the PreOpen() hook
	PostOpen func(ctx interface{}, conn driver.Conn) error

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
	var err error
	var ctx interface{}

	var conn driver.Conn
	if h := p.Hooks.PostOpen; h != nil {
		// Setup PostOpen. This needs to be a closure like this
		// or otherwise changes to the `ctx` and `conn` parameters
		// within this Open() method does not get applied at the
		// time defer is fired
		defer func() { h(ctx, conn) }()
	}

	if h := p.Hooks.PreOpen; h != nil {
		if ctx, err = h(name); err != nil {
			return nil, err
		}
	}
	conn, err = p.Driver.Open(name)
	if err != nil {
		return nil, err
	}

	conn = &Conn{
		Conn:  conn,
		Proxy: p,
	}

	if hook := p.Hooks.Open; hook != nil {
		if err = hook(ctx, conn); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}
