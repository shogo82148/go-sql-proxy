// +build go1.9

package proxy

import "database/sql/driver"

// CheckNamedValue for implementing NamedValueChecker
func (stmt *Stmt) CheckNamedValue(nv *driver.NamedValue) (err error) {
	if nvc, ok := stmt.Stmt.(driver.NamedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// When converting data in sql/driver/convert.go, it is checked first whether the `stmt`
	// implements `NamedValueChecker`, and then checks if `conn` implements NamedValueChecker.
	// In the case of "go-sql-proxy", the `proxy.Stmt` "implements" `CheckNamedValue` here,
	// so we also check both `stmt` and `conn` inside here.
	if nvc, ok := stmt.Conn.Conn.(driver.NamedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// fallback to default
	return defaultCheckNamedValue(nv)
}
