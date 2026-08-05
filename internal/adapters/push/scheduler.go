package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/notification"
	domainuser "rinofinance-api/internal/domain/user"
)

const (
	alertHour   = 8
	updateHour  = 20
	updateTitle = "Hora de atualizar suas finanças 💸"
)

var brt = time.FixedZone("BRT", -3*60*60)

type Scheduler struct {
	subscriptions notification.Repository
	users         domainuser.Repository
	cards         domaincard.CardRepository
	vapidPublic   string
	vapidPrivate  string
	vapidEmail    string
}

func NewScheduler(
	subscriptions notification.Repository,
	users domainuser.Repository,
	cards domaincard.CardRepository,
	vapidPublic, vapidPrivate, vapidEmail string,
) *Scheduler {
	return &Scheduler{
		subscriptions: subscriptions,
		users:         users,
		cards:         cards,
		vapidPublic:   vapidPublic,
		vapidPrivate:  vapidPrivate,
		vapidEmail:    vapidEmail,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	if s.vapidPublic == "" || s.vapidPrivate == "" {
		log.Println("push scheduler: VAPID keys não configuradas — desativado")
		return
	}
	log.Println("push scheduler: iniciado")
	go func() {
		for {
			next := time.Now().Truncate(time.Minute).Add(time.Minute)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(next)):
			}

			now := time.Now().In(brt)
			if now.Minute() != 0 {
				continue
			}
			switch now.Hour() {
			case alertHour:
				s.sendCardAlerts(ctx, now)
			case updateHour:
				s.run(ctx, now, "atualizar", s.updateFor)
			}
		}
	}()
}

type message struct {
	title string
	body  string
	send  bool
}

func (s *Scheduler) run(ctx context.Context, now time.Time, slot string, build func(context.Context, uuid.UUID, time.Time) message) {
	subs, err := s.subscriptions.ListAll(ctx)
	if err != nil {
		log.Printf("push scheduler: erro ao listar inscrições: %v", err)
		return
	}

	byUser := make(map[uuid.UUID][]*notification.PushSubscription)
	for _, sub := range subs {
		byUser[sub.UserID] = append(byUser[sub.UserID], sub)
	}
	log.Printf("push scheduler: slot=%s para %d usuário(s)", slot, len(byUser))

	for userID, userSubs := range byUser {
		msg := build(ctx, userID, now)
		if !msg.send {
			continue
		}
		s.deliver(ctx, userSubs, msg)
	}
}

func (s *Scheduler) deliver(ctx context.Context, subs []*notification.PushSubscription, msg message) {
	for _, sub := range subs {
		status, err := send(sub, msg.title, msg.body, s.vapidPublic, s.vapidPrivate, s.vapidEmail)
		if err != nil {
			log.Printf("push scheduler: erro ao enviar: %v", err)
		}
		if status == 404 || status == 410 {
			if delErr := s.subscriptions.DeleteByID(ctx, sub.ID); delErr != nil {
				log.Printf("push scheduler: erro ao remover inscrição expirada: %v", delErr)
			}
		}
	}
}

func (s *Scheduler) updateFor(ctx context.Context, userID uuid.UUID, _ time.Time) message {
	return message{title: updateTitle, body: updateBody(s.userName(ctx, userID)), send: true}
}

func (s *Scheduler) sendCardAlerts(ctx context.Context, now time.Time) {
	subs, err := s.subscriptions.ListAll(ctx)
	if err != nil {
		log.Printf("push scheduler: erro ao listar inscrições: %v", err)
		return
	}

	byUser := make(map[uuid.UUID][]*notification.PushSubscription)
	for _, sub := range subs {
		byUser[sub.UserID] = append(byUser[sub.UserID], sub)
	}

	for userID, userSubs := range byUser {
		cards, err := s.cards.ListByUser(ctx, userID)
		if err != nil {
			continue
		}
		alerts := cardAlerts(cards, now)
		if len(alerts) > 0 {
			log.Printf("push scheduler: %d alerta(s) de fatura para usuário %s", len(alerts), userID)
		}
		for _, alert := range alerts {
			s.deliver(ctx, userSubs, message{title: alert.title, body: alert.body, send: true})
		}
	}
}

func (s *Scheduler) SendNow(ctx context.Context, userID uuid.UUID) error {
	subs, err := s.subscriptions.ListAll(ctx)
	if err != nil {
		return err
	}
	var userSubs []*notification.PushSubscription
	for _, sub := range subs {
		if sub.UserID == userID {
			userSubs = append(userSubs, sub)
		}
	}
	if len(userSubs) == 0 {
		return nil
	}

	now := time.Now().In(brt)
	if cards, err := s.cards.ListByUser(ctx, userID); err == nil {
		if alerts := cardAlerts(cards, now); len(alerts) > 0 {
			for _, alert := range alerts {
				s.deliver(ctx, userSubs, message{title: alert.title, body: alert.body, send: true})
			}
			return nil
		}
	}
	s.deliver(ctx, userSubs, s.updateFor(ctx, userID, now))
	return nil
}

func (s *Scheduler) userName(ctx context.Context, userID uuid.UUID) string {
	if user, err := s.users.FindByID(ctx, userID); err == nil {
		return firstName(user.Name)
	}
	return ""
}

func updateBody(name string) string {
	if name != "" {
		return name + ", registre os gastos de hoje e mantenha tudo em dia."
	}
	return "Registre os gastos de hoje e mantenha tudo em dia."
}

func send(sub *notification.PushSubscription, title, body, pubKey, privKey, email string) (int, error) {
	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.P256DH, Auth: sub.Auth},
	}, &webpush.Options{
		VAPIDPublicKey:  pubKey,
		VAPIDPrivateKey: privKey,
		Subscriber:      email,
		TTL:             3600,
	})
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("push rejeitado: HTTP %d — %s", resp.StatusCode, string(respBody))
	}
	return resp.StatusCode, nil
}

func firstName(name string) string {
	for i, r := range name {
		if r == ' ' {
			return name[:i]
		}
	}
	return name
}
