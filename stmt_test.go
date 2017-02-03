// +build go1.8

package proxy

import "database/sql/driver"

var _ driver.Stmt = &Stmt{}
var _ driver.StmtExecContext = &Stmt{}
var _ driver.StmtQueryContext = &Stmt{}
