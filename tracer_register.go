package proxy

import (
	"database/sql"
	"log"
	"strings"
)

// RegisterTracer creates proxies that log queries from the sql drivers already registered,
// and registers the proxies as sql driver.
// The proxies' names have suffix ":trace".
func RegisterTracer() {
	for _, driver := range sql.Drivers() {
		if strings.HasSuffix(driver, ":trace") || strings.HasSuffix(driver, ":proxy") {
			continue
		}
		db, err := sql.Open(driver, "")
		if err != nil {
			continue
		}
		defer db.Close()
		sql.Register(driver+":trace", NewTraceProxy(db.Driver(), logger{}))
	}
}

type logger struct{}

// Output outputs the log by log package.
func (logger) Output(calldepth int, s string) error {
	return log.Output(calldepth, s)
}
