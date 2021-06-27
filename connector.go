//go:build go1.10
// +build go1.10

package proxy

import (
	"context"
	"database/sql/driver"
	"io"
)

// Connector adds hook points into "database/sql/driver".Connector.
type Connector struct {
	Proxy     *Proxy
	Connector driver.Connector
	Name      string
}

// Connect returns a connection to the database which wrapped by Conn.
// It will triggers PreOpen, Open, PostOpen hooks.
func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	var err error
	var myctx interface{}
	var conn driver.Conn
	var myconn *Conn
	hooks := c.Proxy.getHooks(ctx)

	if hooks != nil {
		// Setup PostOpen. This needs to be a closure like this
		// or otherwise changes to the `ctx` and `conn` parameters
		// within this Open() method does not get applied at the
		// time defer is fired
		defer func() { hooks.postOpen(ctx, myctx, myconn, err) }()
		if myctx, err = hooks.preOpen(ctx, c.Name); err != nil {
			return nil, err
		}
	}
	conn, err = c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}

	myconn = &Conn{
		Conn:  conn,
		Proxy: c.Proxy,
	}

	if hooks != nil {
		if err = hooks.open(ctx, myctx, myconn); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return myconn, nil
}

// Driver returns the underlying Driver of the Connector.
func (c *Connector) Driver() driver.Driver {
	return c.Proxy
}

// Close closes the c.Connector if it implements the io.Closer interface.
// It is called by the DB.Close method from Go 1.17.
func (c *Connector) Close() error {
	if c, ok := c.Connector.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewConnector creates new proxied Connector.
func NewConnector(c driver.Connector, hs ...*HooksContext) driver.Connector {
	p := NewProxyContext(c.Driver(), hs...)
	return &Connector{
		Proxy:     p,
		Connector: c,
		Name:      "",
	}
}

type fallbackConnector struct {
	driver driver.Driver
	name   string
}

func (c *fallbackConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.driver.Open(c.name)
	if err != nil {
		return nil, err
	}
	select {
	default:
	case <-ctx.Done():
		conn.Close()
		return nil, ctx.Err()
	}
	return conn, nil
}

func (c *fallbackConnector) Driver() driver.Driver {
	return c.driver
}

// OpenConnector creates a new connector which is wrapped by Connector.
// It will triggers PreOpen, Open, PostOpen hooks.
func (p *Proxy) OpenConnector(name string) (driver.Connector, error) {
	if d, ok := p.Driver.(driver.DriverContext); ok {
		c, err := d.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &Connector{
			Proxy:     p,
			Connector: c,
			Name:      name,
		}, nil
	}
	return &Connector{
		Proxy: p,
		Connector: &fallbackConnector{
			driver: p.Driver,
			name:   name,
		},
		Name: name,
	}, nil
}
