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
	hooks := tx.Proxy.getHooks(tx.ctx)
	if hooks != nil {
		defer func() { hooks.postCommit(tx.ctx, ctx, tx, err) }()
		if ctx, err = hooks.preCommit(tx.ctx, tx); err != nil {
			return err
		}
	}

	if err = tx.Tx.Commit(); err != nil {
		return err
	}

	if hooks != nil {
		return hooks.commit(tx.ctx, ctx, tx)
	}
	return nil
}

// Rollback rollbacks the transaction.
// It will trigger PreRollback, Rollback, PostRollback hooks.
func (tx *Tx) Rollback() error {
	var err error
	var ctx interface{}
	hooks := tx.Proxy.getHooks(tx.ctx)
	if hooks != nil {
		defer func() { hooks.postRollback(tx.ctx, ctx, tx, err) }()
		if ctx, err = hooks.preRollback(tx.ctx, tx); err != nil {
			return err
		}
	}

	if err = tx.Tx.Rollback(); err != nil {
		return err
	}

	if hooks != nil {
		return hooks.rollback(tx.ctx, ctx, tx)
	}
	return nil
}
