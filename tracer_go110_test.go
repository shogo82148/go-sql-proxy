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

var illegalSQLError = `tracer_go110_test.go:\d+: Exec 0x[0-9a-f]+: ILLEGAL SQL; args = \[\]; err = "near \\"ILLEGAL\\": syntax error" `

func TestTraceProxy(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(log.Lshortfile)
	proxy.RegisterTracer()

	db, err := sql.Open("fakedb:trace", `{"name":"trace"}`)
	if err != nil {
		t.Fatalf("Open filed: %v", err)
	}

	_, err = db.Exec(
		"CREATE TABLE t1 (id INTEGER PRIMARY KEY)",
	)
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	err = func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if _, err := tx.Exec("INSERT INTO t1 (id) VALUES(?)", 1); err != nil {
			return err
		}
		return tx.Commit()
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
		regexp.MustCompile(`tracer_go110_test.go:29: Open 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_go110_test.go:29: Exec 0x[0-9a-f]+: CREATE TABLE t1 \(id INTEGER PRIMARY KEY\); args = \[\] ` + timeComponent),
		regexp.MustCompile(`.*:\d+: ResetSession 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_go110_test.go:37: Begin 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_go110_test.go:42: Exec 0x[0-9a-f]+: INSERT INTO t1 \(id\) VALUES\(\?\); args = \[1\] ` + timeComponent),
		regexp.MustCompile(`tracer_go110_test.go:45: Commit 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`.*:\d+: ResetSession 0x[0-9a-f]+ ` + timeComponent),
		regexp.MustCompile(`tracer_go110_test.go:51: Close 0x[0-9a-f]+ ` + timeComponent),
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
