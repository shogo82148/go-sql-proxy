// +build go1.8

package proxy_test

func init() {
	illegalSQLError = `tracer_test.go:53: Exec: ILLEGAL SQL; args = \[\]; err = "near \\"ILLEGAL\\": syntax error" `
}
