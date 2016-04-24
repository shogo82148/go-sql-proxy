package proxy

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/shogo82148/txmanager"
)

type proxyTest struct {
	*testing.T
	recorded []string
	seq      chan int
}

var tseq chan int

func init() {
	tseq = make(chan int)
	go func() {
		for i := 1; ; i++ {
			tseq <- i
		}
	}()
}

func newProxyTest(t *testing.T) *proxyTest {
	return &proxyTest{T: t, seq: tseq}
}

func (t *proxyTest) record(in string) {
	t.recorded = append(t.recorded, in)
}

func (t proxyTest) verify(expected []string) error {
	if len(expected) != len(t.recorded) {
		return fmt.Errorf("Number of steps do not match (%d != %d)", len(expected), len(t.recorded))
	}

	for i := range expected {
		if t.recorded[i] != expected[i] {
			return fmt.Errorf("Expected '%s', got '%s' for step %d",
				expected[i],
				t.recorded[i],
				i,
			)
		}
	}

	return nil
}

func (t *proxyTest) testOpen(makeHooks func(*proxyTest) *Hooks, expected []string) {
	driverName := fmt.Sprintf("sqlite-proxy-test-open-%d", <-t.seq)
	t.recorded = []string(nil)
	h := makeHooks(t)

	sql.Register(driverName, NewProxy(&sqlite3.SQLiteDriver{}, h))

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatalf("%s: Open failed: %v", driverName, err)
	}
	// The sql driver's Open() is not called until some sort of
	// statement is executed.
	db.Query("SELECT 1")
	if err := t.verify(expected); err != nil {
		t.Errorf("%s: %s", driverName, err)
		return
	}
}

func (t *proxyTest) testExec(makeHooks func(*proxyTest) *Hooks, expected []string) {
	driverName := fmt.Sprintf("sqlite-proxy-test-exec-%d", <-t.seq)
	t.recorded = []string(nil)
	h := makeHooks(t)

	sql.Register(driverName, NewProxy(&sqlite3.SQLiteDriver{}, h))

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	_, err = db.Exec(
		"CREATE TABLE t1 (id INTEGER PRIMARY KEY)",
	)
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	if err := t.verify(expected); err != nil {
		t.Errorf("%s: %s", driverName, err)
		return
	}
}

func (t *proxyTest) testQuery(makeHooks func(*proxyTest) *Hooks, expected []string) {
	driverName := fmt.Sprintf("sqlite-proxy-test-query-%d", <-t.seq)
	t.recorded = []string(nil)
	h := makeHooks(t)

	sql.Register(driverName, NewProxy(&sqlite3.SQLiteDriver{}, h))

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	_, err = db.Query("SELECT 1")
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	if err := t.verify(expected); err != nil {
		t.Errorf("%s: %s", driverName, err)
		return
	}
}

func TestOpenAll(t *testing.T) {
	var elapsed time.Duration
	newProxyTest(t).testOpen(
		func(p *proxyTest) *Hooks {
			return &Hooks{
				PreOpen: func(name string) (interface{}, error) {
					p.record("PreOpen")
					return time.Now(), nil
				},
				Open: func(_ interface{}, _ driver.Conn) error {
					p.record("Open")
					return nil
				},
				PostOpen: func(ctx interface{}, _ driver.Conn) error {
					elapsed = time.Since(ctx.(time.Time))
					p.record("PostOpen")
					return nil
				},
			}
		},
		[]string{
			"PreOpen",
			"Open",
			"PostOpen",
		},
	)

	if elapsed == 0 {
		t.Errorf("'elapsed' should not be zero")
		return
	}
}

func TestNoPreOpen(t *testing.T) {
	newProxyTest(t).testOpen(
		func(p *proxyTest) *Hooks {
			return &Hooks{
				Open: func(_ interface{}, _ driver.Conn) error {
					p.record("Open")
					return nil
				},
				PostOpen: func(_ interface{}, _ driver.Conn) error {
					p.record("PostOpen")
					return nil
				},
			}
		},
		[]string{
			"Open",
			"PostOpen",
		},
	)
}

func TestNoPostOpen(t *testing.T) {
	newProxyTest(t).testOpen(
		func(p *proxyTest) *Hooks {
			return &Hooks{
				PreOpen: func(_ string) (interface{}, error) {
					p.record("PreOpen")
					return nil, nil
				},
				Open: func(_ interface{}, _ driver.Conn) error {
					p.record("Open")
					return nil
				},
			}
		},
		[]string{
			"PreOpen",
			"Open",
		},
	)
}

