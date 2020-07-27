// +build go1.10

package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

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
				"[PreExec]\n[Exec]\n[PostExec]\n" +
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
				"[PreExec]\n[Exec]\n[PostExec]\n" +
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
			hooksLog: "[PreClose]\n[Close]\n[PostClose]\n",
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
			hooksLog: "[PreClose]\n[Close]\n[PostClose]\n",
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

			// Run test queries via a proxy
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
