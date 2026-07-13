package dto

import (
	"time"

	"github.com/google/uuid"

	appcard "rinofinance-api/internal/application/card"
	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// DateOnlyLayout is the wire format used for date-only fields
// (FirstInstallmentDate), matching what an HTML <input type="date"> sends.
const DateOnlyLayout = "2006-01-02"

// CardRequest is the payload for creating/updating a credit card.
type CardRequest struct {
	Name        string       `json:"name"`
	Color       string       `json:"color"`
	LogoURL     string       `json:"logoUrl"`
	ImageURL    string       `json:"imageUrl"`
	CreditLimit shared.Money `json:"creditLimit"`
	DueDay      int          `json:"dueDay"`
}

// Details maps the request's optional visual/financial fields onto the
// application layer's CardDetails, shared by create and update.
func (r CardRequest) Details() appcard.CardDetails {
	return appcard.CardDetails{
		Color:       r.Color,
		LogoURL:     r.LogoURL,
		ImageURL:    r.ImageURL,
		CreditLimit: r.CreditLimit,
		DueDay:      r.DueDay,
	}
}

// InstallmentPurchaseRequest is the payload for creating/updating an
// installment purchase.
type InstallmentPurchaseRequest struct {
	Name                 string       `json:"name"`
	InstallmentAmount    shared.Money `json:"installmentAmount"`
	TotalInstallments    int          `json:"totalInstallments"`
	FirstInstallmentDate string       `json:"firstInstallmentDate"`
	// Domain is the merchant's website (e.g. "netflix.com"), used by the
	// frontend to fetch a brand logo via logo.dev.
	Domain     string `json:"domain"`
	CategoryID string `json:"categoryId"`
}

// ParseFirstInstallmentDate parses the request's date-only string.
func (r InstallmentPurchaseRequest) ParseFirstInstallmentDate() (time.Time, error) {
	return time.Parse(DateOnlyLayout, r.FirstInstallmentDate)
}

// SubscriptionRequest is the payload for creating/updating a subscription.
type SubscriptionRequest struct {
	Name          string       `json:"name"`
	MonthlyAmount shared.Money `json:"monthlyAmount"`
	// Domain is the service's website (e.g. "netflix.com"), used by the
	// frontend to fetch a brand logo via logo.dev.
	Domain     string `json:"domain"`
	CategoryID string `json:"categoryId"`
}

// ImportFaturaRequest is the payload for POST /api/cards/{cardId}/import:
// the frontend parses the PDF statement client-side and sends the already
// classified items (installment purchases and subscriptions).
type ImportFaturaRequest struct {
	InstallmentPurchases []ImportInstallmentItem  `json:"installmentPurchases"`
	Subscriptions        []ImportSubscriptionItem `json:"subscriptions"`
}

// ImportInstallmentItem is one installment purchase parsed from a
// statement (an "avulsa" one-off is just totalInstallments = 1).
type ImportInstallmentItem struct {
	Name                 string       `json:"name"`
	InstallmentAmount    shared.Money `json:"installmentAmount"`
	TotalInstallments    int          `json:"totalInstallments"`
	FirstInstallmentDate string       `json:"firstInstallmentDate"`
	Domain               string       `json:"domain"`
}

// ImportSubscriptionItem is one subscription parsed from a statement.
type ImportSubscriptionItem struct {
	Name          string       `json:"name"`
	MonthlyAmount shared.Money `json:"monthlyAmount"`
	Domain        string       `json:"domain"`
}

// ImportFaturaResponse reports how many items were created.
type ImportFaturaResponse struct {
	InstallmentPurchases int `json:"installmentPurchases"`
	Subscriptions        int `json:"subscriptions"`
}

// ClearCardRequest is the payload for POST /api/cards/{cardId}/clear: the
// IDs of the installment purchases and subscriptions the user chose to
// remove.
type ClearCardRequest struct {
	InstallmentPurchaseIDs []string `json:"installmentPurchaseIds"`
	SubscriptionIDs        []string `json:"subscriptionIds"`
}

// ClearCardResponse reports how many items were removed.
type ClearCardResponse struct {
	InstallmentPurchases int `json:"installmentPurchases"`
	Subscriptions        int `json:"subscriptions"`
}

// ReorderCardsRequest is the payload for PUT /api/cards/order: the card
// IDs in their new display order.
type ReorderCardsRequest struct {
	CardIDs []string `json:"cardIds"`
}

// CardResponse is the public representation of a CreditCard on its own
// (without items), used for create/update/delete responses.
type CardResponse struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Color       string       `json:"color,omitempty"`
	LogoURL     string       `json:"logoUrl,omitempty"`
	ImageURL    string       `json:"imageUrl,omitempty"`
	CreditLimit shared.Money `json:"creditLimit"`
	DueDay      int          `json:"dueDay,omitempty"`
}

// NewCardResponse builds a CardResponse from the domain CreditCard.
func NewCardResponse(c *domaincard.CreditCard) CardResponse {
	return CardResponse{
		ID:          c.ID,
		Name:        c.Name,
		Color:       c.Color,
		LogoURL:     c.LogoURL,
		ImageURL:    c.ImageURL,
		CreditLimit: c.CreditLimit,
		DueDay:      c.DueDay,
	}
}

