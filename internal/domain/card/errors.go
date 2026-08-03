package card

import "errors"

var (
	ErrInvalidInstallmentCount = errors.New("quantidade de parcelas deve ser maior que zero")

	ErrInvalidFirstInstallmentDate = errors.New("data da primeira parcela é obrigatória")
)
