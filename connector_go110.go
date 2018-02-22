// +build go1.10

package proxy

import (
	"context"
	"database/sql/driver"
)

type connector struct {
	p         *Proxy
	connector driver.Connector
}

// NewConnector creates new connector.
// It adds hook points to other sql drivers.
func NewConnector(c driver.Connector, hs ...*HooksContext) driver.Connector {
	return &connector{
		p:         NewProxyContext(c.Driver(), hs...),
		connector: c,
	}
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	var err error
	var conn driver.Conn

	conn, err = c.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &Conn{
		Conn:  conn,
		Proxy: c.p,
	}, nil
}

func (c *connector) Driver() driver.Driver {
	return c.p
}
