// a proxy package is a proxy driver for database/sql.

package proxy

import (
	"context"
	"database/sql/driver"
	"errors"
)

// hooks is callback functions for the proxy.
// it is private because it doesn't guarantee backward compatibility.
type hooks interface {
	prePing(c context.Context, conn *Conn) (interface{}, error)
	ping(c context.Context, ctx interface{}, conn *Conn) error
	postPing(c context.Context, ctx interface{}, conn *Conn, err error) error
	preOpen(c context.Context, name string) (interface{}, error)
	open(c context.Context, ctx interface{}, conn *Conn) error
	postOpen(c context.Context, ctx interface{}, conn *Conn, err error) error
	preExec(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error)
	exec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result) error
	postExec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result, err error) error
	preQuery(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error)
	query(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows) error
	postQuery(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows, err error) error
	preBegin(c context.Context, conn *Conn) (interface{}, error)
	begin(c context.Context, ctx interface{}, conn *Conn) error
	postBegin(c context.Context, ctx interface{}, conn *Conn, err error) error
	preCommit(c context.Context, tx *Tx) (interface{}, error)
	commit(c context.Context, ctx interface{}, tx *Tx) error
	postCommit(c context.Context, ctx interface{}, tx *Tx, err error) error
	preRollback(c context.Context, tx *Tx) (interface{}, error)
	rollback(c context.Context, ctx interface{}, tx *Tx) error
	postRollback(c context.Context, ctx interface{}, tx *Tx, err error) error
	preClose(c context.Context, conn *Conn) (interface{}, error)
	close(c context.Context, ctx interface{}, conn *Conn) error
	postClose(c context.Context, ctx interface{}, conn *Conn, err error) error
	preResetSession(c context.Context, conn *Conn) (interface{}, error)
	resetSession(c context.Context, ctx interface{}, conn *Conn) error
	postResetSession(c context.Context, ctx interface{}, conn *Conn, err error) error
}

