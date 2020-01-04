package proxy_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	proxy "github.com/shogo82148/go-sql-proxy"
)

func ExampleRegisterProxy() {
	proxy.RegisterProxy()
	db, err := sql.Open("fakedb:proxy", `{"name":"trace"}`)
	if err != nil {
		log.Fatal(err)
	}

	// proxy.WithHooks enables the hooks in this context.
	ctx := proxy.WithHooks(context.Background(), &proxy.HooksContext{
		Ping: func(c context.Context, ctx interface{}, conn *proxy.Conn) error {
			fmt.Println("Ping")
			return nil
		},
	})
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	// Output:
	// Ping

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}
