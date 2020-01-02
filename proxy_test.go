// +build go1.10

package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	if ctx2, err := h.prePing(c, nil); ctx2 != ctx || err != nil {
		t.Errorf("prePing returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.ping(c, ctx, nil); err != nil {
		t.Error("ping returns error: ", err)
	}
	if err := h.postPing(c, ctx, nil, nil); err != nil {
		t.Error("postPing returns error: ", err)
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
	if ctx2, err := h.preClose(c, nil); ctx2 != ctx || err != nil {
		t.Errorf("preClose returns unexpected values: got (%v, %v) want (%v, nil)", ctx2, err, ctx)
	}
	if err := h.close(c, ctx, nil); err != nil {
		t.Error("close returns error: ", err)
	}
	if err := h.postClose(c, ctx, nil, nil); err != nil {
		t.Error("postClose returns error: ", err)
	}
	if ctx2, err := h.preResetSession(c, nil); ctx2 != ctx || err != nil {
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
	}, ctx0)
}

func TestFakeDB(t *testing.T) {
	testName := t.Name()
	testCases := []struct {
		opt      *fakeConnOption
		hooksLog string
		f        func(db *sql.DB) error
	}{
		// the target driver is minimum implementation
		{
			opt: &fakeConnOption{
				Name: "pingAll",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PrePing]\n[Ping]\n[PostPing]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				return db.Ping()
			},
		},
		{
			opt: &fakeConnOption{
				Name: "execAll",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreExec]\n[Exec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", 123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "execError",
				FailExec: true,
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreExec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", 123456789)
				if err == nil {
					return errors.New("excepted error, but not")
				}
				return nil
			},
		},
		{
			opt: &fakeConnOption{
				Name: "execError-NamedValue",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreExec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				// this Exec will fail, because the driver doesn't support sql.Named()
				_, err := db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", sql.Named("foo", 123456789))
				if err == nil {
					return errors.New("expected error, but not")
				}
				return nil
			},
		},
		{
			opt: &fakeConnOption{
				Name: "queryAll",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreQuery]\n[Query]\n[PostQuery]\n",
			f: func(db *sql.DB) error {
				_, err := db.Query("SELECT * FROM test WHERE id = ?", 123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:      "queryError",
				FailQuery: true,
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreQuery]\n[PostQuery]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Query("SELECT * FROM test WHERE id = ?", 123456789)
				if err == nil {
					return errors.New("expected error, but not")
				}
				return nil
			},
		},
		{
			opt: &fakeConnOption{
				Name: "prepare",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				stmt, err := db.Prepare("SELECT * FROM test WHERE id = ?")
				if err != nil {
					return nil
				}
				return stmt.Close()
			},
		},
		{
			opt: &fakeConnOption{
				Name: "commit",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[Begin]\n[PostBegin]\n" +
				"[PreCommit]\n[Commit]\n[PostCommit]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				tx, err := db.Begin()
				if err != nil {
					return err
				}
				return tx.Commit()
			},
		},
		{
			opt: &fakeConnOption{
				Name: "rollback",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[Begin]\n[PostBegin]\n" +
				"[PreRollback]\n[Rollback]\n[PostRollback]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				tx, err := db.Begin()
				if err != nil {
					return err
				}
				return tx.Rollback()
			},
		},
		{
			opt: &fakeConnOption{
				Name: "begin-isolation",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[PostBegin]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				_, err := db.BeginTx(ctx, &sql.TxOptions{
					Isolation: sql.LevelLinearizable,
				})
				if err == nil {
					// because the driver does not support sql.LevelLinearizable
					return errors.New("expected error, but not")
				}
				return nil
			},
		},
		{
			opt: &fakeConnOption{
				Name: "begin-readonly",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[PostBegin]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				_, err := db.BeginTx(ctx, &sql.TxOptions{
					ReadOnly: true,
				})
				if err == nil {
					// because the driver does not support read-only transaction
					return errors.New("expected error, but not")
				}
				return nil
			},
		},

		// the Conn of the target driver implements Execer and Queryer
		{
			opt: &fakeConnOption{
				Name:     "execAll-ext",
				ConnType: "fakeConnExt",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreExec]\n[Exec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", 123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "execError-ext",
				ConnType: "fakeConnExt",
				FailExec: true,
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreExec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", 123456789)
				if err == nil {
					return errors.New("excepted error, but not")
				}
				return nil
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "queryAll-ext",
				ConnType: "fakeConnExt",
			},
			hooksLog: "[PreOpen]\n" +
				"[Open]\n[PostOpen]\n[PreQuery]\n[Query]\n[PostQuery]\n",
			f: func(db *sql.DB) error {
				_, err := db.Query("SELECT * FROM test WHERE id = ?", 123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:      "queryError-ext",
				ConnType:  "fakeConnExt",
				FailQuery: true,
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreQuery]\n[PostQuery]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Query("SELECT * FROM test WHERE id = ?", 123456789)
				if err == nil {
					return errors.New("expected error, but not")
				}
				return nil
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "prepare-ext",
				ConnType: "fakeConnExt",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreExec]\n[Exec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				stmt, err := db.Prepare("SELECT * FROM test WHERE id = ?")
				if err != nil {
					return err
				}
				defer stmt.Close()
				_, err = stmt.Exec(123456789)
				return err
			},
		},

		// the Conn of the target driver supports the context.
		{
			opt: &fakeConnOption{
				Name:     "pingAll-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PrePing]\n[Ping]\n[PostPing]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				return db.Ping()
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "execAll-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreExec]\n[Exec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", 123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "execAll-NamedValue-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreExec]\n[Exec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				_, err := db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)", sql.Named("foo", 123456789))
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "queryAll-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreQuery]\n[Query]\n[PostQuery]\n",
			f: func(db *sql.DB) error {
				_, err := db.Query("SELECT * FROM test WHERE id = ?", 123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "prepare-exec-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreExec]\n[Exec]\n[PostExec]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				stmt, err := db.Prepare("CREATE TABLE t1 (id INTEGER PRIMARY KEY)")
				if err != nil {
					return nil
				}
				defer stmt.Close()
				_, err = stmt.Exec(123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "prepare-query-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreQuery]\n[Query]\n[PostQuery]\n",
			f: func(db *sql.DB) error {
				stmt, err := db.Prepare("SELECT * FROM test WHERE id = ?")
				if err != nil {
					return nil
				}
				defer stmt.Close()
				rows, err := stmt.Query(123456789)
				if err != nil {
					return err
				}
				// skip close in this test, while you must close the rows in your product.
				// because the result from fakeDB is broken.
				// rows.Close()
				_ = rows
				return nil
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "commit-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[Begin]\n[PostBegin]\n" +
				"[PreCommit]\n[Commit]\n[PostCommit]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				tx, err := db.Begin()
				if err != nil {
					return err
				}
				return tx.Commit()
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "rollback-ctx",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[Begin]\n[PostBegin]\n" +
				"[PreRollback]\n[Rollback]\n[PostRollback]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				tx, err := db.Begin()
				if err != nil {
					return err
				}
				return tx.Rollback()
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "begin-ctx-isolation",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[Begin]\n[PostBegin]\n" +
				"[PreCommit]\n[Commit]\n[PostCommit]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				tx, err := db.BeginTx(ctx, &sql.TxOptions{
					Isolation: sql.LevelLinearizable,
				})
				if err != nil {
					return err
				}
				return tx.Commit()
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "begin-ctx-readonly",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreOpen]\n[Open]\n[PostOpen]\n" +
				"[PreBegin]\n[Begin]\n[PostBegin]\n" +
				"[PreCommit]\n[Commit]\n[PostCommit]\n" +
				"[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				tx, err := db.BeginTx(ctx, &sql.TxOptions{
					ReadOnly: true,
				})
				if err != nil {
					return err
				}
				return tx.Commit()
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "context-with-no-hooks",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				// remove the hooks from the current context.
				// Exec will not be logged.
				ctx := WithHooks(context.Background())
				_, err := db.ExecContext(ctx, "CREATE TABLE t1 (id INTEGER PRIMARY KEY)", 123456789)
				return err
			},
		},
		{
			opt: &fakeConnOption{
				Name:     "context-with-hooks",
				ConnType: "fakeConnCtx",
			},
			hooksLog: "[PreResetSession]\n[ResetSession]\n[PostResetSession]\n" +
				"[PreClose]\n[Close]\n[PostClose]\n",
			f: func(db *sql.DB) error {
				buf := &bytes.Buffer{}
				ctx := context.WithValue(context.Background(), contextHooksKey{}, newLoggingHook(buf))
				_, err := db.ExecContext(ctx, "CREATE TABLE t1 (id INTEGER PRIMARY KEY)", 123456789)
				if err != nil {
					return err
				}
				if _, ok := db.Driver().(*Proxy); ok {
					got := buf.String()
					want := "[PreOpen]\n[Open]\n[PostOpen]\n[PreExec]\n[Exec]\n[PostExec]\n"
					if got != want {
						return fmt.Errorf("want %s, got %s", want, got)
					}
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.opt.Name, func(t *testing.T) {
			// install a proxy
			name := tc.opt.Name
			buf := &bytes.Buffer{}
			driverName := fmt.Sprintf("%s-proxy-%s", testName, name)
			sql.Register(driverName, &Proxy{
				Driver: fdriver,
				hooks:  newLoggingHook(buf),
			})

			// Run test queries directly
			tc.opt.Name = fmt.Sprintf("%s-%s", testName, name)
			dbName, err := json.Marshal(tc.opt)
			if err != nil {
				t.Fatal(err)
			}
			db, err := sql.Open("fakedb", string(dbName))
			if err != nil {
				t.Fatal(err)
			}
			if err = tc.f(db); err != nil {
				t.Error(err)
			}
			db.Close()

			// Run test queris via a proxy
			tc.opt.Name = fmt.Sprintf("%s-proxy-%s", testName, name)
			dbProxyName, err := json.Marshal(tc.opt)
			dbProxy, err := sql.Open(driverName, string(dbProxyName))
			if err != nil {
				t.Fatal(err)
			}
			if err = tc.f(dbProxy); err != nil {
				t.Error(err)
			}
			dbProxy.Close()

			// check the logs
			want := fdriver.DB(string(dbName)).LogToString()
			got := fdriver.DB(string(dbProxyName)).LogToString()
			if want != got {
				t.Errorf("want %s, got %s", want, got)
			}
			if tc.hooksLog != buf.String() {
				t.Errorf("want %s, got %s", tc.hooksLog, buf.String())
			}
			t.Log("Driver log:", got)
			t.Log("Hook log:", buf.String())
		})
	}
}
