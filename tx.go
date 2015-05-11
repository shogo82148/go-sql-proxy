package proxy

import "database/sql/driver"

type Tx struct {
	Tx    driver.Tx
	Proxy *Proxy
}

func (tx *Tx) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		return err
	}

	if hook := tx.Proxy.Hooks.Commit; hook != nil {
		return hook(tx)
	}

	return nil
}

func (tx *Tx) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil {
		return err
	}

	if hook := tx.Proxy.Hooks.Rollback; hook != nil {
		return hook(tx)
	}

	return nil
}
