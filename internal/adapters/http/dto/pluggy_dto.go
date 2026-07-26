package dto

import apppluggy "rinofinance-api/internal/application/pluggy"

// PluggySyncRequest is the payload for POST /api/pluggy/sync: the Pluggy
// connection (item) id to import into the user's accounts.
type PluggySyncRequest struct {
	ItemID string `json:"itemId"`
}

// PluggySyncResponse summarizes what a sync produced.
type PluggySyncResponse struct {
	AccountsSynced       int `json:"accountsSynced"`
	TransactionsImported int `json:"transactionsImported"`
	TransactionsSkipped  int `json:"transactionsSkipped"`
}

// NewPluggySyncResponse builds the response from the use case result.
func NewPluggySyncResponse(r apppluggy.SyncResult) PluggySyncResponse {
	return PluggySyncResponse{
		AccountsSynced:       r.AccountsSynced,
		TransactionsImported: r.TransactionsImported,
		TransactionsSkipped:  r.TransactionsSkipped,
	}
}
