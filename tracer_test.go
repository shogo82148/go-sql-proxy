package proxy_test

import (
	"bufio"
	"bytes"
	"database/sql"
	"log"
	"regexp"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/shogo82148/go-sql-proxy"
	"github.com/shogo82148/txmanager"
)

var illegalSQLError = `tracer_test.go:53: Exec: ILLEGAL SQL; args = \[\] `

func TestTraceProxy(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", log.Lshortfile)
	sql.Register("sqlite3-trace-proxy", proxy.NewTraceProxy(&sqlite3.SQLiteDriver{}, logger))

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

	_, err = dbm.Exec("ILLEGAL SQL")
	if err == nil {
		t.Error("got no error, want error")
	}

	timeComponent := `\(\d+(?:\.\d+)?[^\)]+\)`
	expected := []*regexp.Regexp{
		// Fake time compinent with (\d+\.\d+[^\)]+)
		regexp.MustCompile(`tracer_test.go:\d+: Open ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:\d+: Exec: CREATE TABLE t1 \(id INTEGER PRIMARY KEY\); args = \[\] ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:\d+: Begin ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:\d+: Exec: INSERT INTO t1 \(id\) VALUES\(\?\); args = \[1\] ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:\d+: Commit ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:\d+: Query: SELECT id FROM t1 WHERE id = \?; args = \[1\] ` + timeComponent),
		regexp.MustCompile(illegalSQLError + timeComponent),
	}

	scanner := bufio.NewScanner(buf)
	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		t.Log(line)
		if i >= len(expected) {
			t.Errorf("Got more lines than expected (%s)", line)
			break
		}

		if !expected[i].MatchString(line) {
			t.Errorf("\ngot: %s\nwant: %s", line, expected[i])
		}
		i++
	}
}
