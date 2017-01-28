// +build go1.8

package proxy

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"
)

type fakeDriver struct {
	mu  sync.Mutex
	dbs map[string]*fakeDB
}

type fakeDB struct {
	mu  sync.Mutex
	log *bytes.Buffer
}

type fakeConn struct {
	db *fakeDB
}

type fakeTx struct {
}

type fakeStmt struct {
}

var fdriver = &fakeDriver{}

func init() {
	sql.Register("fakedb", fdriver)
}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	db, ok := d.dbs[name]
	if !ok {
		db = &fakeDB{
			log: &bytes.Buffer{},
		}
		if d.dbs == nil {
			d.dbs = make(map[string]*fakeDB)
		}
		d.dbs[name] = db
	}
	return &fakeConn{
		db: db,
	}, nil
}

func (d *fakeDriver) DB(name string) *fakeDB {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dbs[name]
}

// Log write the params to the log.
func (db *fakeDB) Log(params ...interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()
	fmt.Fprintln(db.log, params...)
}

// LogToString converts the log into string.
func (db *fakeDB) LogToString() string {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.log.String()
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	c.db.Log("[Prepare]", query)
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
