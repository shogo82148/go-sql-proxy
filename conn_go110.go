// +build go1.10

package proxy

import (
	"context"
	"database/sql/driver"
)

// ResetSession resets the state of Conn.
func (conn *Conn) ResetSession(ctx context.Context) error {
	var err error
	var myctx interface{}
	hooks := conn.Proxy.getHooks(ctx)

	if hooks != nil {
		defer func() { hooks.postResetSession(ctx, myctx, conn, err) }()
		if myctx, err = hooks.preResetSession(ctx, conn); err != nil {
			return err
		}
	}

	if sr, ok := conn.Conn.(driver.SessionResetter); ok {
		err = sr.ResetSession(ctx)
		if err != nil {
			return err
		}
	}

	if hooks != nil {
		err = hooks.resetSession(ctx, myctx, conn)
	}
	return err
}
