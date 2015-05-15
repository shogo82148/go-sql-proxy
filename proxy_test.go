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

type steps struct {
	recorded []string
}

func newSteps() *steps {
	return &steps{recorded: []string{}}
}

func (s *steps) record(in string) {
	s.recorded = append(s.recorded, in)
}

func (s *steps) verify(expected []string) error {
	if len(expected) != len(s.recorded) {
		return fmt.Errorf("Number of steps do not match (%d != %d)", len(expected), len(s.recorded))
	}

	for i := range expected {
		if s.recorded[i] != expected[i] {
			return fmt.Errorf("Expected '%s', got '%s' for step %d",
				expected[i],
				s.recorded[i],
				i,
			)
		}
	}

	return nil
}

func testOpen(t *testing.T, s *steps, driverName string, hooks *Hooks, expected []string) {
	sql.Register(driverName, NewProxy(&sqlite3.SQLiteDriver{}, hooks))

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatalf("%s: Open failed: %v", driverName, err)
	}
	// The sql driver's Open() is not called until some sort of
	// statement is executed.
	db.Query("SELECT 1")
	if err := s.verify(expected); err != nil {
		t.Errorf("%s: %s", driverName, err)
		return
	}
}

func TestOpenAll(t *testing.T) {
	// driverNames must be unique for every call to sql.Register()
	var elapsed time.Duration

	s := newSteps()
	testOpen(
		t,
		s,
		"sqlite3-proxy-test-open",
		&Hooks{
			PreOpen: func(name string) (interface{}, error) {
				s.record("PreOpen")
				return time.Now(), nil
			},
			Open: func(_ interface{}, _ driver.Conn) error {
				s.record("Open")
				return nil
			},
			PostOpen: func(ctx interface{}, _ driver.Conn) error {
				elapsed = time.Since(ctx.(time.Time))
				s.record("PostOpen")
				return nil
			},
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
	s := newSteps()
	testOpen(
		t,
		s,
		"sqlite3-proxy-test-no-preopen",
		&Hooks{
			Open: func(_ interface{}, _ driver.Conn) error {
				s.record("Open")
				return nil
			},
			PostOpen: func(_ interface{}, _ driver.Conn) error {
				s.record("PostOpen")
				return nil
			},
		},
		[]string{
			"Open",
			"PostOpen",
		},
	)
}

func TestNoPostOpen(t *testing.T) {
	s := newSteps()
	testOpen(
		t,
		s,
		"sqlite3-proxy-test-no-postopen",
		&Hooks{
			PreOpen: func(_ string) (interface{}, error) {
				s.record("PreOpen")
				return nil, nil
			},
			Open: func(_ interface{}, _ driver.Conn) error {
				s.record("Open")
				return nil
			},
		},
		[]string{
			"PreOpen",
			"Open",
		},
	)
}

func testExec(t *testing.T, s *steps, driverName string, hooks *Hooks, expected []string) {
	sql.Register(driverName, NewProxy(&sqlite3.SQLiteDriver{}, hooks))

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

	if err := s.verify(expected); err != nil {
		t.Errorf("%s: %s", driverName, err)
		return
	}
}

func TestExecAll(t *testing.T) {
	s := newSteps()
	var elapsed time.Duration

	testExec(
		t,
		s,
		"sqlite3-proxy-test-exec",
		&Hooks{
			PreExec: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
				s.record("PreExec")
				return time.Now(), nil
			},
			Exec: func(_ interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
				s.record("Exec")
				return nil
			},
			PostExec: func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
				s.record("PostExec")
				elapsed = time.Since(ctx.(time.Time))
				return nil
			},
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
	s := newSteps()

	testExec(
		t,
		s,
		"sqlite3-proxy-test-no-preexec",
		&Hooks{
			Exec: func(_ interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
				s.record("Exec")
				return nil
			},
			PostExec: func(ctx interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
				s.record("PostExec")
				return nil
			},
		},
		[]string{
			"Exec",
			"PostExec",
		},
	)
}

func TestExecNoPostExec(t *testing.T) {
	s := newSteps()

	testExec(
		t,
		s,
		"sqlite3-proxy-test-no-postexec",
		&Hooks{
			PreExec: func(stmt *Stmt, args []driver.Value) (interface{}, error) {
				s.record("PreExec")
				return nil, nil
			},
			Exec: func(_ interface{}, stmt *Stmt, args []driver.Value, result driver.Result) error {
				s.record("Exec")
				return nil
			},
		},
		[]string{
			"PreExec",
			"Exec",
		},
	)
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
		Query: func(stmt *Stmt, args []driver.Value, rows driver.Rows) error {
			t.Logf("Query: %s; args = %v", stmt.QueryString, args)
			statements = append(statements, fmt.Sprintf("Query: %s; args = %v", stmt.QueryString, args))
			return nil
		},
		Begin: func(conn *Conn) error {
			t.Log("Begin")
			statements = append(statements, "Begin")
			return nil
		},
		Commit: func(tx *Tx) error {
			t.Log("Commit")
			statements = append(statements, "Commit")
			return nil
		},
		Rollback: func(tx *Tx) error {
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