// HooksContext is callback functions with context.Context for the proxy.
type HooksContext struct {
	// PrePing is a callback that gets called prior to calling
	// `Conn.Ping`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Conn.Ping` and `Hooks.Ping` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Ping` and
	// `Hooks.PostPing` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Conn.Ping` method and `Hooks.Ping`
	// methods are not called.
	PrePing func(c context.Context, conn *Conn) (interface{}, error)

	// Ping is called after the underlying driver's `Conn.Exec` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PrePing` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.Ping` method.
	Ping func(c context.Context, ctx interface{}, conn *Conn) error

	// PostPing is a callback that gets called at the end of
	// the call to `Conn.Ping`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PrePing` method, and may be nil.
	PostPing func(c context.Context, ctx interface{}, conn *Conn, err error) error

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
	PreOpen func(c context.Context, name string) (interface{}, error)

	// Open is called after the underlying driver's `Driver.Open` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreOpen` method, and may be nil.
	//
	// If this callback returns an error, then the `conn` object is
	// closed by calling the `Close` method, and the error from this
	// callback is returned by the `db.Open` method.
	Open func(c context.Context, ctx interface{}, conn *Conn) error

	// PostOpen is a callback that gets called at the end of
	// the call to `db.Open(). It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreOpen` method, and may be nil.
	PostOpen func(c context.Context, ctx interface{}, conn *Conn, err error) error

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
	PreExec func(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error)

	// Exec is called after the underlying driver's `Driver.Exec` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreExec` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Stmt.Exec` method.
	Exec func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result) error

	// PostExec is a callback that gets called at the end of
	// the call to `Stmt.Exec`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreExec` method, and may be nil.
	PostExec func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result, err error) error

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
	PreQuery func(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error)

	// Query is called after the underlying driver's `Stmt.Query` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreQuery` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Stmt.Query` method.
	Query func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows) error

	// PostQuery is a callback that gets called at the end of
	// the call to `Stmt.Query`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreQuery` method, and may be nil.
	PostQuery func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows, err error) error

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
	PreBegin func(c context.Context, conn *Conn) (interface{}, error)

	// Begin is called after the underlying driver's `Conn.Begin` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreBegin` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.Begin` method.
	Begin func(c context.Context, ctx interface{}, conn *Conn) error

	// PostBegin is a callback that gets called at the end of
	// the call to `Conn.Begin`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreBegin` method, and may be nil.
	PostBegin func(c context.Context, ctx interface{}, conn *Conn, err error) error

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
	PreCommit func(c context.Context, tx *Tx) (interface{}, error)

	// Commit is called after the underlying driver's `Tx.Commit` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreCommit` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Tx.Commit` method.
	Commit func(c context.Context, ctx interface{}, tx *Tx) error

	// PostCommit is a callback that gets called at the end of
	// the call to `Tx.Commit`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreCommit` method, and may be nil.
	PostCommit func(c context.Context, ctx interface{}, tx *Tx, err error) error

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
	PreRollback func(c context.Context, tx *Tx) (interface{}, error)

	// Rollback is called after the underlying driver's `Tx.Rollback` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreRollback` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Tx.Rollback` method.
	Rollback func(c context.Context, ctx interface{}, tx *Tx) error

	// PostRollback is a callback that gets called at the end of
	// the call to `Tx.Rollback`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreRollback` method, and may be nil.
	PostRollback func(c context.Context, ctx interface{}, tx *Tx, err error) error

	// PreClose is a callback that gets called prior to calling
	// `Conn.Close`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Conn.Close` and `Hooks.Close` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Close` and
	// `Hooks.PostClose` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Conn.Close` method and `Hooks.Close`
	// methods are not called.
	PreClose func(c context.Context, conn *Conn) (interface{}, error)

	// Close is called after the underlying driver's `Conn.Close` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreClose` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.Close` method.
	Close func(c context.Context, ctx interface{}, conn *Conn) error

	// PostClose is a callback that gets called at the end of
	// the call to `Conn.Close`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreClose` method, and may be nil.
	PostClose func(c context.Context, ctx interface{}, conn *Conn, err error) error

	// PreResetSession is a callback that gets called prior to calling
	// `Conn.ResetSession`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Conn.ResetSession` and `Hooks.ResetSession` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.ResetSession` and
	// `Hooks.PostResetSession` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Conn.ResetSession` method and `Hooks.ResetSession`
	// methods are not called.
	PreResetSession func(c context.Context, conn *Conn) (interface{}, error)

	// ResetSession is called after the underlying driver's `Conn.ResetSession` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreResetSession` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.ResetSession` method.
	ResetSession func(c context.Context, ctx interface{}, conn *Conn) error

	// PostResetSession is a callback that gets called at the end of
	// the call to `Conn.ResetSession`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreResetSession` method, and may be nil.
	PostResetSession func(c context.Context, ctx interface{}, conn *Conn, err error) error
}

func (h *HooksContext) prePing(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PrePing == nil {
		return nil, nil
	}
	return h.PrePing(c, conn)
}

func (h *HooksContext) ping(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Ping == nil {
		return nil
	}
	return h.Ping(c, ctx, conn)
}

func (h *HooksContext) postPing(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostPing == nil {
		return nil
	}
	return h.PostPing(c, ctx, conn, err)
}

func (h *HooksContext) preOpen(c context.Context, name string) (interface{}, error) {
	if h == nil || h.PreOpen == nil {
		return nil, nil
	}
	return h.PreOpen(c, name)
}

func (h *HooksContext) open(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Open == nil {
		return nil
	}
	return h.Open(c, ctx, conn)
}

func (h *HooksContext) postOpen(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostOpen == nil {
		return nil
	}
	return h.PostOpen(c, ctx, conn, err)
}

func (h *HooksContext) preExec(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	if h == nil || h.PreExec == nil {
		return nil, nil
	}
	return h.PreExec(c, stmt, args)
}

func (h *HooksContext) exec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result) error {
	if h == nil || h.Exec == nil {
		return nil
	}
	return h.Exec(c, ctx, stmt, args, result)
}

func (h *HooksContext) postExec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result, err error) error {
	if h == nil || h.PostExec == nil {
		return nil
	}
	return h.PostExec(c, ctx, stmt, args, result, err)
}

func (h *HooksContext) preQuery(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	if h == nil || h.PreQuery == nil {
		return nil, nil
	}
	return h.PreQuery(c, stmt, args)
}

