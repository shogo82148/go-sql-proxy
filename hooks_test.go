package proxy

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"
	"time"
)

func testHooksInterface(t *testing.T, h hooks, ctx interface{}) {
	c := context.Background()
	if ctx2, err := h.preOpen(c, ""); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preOpen returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.open(c, ctx, nil); err != nil {
		t.Error("open returns error: ", err)
	}
	if err := h.postOpen(c, ctx, nil, nil); err != nil {
		t.Error("postOpen returns error: ", err)
	}
	if ctx2, err := h.prePing(c, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("prePing returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.ping(c, ctx, nil); err != nil {
		t.Error("ping returns error: ", err)
	}
	if err := h.postPing(c, ctx, nil, nil); err != nil {
		t.Error("postPing returns error: ", err)
	}
	if ctx2, err := h.preExec(c, nil, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preExec returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.exec(c, ctx, nil, nil, nil); err != nil {
		t.Error("exec returns error: ", err)
	}
	if err := h.postExec(c, ctx, nil, nil, nil, nil); err != nil {
		t.Error("postExec returns error: ", err)
	}
	if ctx2, err := h.preQuery(c, nil, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preQuery returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.query(c, ctx, nil, nil, nil); err != nil {
		t.Error("query returns error: ", err)
	}
	if err := h.postQuery(c, ctx, nil, nil, nil, nil); err != nil {
		t.Error("postQuery returns error: ", err)
	}
	if ctx2, err := h.preBegin(c, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preBegin returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.begin(c, ctx, nil); err != nil {
		t.Error("begin returns error: ", err)
	}
	if err := h.postBegin(c, ctx, nil, nil); err != nil {
		t.Error("postBegin returns error: ", err)
	}
	if ctx2, err := h.preCommit(c, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preCommit returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.commit(c, ctx, nil); err != nil {
		t.Error("commit returns error: ", err)
	}
	if err := h.postCommit(c, ctx, nil, nil); err != nil {
		t.Error("postCommit returns error: ", err)
	}
	if ctx2, err := h.preRollback(c, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preRollback returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.rollback(c, ctx, nil); err != nil {
		t.Error("rollback returns error: ", err)
	}
	if err := h.postRollback(c, ctx, nil, nil); err != nil {
		t.Error("postRollback returns error: ", err)
	}
	if ctx2, err := h.preClose(c, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preClose returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.close(c, ctx, nil); err != nil {
		t.Error("close returns error: ", err)
	}
	if err := h.postClose(c, ctx, nil, nil); err != nil {
		t.Error("postClose returns error: ", err)
	}
	if ctx2, err := h.preResetSession(c, nil); !reflect.DeepEqual(ctx2, ctx) || err != nil {
		t.Errorf("preResetSession returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.resetSession(c, ctx, nil); err != nil {
		t.Error("resetSession returns error: ", err)
	}
	if err := h.postResetSession(c, ctx, nil, nil); err != nil {
		t.Error("postResetSession returns error: ", err)
	}
}

func TestNilHooksContext(t *testing.T) {
	// nil HooksContext will not panic and have no effect
	testHooksInterface(t, (*HooksContext)(nil), nil)
}

func TestZeroHooksContext(t *testing.T) {
	// zero HooksContext will not panic and have no effect
	testHooksInterface(t, &HooksContext{}, nil)
}

func newTestHooksContext(t *testing.T) (*HooksContext, interface{}) {
	dummy := time.Now().UnixNano()
	ctx0 := &dummy
	checkCtx := func(name string, ctx interface{}) {
		if ctx != ctx0 {
			t.Errorf("unexpected ctx: got %v want %v in %s", ctx, ctx0, name)
		}
	}
	return &HooksContext{
		PrePing: func(c context.Context, conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		Ping: func(c context.Context, ctx interface{}, conn *Conn) error {
			checkCtx("Ping", ctx)
			return nil
		},
		PostPing: func(c context.Context, ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostPing", ctx)
			return err
		},
		PreOpen: func(c context.Context, name string) (interface{}, error) {
			return ctx0, nil
		},
		Open: func(c context.Context, ctx interface{}, conn *Conn) error {
			checkCtx("Open", ctx)
			return nil
		},
		PostOpen: func(c context.Context, ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostOpen", ctx)
			return err
		},
		PreExec: func(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
			return ctx0, nil
		},
		Exec: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result) error {
			checkCtx("Exec", ctx)
			return nil
		},
		PostExec: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, result driver.Result, err error) error {
			checkCtx("PostExec", ctx)
			return err
		},
		PreQuery: func(c context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
			return ctx0, nil
		},
		Query: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows) error {
			checkCtx("Query", ctx)
			return nil
		},
		PostQuery: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, rows driver.Rows, err error) error {
			checkCtx("PostQuery", ctx)
			return err
		},
		PreBegin: func(c context.Context, conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		Begin: func(c context.Context, ctx interface{}, conn *Conn) error {
			checkCtx("Begin", ctx)
			return nil
		},
		PostBegin: func(c context.Context, ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostBegin", ctx)
			return err
		},
		PreCommit: func(c context.Context, tx *Tx) (interface{}, error) {
			return ctx0, nil
		},
		Commit: func(c context.Context, ctx interface{}, tx *Tx) error {
			checkCtx("Commit", ctx)
			return nil
		},
		PostCommit: func(c context.Context, ctx interface{}, tx *Tx, err error) error {
			checkCtx("PostCommit", ctx)
			return err
		},
		PreRollback: func(c context.Context, tx *Tx) (interface{}, error) {
			return ctx0, nil
		},
		Rollback: func(c context.Context, ctx interface{}, tx *Tx) error {
			checkCtx("Rollback", ctx)
			return nil
		},
		PostRollback: func(c context.Context, ctx interface{}, tx *Tx, err error) error {
			checkCtx("PostRollback", ctx)
			return err
		},
		PreClose: func(c context.Context, conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		Close: func(c context.Context, ctx interface{}, conn *Conn) error {
			checkCtx("Close", ctx)
			return nil
		},
		PostClose: func(c context.Context, ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostClose", ctx)
			return err
		},
		PreResetSession: func(c context.Context, conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		ResetSession: func(c context.Context, ctx interface{}, conn *Conn) error {
			checkCtx("ResetSession", ctx)
			return nil
		},
		PostResetSession: func(c context.Context, ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostResetSession", ctx)
			return err
		},
	}, ctx0
}

func TestHooksContext(t *testing.T) {
	hooks, ctx0 := newTestHooksContext(t)
	testHooksInterface(t, hooks, ctx0)
}

func TestNilHooks(t *testing.T) {
	// nil Hooks will not panic and have no effect
	testHooksInterface(t, (*Hooks)(nil), nil)
}

func TestZeroHooks(t *testing.T) {
	// zero Hooks will not panic and have no effect
	testHooksInterface(t, &Hooks{}, nil)
}

func newTestHooks(t *testing.T) (*Hooks, interface{}) {
	dummy := time.Now().UnixNano()
	ctx0 := &dummy
	checkCtx := func(name string, ctx interface{}) {
		if ctx != ctx0 {
			t.Errorf("unexpected ctx: got %v want %v in %s", ctx, ctx0, name)
		}
	}
	return &Hooks{
		PrePing: func(conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		Ping: func(ctx interface{}, conn *Conn) error {
			checkCtx("Ping", ctx)
			return nil
		},
		PostPing: func(ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostPing", ctx)
			return err
		},
		PreOpen: func(name string) (interface{}, error) {
			return ctx0, nil
		},
		Open: func(ctx interface{}, conn *Conn) error {
			checkCtx("Open", ctx)
			return nil
		},
		PostOpen: func(ctx interface{}, conn *Conn) error {
			checkCtx("PostOpen", ctx)
			return nil
		},
		PreExec: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
			return ctx0, nil
		},
		Exec: func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
			checkCtx("Exec", ctx)
			return nil
		},
		PostExec: func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
			checkCtx("PostExec", ctx)
			return nil
		},
		PreQuery: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
			return ctx0, nil
		},
		Query: func(ctx interface{}, stmt *Stmt, args []driver.Value, rows driver.Rows) error {
			checkCtx("Query", ctx)
			return nil
		},
		PostQuery: func(ctx interface{}, stmt *Stmt, args []driver.Value, rows driver.Rows) error {
			checkCtx("PostQuery", ctx)
			return nil
		},
		PreBegin: func(conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		Begin: func(ctx interface{}, conn *Conn) error {
			checkCtx("Begin", ctx)
			return nil
		},
		PostBegin: func(ctx interface{}, conn *Conn) error {
			checkCtx("PostBegin", ctx)
			return nil
		},
		PreCommit: func(tx *Tx) (interface{}, error) {
			return ctx0, nil
		},
		Commit: func(ctx interface{}, tx *Tx) error {
			checkCtx("Commit", ctx)
			return nil
		},
		PostCommit: func(ctx interface{}, tx *Tx) error {
			checkCtx("PostCommit", ctx)
			return nil
		},
		PreRollback: func(tx *Tx) (interface{}, error) {
			return ctx0, nil
		},
		Rollback: func(ctx interface{}, tx *Tx) error {
			checkCtx("Rollback", ctx)
			return nil
		},
		PostRollback: func(ctx interface{}, tx *Tx) error {
			checkCtx("PostRollback", ctx)
			return nil
		},
		PreClose: func(conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		Close: func(ctx interface{}, conn *Conn) error {
			checkCtx("Close", ctx)
			return nil
		},
		PostClose: func(ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostClose", ctx)
			return err
		},
		PreResetSession: func(conn *Conn) (interface{}, error) {
			return ctx0, nil
		},
		ResetSession: func(ctx interface{}, conn *Conn) error {
			checkCtx("ResetSession", ctx)
			return nil
		},
		PostResetSession: func(ctx interface{}, conn *Conn, err error) error {
			checkCtx("PostResetSession", ctx)
			return err
		},
	}, ctx0
}

func TestHooks(t *testing.T) {
	hooks, ctx0 := newTestHooks(t)
	testHooksInterface(t, hooks, ctx0)
}

func TestNilMultipleHooks(t *testing.T) {
	// nil HooksContext will not panic and have no effect
	testHooksInterface(t, multipleHooks(nil), nil)
}

func TestZeroMultipleHooks(t *testing.T) {
	// zero HooksContext will not panic and have no effect
	testHooksInterface(t, multipleHooks{}, nil)
}

func TestMultipleHooks(t *testing.T) {
	hooks1, ctx1 := newTestHooksContext(t)
	hooks2, ctx2 := newTestHooks(t)
	hooks := multipleHooks{hooks1, hooks2}
	ctx0 := []interface{}{ctx1, ctx2}
	testHooksInterface(t, hooks, ctx0)
}

func TestWithHooks(t *testing.T) {
	ctx := WithHooks(context.Background(), &HooksContext{}, &HooksContext{})
	hooks := contextHooks(ctx)
	conn := &Conn{}
	myctx, err := hooks.prePing(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	if err := hooks.ping(ctx, myctx, conn); err != nil {
		t.Fatal(err)
	}
	if err := hooks.postPing(ctx, myctx, conn, nil); err != nil {
		t.Fatal(err)
	}
}

func TestWithHooks_error(t *testing.T) {
	var count int
	var errPrePing = errors.New("pre-ping error")
	var errPing = errors.New("ping error")
	var errPostPing = errors.New("post-ping error")
	ctx := WithHooks(context.Background(), &HooksContext{
		PrePing: func(c context.Context, conn *Conn) (interface{}, error) {
			count++
			if count != 1 {
				t.Errorf("want count is %d, got %d", 1, count)
			}
			return nil, errPrePing
		},
		Ping: func(c context.Context, ctx interface{}, conn *Conn) error {
			count++
			if count != 3 {
				t.Errorf("want count is %d, got %d", 3, count)
			}
			return errPing
		},
		PostPing: func(c context.Context, ctx interface{}, conn *Conn, err error) error {
			count++
			if count != 5 {
				t.Errorf("want count is %d, got %d", 5, count)
			}
			return errors.New("some error")
		},
	}, &HooksContext{
		PrePing: func(c context.Context, conn *Conn) (interface{}, error) {
			count++
			if count != 2 {
				t.Errorf("want count is %d, got %d", 2, count)
			}
			return nil, errors.New("some error")
		},
		Ping: func(c context.Context, ctx interface{}, conn *Conn) error {
			// never reach here because previous ping hook fails.
			panic("never reach")
		},
		PostPing: func(c context.Context, ctx interface{}, conn *Conn, err error) error {
			count++

			// PostPing functions are called by reverse order.
			if count != 4 {
				t.Errorf("want count is %d, got %d", 4, count)
			}
			return errPostPing
		},
	})

	hooks := contextHooks(ctx)
	conn := &Conn{}
	myctx, err := hooks.prePing(ctx, conn)
	if err != errPrePing {
		t.Errorf("want %v, got %v", errPrePing, err)
	}
	err = hooks.ping(ctx, myctx, conn)
	if err != errPing {
		t.Errorf("want %v, got %v", errPing, err)
	}
	err = hooks.postPing(ctx, myctx, conn, err)
	if err != errPostPing {
		t.Errorf("want %v, got %v", errPostPing, err)
	}
	if count != 5 {
		t.Errorf("want %d, got %d", 5, count)
	}
}
