// +build go1.9

package proxy

import "database/sql/driver"

// copied from sql/driver/convert.go
// defaultCheckNamedValue wraps the default ColumnConverter to have the same
// function signature as the CheckNamedValue in the driver.NamedValueChecker
// interface.
func defaultCheckNamedValue(nv *driver.NamedValue) (err error) {
	nv.Value, err = driver.DefaultParameterConverter.ConvertValue(nv.Value)
	return err
}

// CheckNamedValue for implementing NamedValueChecker
// This function may be unnecessary because `proxy.Stmt` already implements `NamedValueChecker`,
// but it is implemented just in case.
func (conn *Conn) CheckNamedValue(nv *driver.NamedValue) (err error) {
	if nvc, ok := conn.Conn.(driver.NamedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// fallback to default
	return defaultCheckNamedValue(nv)
}