func (h *HooksContext) query(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows) error {
	if h == nil || h.Query == nil {
		return nil
	}
	return h.Query(c, ctx, stmt, args, rows)
}

func (h *HooksContext) postQuery(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows, err error) error {
	if h == nil || h.PostQuery == nil {
		return nil
	}
	return h.PostQuery(c, ctx, stmt, args, rows, err)
}

func (h *HooksContext) preBegin(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PreBegin == nil {
		return nil, nil
	}
	return h.PreBegin(c, conn)
}

func (h *HooksContext) begin(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Open == nil {
		return nil
	}
	return h.Begin(c, ctx, conn)
}

func (h *HooksContext) postBegin(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostBegin == nil {
		return nil
	}
	return h.PostBegin(c, ctx, conn, err)
}

func (h *HooksContext) preCommit(c context.Context, tx *Tx) (interface{}, error) {
	if h == nil || h.PreCommit == nil {
		return nil, nil
	}
	return h.PreCommit(c, tx)
}

func (h *HooksContext) commit(c context.Context, ctx interface{}, tx *Tx) error {
	if h == nil || h.Commit == nil {
		return nil
	}
	return h.Commit(c, ctx, tx)
}

func (h *HooksContext) postCommit(c context.Context, ctx interface{}, tx *Tx, err error) error {
	if h == nil || h.PostCommit == nil {
		return nil
	}
	return h.PostCommit(c, ctx, tx, err)
}

func (h *HooksContext) preRollback(c context.Context, tx *Tx) (interface{}, error) {
	if h == nil || h.PreRollback == nil {
		return nil, nil
	}
	return h.PreRollback(c, tx)
}

func (h *HooksContext) rollback(c context.Context, ctx interface{}, tx *Tx) error {
	if h == nil || h.Rollback == nil {
		return nil
	}
	return h.Rollback(c, ctx, tx)
}

func (h *HooksContext) postRollback(c context.Context, ctx interface{}, tx *Tx, err error) error {
	if h == nil || h.PostRollback == nil {
		return nil
	}
	return h.PostRollback(c, ctx, tx, err)
}

func (h *HooksContext) preClose(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PreClose == nil {
		return nil, nil
	}
	return h.PreClose(c, conn)
}

func (h *HooksContext) close(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Close == nil {
		return nil
	}
	return h.Close(c, ctx, conn)
}

func (h *HooksContext) postClose(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostClose == nil {
		return nil
	}
	return h.PostClose(c, ctx, conn, err)
}

func (h *HooksContext) preResetSession(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PreResetSession == nil {
		return nil, nil
	}
	return h.PreResetSession(c, conn)
}

func (h *HooksContext) resetSession(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.ResetSession == nil {
		return nil
	}
	return h.ResetSession(c, ctx, conn)
}

func (h *HooksContext) postResetSession(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostResetSession == nil {
		return nil
	}
	return h.PostResetSession(c, ctx, conn, err)
}

