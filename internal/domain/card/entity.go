// Package card models Aba 2 (Controle de Cartões): credit cards, their
// installment purchases and monthly subscriptions. Remaining installments
// are derived from FirstInstallmentDate + TotalInstallments compared
// against a reference month, never stored as a manually decremented
// counter — this avoids drift if a month is skipped or the totals are
// recomputed after the fact.
package card

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// monthIndex collapses a date to a comparable whole-month ordinal, so two
// dates can be compared purely by calendar month (year*12 + month).
func monthIndex(t time.Time) int {
	return t.Year()*12 + int(t.Month())
}

// firstOfMonthUTC normalizes a date to midnight on the first day of its month
// (UTC) — the canonical form stored for a "canceled from" marker.
func firstOfMonthUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// CreditCard is the aggregate root grouping installment purchases and
// subscriptions billed together (e.g. "Nubank", "Inter").
type CreditCard struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	// Color is a user-chosen hex color (e.g. "#8A05BE") used as the UI
	// accent for this card's block, so a card can visually match its
	// real-world brand (Nubank purple, Inter orange, etc.) without the
	// app needing to know about specific banks.
	Color string
	// LogoURL optionally holds a small logo image as a data: URL (mirrors
	// User.AvatarURL — resized client-side, no object storage needed for
	// this version). Empty until the user uploads one, in which case the
	// UI falls back to a colored initial badge.
	LogoURL string
	// ImageURL optionally holds the full "card art" image as a data: URL,
	// shown as the card's visual in the overview grid. Empty falls back to
	// a color gradient with the card's name.
	ImageURL string
	// CreditLimit is the card's total credit limit. Zero means "not
	// configured", which hides the usage bar in the UI.
	CreditLimit shared.Money
	// DueDay is the invoice due day of month (1–31). Zero means unset,
	// hiding the "vence em X dias" countdown in the UI.
	DueDay int
	// ClosingDay is the invoice closing day of month (1–31) — when the
	// current bill closes and a new one starts accumulating. Zero means
	// unset. Combined with DueDay it drives the "melhor dia de compra"
	// (best purchase day = the day after the bill closes).
	ClosingDay int
	// Position is the card's manual sort order in the overview grid (lower
	// comes first). New cards are appended to the end; the user can drag to
	// reorder.
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewCreditCard builds a new CreditCard for a user.
func NewCreditCard(userID uuid.UUID, name string) (*CreditCard, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}

	now := time.Now().UTC()
	return &CreditCard{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Rename updates the credit card's display name.
func (c *CreditCard) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	c.Name = name
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// SetColor replaces the card's UI accent color. Passing an empty string
// clears it, falling back to a neutral default in the UI.
func (c *CreditCard) SetColor(color string) {
	c.Color = color
	c.UpdatedAt = time.Now().UTC()
}

// SetLogo replaces the card's logo image. Passing an empty string clears
// it, falling back to initials in the UI.
func (c *CreditCard) SetLogo(logoURL string) {
	c.LogoURL = logoURL
	c.UpdatedAt = time.Now().UTC()
}

// SetImage replaces the card's full "card art" image (a data: URL) shown
// as the card visual in the overview. Passing an empty string clears it,
// falling back to a color gradient with the card's name.
func (c *CreditCard) SetImage(imageURL string) {
	c.ImageURL = imageURL
	c.UpdatedAt = time.Now().UTC()
}

// SetCreditLimit sets the card's total credit limit. Zero means "not
// configured", hiding the usage bar in the UI.
func (c *CreditCard) SetCreditLimit(limit shared.Money) {
	c.CreditLimit = limit
	c.UpdatedAt = time.Now().UTC()
}

// SetDueDay sets the invoice due day of month (1–31). Any value outside
// that range is treated as unset (0), hiding the countdown in the UI.
func (c *CreditCard) SetDueDay(day int) {
	if day < 1 || day > 31 {
		day = 0
	}
	c.DueDay = day
	c.UpdatedAt = time.Now().UTC()
}

// SetClosingDay sets the invoice closing day of month (1–31). Any value
// outside that range is treated as unset (0).
func (c *CreditCard) SetClosingDay(day int) {
	if day < 1 || day > 31 {
		day = 0
	}
	c.ClosingDay = day
	c.UpdatedAt = time.Now().UTC()
}

// SetPosition sets the card's manual sort order in the overview grid.
func (c *CreditCard) SetPosition(position int) {
	c.Position = position
	c.UpdatedAt = time.Now().UTC()
}

// InstallmentPurchase is a purchase split into fixed monthly installments
// (e.g. "Notebook 10x"). RemainingInstallments and RemainingTotal are
// always computed relative to a reference month rather than stored, per
// the product decision to avoid manual month-to-month decrementing.
type InstallmentPurchase struct {
	ID                   uuid.UUID
	CardID               uuid.UUID
	Name                 string
	InstallmentAmount    shared.Money
	TotalInstallments    int
	FirstInstallmentDate time.Time
	// Domain optionally holds the merchant's website (e.g. "netflix.com"),
	// used by the frontend to fetch a brand logo via logo.dev instead of
	// requiring a manual image upload for every purchase.
	Domain string
	// Flagged marks a purchase the user wants to keep an eye on; the UI
	// paints the whole row red. It's a pure status marker, independent of
	// any calculation.
	Flagged bool
	// ExcludedFromOwed lets the user drop this purchase from the "total que
	// devo" (payoff) sum shown on the card — e.g. a shared/reimbursed
	// purchase they don't consider their own debt. It does not affect the
	// monthly bill, just the debt total the UI computes.
	ExcludedFromOwed bool
	// CategoryID optionally links this purchase to a user category. Nil
	// means uncategorized.
	CategoryID *uuid.UUID
	// CanceledFrom, when set, is the first month (normalized to its first
	// day) in which this purchase is no longer billed — used by "limpar
	// fatura" / the row's "excluir" to stop an item from the current month
	// onward WITHOUT rewriting the past: months before it keep billing it.
	// Nil means it follows its natural installment schedule.
	CanceledFrom *time.Time
	// EffectiveFrom, when set, is the first month this purchase becomes
	// visible — even if its installment schedule reaches further back. Set on
	// import to the fatura's month so a mid-way parcela (e.g. 2/3) shows only
	// from the imported bill onward, never populating earlier months it was
	// never tracked in. Nil = visible per its natural schedule.
	EffectiveFrom *time.Time
	// Position is the manual sort order within the card's list of purchases.
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewInstallmentPurchase builds a new installment purchase under a card.
func NewInstallmentPurchase(
	cardID uuid.UUID,
	name string,
	installmentAmount shared.Money,
	totalInstallments int,
	firstInstallmentDate time.Time,
) (*InstallmentPurchase, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if installmentAmount.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}
	if totalInstallments < 1 {
		return nil, ErrInvalidInstallmentCount
	}
	if firstInstallmentDate.IsZero() {
		return nil, ErrInvalidFirstInstallmentDate
	}

	now := time.Now().UTC()
	return &InstallmentPurchase{
		ID:                   uuid.New(),
		CardID:               cardID,
		Name:                 name,
		InstallmentAmount:    installmentAmount,
		TotalInstallments:    totalInstallments,
		FirstInstallmentDate: firstInstallmentDate,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// monthsElapsed returns how many whole calendar months separate the first
// installment from the reference date. It can be negative (purchase starts
// in a future month) or exceed TotalInstallments (purchase already paid
// off).
func (p *InstallmentPurchase) monthsElapsed(reference time.Time) int {
	first := p.FirstInstallmentDate
	return (reference.Year()-first.Year())*12 + int(reference.Month()) - int(first.Month())
}

// canceledBy reports whether the reference month is at or after this
// purchase's cancellation month (so it no longer bills).
func (p *InstallmentPurchase) canceledBy(reference time.Time) bool {
	return p.CanceledFrom != nil && monthIndex(reference) >= monthIndex(*p.CanceledFrom)
}

// EndFrom stops this purchase from the given month onward (normalized to the
// first of that month), preserving every earlier month's bill. Passing a
// month at or before it started effectively hides it everywhere.
func (p *InstallmentPurchase) EndFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	p.CanceledFrom = &m
	p.UpdatedAt = time.Now().UTC()
}

// startedBy reports whether the purchase is visible in the reference month —
// its EffectiveFrom lower bound (if any) has been reached.
func (p *InstallmentPurchase) startedBy(reference time.Time) bool {
	return p.EffectiveFrom == nil || monthIndex(reference) >= monthIndex(*p.EffectiveFrom)
}

// SetEffectiveFrom bounds this purchase to be visible only from `month` on.
func (p *InstallmentPurchase) SetEffectiveFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	p.EffectiveFrom = &m
	p.UpdatedAt = time.Now().UTC()
}

// IsActiveOn reports whether this purchase still bills an installment in the
// reference month (visible, started, not paid off, not canceled by then).
func (p *InstallmentPurchase) IsActiveOn(reference time.Time) bool {
	elapsed := p.monthsElapsed(reference)
	return elapsed >= 0 && elapsed < p.TotalInstallments && p.startedBy(reference) && !p.canceledBy(reference)
}

// RemainingInstallments returns how many installments (including the
// reference month's, if active) are still owed as of the reference month —
// capped at the cancellation month when the purchase was ended early.
func (p *InstallmentPurchase) RemainingInstallments(reference time.Time) int {
	if p.canceledBy(reference) {
		return 0
	}
	elapsed := p.monthsElapsed(reference)
	remaining := p.TotalInstallments - elapsed
	if remaining > p.TotalInstallments {
		remaining = p.TotalInstallments
	}
	// An early cancellation stops billing at CanceledFrom: only the months
	// from the reference up to (but excluding) that month still count.
	if p.CanceledFrom != nil {
		untilCancel := monthIndex(*p.CanceledFrom) - monthIndex(reference)
		if untilCancel < remaining {
			remaining = untilCancel
		}
	}
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// RemainingTotal implements the spec's "Total Restante" field: Valor da
// Parcela × Parcelas Restantes.
func (p *InstallmentPurchase) RemainingTotal(reference time.Time) shared.Money {
	return p.InstallmentAmount.MulInt(p.RemainingInstallments(reference))
}

// MonthlyChargeAmount returns the amount this purchase contributes to the
// card's bill for the reference month: the installment amount while
// active, or zero before it starts / after it's paid off.
func (p *InstallmentPurchase) MonthlyChargeAmount(reference time.Time) shared.Money {
	if !p.IsActiveOn(reference) {
		return shared.Zero
	}
	return p.InstallmentAmount
}

// Rename updates the purchase's display name.
func (p *InstallmentPurchase) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	p.Name = name
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// SetDomain replaces the merchant domain used to fetch this purchase's
// brand logo. Passing an empty string clears it, falling back to a
// generic icon in the UI.
func (p *InstallmentPurchase) SetDomain(domain string) {
	p.Domain = domain
	p.UpdatedAt = time.Now().UTC()
}

// SetFlagged marks (or unmarks) this purchase for attention, painting its
// row red in the UI.
func (p *InstallmentPurchase) SetFlagged(flagged bool) {
	p.Flagged = flagged
	p.UpdatedAt = time.Now().UTC()
}

// SetExcludedFromOwed includes (false) or excludes (true) this purchase
// from the card's "total que devo" payoff sum.
func (p *InstallmentPurchase) SetExcludedFromOwed(excluded bool) {
	p.ExcludedFromOwed = excluded
	p.UpdatedAt = time.Now().UTC()
}

// SetCategory links (or clears, when nil) the purchase's category.
func (p *InstallmentPurchase) SetCategory(categoryID *uuid.UUID) {
	p.CategoryID = categoryID
	p.UpdatedAt = time.Now().UTC()
}

// SetPosition sets the manual sort order within the card's purchase list.
func (p *InstallmentPurchase) SetPosition(position int) {
	p.Position = position
	p.UpdatedAt = time.Now().UTC()
}

// Subscription is a fixed monthly charge under a card (e.g. "Spotify").
type Subscription struct {
	ID            uuid.UUID
	CardID        uuid.UUID
	Name          string
	MonthlyAmount shared.Money
	// Domain optionally holds the service's website (e.g. "netflix.com"),
	// used by the frontend to fetch a brand logo via logo.dev.
	Domain string
	// CategoryID optionally links this subscription to a user category. Nil
	// means uncategorized.
	CategoryID *uuid.UUID
	// CanceledFrom, when set, is the first month (normalized to its first
	// day) from which this subscription is no longer billed. Earlier months
	// still show it, so "encerrar" preserves the past. Nil means ongoing.
	CanceledFrom *time.Time
	// EffectiveFrom, when set, is the first month this subscription becomes
	// visible — set on import so it doesn't retroactively appear in months
	// before the imported bill. Nil means active from any month.
	EffectiveFrom *time.Time
	// Position is the manual sort order within the card's subscription list.
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSubscription builds a new subscription under a card.
func NewSubscription(cardID uuid.UUID, name string, monthlyAmount shared.Money) (*Subscription, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if monthlyAmount.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}

	now := time.Now().UTC()
	return &Subscription{
		ID:            uuid.New(),
		CardID:        cardID,
		Name:          name,
		MonthlyAmount: monthlyAmount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Rename updates the subscription's display name.
func (s *Subscription) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	s.Name = name
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// SetDomain replaces the merchant domain used to fetch this
// subscription's brand logo. Passing an empty string clears it, falling
// back to a generic icon in the UI.
func (s *Subscription) SetDomain(domain string) {
	s.Domain = domain
	s.UpdatedAt = time.Now().UTC()
}

// SetCategory links (or clears, when nil) the subscription's category.
func (s *Subscription) SetCategory(categoryID *uuid.UUID) {
	s.CategoryID = categoryID
	s.UpdatedAt = time.Now().UTC()
}

// SetPosition sets the manual sort order within the card's subscription list.
func (s *Subscription) SetPosition(position int) {
	s.Position = position
	s.UpdatedAt = time.Now().UTC()
}

// UpdateMonthlyAmount changes the fixed monthly charge.
func (s *Subscription) UpdateMonthlyAmount(amount shared.Money) error {
	if amount.IsNegative() {
		return shared.ErrNegativeAmount
	}
	s.MonthlyAmount = amount
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// EndFrom stops this subscription from the given month onward (normalized to
// the first of that month), preserving every earlier month's bill.
func (s *Subscription) EndFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	s.CanceledFrom = &m
	s.UpdatedAt = time.Now().UTC()
}

// SetEffectiveFrom bounds this subscription to be visible only from `month` on.
func (s *Subscription) SetEffectiveFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	s.EffectiveFrom = &m
	s.UpdatedAt = time.Now().UTC()
}

// IsActiveOn reports whether this subscription bills in the reference month:
// it has become visible (EffectiveFrom) and hasn't been canceled by then.
func (s *Subscription) IsActiveOn(reference time.Time) bool {
	started := s.EffectiveFrom == nil || monthIndex(reference) >= monthIndex(*s.EffectiveFrom)
	notCanceled := s.CanceledFrom == nil || monthIndex(reference) < monthIndex(*s.CanceledFrom)
	return started && notCanceled
}

// MonthlyChargeAmount returns the amount this subscription contributes to the
// card's bill for the reference month: its fixed amount while active, or zero
// once canceled.
func (s *Subscription) MonthlyChargeAmount(reference time.Time) shared.Money {
	if !s.IsActiveOn(reference) {
		return shared.Zero
	}
	return s.MonthlyAmount
}

// MonthlyTotal sums the current-month charge across a card's installment
// purchases and subscriptions. It is the single source of truth consumed
// by both the "por cartão" and "Total Geral" figures in Aba 2, and by the
// linked Expense.SyncAmountFromCard call in Aba 1.
func MonthlyTotal(reference time.Time, purchases []*InstallmentPurchase, subscriptions []*Subscription) shared.Money {
	total := shared.Zero
	for _, p := range purchases {
		total = total.Add(p.MonthlyChargeAmount(reference))
	}
	for _, s := range subscriptions {
		total = total.Add(s.MonthlyChargeAmount(reference))
	}
	return total
}
