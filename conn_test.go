package proxy

import "database/sql/driver"

var _ driver.Conn = (*Conn)(nil)
var _ driver.ConnBeginTx = (*Conn)(nil)
var _ driver.ConnPrepareContext = (*Conn)(nil)
var _ driver.Execer = (*Conn)(nil)
var _ driver.ExecerContext = (*Conn)(nil)
var _ driver.Pinger = (*Conn)(nil)
var _ driver.Queryer = (*Conn)(nil)
var _ driver.QueryerContext = (*Conn)(nil)
var _ namedValueChecker = (*Conn)(nil)
var _ sessionResetter = (*Conn)(nil)
