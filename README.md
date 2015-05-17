# go-sql-proxy

[![Build Status](https://travis-ci.org/shogo82148/go-sql-proxy.svg?branch=master)](https://travis-ci.org/shogo82148/go-sql-proxy)

A proxy package is a proxy driver for dabase/sql.
You can hook SQL execution.

First, register new proxy driver.

``` go
hooks := &proxy.Hooks{
	// Hook functions here(Open, Exec, Query, etc.)
}
sql.Register("new-proxy-name", proxy.NewProxy(&another.Driver{}, hooks))
```

And then, open new database handle with the registered proxy driver.

``` go
db, err := sql.Open("new-proxy-name", dataSourceName)
```


# EXAMPLE: SQL tracer

``` go
package main

import (
	"database/sql"
	"database/sql/driver"
	"log"

	"github.com/mattn/go-sqlite3"
	"github.com/shogo82148/go-sql-proxy"
)

func main() {
	sql.Register("sqlite3-proxy", proxy.NewProxy(&sqlite3.SQLiteDriver{}, &proxy.Hooks{
		Open: func(_ interface{}, conn *proxy.Conn) error {
			log.Println("Open")
			return nil
		},
		Exec: func(_ interface{}, stmt *proxy.Stmt, args []driver.Value, result driver.Result) error {
			log.Printf("Exec: %s; args = %v\n", stmt.QueryString, args)
			return nil
		},
		Query: func(_ interface{}, stmt *proxy.Stmt, args []driver.Value, rows driver.Rows) error {
			log.Printf("Query: %s; args = %v\n", stmt.QueryString, args)
			return nil
		},
		Begin: func(conn *proxy.Conn) error {
			log.Println("Begin")
			return nil
		},
		Commit: func(tx *proxy.Tx) error {
			log.Println("Commit")
			return nil
		},
		Rollback: func(tx *proxy.Tx) error {
			log.Println("Rollback")
			return nil
		},
	}))

	db, err := sql.Open("sqlite3-proxy", ":memory:")
	if err != nil {
		log.Fatalf("Open filed: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		"CREATE TABLE t1 (id INTEGER PRIMARY KEY)",
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

# LICENSE

This software is released under the MIT License, see LICENSE file.

## godoc

See [godoc](https://godoc.org/github.com/shogo82148/go-sql-proxy) for more imformation.
