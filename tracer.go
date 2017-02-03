// +build go1.8

package proxy

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Outputter is what is used by the tracing proxy created via `NewTraceProxy`.
// Anything that implements a `log.Logger` style `Output` method will satisfy
// this interface.
type Outputter interface {
	Output(calldepth int, s string) error
}

// Filter is used by the tracing proxy for skipping database libraries (e.g. O/R mapper).
type Filter interface {
	DoOutput(packageName string) bool
}

// PackageFilter is an implementation of Filter.
type PackageFilter map[string]struct{}

// DoOutput returns false if the package is in the ignored list.
func (f PackageFilter) DoOutput(packageName string) bool {
	_, ok := f[packageName]
	return !ok
}

// Ignore add the package into the ignored list.
func (f PackageFilter) Ignore(packageName string) {
	f[packageName] = struct{}{}
}

var DefaultPackageFilter = PackageFilter{
	"database/sql":                       struct{}{},
	"github.com/shogo82148/txmanager":    struct{}{},
	"github.com/shogo82148/go-sql-proxy": struct{}{},
}

func findCaller(f Filter) int {
	// i starts 4. 0: findCaller, 1: hooks, 2: proxy-funcs, 3: database/sql, and equals or greater than 4: user-funcs
	for i := 4; ; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
		parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		pl := len(parts)
		packageName := ""
		for j := pl - 1; j > 0; j-- { // find a type name
			if parts[j][0] == '(' {
				packageName = strings.Join(parts[0:j], ".")
				break
			}
		}
		if packageName == "" {
			packageName = strings.Join(parts[0:pl-1], ".")
		}

		if f.DoOutput(packageName) {
			return i
		}
	}
	return 0
}

// NewTraceProxy generates a proxy that logs queries.
func NewTraceProxy(d driver.Driver, o Outputter) *Proxy {
	return &Proxy{
		Driver: d,
		Hooks:  NewTraceHooks(o, nil),
	}
}

// NewTraceProxyWithFilter generates a proxy that logs queries.
func NewTraceProxyWithFilter(d driver.Driver, o Outputter, f Filter) *Proxy {
	return &Proxy{
		Driver: d,
		Hooks:  NewTraceHooks(o, f),
	}
}

func NewTraceHooks(o Outputter, f Filter) *HooksContext {
	if f == nil {
		f = DefaultPackageFilter
	}
	return &HooksContext{
		PreOpen: func(_ context.Context, _ string) (interface{}, error) {
			return time.Now(), nil
		},
		PostOpen: func(_ context.Context, ctx interface{}, _ driver.Conn, _ error) error {
			o.Output(
				findCaller(f),
				fmt.Sprintf(
					"Open (%s)",
					time.Since(ctx.(time.Time)),
				),
			)
			return nil
		},
		PreExec: func(_ context.Context, _ *Stmt, _ []driver.NamedValue) (interface{}, error) {
			return time.Now(), nil
		},
		PostExec: func(_ context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, _ driver.Result, _ error) error {
			o.Output(
				findCaller(f),
				fmt.Sprintf(
					"Exec: %s; args = [%s] (%s)",
					stmt.QueryString,
					namedValuesToString(args),
					time.Since(ctx.(time.Time)),
				),
			)
			return nil
		},
		PreQuery: func(_ context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
			return time.Now(), nil
		},
		PostQuery: func(_ context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, _ driver.Rows, _ error) error {
			o.Output(
				findCaller(f),
				fmt.Sprintf(
					"Query: %s; args = [%s] (%s)",
					stmt.QueryString,
					namedValuesToString(args),
					time.Since(ctx.(time.Time)),
				),
			)
			return nil
		},
		PreBegin: func(_ context.Context, _ *Conn) (interface{}, error) {
			return time.Now(), nil
		},
		PostBegin: func(_ context.Context, ctx interface{}, _ *Conn, _ error) error {
			o.Output(
				findCaller(f),
				fmt.Sprintf("Begin (%s)", time.Since(ctx.(time.Time))),
			)
			return nil
		},
		PreCommit: func(_ context.Context, _ *Tx) (interface{}, error) {
			return time.Now(), nil
		},
		PostCommit: func(_ context.Context, ctx interface{}, _ *Tx, _ error) error {
			o.Output(
				findCaller(f),
				fmt.Sprintf("Commit (%s)", time.Since(ctx.(time.Time))),
			)
			return nil
		},
		PreRollback: func(_ context.Context, _ *Tx) (interface{}, error) {
			return time.Now(), nil
		},
		PostRollback: func(_ context.Context, ctx interface{}, _ *Tx, _ error) error {
			o.Output(
				findCaller(f),
				fmt.Sprintf("Rollback (%s)", time.Since(ctx.(time.Time))),
			)
			return nil
		},
	}
}

func namedValuesToString(args []driver.NamedValue) string {
	buf := &bytes.Buffer{}
	for _, arg := range args {
		if len(arg.Name) > 0 {
			fmt.Fprintf(buf, "%s:%#v, ", arg.Name, arg.Value)
		} else {
			fmt.Fprintf(buf, "%#v, ", arg.Value)
		}
	}
	b := buf.Bytes()
	if len(b) < 2 {
		return ""
	}
	str := string(b[:len(b)-2]) // ignore last ','
	return str
}
