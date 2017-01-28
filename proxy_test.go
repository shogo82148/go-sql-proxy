// +build go1.8

package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
)

func testHooksInterface(t *testing.T, h hooks, ctx interface{}) {
	c := context.Background()
	if ctx2, err := h.preOpen(c, ""); ctx2 != ctx || err != nil {
		t.Errorf("preOpen returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.open(c, ctx, nil); err != nil {
		t.Error("open returns error: ", err)
	}
	if err := h.postOpen(c, ctx, nil, nil); err != nil {
		t.Error("postOpen returns error: ", err)
	}
	if ctx2, err := h.preExec(c, nil, nil); ctx2 != ctx || err != nil {
		t.Errorf("preExec returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.exec(c, ctx, nil, nil, nil); err != nil {
		t.Error("exec returns error: ", err)
	}
	if err := h.postExec(c, ctx, nil, nil, nil, nil); err != nil {
		t.Error("postExec returns error: ", err)
	}
	if ctx2, err := h.preQuery(c, nil, nil); ctx2 != ctx || err != nil {
		t.Errorf("preQuery returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.query(c, ctx, nil, nil, nil); err != nil {
		t.Error("query returns error: ", err)
	}
	if err := h.postQuery(c, ctx, nil, nil, nil, nil); err != nil {
		t.Error("postQuery returns error: ", err)
	}
	if ctx2, err := h.preBegin(c, nil); ctx2 != ctx || err != nil {
		t.Errorf("preBegin returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.begin(c, ctx, nil); err != nil {
		t.Error("begin returns error: ", err)
	}
	if err := h.postBegin(c, ctx, nil, nil); err != nil {
		t.Error("postBegin returns error: ", err)
	}
	if ctx2, err := h.preCommit(c, nil); ctx2 != ctx || err != nil {
		t.Errorf("preCommit returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.commit(c, ctx, nil); err != nil {
		t.Error("commit returns error: ", err)
	}
	if err := h.postCommit(c, ctx, nil, nil); err != nil {
		t.Error("postCommit returns error: ", err)
	}
	if ctx2, err := h.preRollback(c, nil); ctx2 != ctx || err != nil {
		t.Errorf("preRollback returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.rollback(c, ctx, nil); err != nil {
		t.Error("rollback returns error: ", err)
	}
	if err := h.postRollback(c, ctx, nil, nil); err != nil {
		t.Error("postRollback returns error: ", err)
	}
}

func TestNilHooksContext(t *testing.T) {
	// nil HooksContext will not panic and have no effec
	testHooksInterface(t, (*HooksContext)(nil), nil)
}

func TestZeroHooksContext(t *testing.T) {
	// zero HooksContext will not panic and have no effec
	testHooksInterface(t, &HooksContext{}, nil)
}

func TestHooksContext(t *testing.T) {
	dummy := 0
	ctx0 := &dummy
	checkCtx := func(name string, ctx interface{}) {
		if ctx != ctx0 {
			t.Errorf("unexpected ctx: got %v want %v in %s", ctx, ctx0, name)
		}
	}
	testHooksInterface(t, &HooksContext{
		PreOpen: func(c context.Context, name string) (interface{}, error) {
			return ctx0, nil
		},
		Open: func(c context.Context, ctx interface{}, conn driver.Conn) error {
			checkCtx("Open", ctx)
			return nil
		},
		PostOpen: func(c context.Context, ctx interface{}, conn driver.Conn, err error) error {
			checkCtx("PostOpen", ctx)
			return err
		},
		PreExec: func(c context.Context, stmt *Stmt, args []driver.Value) (interface{}, error) {
			return ctx0, nil
		},
		Exec: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
			checkCtx("Exec", ctx)
			return nil
		},
		PostExec: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result, err error) error {
			checkCtx("PostExec", ctx)
			return err
		},
		PreQuery: func(c context.Context, stmt *Stmt, args []driver.Value) (interface{}, error) {
			return ctx0, nil
		},
		Query: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.Value, rows driver.Rows) error {
			checkCtx("Query", ctx)
			return nil
		},
		PostQuery: func(c context.Context, ctx interface{}, stmt *Stmt, args []driver.Value, rows driver.Rows, err error) error {
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
	}, ctx0)
}

func TestNilHooks(t *testing.T) {
	// nil Hooks will not panic and have no effect
	testHooksInterface(t, (*Hooks)(nil), nil)
}

func TestZeroHooks(t *testing.T) {
	// zero Hooks will not panic and have no effect
	testHooksInterface(t, &Hooks{}, nil)
}

func TestHooks(t *testing.T) {
	dummy := 0
	ctx0 := &dummy
	checkCtx := func(name string, ctx interface{}) {
		if ctx != ctx0 {
			t.Errorf("unexpected ctx: got %v want %v in %s", ctx, ctx0, name)
		}
	}
	testHooksInterface(t, &Hooks{
		PreOpen: func(name string) (interface{}, error) {
			return ctx0, nil
		},
		Open: func(ctx interface{}, conn driver.Conn) error {
			checkCtx("Open", ctx)
			return nil
		},
		PostOpen: func(ctx interface{}, conn driver.Conn) error {
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
	}, ctx0)
}

func TestFakeDB(t *testing.T) {
	testName := t.Name()
	testCases := []struct {
		name     string
		hooksLog string
		f        func(db *sql.DB) error
	}{
		{
			name: "prepare",
			hooksLog: "[PreOpen] " + testName + "-proxy-prepare\n" +
				"[Open]\n[PostOpen]\n",
			f: func(db *sql.DB) error {
				db.Prepare("HOGE")
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// install a proxy
			buf := &bytes.Buffer{}
			driverName := fmt.Sprintf("%s-proxy-%s", testName, tc.name)
			sql.Register(driverName, &Proxy{
				Driver: fdriver,
				Hooks:  newLoggingHook(buf),
			})

			// Run test queries directory
			dbName := fmt.Sprintf("%s-%s", testName, tc.name)
			db, err := sql.Open("fakedb", dbName)
			if err != nil {
				t.Fatal(err)
			}
			if err = tc.f(db); err != nil {
				t.Fatal(err)
			}

			// Run test queris via a proxy
			dbProxyName := fmt.Sprintf("%s-proxy-%s", testName, tc.name)
			dbProxy, err := sql.Open(driverName, dbProxyName)
			if err != nil {
				t.Fatal(err)
			}
			if err = tc.f(dbProxy); err != nil {
				t.Fatal(err)
			}

			// check the logs
			want := fdriver.DB(dbName).LogToString()
			got := fdriver.DB(dbProxyName).LogToString()
			if want != got {
				t.Errorf("want %s, got %s", want, got)
			}
			if tc.hooksLog != buf.String() {
				t.Errorf("want %s, got %s", tc.hooksLog, buf.String())
			}
		})
	}
}
