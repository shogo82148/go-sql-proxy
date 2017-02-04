// +build go1.8

package proxy

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
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
	return &HooksContext{
		PreOpen: func(_ context.Context, _ string) (interface{}, error) {
			return time.Now(), nil
		},
		PostOpen: func(_ context.Context, ctx interface{}, _ driver.Conn, _ error) error {
			d := time.Since(ctx.(time.Time))
			if d < opt.SlowQuery {
				return nil
			}
			o.Output(
				findCaller(f),
				fmt.Sprintf("Open (%s)", d),
			)
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
			strErr := ""
			if err != nil {
				strErr = fmt.Sprintf("; err = %#v", err.Error())
			}
			o.Output(
				findCaller(f),
				fmt.Sprintf(
					"Exec: %s; args = [%s]%s (%s)",
					stmt.QueryString,
					namedValuesToString(args),
					strErr,
					d,
				),
			)
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
			strErr := ""
			if err != nil {
				strErr = fmt.Sprintf("; err = %#v", err.Error())
			}
			o.Output(
				findCaller(f),
				fmt.Sprintf(
					"Query: %s; args = [%s]%s (%s)",
					stmt.QueryString,
					namedValuesToString(args),
					strErr,
					d,
				),
			)
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
			strErr := ""
			if err != nil {
				strErr = fmt.Sprintf("; err = %#v", err.Error())
			}
			o.Output(
				findCaller(f),
				fmt.Sprintf("Begin%s (%s)", strErr, d),
			)
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
			strErr := ""
			if err != nil {
				strErr = fmt.Sprintf("; err = %#v", err.Error())
			}
			o.Output(
				findCaller(f),
				fmt.Sprintf("Commit%s (%s)", strErr, d),
			)
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
			strErr := ""
			if err != nil {
				strErr = fmt.Sprintf("; err = %#v", err.Error())
			}
			o.Output(
				findCaller(f),
				fmt.Sprintf("Rollback%s (%s)", strErr, d),
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
