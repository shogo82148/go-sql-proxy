package proxy

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"runtime"
	"sync"
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

// DefaultPackageFilter ignores some database util package.
var DefaultPackageFilter = PackageFilter{
	"database/sql":                       struct{}{},
	"github.com/shogo82148/txmanager":    struct{}{},
	"github.com/shogo82148/go-sql-proxy": struct{}{},
}

// TracerOptions holds the tarcing option.
type TracerOptions struct {
	// Outputter is the output of the log.
	// If is nil nil, log.Output is used.
	Outputter Outputter

	// Filter is used by the tracing proxy for skipping database libraries (e.g. O/R mapper).
	// If it is nil, DefaultPackageFilter is used.
	Filter Filter

	// SlowQuery is a threshold duration to output into log.
	// output all queries if SlowQuery is zero.
	SlowQuery time.Duration
}

// NewTraceProxy generates a proxy that logs queries.
func NewTraceProxy(d driver.Driver, o Outputter) *Proxy {
	return NewProxyContext(d, NewTraceHooks(TracerOptions{
		Outputter: o,
	}))
}

// NewTraceProxyWithFilter generates a proxy that logs queries.
func NewTraceProxyWithFilter(d driver.Driver, o Outputter, f Filter) *Proxy {
	return NewProxyContext(d, NewTraceHooks(TracerOptions{
		Outputter: o,
		Filter:    f,
	}))
}

func NewTraceHooks(opt TracerOptions) *HooksContext {
	f := opt.Filter
	if f == nil {
		f = DefaultPackageFilter
	}
	o := opt.Outputter
	if o == nil {
		o = logger{}
	}
	pool := &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
	return &HooksContext{
		PreOpen: func(_ context.Context, _ string) (interface{}, error) {
			return time.Now(), nil
		},
		PostOpen: func(_ context.Context, ctx interface{}, _ driver.Conn, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			io.WriteString(buf, "Open")
			if err != nil {
				fmt.Fprintf(buf, "; err = %#v", err.Error())
			}
			io.WriteString(buf, " (")
			io.WriteString(buf, d.String())
			io.WriteString(buf, ")")
			s := buf.String()
			pool.Put(buf)
			o.Output(findCaller(f), s)
			return nil
		},
		PreExec: func(_ context.Context, _ *Stmt, _ []driver.NamedValue) (interface{}, error) {
			return time.Now(), nil
		},
		PostExec: func(_ context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, _ driver.Result, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			io.WriteString(buf, "Exec: ")
			io.WriteString(buf, stmt.QueryString)
			io.WriteString(buf, "; args = [")
			writeNamedValues(buf, args)
			io.WriteString(buf, "]")
			if err != nil {
				fmt.Fprintf(buf, "; err = %#v", err.Error())
			}
			io.WriteString(buf, " (")
			io.WriteString(buf, d.String())
			io.WriteString(buf, ")")
			s := buf.String()
			pool.Put(buf)
			o.Output(findCaller(f), s)
			return nil
		},
		PreQuery: func(_ context.Context, stmt *Stmt, args []driver.NamedValue) (interface{}, error) {
			return time.Now(), nil
		},
		PostQuery: func(_ context.Context, ctx interface{}, stmt *Stmt, args []driver.NamedValue, _ driver.Rows, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			io.WriteString(buf, "Query: ")
			io.WriteString(buf, stmt.QueryString)
			io.WriteString(buf, "; args = [")
			writeNamedValues(buf, args)
			io.WriteString(buf, "]")
			if err != nil {
				fmt.Fprintf(buf, "; err = %#v", err.Error())
			}
			io.WriteString(buf, " (")
			io.WriteString(buf, d.String())
			io.WriteString(buf, ")")
			s := buf.String()
			pool.Put(buf)
			o.Output(findCaller(f), s)
			return nil
		},
		PreBegin: func(_ context.Context, _ *Conn) (interface{}, error) {
			return time.Now(), nil
		},
		PostBegin: func(_ context.Context, ctx interface{}, _ *Conn, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			io.WriteString(buf, "Begin")
			if err != nil {
				fmt.Fprintf(buf, "; err = %#v", err.Error())
			}
			io.WriteString(buf, " (")
			io.WriteString(buf, d.String())
			io.WriteString(buf, ")")
			s := buf.String()
			pool.Put(buf)
			o.Output(findCaller(f), s)
			return nil
		},
		PreCommit: func(_ context.Context, _ *Tx) (interface{}, error) {
			return time.Now(), nil
		},
		PostCommit: func(_ context.Context, ctx interface{}, _ *Tx, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			io.WriteString(buf, "Commit")
			if err != nil {
				fmt.Fprintf(buf, "; err = %#v", err.Error())
			}
			io.WriteString(buf, " (")
			io.WriteString(buf, d.String())
			io.WriteString(buf, ")")
			s := buf.String()
			pool.Put(buf)
			o.Output(findCaller(f), s)
			return nil
		},
		PreRollback: func(_ context.Context, _ *Tx) (interface{}, error) {
			return time.Now(), nil
		},
		PostRollback: func(_ context.Context, ctx interface{}, _ *Tx, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			io.WriteString(buf, "Rollback")
			if err != nil {
				fmt.Fprintf(buf, "; err = %#v", err.Error())
			}
			io.WriteString(buf, " (")
			io.WriteString(buf, d.String())
			io.WriteString(buf, ")")
			s := buf.String()
			pool.Put(buf)
			o.Output(findCaller(f), s)
			return nil
		},
	}
}

func writeNamedValues(w io.Writer, args []driver.NamedValue) {
	for i, arg := range args {
		if i != 0 {
			io.WriteString(w, ", ")
		}
		if len(arg.Name) > 0 {
			io.WriteString(w, arg.Name)
			io.WriteString(w, ":")
		}
		fmt.Fprintf(w, "%#v", arg.Value)
	}
}

func findCaller(f Filter) int {
	// skip starts 4. 0: Callers, 1: findCaller, 2: hooks, 3: proxy-funcs, 4: database/sql, and equals or greater than 5: user-funcs
	skip := 5
	for {
		var rpc [8]uintptr
		n := runtime.Callers(skip, rpc[:])

		for i, pc := range rpc[:n] {
			// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
			name := runtime.FuncForPC(pc).Name()
			dotIdx := 0
			for j := len(name) - 1; j >= 0; j-- {
				if name[j] == '.' {
					dotIdx = j
				} else if name[j] == '/' {
					break
				}
			}
			packageName := name[:dotIdx]
			if f.DoOutput(packageName) {
				// -1 because the meaning of skip differs between Caller and Callers.
				return skip + i - 1
			}
		}
		if n < len(rpc) {
			break
		}
		skip += n
	}
	return 0
}
