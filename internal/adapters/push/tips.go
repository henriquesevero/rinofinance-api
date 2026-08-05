package push

import (
	"fmt"
	"time"

	domaincard "rinofinance-api/internal/domain/card"
)

const (
	dueAlertDays     = 3
	closingAlertDays = 3
)

type cardAlert struct {
	title string
	body  string
}

func dueAlerts(cards []*domaincard.CreditCard, now time.Time) []cardAlert {
	var alerts []cardAlert
	for _, c := range cards {
		if c.DueDay < 1 || c.DueDay > 31 {
			continue
		}
		if d := daysUntilDay(now, c.DueDay); d <= dueAlertDays {
			alerts = append(alerts, cardAlert{
				title: fmt.Sprintf("⚠️ Fatura do %s %s", c.Name, dueLabel(d)),
				body:  "Deixe programado o pagamento pra não pegar juros.",
			})
		}
	}
	return alerts
}

func closingAlerts(cards []*domaincard.CreditCard, now time.Time) []cardAlert {
	var alerts []cardAlert
	for _, c := range cards {
		if c.ClosingDay < 1 || c.ClosingDay > 31 {
			continue
		}
		if d := daysUntilDay(now, c.ClosingDay); d <= closingAlertDays {
			best := bestPurchaseDay(c.ClosingDay)
			alerts = append(alerts, cardAlert{
				title: fmt.Sprintf("💳 Fatura do %s fecha %s", c.Name, whenLabel(d)),
				body:  fmt.Sprintf("Compre a partir do dia %d pra cair só na próxima fatura.", best),
			})
		}
	}
	return alerts
}

func daysUntilDay(now time.Time, day int) int {
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	onDay := func(year int, month time.Month) time.Time {
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(year, month, day, 0, 0, 0, 0, loc)
	}

	target := onDay(now.Year(), now.Month())
	if target.Before(today) {
		nextMonth := now.Month() + 1
		year := now.Year()
		if nextMonth > 12 {
			nextMonth, year = 1, year+1
		}
		target = onDay(year, nextMonth)
	}
	return int(target.Sub(today).Hours()) / 24
}

func bestPurchaseDay(closingDay int) int {
	if closingDay >= 31 {
		return 1
	}
	return closingDay + 1
}

func dueLabel(days int) string {
	switch days {
	case 0:
		return "vence hoje"
	case 1:
		return "vence amanhã"
	default:
		return fmt.Sprintf("vence em %d dias", days)
	}
}

func whenLabel(days int) string {
	switch days {
	case 0:
		return "hoje"
	case 1:
		return "amanhã"
	default:
		return fmt.Sprintf("em %d dias", days)
	}
}