// InstallmentPurchaseResponse is the public representation of an
// InstallmentPurchase, including the fields computed relative to a
// reference month (remaining installments and remaining total).
type InstallmentPurchaseResponse struct {
	ID                    uuid.UUID    `json:"id"`
	Name                  string       `json:"name"`
	InstallmentAmount     shared.Money `json:"installmentAmount"`
	TotalInstallments     int          `json:"totalInstallments"`
	FirstInstallmentDate  string       `json:"firstInstallmentDate"`
	RemainingInstallments int          `json:"remainingInstallments"`
	RemainingTotal        shared.Money `json:"remainingTotal"`
	Domain                string       `json:"domain,omitempty"`
	Flagged               bool         `json:"flagged"`
	CategoryID            *uuid.UUID   `json:"categoryId,omitempty"`
}

// NewInstallmentPurchaseResponse builds an InstallmentPurchaseResponse,
// computing remaining installments/total as of reference.
func NewInstallmentPurchaseResponse(p *domaincard.InstallmentPurchase, reference time.Time) InstallmentPurchaseResponse {
	return InstallmentPurchaseResponse{
		ID:                    p.ID,
		Name:                  p.Name,
		InstallmentAmount:     p.InstallmentAmount,
		TotalInstallments:     p.TotalInstallments,
		FirstInstallmentDate:  p.FirstInstallmentDate.Format(DateOnlyLayout),
		RemainingInstallments: p.RemainingInstallments(reference),
		RemainingTotal:        p.RemainingTotal(reference),
		Domain:                p.Domain,
		Flagged:               p.Flagged,
		CategoryID:            p.CategoryID,
	}
}

// SubscriptionResponse is the public representation of a Subscription.
type SubscriptionResponse struct {
	ID            uuid.UUID    `json:"id"`
	Name          string       `json:"name"`
	MonthlyAmount shared.Money `json:"monthlyAmount"`
	Domain        string       `json:"domain,omitempty"`
	CategoryID    *uuid.UUID   `json:"categoryId,omitempty"`
}

// NewSubscriptionResponse builds a SubscriptionResponse from the domain
// Subscription.
func NewSubscriptionResponse(s *domaincard.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:            s.ID,
		Name:          s.Name,
		MonthlyAmount: s.MonthlyAmount,
		Domain:        s.Domain,
		CategoryID:    s.CategoryID,
	}
}

// CardOverviewResponse is one card's full Aba 2 section: its items and
// current-month total.
type CardOverviewResponse struct {
	ID                   uuid.UUID                     `json:"id"`
	Name                 string                        `json:"name"`
	Color                string                        `json:"color,omitempty"`
	LogoURL              string                        `json:"logoUrl,omitempty"`
	ImageURL             string                        `json:"imageUrl,omitempty"`
	CreditLimit          shared.Money                  `json:"creditLimit"`
	DueDay               int                           `json:"dueDay,omitempty"`
	InstallmentPurchases []InstallmentPurchaseResponse `json:"installmentPurchases"`
	Subscriptions        []SubscriptionResponse        `json:"subscriptions"`
	MonthlyTotal         shared.Money                  `json:"monthlyTotal"`
}

// CardsOverviewResponse is the full payload for GET /api/cards: every
// card's overview plus the combined "Total Geral".
type CardsOverviewResponse struct {
	Cards      []CardOverviewResponse `json:"cards"`
	GrandTotal shared.Money           `json:"grandTotal"`
}

// NewCardsOverviewResponse builds the full Aba 2 payload from the
// application layer's CardOverview slice and grand total.
func NewCardsOverviewResponse(overviews []appcard.CardOverview, grandTotal shared.Money, reference time.Time) CardsOverviewResponse {
	cards := make([]CardOverviewResponse, 0, len(overviews))
	for _, o := range overviews {
		purchases := make([]InstallmentPurchaseResponse, 0, len(o.InstallmentPurchases))
		for _, p := range o.InstallmentPurchases {
			purchases = append(purchases, NewInstallmentPurchaseResponse(p, reference))
		}
		subscriptions := make([]SubscriptionResponse, 0, len(o.Subscriptions))
		for _, s := range o.Subscriptions {
			subscriptions = append(subscriptions, NewSubscriptionResponse(s))
		}
		cards = append(cards, CardOverviewResponse{
			ID:                   o.Card.ID,
			Name:                 o.Card.Name,
			Color:                o.Card.Color,
			LogoURL:              o.Card.LogoURL,
			ImageURL:             o.Card.ImageURL,
			CreditLimit:          o.Card.CreditLimit,
			DueDay:               o.Card.DueDay,
			InstallmentPurchases: purchases,
			Subscriptions:        subscriptions,
			MonthlyTotal:         o.MonthlyTotal,
		})
	}
	return CardsOverviewResponse{Cards: cards, GrandTotal: grandTotal}
}
