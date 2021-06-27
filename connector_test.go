//go:build go1.10
// +build go1.10

package proxy

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

var _ io.Closer = (*Connector)(nil)
var _ driver.Connector = (*Connector)(nil)
var _ driver.DriverContext = (*Proxy)(nil)

var _ io.Closer = (*closerConnector)(nil)
var _ driver.Connector = (*closerConnector)(nil)

type closerConnector struct {
	// the result of Close() method
	errClose error

	// a flag whether Close() method is called
	closed bool
}

func (c *closerConnector) Connect(ctx context.Context) (driver.Conn, error) {
	panic("never used")
}

func (c *closerConnector) Driver() driver.Driver {
	return fdriverctx
}

func (c *closerConnector) Close() error {
	c.closed = true
	return c.errClose
}

func TestConnectorClose(t *testing.T) {
	t.Run("c.Connector doesn't implement io.Closer", func(t *testing.T) {
		c0 := &fakeConnector{
			driver: fdriverctx,
		}
		c1 := NewConnector(c0)
		if err := c1.(io.Closer).Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Closing c.Connector succeeds", func(t *testing.T) {
		c0 := &closerConnector{}
		c1 := NewConnector(c0)
		if err := c1.(io.Closer).Close(); err != nil {
			t.Fatal(err)
		}
		if !c0.closed {
			t.Errorf("c.Connector should be closed, but not")
		}
	})

	t.Run("Closing c.Connector fails", func(t *testing.T) {
		errClose := errors.New("some error while closing")
		c0 := &closerConnector{
			errClose: errClose,
		}
		c1 := NewConnector(c0)
		if err := c1.(io.Closer).Close(); err != errClose {
			t.Errorf("want err is %v, got %v", errClose, err)
		}
		if !c0.closed {
			t.Errorf("c.Connector should be closed, but not")
		}
	})
}
