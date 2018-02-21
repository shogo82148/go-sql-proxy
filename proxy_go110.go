// +build go1.10

package proxy

import (
	"context"
	"database/sql/driver"
)

type driverConnector struct {
	driver driver.Driver
	name   string
}

func (c *driverConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.driver.Open(c.name)
}

func (c *driverConnector) Driver() driver.Driver {
	return c.driver
}

// OpenConnector creates new connection which is wrapped by Conn.
// It will triggers PreOpen, Open, PostOpen hooks.
func (p *Proxy) OpenConnector(name string) (driver.Connector, error) {
	if d, ok := p.Driver.(driver.DriverContext); ok {
		c, err := d.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &connector{
			p:         p,
			connector: c,
		}, nil
	}

	// fallback
	return &driverConnector{
		driver: p,
		name:   name,
	}, nil
}