// Hooks is callback functions for the proxy.
// Deprecated: You should use HooksContext instead.
type Hooks struct {
	// PrePing is a callback that gets called prior to calling
	// `Conn.Ping`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Conn.Ping` and `Hooks.Ping` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Ping` and
	// `Hooks.PostPing` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Conn.Ping` method and `Hooks.Ping`
	// methods are not called.
	PrePing func(conn *Conn) (interface{}, error)

	// Ping is called after the underlying driver's `Conn.Exec` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PrePing` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.Ping` method.
	Ping func(ctx interface{}, conn *Conn) error

	// PostPing is a callback that gets called at the end of
	// the call to `Conn.Ping`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PrePing` method, and may be nil.
	PostPing func(ctx interface{}, conn *Conn, err error) error

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
	Open func(ctx interface{}, conn *Conn) error

	// PostOpen is a callback that gets called at the end of
	// the call to `db.Open(). It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreOpen` method, and may be nil.
	PostOpen func(ctx interface{}, conn *Conn) error

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

	// PreClose is a callback that gets called prior to calling
	// `Conn.Close`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Conn.Close` and `Hooks.Close` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.Close` and
	// `Hooks.PostClose` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Conn.Close` method and `Hooks.Close`
	// methods are not called.
	PreClose func(conn *Conn) (interface{}, error)

	// Close is called after the underlying driver's `Conn.Close` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreClose` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.Close` method.
	Close func(ctx interface{}, conn *Conn) error

	// PostClose is a callback that gets called at the end of
	// the call to `Conn.Close`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreClose` method, and may be nil.
	PostClose func(ctx interface{}, conn *Conn, err error) error

	// PreResetSession is a callback that gets called prior to calling
	// `Conn.ResetSession`, and is ALWAYS called. If this callback returns an
	// error, the underlying driver's `Conn.ResetSession` and `Hooks.ResetSession` methods
	// are not called.
	//
	// The first return value is passed to both `Hooks.ResetSession` and
	// `Hooks.PostResetSession` callbacks. You may specify anything you want.
	// Return nil if you do not need to use it.
	//
	// The second return value is indicates the error found while
	// executing this hook. If this callback returns an error,
	// the underlying driver's `Conn.ResetSession` method and `Hooks.ResetSession`
	// methods are not called.
	PreResetSession func(conn *Conn) (interface{}, error)

	// ResetSession is called after the underlying driver's `Conn.ResetSession` method
	// returns without any errors.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreResetSession` method, and may be nil.
	//
	// If this callback returns an error, then the error from this
	// callback is returned by the `Conn.ResetSession` method.
	ResetSession func(ctx interface{}, conn *Conn) error

	// PostResetSession is a callback that gets called at the end of
	// the call to `Conn.ResetSession`. It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreResetSession` method, and may be nil.
	PostResetSession func(ctx interface{}, conn *Conn, err error) error
}

func namedValuesToValues(args []driver.NamedValue) ([]driver.Value, error) {
	var err error
	ret := make([]driver.Value, len(args))
	for _, arg := range args {
		if len(arg.Name) > 0 {
			err = errors.New("proxy: driver does not support the use of Named Parameters")
		}
		ret[arg.Ordinal-1] = arg.Value
	}
	return ret, err
}

func valuesToNamedValues(args []driver.Value) []driver.NamedValue {
	ret := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		ret[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	return ret
}

func (h *Hooks) prePing(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PrePing == nil {
		return nil, nil
	}
	return h.PrePing(conn)
}

func (h *Hooks) ping(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Ping == nil {
		return nil
	}
	return h.Ping(ctx, conn)
}

func (h *Hooks) postPing(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostPing == nil {
		return nil
	}
	return h.PostPing(ctx, conn, err)
}

func (h *Hooks) preOpen(c context.Context, name string) (interface{}, error) {
	if h == nil || h.PreOpen == nil {
		return nil, nil
	}
	return h.PreOpen(name)
}

func (h *Hooks) open(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Open == nil {
		return nil
	}
	return h.Open(ctx, conn)
}

func (h *Hooks) postOpen(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostOpen == nil {
		return nil
	}
	return h.PostOpen(ctx, conn)
}

func (h *Hooks) preExec(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	if h == nil || h.PreExec == nil {
		return nil, nil
	}
	dargs, _ := namedValuesToValues(args)
	return h.PreExec(stmt, dargs)
}

func (h *Hooks) exec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result) error {
	if h == nil || h.Exec == nil {
		return nil
	}
	dargs, _ := namedValuesToValues(args)
	return h.Exec(ctx, stmt, dargs, result)
}

func (h *Hooks) postExec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result, err error) error {
	if h == nil || h.PostExec == nil {
		return nil
	}
	dargs, _ := namedValuesToValues(args)
	return h.PostExec(ctx, stmt, dargs, result)
}

func (h *Hooks) preQuery(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	if h == nil || h.PreQuery == nil {
		return nil, nil
	}
	dargs, _ := namedValuesToValues(args)
	return h.PreQuery(stmt, dargs)
}

func (h *Hooks) query(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows) error {
	if h == nil || h.Query == nil {
		return nil
	}
	dargs, _ := namedValuesToValues(args)
	return h.Query(ctx, stmt, dargs, rows)
}

func (h *Hooks) postQuery(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows, err error) error {
	if h == nil || h.PostQuery == nil {
		return nil
	}
	dargs, _ := namedValuesToValues(args)
	return h.PostQuery(ctx, stmt, dargs, rows)
}

