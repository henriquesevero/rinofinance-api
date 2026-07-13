package account

import "errors"

// ErrInvalidPurchaseDate indicates a debit purchase was created or updated
// without a valid date.
var ErrInvalidPurchaseDate = errors.New("data da compra inválida")
