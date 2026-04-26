package services

import (
	"encoding/json"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/ovumcy/ovumcy-web/internal/db"
)

type PushService struct {
	repo         *db.PushSubscriptionRepository
	vapidPublic  string
	vapidPrivate string
}

func NewPushService(repo *db.PushSubscriptionRepository, vapidPublic, vapidPrivate string) *PushService {
	return &PushService{
		repo:         repo,
		vapidPublic:  vapidPublic,
		vapidPrivate: vapidPrivate,
	}
}

func (s *PushService) IsConfigured() bool {
	return s.vapidPublic != "" && s.vapidPrivate != ""
}

type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	URL   string `json:"url,omitempty"`
}

func (s *PushService) SendNotification(userID uint, payload PushPayload) error {
	if !s.IsConfigured() {
		return nil // Gracefully do nothing if VAPID keys aren't set
	}

	subs, err := s.repo.FindByUserID(userID)
	if err != nil {
		return fmt.Errorf("find subscriptions: %w", err)
	}

	if len(subs) == 0 {
		return nil
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	for _, sub := range subs {
		// Convert our model to the webpush format
		wpSub := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}

		// Send Notification
		res, err := webpush.SendNotification(payloadBytes, wpSub, &webpush.Options{
			Subscriber:      "mailto:admin@ovumcy.com",
			VAPIDPublicKey:  s.vapidPublic,
			VAPIDPrivateKey: s.vapidPrivate,
			TTL:             30,
		})
		if err != nil {
			// Webpush error handling: if 410 Gone, the subscription is no longer valid
			if res != nil && res.StatusCode == 410 {
				_ = s.repo.DeleteByEndpoint(sub.Endpoint)
			}
			continue // Log error in real app, but continue pushing to other devices
		}
		if res != nil {
			_ = res.Body.Close()
		}
	}

	return nil
}
