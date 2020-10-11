# go-sql-proxy

[![Build Status](https://github.com/shogo82148/go-sql-proxy/workflows/Test/badge.svg?branch=main)](https://github.com/shogo82148/go-sql-proxy/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/shogo82148/go-sql-proxy)](https://goreportcard.com/report/github.com/shogo82148/go-sql-proxy)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/shogo82148/go-sql-proxy?tab=doc)](https://pkg.go.dev/github.com/shogo82148/go-sql-proxy?tab=doc)
[![Coverage Status](https://coveralls.io/repos/github/shogo82148/go-sql-proxy/badge.svg?branch=main)](https://coveralls.io/github/shogo82148/go-sql-proxy?branch=main)

The proxy package is a proxy driver for the database/sql package.
You can hook SQL executions.
It supports Go 1.8 or laster.

## SYNOPSIS

### Use Ready‐Made SQL tracer

`proxy.RegisterTracer` is a shortcut for registering a SQL query tracer.

``` go
package main

import (
	"context"
	"database/sql"

	"github.com/shogo82148/go-sql-proxy"
)

func main() {
	proxy.RegisterTracer()

	db, _ := sql.Open("origin:trace", "data source")
	db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)")
	// STDERR: main.go:14: Exec: CREATE TABLE t1 (id INTEGER PRIMARY KEY); args = [] (0s)
}
```

Use `proxy.NewTraceProxy` to change the log output destination.

``` go
logger := New(os.Stderr, "", LstdFlags)
tracer := NewTraceProxy(&another.Driver{}, logger)
sql.Register("origin:trace", tracer)
db, err := sql.Open("origin:tracer", "data source")
```


### Use with the context package

From Go version 1.8 onward, database/sql supports the context package.
You can register your hooks into the context.

``` go
package main

import (
	"context"
	"database/sql"

	"github.com/shogo82148/go-sql-proxy"
)

var tracer = proxy.NewTraceHooks(proxy.TracerOptions{})

func main() {
	proxy.RegisterProxy()

	db, _ := sql.Open("origin:proxy", "data source")

	// The tracer is enabled in this context.
	ctx := proxy.WithHooks(context.Background(), tracer)
	db.ExecContext(ctx, "CREATE TABLE t1 (id INTEGER PRIMARY KEY)")
}
```

### Create your own hooks

First, register new proxy driver.

``` go
hooks := &proxy.HooksContext{
	// Hook functions here(Open, Exec, Query, etc.)
	// See godoc for more details
}
sql.Register("new-proxy-name", proxy.NewProxyContext(&origin.Driver{}, hooks))
```

And then, open new database handle with the registered proxy driver.

``` go
db, err := sql.Open("new-proxy-name", "data source")
```

## EXAMPLES

### EXAMPLE: SQL tracer

``` go
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"

	"github.com/mattn/go-sqlite3"
	"github.com/shogo82148/go-sql-proxy"
)

func main() {
	sql.Register("sqlite3-proxy", proxy.NewProxyContext(&sqlite3.SQLiteDriver{}, &proxy.HooksContext{
		Open: func(_ context.Context, _ interface{}, conn *proxy.Conn) error {
			log.Println("Open")
			return nil
		},
		Exec: func(_ context.Context, _ interface{}, stmt *proxy.Stmt, args []driver.NamedValue, result driver.Result) error {
			log.Printf("Exec: %s; args = %v\n", stmt.QueryString, args)
			return nil
		},
		Query: func(_ context.Context, _ interface{}, stmt *proxy.Stmt, args []driver.NamedValue, rows driver.Rows) error {
			log.Printf("Query: %s; args = %v\n", stmt.QueryString, args)
			return nil
		},
		Begin: func(_ context.Context, _ interface{}, conn *proxy.Conn) error {
			log.Println("Begin")
			return nil
		},
		Commit: func(_ context.Context, _ interface{}, tx *proxy.Tx) error {
			log.Println("Commit")
			return nil
		},
		Rollback: func(_ context.Context, _ interface{}, tx *proxy.Tx) error {
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

### EXAMPLE: elapsed time logger

``` go
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/shogo82148/go-sql-proxy"
)

func main() {
	sql.Register("sqlite3-proxy", proxy.NewProxyContext(&sqlite3.SQLiteDriver{}, &proxy.HooksContext{
		PreExec: func(_ context.Context, _ *proxy.Stmt, _ []driver.NamedValue) (interface{}, error) {
			// The first return value(time.Now()) is passed to both `Hooks.Exec` and `Hook.ExecPost` callbacks.
			return time.Now(), nil
		},
		PostExec: func(_ context.Context, ctx interface{}, stmt *proxy.Stmt, args []driver.NamedValue, _ driver.Result, _ error) error {
			// The `ctx` parameter is the return value supplied from the `Hooks.PreExec` method, and may be nil.
			log.Printf("Query: %s; args = %v (%s)\n", stmt.QueryString, args, time.Since(ctx.(time.Time)))
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


## LICENSE

This software is released under the MIT License, see LICENSE file.

## godoc

See [godoc](https://pkg.go.dev/github.com/shogo82148/go-sql-proxy?tab=doc) for more information.