func (h *Hooks) preBegin(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PreBegin == nil {
		return nil, nil
	}
	return h.PreBegin(conn)
}

func (h *Hooks) begin(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Open == nil {
		return nil
	}
	return h.Begin(ctx, conn)
}

func (h *Hooks) postBegin(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostBegin == nil {
		return nil
	}
	return h.PostBegin(ctx, conn)
}

func (h *Hooks) preCommit(c context.Context, tx *Tx) (interface{}, error) {
	if h == nil || h.PreCommit == nil {
		return nil, nil
	}
	return h.PreCommit(tx)
}

func (h *Hooks) commit(c context.Context, ctx interface{}, tx *Tx) error {
	if h == nil || h.Commit == nil {
		return nil
	}
	return h.Commit(ctx, tx)
}

func (h *Hooks) postCommit(c context.Context, ctx interface{}, tx *Tx, err error) error {
	if h == nil || h.PostCommit == nil {
		return nil
	}
	return h.PostCommit(ctx, tx)
}

func (h *Hooks) preRollback(c context.Context, tx *Tx) (interface{}, error) {
	if h == nil || h.PreRollback == nil {
		return nil, nil
	}
	return h.PreRollback(tx)
}

func (h *Hooks) rollback(c context.Context, ctx interface{}, tx *Tx) error {
	if h == nil || h.Rollback == nil {
		return nil
	}
	return h.Rollback(ctx, tx)
}

func (h *Hooks) postRollback(c context.Context, ctx interface{}, tx *Tx, err error) error {
	if h == nil || h.PostRollback == nil {
		return nil
	}
	return h.PostRollback(ctx, tx)
}

func (h *Hooks) preClose(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PreClose == nil {
		return nil, nil
	}
	return h.PreClose(conn)
}

func (h *Hooks) close(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.Close == nil {
		return nil
	}
	return h.Close(ctx, conn)
}

func (h *Hooks) postClose(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostClose == nil {
		return nil
	}
	return h.PostClose(ctx, conn, err)
}

func (h *Hooks) preResetSession(c context.Context, conn *Conn) (interface{}, error) {
	if h == nil || h.PreResetSession == nil {
		return nil, nil
	}
	return h.PreResetSession(conn)
}

func (h *Hooks) resetSession(c context.Context, ctx interface{}, conn *Conn) error {
	if h == nil || h.ResetSession == nil {
		return nil
	}
	return h.ResetSession(ctx, conn)
}

func (h *Hooks) postResetSession(c context.Context, ctx interface{}, conn *Conn, err error) error {
	if h == nil || h.PostResetSession == nil {
		return nil
	}
	return h.PostResetSession(ctx, conn, err)
}

type multipleHooks []hooks

func (h multipleHooks) preDo(f func(h hooks) (interface{}, error)) (interface{}, error) {
	if len(h) == 0 {
		return nil, nil
	}
	ctx := make([]interface{}, len(h))
	var err error
	for i, hk := range h {
		ctx0, err0 := f(hk)
		ctx[i] = ctx0
		if err0 != nil && err == nil {
			err = err0
		}
	}
	return ctx, err
}

func (h multipleHooks) do(ctx interface{}, f func(h hooks, ctx interface{}) error) error {
	if len(h) == 0 {
		return nil
	}
	sctx, ok := ctx.([]interface{})
	if !ok {
		return errors.New("invalid context type")
	}
	for i, hk := range h {
		if err := f(hk, sctx[i]); err != nil {
			return err
		}
	}
	return nil
}

func (h multipleHooks) postDo(ctx interface{}, err error, f func(h hooks, ctx interface{}, err error) error) error {
	if len(h) == 0 {
		return nil
	}
	sctx, ok := ctx.([]interface{})
	if !ok {
		return errors.New("invalid context type")
	}
	var reterr error
	for i := len(h) - 1; i >= 0; i-- {
		if err0 := f(h[i], sctx[i], err); err0 != nil {
			if err == nil {
				err = err0
			}
			if reterr == nil {
				reterr = err0
			}
		}
	}
	return reterr
}

func (h multipleHooks) prePing(c context.Context, conn *Conn) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.prePing(c, conn)
	})
}

