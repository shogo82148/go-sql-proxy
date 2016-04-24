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
	// called.
	//
	// The first return value is passed to both `Hooks.Open` and
	// `Hooks.PostOpen` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Driver.Open` method and `Hooks.Open`
	// methods are not called.
	PreOpen func(name string) (interface{}, error)

	// Open is called after the underlying driver's `Driver.Open` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreOpen` method, and may be nil.
	//
	// If this callback returns an error, then the `conn` object is
	// closed by calling the `Close` method, and the error from this
	// callback is returned by the `db.Open` method.
	Open func(ctx interface{}, conn driver.Conn) error

	// PostOpen is a callback that gets called at the end of
	// the call to `db.Open(). It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreOpen` method, and may be nil.
	PostOpen func(ctx interface{}, conn driver.Conn) error

	// PreExec is a callback that gets called prior to calling
	// `Stmt.Exec`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Stmt.Exec` and `Hooks.Exec` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Exec` and
	// `Hooks.PostExec` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Driver.Exec` method and `Hooks.Exec`
	// methods are not called.
	PreExec func(stmt *Stmt, args []driver.Value) (interface{}, error)

	// Exec is called after the underlying driver's `Driver.Exec` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreExec` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Stmt.Exec` method.
	Exec func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error

	// PostExec is a callback that gets called at the end of
	// the call to `Stmt.Exec`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreExec` method, and may be nil.
	PostExec func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error

	// PreQuery is a callback that gets called prior to calling
	// `Stmt.Query`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Stmt.Query` and `Hooks.Query` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Query` and
	// `Hooks.PostQuery` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Stmt.Query` method and `Hooks.Query`
	// methods are not called.
	PreQuery func(stmt *Stmt, args []driver.Value) (interface{}, error)

	// Query is called after the underlying driver's `Stmt.Query` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreQuery` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Stmt.Query` method.
	Query func(ctx interface{}, stmt *Stmt, args []driver.Value, rows driver.Rows) error

	// PostQuery is a callback that gets called at the end of
	// the call to `Stmt.Query`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreQuery` method, and may be nil.
	PostQuery func(ctx interface{}, stmt *Stmt, args []driver.Value, rows driver.Rows) error

	// PreBegin is a callback that gets called prior to calling
	// `Stmt.Begin`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Conn.Begin` and `Hooks.Begin` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Begin` and
	// `Hooks.PostBegin` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Conn.Begin` method and `Hooks.Begin`
	// methods are not called.
	PreBegin func(conn *Conn) (interface{}, error)

	// Begin is called after the underlying driver's `Conn.Begin` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreBegin` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.Begin` method.
	Begin func(ctx interface{}, conn *Conn) error

	// PostBegin is a callback that gets called at the end of
	// the call to `Conn.Begin`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreBegin` method, and may be nil.
	PostBegin func(ctx interface{}, conn *Conn) error

	// PreCommit is a callback that gets called prior to calling
	// `Tx.Commit`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Tx.Commit` and `Hooks.Commit` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Commit` and
	// `Hooks.PostCommit` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Tx.Commit` method and `Hooks.Commit`
	// methods are not called.
	PreCommit func(tx *Tx) (interface{}, error)

	// Commit is called after the underlying driver's `Tx.Commit` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreCommit` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Tx.Commit` method.
	Commit func(ctx interface{}, tx *Tx) error

	// PostCommit is a callback that gets called at the end of
	// the call to `Tx.Commit`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreCommit` method, and may be nil.
	PostCommit func(ctx interface{}, tx *Tx) error

	// PreRollback is a callback that gets called prior to calling
	// `Tx.Rollback`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Tx.Rollback` and `Hooks.Rollback` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Rollback` and
	// `Hooks.PostRollback` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Tx.Rollback` method and `Hooks.Rollback`
	PreRollback func(tx *Tx) (interface{}, error)

	// Rollback is called after the underlying driver's `Tx.Rollback` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreRollback` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Tx.Rollback` method.
	Rollback func(ctx interface{}, tx *Tx) error

	// PostRollback is a callback that gets called at the end of
	// the call to `Tx.Rollback`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreRollback` method, and may be nil.
	PostRollback func(ctx interface{}, tx *Tx) error
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
