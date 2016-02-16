// +build go1.5

package proxy

import (
	"database/sql"
	"log"
	"strings"
)

type logger struct{}

func (_ logger) Output(calldepth int, s string) error {
	return log.Output(calldepth, s)
}

// RegisterTracer creates proxies that logs queries from the sql drivers already registered,
// and registers the proxies as sql driver.
// The proxies' names have suffix ":trace".
func RegisterTracer() {
	for _, driver := range sql.Drivers() {
		if strings.HasSuffix(driver, ":trace") {
			continue
		}
		db, _ := sql.Open(driver, "")
		defer db.Close()
		sql.Register(driver+":trace", NewTraceProxy(db.Driver(), logger{}))
	}
}
