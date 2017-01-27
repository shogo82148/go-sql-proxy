// +build go1.8

package proxy

import (
	"database/sql"
	"database/sql/driver"
)

type fakeDriver struct {
}

type fakeDB struct {
}

type fakeConn struct {
}

type fakeTx struct {
}

type fakeStmt struct {
}

var fdriver driver.Driver = &fakeDriver{}

func init() {
	sql.Register("fakedb", fdriver)
}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{}, nil
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{}, nil
}

func (c *fakeConn) Close() error {
	return nil
}

func (c *fakeConn) Begin() (driver.Tx, error) {
	return &fakeTx{}, nil
}

func (tx *fakeTx) Commit() error {
	return nil
}

func (tx *fakeTx) Rollback() error {
	return nil
}

func (stmt *fakeStmt) Close() error {
	return nil
}

func (stmt *fakeStmt) NumInput() int {
	return 0
}

func (stmt *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (stmt *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, nil
}
