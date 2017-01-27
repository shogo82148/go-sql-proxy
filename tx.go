// +build go1.8

package proxy

import (
	"context"
	"database/sql/driver"
)

// Tx adds hook points into "database/sql/driver".Tx.
type Tx struct {
	Tx    driver.Tx
	Proxy *Proxy
	ctx   context.Context
}

// Commit commits the transaction.
// It will trigger PreCommit, Commit, PostCommit hooks.
func (tx *Tx) Commit() error {
	var err error
	var ctx interface{}

	defer func() { tx.Proxy.Hooks.postCommit(tx.ctx, ctx, tx, err) }()
	if ctx, err = tx.Proxy.Hooks.preCommit(tx.ctx, tx); err != nil {
		return err
	}

	if err = tx.Tx.Commit(); err != nil {
		return err
	}

	return tx.Proxy.Hooks.commit(tx.ctx, ctx, tx)
}

// Rollback rollbacks the transaction.
// It will trigger PreRollback, Rollback, PostRollback hooks.
func (tx *Tx) Rollback() error {
	var err error
	var ctx interface{}

	defer func() { tx.Proxy.Hooks.postRollback(tx.ctx, ctx, tx, err) }()
	if ctx, err = tx.Proxy.Hooks.preRollback(tx.ctx, tx); err != nil {
		return err
	}

	if err = tx.Tx.Rollback(); err != nil {
		return err
	}

	return tx.Proxy.Hooks.rollback(tx.ctx, ctx, tx)
}
