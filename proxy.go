// a proxy package is a proxy driver for database/sql.

package proxy

import (
	"context"
	"database/sql/driver"
)

// namedValueChecker is the same as driver.NamedValueChecker.
// Copied from database/sql/driver/driver.go for supporting Go 1.8.
type namedValueChecker interface {
	// CheckNamedValue is called before passing arguments to the driver
	// and is called in place of any ColumnConverter. CheckNamedValue must do type
	// validation and conversion as appropriate for the driver.
	CheckNamedValue(*driver.NamedValue) error
}

// Proxy is a sql driver.
// It adds hook points to other sql drivers.
type Proxy struct {
	Driver driver.Driver
	hooks  hooks
}

// NewProxy creates new Proxy driver.
// DEPRECATED: You should use NewProxyContext instead.
func NewProxy(driver driver.Driver, hs ...*Hooks) *Proxy {
	switch {
	case len(hs) == 0:
		return &Proxy{
			Driver: driver,
		}
	case len(hs) == 1 && hs[0] != nil:
		return &Proxy{
			Driver: driver,
			hooks:  hs[0],
		}
	}

	hooksSlice := make([]hooks, 0, len(hs))
	for _, hk := range hs {
		if hk != nil {
			hooksSlice = append(hooksSlice, hk)
		}
	}
	return &Proxy{
		Driver: driver,
		hooks:  multipleHooks(hooksSlice),
	}
}

// NewProxyContext creates new Proxy driver.
func NewProxyContext(driver driver.Driver, hs ...*HooksContext) *Proxy {
	switch {
	case len(hs) == 0:
		return &Proxy{
			Driver: driver,
		}
	case len(hs) == 1 && hs[0] != nil:
		return &Proxy{
			Driver: driver,
			hooks:  hs[0],
		}
	}

	hooksSlice := make([]hooks, 0, len(hs))
	for _, hk := range hs {
		if hk != nil {
			hooksSlice = append(hooksSlice, hk)
		}
	}
	return &Proxy{
		Driver: driver,
		hooks:  multipleHooks(hooksSlice),
	}
}

func (p *Proxy) getHooks(ctx context.Context) hooks {
	if h, ok := ctx.Value(contextHooksKey{}).(hooks); ok {
		// Make the caller nil check easy.
		if h == (*Hooks)(nil) || h == (*HooksContext)(nil) {
			return nil
		}
		return h
	}
	return p.hooks
}

// Open creates new connection which is wrapped by Conn.
// It will triggers PreOpen, Open, PostOpen hooks.
func (p *Proxy) Open(name string) (driver.Conn, error) {
	c := context.Background()
	var err error
	var ctx interface{}
	var conn driver.Conn
	var myconn *Conn

	if p.hooks != nil {
		// Setup PostOpen. This needs to be a closure like this
		// or otherwise changes to the `ctx` and `conn` parameters
		// within this Open() method does not get applied at the
		// time defer is fired
		defer func() { p.hooks.postOpen(c, ctx, myconn, err) }()

		if ctx, err = p.hooks.preOpen(c, name); err != nil {
			return nil, err
		}
	}
	conn, err = p.Driver.Open(name)
	if err != nil {
		return nil, err
	}

	myconn = &Conn{
		Conn:  conn,
		Proxy: p,
	}

	if p.hooks != nil {
		if err = p.hooks.open(c, ctx, myconn); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return myconn, nil
}
