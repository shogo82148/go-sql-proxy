package proxy

import (
	"database/sql"
	"strings"
)

// RegisterProxy creates proxies that do not anything by default,
// and registers the proxies as sql driver.
// Use `proxy.WithHooks(ctx, hooks)` to hook query execution.
// The proxies' names have suffix ":proxy".
func RegisterProxy() {
	for _, driver := range sql.Drivers() {
		if strings.HasSuffix(driver, ":trace") || strings.HasSuffix(driver, ":proxy") {
			continue
		}
		db, err := sql.Open(driver, "")
		if err != nil {
			continue
		}
		defer db.Close()
		sql.Register(driver+":proxy", NewProxyContext(db.Driver()))
	}
}
