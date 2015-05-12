package proxy

import (
	"bytes"
	"database/sql"
	"log"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/shogo82148/txmanager"
)

func TestTraceProxy(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", log.Lshortfile)
	sql.Register("sqlite3-trace-proxy", NewTraceProxy(&sqlite3.SQLiteDriver{}, logger))

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

	want := `tracer_test.go:25: Open
tracer_test.go:25: Exec: CREATE TABLE t1 (id INTEGER PRIMARY KEY); args = []
txmanager.go:177: Begin
txmanager.go:182: Exec: INSERT INTO t1 (id) VALUES(?); args = [1]
tracer_test.go:34: Commit
tracer_test.go:39: Query: SELECT id FROM t1 WHERE id = ?; args = [1]
`
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}
