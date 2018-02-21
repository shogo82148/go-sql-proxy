// +build go1.10

package proxy

import (
	"context"
	"database/sql/driver"
)

type connector struct {
	connector driver.Connector
	hooks     hooks
}

// NewConnector creates new connector.
// It adds hook points to other sql drivers.
func NewConnector(c driver.Connector, hs ...*HooksContext) driver.Connector {
	switch {
	case len(hs) == 0:
		return &connector{
			connector: c,
			hooks:     (*Hooks)(nil),
		}
	case len(hs) == 1 && hs[0] != nil:
		return &connector{
			connector: c,
			hooks:     hs[0],
		}
	}

	hooksSlice := make([]hooks, 0, len(hs))
	for _, hk := range hs {
		if hk != nil {
			hooksSlice = append(hooksSlice, hk)
		}
	}
	return &connector{
		connector: c,
		hooks:     multipleHooks(hooksSlice),
	}
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	proxy := newProxyContext(c.connector.Driver(), c.hooks)
	return &Conn{
		Conn:  conn,
		Proxy: proxy,
	}, nil
}

func (c *connector) Driver() driver.Driver {
	return newProxyContext(c.connector.Driver(), c.hooks)
}
