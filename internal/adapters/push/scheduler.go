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
	appTitle = "RinoFinance"
	sendHour = 20
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
			if now.Hour() == sendHour && now.Minute() == 0 {
				s.sendDailyReminders(ctx, now)
			}
		}
	}()
}

func (s *Scheduler) sendDailyReminders(ctx context.Context, now time.Time) {
	subs, err := s.subscriptions.ListAll(ctx)
	if err != nil {
		log.Printf("push scheduler: erro ao listar inscrições: %v", err)
		return
	}
	if len(subs) == 0 {
		return
	}

	byUser := make(map[uuid.UUID][]*notification.PushSubscription)
	for _, sub := range subs {
		byUser[sub.UserID] = append(byUser[sub.UserID], sub)
	}
	log.Printf("push scheduler: enviando lembrete diário para %d usuário(s)", len(byUser))

	for userID, userSubs := range byUser {
		body := s.buildMessage(ctx, userID, now)
		for _, sub := range userSubs {
			status, err := send(sub, appTitle, body, s.vapidPublic, s.vapidPrivate, s.vapidEmail)
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
}

func (s *Scheduler) SendNow(ctx context.Context, userID uuid.UUID) error {
	subs, err := s.subscriptions.ListAll(ctx)
	if err != nil {
		return err
	}
	body := s.buildMessage(ctx, userID, time.Now().In(brt))
	for _, sub := range subs {
		if sub.UserID != userID {
			continue
		}
		status, err := send(sub, appTitle, body, s.vapidPublic, s.vapidPrivate, s.vapidEmail)
		if (status == 404 || status == 410) && err != nil {
			_ = s.subscriptions.DeleteByID(ctx, sub.ID)
		}
	}
	return nil
}

func (s *Scheduler) buildMessage(ctx context.Context, userID uuid.UUID, now time.Time) string {
	name := ""
	if user, err := s.users.FindByID(ctx, userID); err == nil {
		name = firstName(user.Name)
	}

	greeting := "Hora de atualizar suas finanças 💸"
	if name != "" {
		greeting = name + ", hora de atualizar suas finanças 💸"
	}

	cards, err := s.cards.ListByUser(ctx, userID)
	if err == nil {
		if tip := cardTip(cards, now); tip != "" {
			return greeting + " " + tip
		}
	}
	return greeting + " Registre os gastos de hoje e mantenha tudo em dia."
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
