//go:build go1.10
// +build go1.10

package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type fakeDriverCtx fakeDriver
type fakeConnector struct {
	driver *fakeDriverCtx
	opt    *fakeConnOption
	db     *fakeDB
}

var fdriverctx = &fakeDriverCtx{}
var _ driver.DriverContext = (*fakeDriverCtx)(nil)
var _ driver.Connector = (*fakeConnector)(nil)

func init() {
	sql.Register("fakedbctx", fdriverctx)
}

func (d *fakeDriverCtx) Open(name string) (driver.Conn, error) {
	return nil, errors.New("not implemented")
}

func (d *fakeDriverCtx) OpenConnector(name string) (driver.Connector, error) {
	var opt fakeConnOption
	err := json.Unmarshal([]byte(name), &opt)
	if err != nil {
		return nil, err
	}

	// validate options
	switch opt.ConnType {
	case "", "fakeConn", "fakeConnExt", "fakeConnCtx":
		// validation OK
	default:
		return nil, errors.New("known ConnType")
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	db, ok := d.dbs[opt.Name]
	if !ok {
		db = &fakeDB{
			log: &bytes.Buffer{},
		}
		if d.dbs == nil {
			d.dbs = make(map[string]*fakeDB)
		}
		d.dbs[name] = db
	}

	return &fakeConnector{
		driver: d,
		opt:    &opt,
		db:     db,
	}, nil
}

func (d *fakeDriverCtx) DB(name string) *fakeDB {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dbs[name]
}

func (c *fakeConnector) Connect(ctx context.Context) (driver.Conn, error) {
	var conn driver.Conn
	switch c.opt.ConnType {
	case "", "fakeConn":
		conn = &fakeConn{
			db:  c.db,
			opt: c.opt,
		}
	case "fakeConnExt":
		conn = &fakeConnExt{
			db:  c.db,
			opt: c.opt,
		}
	case "fakeConnCtx":
		conn = &fakeConnCtx{
			db:  c.db,
			opt: c.opt,
		}
	default:
		return nil, errors.New("known ConnType")
	}

	return conn, nil
}

func (c *fakeConnector) Driver() driver.Driver {
	return c.driver
}
