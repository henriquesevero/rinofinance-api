package income

import "errors"

var (
	ErrAmountManagedByAccount = errors.New("valor desta entrada é controlado automaticamente pela conta vinculada")

	ErrNotAccountLinked = errors.New("entrada não está vinculada a nenhuma conta")
)
