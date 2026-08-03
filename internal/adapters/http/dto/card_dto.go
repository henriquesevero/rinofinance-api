package dto

import (
	"time"

	"github.com/google/uuid"

	appcard "rinofinance-api/internal/application/card"
	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

const DateOnlyLayout = "2006-01-02"

type CardRequest struct {
	Name        string       `json:"name"`
	Color       string       `json:"color"`
	LogoURL     string       `json:"logoUrl"`
	ImageURL    string       `json:"imageUrl"`
	CreditLimit shared.Money `json:"creditLimit"`
	DueDay      int          `json:"dueDay"`
	ClosingDay  int          `json:"closingDay"`
}

func (r CardRequest) Details() appcard.CardDetails {
	return appcard.CardDetails{
		Color:       r.Color,
		LogoURL:     r.LogoURL,
		ImageURL:    r.ImageURL,
		CreditLimit: r.CreditLimit,
		DueDay:      r.DueDay,
		ClosingDay:  r.ClosingDay,
	}
}

type InstallmentPurchaseRequest struct {
	Name                 string       `json:"name"`
	InstallmentAmount    shared.Money `json:"installmentAmount"`
	TotalInstallments    int          `json:"totalInstallments"`
	FirstInstallmentDate string       `json:"firstInstallmentDate"`

	Domain     string `json:"domain"`
	CategoryID string `json:"categoryId"`
}

func (r InstallmentPurchaseRequest) ParseFirstInstallmentDate() (time.Time, error) {
	return time.Parse(DateOnlyLayout, r.FirstInstallmentDate)
}

type SubscriptionRequest struct {
	Name          string       `json:"name"`
	MonthlyAmount shared.Money `json:"monthlyAmount"`

	Domain     string `json:"domain"`
	CategoryID string `json:"categoryId"`
}

type ImportFaturaRequest struct {
	InstallmentPurchases []ImportInstallmentItem  `json:"installmentPurchases"`
	Subscriptions        []ImportSubscriptionItem `json:"subscriptions"`

	ReferenceMonth string `json:"referenceMonth"`
}

type ImportInstallmentItem struct {
	Name                 string       `json:"name"`
	InstallmentAmount    shared.Money `json:"installmentAmount"`
	TotalInstallments    int          `json:"totalInstallments"`
	FirstInstallmentDate string       `json:"firstInstallmentDate"`
	Domain               string       `json:"domain"`
	CategoryID           string       `json:"categoryId"`
}

type ImportSubscriptionItem struct {
	Name          string       `json:"name"`
	MonthlyAmount shared.Money `json:"monthlyAmount"`
	Domain        string       `json:"domain"`
	CategoryID    string       `json:"categoryId"`
}

type ImportFaturaResponse struct {
	InstallmentPurchases int `json:"installmentPurchases"`
	Subscriptions        int `json:"subscriptions"`
}

type ClearCardRequest struct {
	InstallmentPurchaseIDs []string `json:"installmentPurchaseIds"`
	SubscriptionIDs        []string `json:"subscriptionIds"`

	Mode string `json:"mode"`
}

type ClearCardResponse struct {
	InstallmentPurchases int `json:"installmentPurchases"`
	Subscriptions        int `json:"subscriptions"`
}

type ReorderCardsRequest struct {
	CardIDs []string `json:"cardIds"`
}

type CardResponse struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Color       string       `json:"color,omitempty"`
	LogoURL     string       `json:"logoUrl,omitempty"`
	ImageURL    string       `json:"imageUrl,omitempty"`
	CreditLimit shared.Money `json:"creditLimit"`
	DueDay      int          `json:"dueDay,omitempty"`
	ClosingDay  int          `json:"closingDay,omitempty"`
}

func NewCardResponse(c *domaincard.CreditCard) CardResponse {
	return CardResponse{
		ID:          c.ID,
		Name:        c.Name,
		Color:       c.Color,
		LogoURL:     c.LogoURL,
		ImageURL:    c.ImageURL,
		CreditLimit: c.CreditLimit,
		DueDay:      c.DueDay,
		ClosingDay:  c.ClosingDay,
	}
}

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
	ExcludedFromOwed      bool         `json:"excludedFromOwed"`
	CategoryID            *uuid.UUID   `json:"categoryId,omitempty"`

	CanceledFrom string `json:"canceledFrom,omitempty"`

	EffectiveFrom string `json:"effectiveFrom,omitempty"`
}

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
		ExcludedFromOwed:      p.ExcludedFromOwed,
		CategoryID:            p.CategoryID,
		CanceledFrom:          formatMonthPtr(p.CanceledFrom),
		EffectiveFrom:         formatMonthPtr(p.EffectiveFrom),
	}
}

func formatMonthPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(DateOnlyLayout)
}

type SubscriptionResponse struct {
	ID            uuid.UUID    `json:"id"`
	Name          string       `json:"name"`
	MonthlyAmount shared.Money `json:"monthlyAmount"`
	Domain        string       `json:"domain,omitempty"`
	CategoryID    *uuid.UUID   `json:"categoryId,omitempty"`
	CanceledFrom  string       `json:"canceledFrom,omitempty"`
	EffectiveFrom string       `json:"effectiveFrom,omitempty"`
}

func NewSubscriptionResponse(s *domaincard.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:            s.ID,
		Name:          s.Name,
		MonthlyAmount: s.MonthlyAmount,
		Domain:        s.Domain,
		CategoryID:    s.CategoryID,
		CanceledFrom:  formatMonthPtr(s.CanceledFrom),
		EffectiveFrom: formatMonthPtr(s.EffectiveFrom),
	}
}

type CardOverviewResponse struct {
	ID                   uuid.UUID                     `json:"id"`
	Name                 string                        `json:"name"`
	Color                string                        `json:"color,omitempty"`
	LogoURL              string                        `json:"logoUrl,omitempty"`
	ImageURL             string                        `json:"imageUrl,omitempty"`
	CreditLimit          shared.Money                  `json:"creditLimit"`
	DueDay               int                           `json:"dueDay,omitempty"`
	ClosingDay           int                           `json:"closingDay,omitempty"`
	InstallmentPurchases []InstallmentPurchaseResponse `json:"installmentPurchases"`
	Subscriptions        []SubscriptionResponse        `json:"subscriptions"`
	MonthlyTotal         shared.Money                  `json:"monthlyTotal"`
}

type CardsOverviewResponse struct {
	Cards      []CardOverviewResponse `json:"cards"`
	GrandTotal shared.Money           `json:"grandTotal"`
}

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
			ClosingDay:           o.Card.ClosingDay,
			InstallmentPurchases: purchases,
			Subscriptions:        subscriptions,
			MonthlyTotal:         o.MonthlyTotal,
		})
	}
	return CardsOverviewResponse{Cards: cards, GrandTotal: grandTotal}
}
