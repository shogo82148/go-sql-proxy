// +build go1.10

package proxy_test

import (
	"bufio"
	"bytes"
	"database/sql"
	"log"
	"regexp"
	"testing"

	proxy "github.com/shogo82148/go-sql-proxy"
)

func TestTraceProxy(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(log.Lshortfile)
	proxy.RegisterTracer()

	db, err := sql.Open("fakedb:trace", `{"name":"trace"}`)
	if err != nil {
		t.Fatalf("Open filed: %v", err)
	}

	_, err = db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	// test for transactions
	err = func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if _, err := tx.Exec("INSERT INTO t1 (id) VALUES(?)", 1); err != nil {
			return err
		}
		rows, err := tx.Query("SELECT id FROM t1 WHERE id = ?", 1)
		if err != nil {
			return err
		}
		for rows.Next() {
		}
		rows.Close()
		return tx.Commit()
	}()
	if err != nil {
		t.Fatalf("do failed: %v", err)
	}

	// test for rollback
	err = func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		return tx.Rollback()
	}()
	if err != nil {
		t.Fatalf("do failed: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	timeComponent := `\(\d+(?:\.\d+)?[^\)]+\)`
	expected := []*regexp.Regexp{
		// Fake time component with (\d+\.\d+[^\)]+)
		regexp.MustCompile(`tracer_test.go:27: Open 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:27: Exec 0x[0-9a-f]+: CREATE TABLE t1 \(id INTEGER PRIMARY KEY\); args = \[\] ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:34: Begin 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:39: Exec 0x[0-9a-f]+: INSERT INTO t1 \(id\) VALUES\(\?\); args = \[1\] ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:42: Query 0x[0-9a-f]+: SELECT id FROM t1 WHERE id = \?; args = \[1\] ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:49: Commit 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:57: Begin 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_test.go:61: Rollback 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`.*:\d+: Close 0x[0-9a-f]+ ` + timeComponent),
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
