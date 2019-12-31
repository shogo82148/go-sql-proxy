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
	"github.com/shogo82148/go-sql-proxy": struct{}{},
}

// TracerOptions holds the tracing option.
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

// NewTraceHooks creates new HooksContext which trace SQL queries.
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
		PostOpen: func(_ context.Context, ctx interface{}, conn *Conn, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			if conn != nil {
				fmt.Fprintf(buf, "Open %p", conn.Conn)
			} else {
				fmt.Fprint(buf, "Open nil")
			}
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
			fmt.Fprintf(buf, "Exec %p: ", stmt.Conn.Conn)
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
			fmt.Fprintf(buf, "Query %p: ", stmt.Conn.Conn)
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
		PostBegin: func(_ context.Context, ctx interface{}, conn *Conn, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			fmt.Fprintf(buf, "Begin %p", conn.Conn)
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
		PostCommit: func(_ context.Context, ctx interface{}, tx *Tx, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			fmt.Fprintf(buf, "Commit %p", tx.Conn.Conn)
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
		PostRollback: func(_ context.Context, ctx interface{}, tx *Tx, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			fmt.Fprintf(buf, "Rollback %p", tx.Conn.Conn)
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
		PreClose: func(_ context.Context, _ *Conn) (interface{}, error) {
			return time.Now(), nil
		},
		PostClose: func(_ context.Context, ctx interface{}, conn *Conn, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			fmt.Fprintf(buf, "Close %p", conn.Conn)
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
		PreResetSession: func(_ context.Context, _ *Conn) (interface{}, error) {
			return time.Now(), nil
		},
		PostResetSession: func(_ context.Context, ctx interface{}, conn *Conn, err error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			fmt.Fprintf(buf, "ResetSession %p", conn.Conn)
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
		var i int
		n := runtime.Callers(skip, rpc[:])
		frames := runtime.CallersFrames(rpc[:])
		for i = 0; ; i++ {
			frame, more := frames.Next()
			if !more {
				break
			}
			name := frame.Function
			if name == "" {
				continue
			}
			// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
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
				return skip + i
			}
		}
		if n < len(rpc) {
			break
		}
		skip += i
	}
	return 0
}
