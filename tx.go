package proxy

import "database/sql/driver"

type Tx struct {
	Tx    driver.Tx
	Proxy *Proxy
}

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
