package card

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

func monthIndex(t time.Time) int {
	return t.Year()*12 + int(t.Month())
}

func firstOfMonthUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

type CreditCard struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string

	Color string

	LogoURL string

	ImageURL string

	CreditLimit shared.Money

	DueDay int

	ClosingDay int

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func (c *CreditCard) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	c.Name = name
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *CreditCard) SetColor(color string) {
	c.Color = color
	c.UpdatedAt = time.Now().UTC()
}

func (c *CreditCard) SetLogo(logoURL string) {
	c.LogoURL = logoURL
	c.UpdatedAt = time.Now().UTC()
}

func (c *CreditCard) SetImage(imageURL string) {
	c.ImageURL = imageURL
	c.UpdatedAt = time.Now().UTC()
}

func (c *CreditCard) SetCreditLimit(limit shared.Money) {
	c.CreditLimit = limit
	c.UpdatedAt = time.Now().UTC()
}

func (c *CreditCard) SetDueDay(day int) {
	if day < 1 || day > 31 {
		day = 0
	}
	c.DueDay = day
	c.UpdatedAt = time.Now().UTC()
}

func (c *CreditCard) SetClosingDay(day int) {
	if day < 1 || day > 31 {
		day = 0
	}
	c.ClosingDay = day
	c.UpdatedAt = time.Now().UTC()
}

func (c *CreditCard) SetPosition(position int) {
	c.Position = position
	c.UpdatedAt = time.Now().UTC()
}

type InstallmentPurchase struct {
	ID                   uuid.UUID
	CardID               uuid.UUID
	Name                 string
	InstallmentAmount    shared.Money
	TotalInstallments    int
	FirstInstallmentDate time.Time

	Domain string

	Flagged bool

	ExcludedFromOwed bool

	CategoryID *uuid.UUID

	CanceledFrom *time.Time

	EffectiveFrom *time.Time

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func (p *InstallmentPurchase) monthsElapsed(reference time.Time) int {
	first := p.FirstInstallmentDate
	return (reference.Year()-first.Year())*12 + int(reference.Month()) - int(first.Month())
}

func (p *InstallmentPurchase) canceledBy(reference time.Time) bool {
	return p.CanceledFrom != nil && monthIndex(reference) >= monthIndex(*p.CanceledFrom)
}

func (p *InstallmentPurchase) EndFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	p.CanceledFrom = &m
	p.UpdatedAt = time.Now().UTC()
}

func (p *InstallmentPurchase) startedBy(reference time.Time) bool {
	return p.EffectiveFrom == nil || monthIndex(reference) >= monthIndex(*p.EffectiveFrom)
}

func (p *InstallmentPurchase) SetEffectiveFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	p.EffectiveFrom = &m
	p.UpdatedAt = time.Now().UTC()
}

func (p *InstallmentPurchase) IsActiveOn(reference time.Time) bool {
	elapsed := p.monthsElapsed(reference)
	return elapsed >= 0 && elapsed < p.TotalInstallments && p.startedBy(reference) && !p.canceledBy(reference)
}

func (p *InstallmentPurchase) RemainingInstallments(reference time.Time) int {
	if p.canceledBy(reference) {
		return 0
	}
	elapsed := p.monthsElapsed(reference)
	remaining := p.TotalInstallments - elapsed
	if remaining > p.TotalInstallments {
		remaining = p.TotalInstallments
	}

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

func (p *InstallmentPurchase) RemainingTotal(reference time.Time) shared.Money {
	return p.InstallmentAmount.MulInt(p.RemainingInstallments(reference))
}

func (p *InstallmentPurchase) MonthlyChargeAmount(reference time.Time) shared.Money {
	if !p.IsActiveOn(reference) {
		return shared.Zero
	}
	return p.InstallmentAmount
}

func (p *InstallmentPurchase) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	p.Name = name
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *InstallmentPurchase) SetDomain(domain string) {
	p.Domain = domain
	p.UpdatedAt = time.Now().UTC()
}

func (p *InstallmentPurchase) SetFlagged(flagged bool) {
	p.Flagged = flagged
	p.UpdatedAt = time.Now().UTC()
}

func (p *InstallmentPurchase) SetExcludedFromOwed(excluded bool) {
	p.ExcludedFromOwed = excluded
	p.UpdatedAt = time.Now().UTC()
}

func (p *InstallmentPurchase) SetCategory(categoryID *uuid.UUID) {
	p.CategoryID = categoryID
	p.UpdatedAt = time.Now().UTC()
}

func (p *InstallmentPurchase) SetPosition(position int) {
	p.Position = position
	p.UpdatedAt = time.Now().UTC()
}

type Subscription struct {
	ID            uuid.UUID
	CardID        uuid.UUID
	Name          string
	MonthlyAmount shared.Money

	Domain string

	CategoryID *uuid.UUID

	CanceledFrom *time.Time

	EffectiveFrom *time.Time

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func (s *Subscription) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	s.Name = name
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Subscription) SetDomain(domain string) {
	s.Domain = domain
	s.UpdatedAt = time.Now().UTC()
}

func (s *Subscription) SetCategory(categoryID *uuid.UUID) {
	s.CategoryID = categoryID
	s.UpdatedAt = time.Now().UTC()
}

func (s *Subscription) SetPosition(position int) {
	s.Position = position
	s.UpdatedAt = time.Now().UTC()
}

func (s *Subscription) UpdateMonthlyAmount(amount shared.Money) error {
	if amount.IsNegative() {
		return shared.ErrNegativeAmount
	}
	s.MonthlyAmount = amount
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Subscription) EndFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	s.CanceledFrom = &m
	s.UpdatedAt = time.Now().UTC()
}

func (s *Subscription) SetEffectiveFrom(month time.Time) {
	m := firstOfMonthUTC(month)
	s.EffectiveFrom = &m
	s.UpdatedAt = time.Now().UTC()
}

func (s *Subscription) IsActiveOn(reference time.Time) bool {
	started := s.EffectiveFrom == nil || monthIndex(reference) >= monthIndex(*s.EffectiveFrom)
	notCanceled := s.CanceledFrom == nil || monthIndex(reference) < monthIndex(*s.CanceledFrom)
	return started && notCanceled
}

func (s *Subscription) MonthlyChargeAmount(reference time.Time) shared.Money {
	if !s.IsActiveOn(reference) {
		return shared.Zero
	}
	return s.MonthlyAmount
}

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
