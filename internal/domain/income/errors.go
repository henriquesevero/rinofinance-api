package income

import "errors"

var (
	// ErrAmountManagedByAccount indicates a caller tried to manually set the
	// amount of an income that is linked to a bank account.
	ErrAmountManagedByAccount = errors.New("valor desta entrada é controlado automaticamente pela conta vinculada")

	// ErrNotAccountLinked indicates a caller tried to sync an account
	// balance into an income that has no AccountID set.
	ErrNotAccountLinked = errors.New("entrada não está vinculada a nenhuma conta")
)
