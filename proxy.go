// +build go1.8
// a proxy package is a proxy driver for dabase/sql.

package proxy

import (
	"context"
	"database/sql/driver"
)

// Proxy is a sql driver.
// It adds hook points to other sql drivers.
type Proxy struct {
	Driver driver.Driver
	Hooks  hooks
}

// hooks is callback functions for the proxy.
// it is private because it doesn't guarantee backward compatibility.
type hooks interface {
	prePing(c context.Context, conn *Conn) (interface{}, error)
	ping(c context.Context, ctx interface{}, conn *Conn) error
	postPing(c context.Context, ctx interface{}, conn *Conn, err error) error
	preOpen(c context.Context, name string) (interface{}, error)
	open(c context.Context, ctx interface{}, conn driver.Conn) error
	postOpen(c context.Context, ctx interface{}, conn driver.Conn, err error) error
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
	// callback is returned by the `Cpnn.Ping` method.
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
	Open func(c context.Context, ctx interface{}, conn driver.Conn) error

	// PostOpen is a callback that gets called at the end of
	// the call to `db.Open(). It is ALWAYS called.
	//
	// The `ctx` parameter is the return value supplied from the
	// `Hooks.PreOpen` method, and may be nil.
	PostOpen func(c context.Context, ctx interface{}, conn driver.Conn, err error) error

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

func (h *HooksContext) open(c context.Context, ctx interface{}, conn driver.Conn) error {
	if h == nil || h.Open == nil {
		return nil
	}
	return h.Open(c, ctx, conn)
}

func (h *HooksContext) postOpen(c context.Context, ctx interface{}, conn driver.Conn, err error) error {
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

// Hooks is callback functions for the proxy.
// Deprecated: You should use HooksContext instead.
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

func namedValuesToValues(args []driver.NamedValue) []driver.Value {
	ret := make([]driver.Value, len(args))
	for _, arg := range args {
		ret[arg.Ordinal-1] = arg.Value
	}
	return ret
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
	return nil, nil
}

func (h *Hooks) ping(c context.Context, ctx interface{}, conn *Conn) error {
	return nil
}

func (h *Hooks) postPing(c context.Context, ctx interface{}, conn *Conn, err error) error {
	return nil
}

func (h *Hooks) preOpen(c context.Context, name string) (interface{}, error) {
	if h == nil || h.PreOpen == nil {
		return nil, nil
	}
	return h.PreOpen(name)
}

func (h *Hooks) open(c context.Context, ctx interface{}, conn driver.Conn) error {
	if h == nil || h.Open == nil {
		return nil
	}
	return h.Open(ctx, conn)
}

func (h *Hooks) postOpen(c context.Context, ctx interface{}, conn driver.Conn, err error) error {
	if h == nil || h.PostOpen == nil {
		return nil
	}
	return h.PostOpen(ctx, conn)
}

func (h *Hooks) preExec(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	if h == nil || h.PreExec == nil {
		return nil, nil
	}
	return h.PreExec(stmt, namedValuesToValues(args))
}

func (h *Hooks) exec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result) error {
	if h == nil || h.Exec == nil {
		return nil
	}
	return h.Exec(ctx, stmt, namedValuesToValues(args), result)
}

func (h *Hooks) postExec(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result, err error) error {
	if h == nil || h.PostExec == nil {
		return nil
	}
	return h.PostExec(ctx, stmt, namedValuesToValues(args), result)
}

func (h *Hooks) preQuery(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
	if h == nil || h.PreQuery == nil {
		return nil, nil
	}
	return h.PreQuery(stmt, namedValuesToValues(args))
}

func (h *Hooks) query(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows) error {
	if h == nil || h.Query == nil {
		return nil
	}
	return h.Query(ctx, stmt, namedValuesToValues(args), rows)
}

func (h *Hooks) postQuery(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows, err error) error {
	if h == nil || h.PostQuery == nil {
		return nil
	}
	return h.PostQuery(ctx, stmt, namedValuesToValues(args), rows)
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

// NewProxy creates new Proxy driver.
func NewProxy(driver driver.Driver, hooks *Hooks) *Proxy {
	return &Proxy{
		Driver: driver,
		Hooks:  hooks,
	}
}

func NewProxyContext(driver driver.Driver, hooks *HooksContext) *Proxy {
	return &Proxy{
		Driver: driver,
		Hooks:  hooks,
	}
}

// Open creates new connection which is wrapped by Conn.
// It will triggers PreOpen, Open, PostOpen hooks.
func (p *Proxy) Open(name string) (driver.Conn, error) {
	c := context.Background()
	var err error
	var ctx interface{}

	var conn driver.Conn

	// Setup PostOpen. This needs to be a closure like this
	// or otherwise changes to the `ctx` and `conn` parameters
	// within this Open() method does not get applied at the
	// time defer is fired
	defer func() { p.Hooks.postOpen(c, ctx, conn, err) }()

	if ctx, err = p.Hooks.preOpen(c, name); err != nil {
		return nil, err
	}
	conn, err = p.Driver.Open(name)
	if err != nil {
		return nil, err
	}

	conn = &Conn{
		Conn:  conn,
		Proxy: p,
	}

	if err = p.Hooks.open(c, ctx, conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
