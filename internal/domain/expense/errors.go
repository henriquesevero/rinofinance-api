package expense

import "errors"

var (
	ErrAmountManagedByCard = errors.New("valor desta saída é controlado automaticamente pelo cartão vinculado")

	ErrNotCardLinked = errors.New("saída não está vinculada a nenhum cartão")

	ErrAmountManagedByAccount = errors.New("valor desta saída é controlado automaticamente pela conta vinculada")

	ErrNotAccountLinked = errors.New("saída não está vinculada a nenhuma conta")
)
