// +build !go1.8

package proxy

import "database/sql/driver"

// Tx adds hook points into "database/sql/driver".Tx.
type Tx struct {
	Tx    driver.Tx
	Proxy *Proxy
}

// Commit commits the transaction.
// It will trigger PreCommit, Commit, PostCommit hooks.
func (tx *Tx) Commit() error {
	var err error
	var ctx interface{}

	if h := tx.Proxy.Hooks.PostCommit; h != nil {
		defer func() { h(ctx, tx) }()
	}

	if h := tx.Proxy.Hooks.PreCommit; h != nil {
		if ctx, err = h(tx); err != nil {
			return err
		}
	}

	if err = tx.Tx.Commit(); err != nil {
		return err
	}

	if hook := tx.Proxy.Hooks.Commit; hook != nil {
		return hook(ctx, tx)
	}

	return nil
}

// Rollback rollbacks the transaction.
// It will trigger PreRollback, Rollback, PostRollback hooks.
func (tx *Tx) Rollback() error {
	var err error
	var ctx interface{}

	if h := tx.Proxy.Hooks.PostRollback; h != nil {
		defer func() { h(ctx, tx) }()
	}

	if h := tx.Proxy.Hooks.PreRollback; h != nil {
		if ctx, err = h(tx); err != nil {
			return err
		}
	}

	if err := tx.Tx.Rollback(); err != nil {
		return err
	}

	if hook := tx.Proxy.Hooks.Rollback; hook != nil {
		return hook(ctx, tx)
	}

	return nil
}