func TestExecAll(t *testing.T) {
	var elapsed time.Duration

	newProxyTest(t).testExec(
		func(p *proxyTest) *Hooks {
			return &Hooks{
				PreExec: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
					p.record("PreExec")
					return time.Now(), nil
				},
				Exec: func(_ interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
					p.record("Exec")
					return nil
				},
				PostExec: func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
					p.record("PostExec")
					elapsed = time.Since(ctx.(time.Time))
					return nil
				},
			}
		},
		[]string{
			"PreExec",
			"Exec",
			"PostExec",
		},
	)

	if elapsed == 0 {
		t.Errorf("'elapsed' should not be zero")
		return
	}
}

func TestExecNoPreExec(t *testing.T) {
	newProxyTest(t).testExec(
		func(p *proxyTest) *Hooks {
			return &Hooks{
				Exec: func(_ interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
					p.record("Exec")
					return nil
				},
				PostExec: func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
					p.record("PostExec")
					return nil
				},
			}
		},
		[]string{
			"Exec",
			"PostExec",
		},
	)
}

func TestExecNoPostExec(t *testing.T) {
	newProxyTest(t).testExec(
		func(p *proxyTest) *Hooks {
			return &Hooks{
				PreExec: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
					p.record("PreExec")
					return nil, nil
				},
				Exec: func(_ interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
					p.record("Exec")
					return nil
				},
			}
		},
		[]string{
			"PreExec",
			"Exec",
		},
	)
}

func TestQueryAll(t *testing.T) {
	var elapsed time.Duration

	newProxyTest(t).testQuery(
		func(p *proxyTest) *Hooks {
			return &Hooks{
				PreQuery: func(_ *Stmt, _ []driver.Value) (interface{}, error) {
					p.record("PreQuery")
					return time.Now(), nil
				},
				Query: func(_ interface{}, _ *Stmt, _ []driver.Value, _ driver.Rows) error {
					p.record("Query")
					return nil
				},
				PostQuery: func(ctx interface{}, _ *Stmt, _ []driver.Value, _ driver.Rows) error {
					p.record("PostQuery")
					elapsed = time.Since(ctx.(time.Time))
					return nil
				},
			}
		},
		[]string{
			"PreQuery",
			"Query",
			"PostQuery",
		},
	)

	if elapsed == 0 {
		t.Errorf("'elapsed' should not be zero")
		return
	}
}

func TestProxy(t *testing.T) {
	statements := []string{}
	sql.Register("sqlite3-proxy", NewProxy(&sqlite3.SQLiteDriver{}, &Hooks{
		Open: func(_ interface{}, _ driver.Conn) error {
			t.Log("Open")
			statements = append(statements, "Open")
			return nil
		},
		Exec: func(_ interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
			t.Logf("Exec: %s; args = %v", stmt.QueryString, args)
			statements = append(statements, fmt.Sprintf("Exec: %s; args = %v", stmt.QueryString, args))
			return nil
		},
		Query: func(_ interface{}, stmt *Stmt, args []driver.Value, rows driver.Rows) error {
			t.Logf("Query: %s; args = %v", stmt.QueryString, args)
			statements = append(statements, fmt.Sprintf("Query: %s; args = %v", stmt.QueryString, args))
			return nil
		},
		Begin: func(_ interface{}, _ *Conn) error {
			t.Log("Begin")
			statements = append(statements, "Begin")
			return nil
		},
		Commit: func(_ interface{}, _ *Tx) error {
			t.Log("Commit")
			statements = append(statements, "Commit")
			return nil
		},
		Rollback: func(_ interface{}, _ *Tx) error {
			t.Log("Rollback")
			statements = append(statements, "Rollback")
			return nil
		},
	}))

	db, err := sql.Open("sqlite3-proxy", ":memory:")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	_, err = db.Exec(
		"CREATE TABLE t1 (id INTEGER PRIMARY KEY)",
	)
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	dbm := txmanager.NewDB(db)
	txmanager.Do(dbm, func(tx txmanager.Tx) error {
		_, err := tx.Exec("INSERT INTO t1 (id) VALUES(?)", 1)
		return err
	})
	if err != nil {
		t.Fatalf("do failed: %v", err)
	}

	row := dbm.QueryRow("SELECT id FROM t1 WHERE id = ?", 1)
	var id int
	if err = row.Scan(&id); err != nil {
		t.Fatalf("selecting row failed: %v", err)
	}
	if id != 1 {
		t.Errorf("got %d\nwant 1", id)
	}

	want := []string{
		"Open",
		"Exec: CREATE TABLE t1 (id INTEGER PRIMARY KEY); args = []",
		"Begin",
		"Exec: INSERT INTO t1 (id) VALUES(?); args = [1]",
		"Commit",
		"Query: SELECT id FROM t1 WHERE id = ?; args = [1]",
	}
	if !reflect.DeepEqual(statements, want) {
		t.Errorf("got %v\nwant %v", statements, want)
	}
}