func (h multipleHooks) ping(c context.Context, ctx interface{}, conn *Conn) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.ping(c, ctx, conn)
	})
}

func (h multipleHooks) postPing(c context.Context, ctx interface{}, conn *Conn, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postPing(c, ctx, conn, err)
	})
}

func (h multipleHooks) preOpen(c context.Context, name string) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preOpen(c, name)
	})
}

func (h multipleHooks) open(c context.Context, ctx interface{}, conn *Conn) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.open(c, ctx, conn)
	})
}

func (h multipleHooks) postOpen(c context.Context, ctx interface{}, conn *Conn, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postOpen(c, ctx, conn, err)
	})
}

func (h multipleHooks) preExec(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preExec(c, stmt, args)
	})
}

func (h multipleHooks) exec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.exec(c, ctx, stmt, args, result)
	})
}

func (h multipleHooks) postExec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postExec(c, ctx, stmt, args, result, err)
	})
}

func (h multipleHooks) preQuery(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preQuery(c, stmt, args)
	})
}

func (h multipleHooks) query(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.query(c, ctx, stmt, args, rows)
	})
}

func (h multipleHooks) postQuery(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postQuery(c, ctx, stmt, args, rows, err)
	})
}

func (h multipleHooks) preBegin(c context.Context, conn *Conn) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preBegin(c, conn)
	})
}

func (h multipleHooks) begin(c context.Context, ctx interface{}, conn *Conn) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.begin(c, ctx, conn)
	})
}

func (h multipleHooks) postBegin(c context.Context, ctx interface{}, conn *Conn, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postBegin(c, ctx, conn, err)
	})
}

func (h multipleHooks) preCommit(c context.Context, tx *Tx) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preCommit(c, tx)
	})
}

func (h multipleHooks) commit(c context.Context, ctx interface{}, tx *Tx) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.commit(c, ctx, tx)
	})
}

func (h multipleHooks) postCommit(c context.Context, ctx interface{}, tx *Tx, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postCommit(c, ctx, tx, err)
	})
}

func (h multipleHooks) preRollback(c context.Context, tx *Tx) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preRollback(c, tx)
	})
}

func (h multipleHooks) rollback(c context.Context, ctx interface{}, tx *Tx) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.rollback(c, ctx, tx)
	})
}

func (h multipleHooks) postRollback(c context.Context, ctx interface{}, tx *Tx, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postRollback(c, ctx, tx, err)
	})
}

func (h multipleHooks) preClose(c context.Context, conn *Conn) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preClose(c, conn)
	})
}

func (h multipleHooks) close(c context.Context, ctx interface{}, conn *Conn) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.close(c, ctx, conn)
	})
}

func (h multipleHooks) postClose(c context.Context, ctx interface{}, conn *Conn, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postClose(c, ctx, conn, err)
	})
}

func (h multipleHooks) preResetSession(c context.Context, conn *Conn) (interface{}, error) {
	return h.preDo(func(h hooks) (interface{}, error) {
		return h.preResetSession(c, conn)
	})
}

func (h multipleHooks) resetSession(c context.Context, ctx interface{}, conn *Conn) error {
	return h.do(ctx, func(h hooks, ctx interface{}) error {
		return h.resetSession(c, ctx, conn)
	})
}

func (h multipleHooks) postResetSession(c context.Context, ctx interface{}, conn *Conn, err error) error {
	return h.postDo(ctx, err, func(h hooks, ctx interface{}, err error) error {
		return h.postResetSession(c, ctx, conn, err)
	})
}

type contextHooksKey struct{}

// WithHooks returns a copy of parent context in which the hooks associated.
func WithHooks(ctx context.Context, hs ...*HooksContext) context.Context {
	switch len(hs) {
	case 0:
		return context.WithValue(ctx, contextHooksKey{}, (*HooksContext)(nil))
	case 1:
		return context.WithValue(ctx, contextHooksKey{}, hs[0])
	}

	hooksSlice := make([]hooks, len(hs))
	for i, hk := range hs {
		hooksSlice[i] = hk
	}
	return context.WithValue(ctx, contextHooksKey{}, multipleHooks(hooksSlice))
}
