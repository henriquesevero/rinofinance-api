package expense

import "errors"

var (
	// ErrAmountManagedByCard indicates a caller tried to manually set the
	// amount of an expense that is linked to a credit card.
	ErrAmountManagedByCard = errors.New("valor desta saída é controlado automaticamente pelo cartão vinculado")

	// ErrNotCardLinked indicates a caller tried to sync a card total into
	// an expense that has no CardID set.
	ErrNotCardLinked = errors.New("saída não está vinculada a nenhum cartão")

	// ErrAmountManagedByAccount indicates a caller tried to manually set the
	// amount of an expense linked to a bank account's debit purchases.
	ErrAmountManagedByAccount = errors.New("valor desta saída é controlado automaticamente pela conta vinculada")

	// ErrNotAccountLinked indicates a caller tried to sync an account debit
	// total into an expense that has no AccountID set.
	ErrNotAccountLinked = errors.New("saída não está vinculada a nenhuma conta")
)
