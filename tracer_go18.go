// +build go1.8

package proxy

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"sync"
	"time"
)

// TracerOptions holds the tarcing option.
type TracerOptions struct {
	// Outputter is the output of the log.
	Outputter Outputter

	// Filter is used by the tracing proxy for skipping database libraries (e.g. O/R mapper).
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
			o.Output(findCaller(f), buf.String())
			pool.Put(buf)
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
			o.Output(findCaller(f), buf.String())
			pool.Put(buf)
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
			o.Output(findCaller(f), buf.String())
			pool.Put(buf)
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
			o.Output(findCaller(f), buf.String())
			pool.Put(buf)
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
			o.Output(findCaller(f), buf.String())
			pool.Put(buf)
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
			o.Output(findCaller(f), buf.String())
			pool.Put(buf)
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
