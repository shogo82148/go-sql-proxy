// +build go1.10

package proxy

import (
	"context"
	"database/sql/driver"
)

type connector struct {
	connector driver.Connector
	hooks     []*HooksContext
}

// NewConnector creates new connector.
// It adds hook points to other sql drivers.
func NewConnector(c driver.Connector, hs ...*HooksContext) driver.Connector {
	return &connector{
		connector: c,
		hooks:     hs,
	}
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	proxy := NewProxyContext(c.connector.Driver(), c.hooks...)
	return &Conn{
		Conn:  conn,
		Proxy: proxy,
	}, nil
}

func (c *connector) Driver() driver.Driver {
	return NewProxyContext(c.connector.Driver(), c.hooks...)
}
