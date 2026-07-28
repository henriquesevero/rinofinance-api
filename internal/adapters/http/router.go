// Package rest wires the HTTP routing table: it is the outermost layer of
// the primary (driving) adapter, composing the handler and middleware
// packages into a single http.Handler for cmd/api/main.go to serve.
package rest

import (
	"net/http"

	"github.com/google/uuid"

	"rinofinance-api/internal/adapters/http/handler"
	"rinofinance-api/internal/adapters/http/middleware"
	"rinofinance-api/internal/pkg/auth"
)

// Handlers bundles every resource handler the router needs. main.go
// constructs one of these after wiring all use cases and repositories.
type Handlers struct {
	Auth       *handler.AuthHandler
	Account    *handler.AccountHandler
	Income     *handler.IncomeHandler
	Expense    *handler.ExpenseHandler
	Card       *handler.CardHandler
	Investment *handler.InvestmentHandler
	Category   *handler.CategoryHandler
	Wishlist   *handler.WishlistHandler
	Wallet     *handler.WalletHandler
	Dashboard  *handler.DashboardHandler
}

// NewRouter builds the full HTTP routing table, applying CORS to every
// route and JWT authentication to every route except /health and
// /api/auth/*.
func NewRouter(h Handlers, tokens *auth.TokenIssuer, resolveOwner func(uuid.UUID) uuid.UUID, allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"rinofinance-api"}`))
	})

	mux.HandleFunc("POST /api/auth/register", h.Auth.Register)
	mux.HandleFunc("POST /api/auth/login", h.Auth.Login)

	protected := http.NewServeMux()

	protected.HandleFunc("GET /api/dashboard/summary", h.Dashboard.GetSummary)

	protected.HandleFunc("PUT /api/account/profile", h.Account.UpdateProfile)
	protected.HandleFunc("PUT /api/account/email", h.Account.ChangeEmail)
	protected.HandleFunc("PUT /api/account/password", h.Account.ChangePassword)
	protected.HandleFunc("DELETE /api/account", h.Account.DeleteAccount)
	protected.HandleFunc("POST /api/account/share", h.Account.ShareData)
	protected.HandleFunc("POST /api/account/unshare", h.Account.StopSharing)

	protected.HandleFunc("GET /api/incomes", h.Income.List)
	protected.HandleFunc("POST /api/incomes", h.Income.Create)
	protected.HandleFunc("POST /api/incomes/account-linked", h.Income.CreateAccountLinked)
	protected.HandleFunc("PUT /api/incomes/order", h.Income.Reorder)
	protected.HandleFunc("PUT /api/incomes/{id}", h.Income.Update)
	protected.HandleFunc("PATCH /api/incomes/{id}/toggle", h.Income.Toggle)
	protected.HandleFunc("PATCH /api/incomes/{id}/received", h.Income.ToggleReceived)
	protected.HandleFunc("DELETE /api/incomes/{id}", h.Income.Delete)

	protected.HandleFunc("GET /api/expenses", h.Expense.List)
	protected.HandleFunc("POST /api/expenses", h.Expense.Create)
	protected.HandleFunc("POST /api/expenses/card-linked", h.Expense.CreateCardLinked)
	protected.HandleFunc("POST /api/expenses/account-linked", h.Expense.CreateAccountLinked)
	protected.HandleFunc("PUT /api/expenses/order", h.Expense.Reorder)
	protected.HandleFunc("PUT /api/expenses/{id}", h.Expense.Update)
	protected.HandleFunc("PATCH /api/expenses/{id}/toggle", h.Expense.Toggle)
	protected.HandleFunc("PATCH /api/expenses/{id}/paid", h.Expense.TogglePaid)
	protected.HandleFunc("DELETE /api/expenses/{id}", h.Expense.Delete)

	protected.HandleFunc("GET /api/cards", h.Card.ListCards)
	protected.HandleFunc("POST /api/cards", h.Card.CreateCard)
	protected.HandleFunc("PUT /api/cards/order", h.Card.ReorderCards)
	protected.HandleFunc("PUT /api/cards/{id}", h.Card.UpdateCard)
	protected.HandleFunc("DELETE /api/cards/{id}", h.Card.DeleteCard)
	protected.HandleFunc("POST /api/cards/{cardId}/import", h.Card.ImportFatura)
	protected.HandleFunc("POST /api/cards/{cardId}/clear", h.Card.ClearCard)
	protected.HandleFunc("POST /api/cards/{cardId}/installment-purchases", h.Card.CreateInstallmentPurchase)
	protected.HandleFunc("PUT /api/cards/{cardId}/installment-purchases/order", h.Card.ReorderInstallmentPurchases)
	protected.HandleFunc("PUT /api/installment-purchases/{id}", h.Card.UpdateInstallmentPurchase)
	protected.HandleFunc("PATCH /api/installment-purchases/{id}/flag", h.Card.ToggleInstallmentPurchaseFlag)
	protected.HandleFunc("PATCH /api/installment-purchases/{id}/owed-exclusion", h.Card.ToggleInstallmentPurchaseOwedExclusion)
	protected.HandleFunc("DELETE /api/installment-purchases/{id}", h.Card.DeleteInstallmentPurchase)
	protected.HandleFunc("POST /api/cards/{cardId}/subscriptions", h.Card.CreateSubscription)
	protected.HandleFunc("PUT /api/cards/{cardId}/subscriptions/order", h.Card.ReorderSubscriptions)
	protected.HandleFunc("PUT /api/subscriptions/{id}", h.Card.UpdateSubscription)
	protected.HandleFunc("DELETE /api/subscriptions/{id}", h.Card.DeleteSubscription)

	protected.HandleFunc("GET /api/investments", h.Investment.List)
	protected.HandleFunc("POST /api/investments", h.Investment.Create)
	protected.HandleFunc("PUT /api/investments/{id}", h.Investment.Update)
	protected.HandleFunc("PATCH /api/investments/{id}/toggle", h.Investment.Toggle)
	protected.HandleFunc("DELETE /api/investments/{id}", h.Investment.Delete)

	protected.HandleFunc("GET /api/categories", h.Category.List)
	protected.HandleFunc("POST /api/categories", h.Category.Create)
	protected.HandleFunc("PUT /api/categories/order", h.Category.Reorder)
	protected.HandleFunc("PUT /api/categories/{id}", h.Category.Update)
	protected.HandleFunc("DELETE /api/categories/{id}", h.Category.Delete)

	protected.HandleFunc("GET /api/wishlist", h.Wishlist.Overview)
	protected.HandleFunc("GET /api/wishlist/unfurl", h.Wishlist.Unfurl)
	protected.HandleFunc("POST /api/wishlist/sections", h.Wishlist.CreateSection)
	protected.HandleFunc("PUT /api/wishlist/sections/{id}", h.Wishlist.UpdateSection)
	protected.HandleFunc("DELETE /api/wishlist/sections/{id}", h.Wishlist.DeleteSection)
	protected.HandleFunc("POST /api/wishlist/items", h.Wishlist.CreateItem)
	protected.HandleFunc("PUT /api/wishlist/items/{id}", h.Wishlist.UpdateItem)
	protected.HandleFunc("DELETE /api/wishlist/items/{id}", h.Wishlist.DeleteItem)

	protected.HandleFunc("GET /api/accounts", h.Wallet.List)
	protected.HandleFunc("POST /api/accounts", h.Wallet.Create)
	protected.HandleFunc("PUT /api/accounts/order", h.Wallet.Reorder)
	protected.HandleFunc("PUT /api/accounts/{id}", h.Wallet.Update)
	protected.HandleFunc("DELETE /api/accounts/{id}", h.Wallet.Delete)
	protected.HandleFunc("POST /api/accounts/{accountId}/purchases", h.Wallet.CreatePurchase)
	protected.HandleFunc("PUT /api/account-purchases/{id}", h.Wallet.UpdatePurchase)
	protected.HandleFunc("DELETE /api/account-purchases/{id}", h.Wallet.DeletePurchase)

	mux.Handle("/api/", middleware.Auth(tokens, resolveOwner)(protected))

	return middleware.CORS(allowedOrigins)(mux)
}
