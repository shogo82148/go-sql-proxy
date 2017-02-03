// +build go1.8

package proxy

import "database/sql/driver"

var _ driver.Conn = &Conn{}

//var _ driver.ConnBeginTx = &Conn{}
var _ driver.ConnPrepareContext = &Conn{}
var _ driver.Execer = &Conn{}
var _ driver.ExecerContext = &Conn{}
var _ driver.Pinger = &Conn{}
var _ driver.Queryer = &Conn{}
var _ driver.QueryerContext = &Conn{}
