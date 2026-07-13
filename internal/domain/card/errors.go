package card

import "errors"

var (
	// ErrInvalidInstallmentCount indicates TotalInstallments was less than 1.
	ErrInvalidInstallmentCount = errors.New("quantidade de parcelas deve ser maior que zero")

	// ErrInvalidFirstInstallmentDate indicates FirstInstallmentDate was the
	// zero value.
	ErrInvalidFirstInstallmentDate = errors.New("data da primeira parcela é obrigatória")
)
