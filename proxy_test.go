package proxy

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/shogo82148/txmanager"
)

func TestProxy(t *testing.T) {
	statements := []string{}
	sql.Register("sqlite3-proxy", NewProxy(&sqlite3.SQLiteDriver{}, &Hooks{
		Open: func(conn *Conn) error {
			t.Log("Open")
			statements = append(statements, "Open")
			return nil
		},
		Exec: func(stmt *Stmt, args []driver.Value, result driver.Result) error {
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
		t.Fatalf("Open filed: %v", err)
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

func TestTraceProxy(t *testing.T) {
	statements := []string{}
	sql.Register("sqlite3-trace-proxy", NewTraceProxy(&sqlite3.SQLiteDriver{}, log.New(os.Stderr, "", log.Lshortfile)))

	db, err := sql.Open("sqlite3-trace-proxy", ":memory:")
	if err != nil {
		t.Fatalf("Open filed: %v", err)
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
