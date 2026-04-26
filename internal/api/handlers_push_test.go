package api

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ovumcy/ovumcy-web/internal/models"
)

func TestPushSubscribeAcceptsCSRFProtectedFormPayload(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "push-subscribe@example.com")

	response := settingsFormRequestWithCSRF(t, ctx, http.MethodPost, "/api/settings/push/subscribe", url.Values{
		"endpoint": {"https://push.example/subscription-1"},
		"p256dh":   {"client-key"},
		"auth":     {"auth-secret"},
	}, map[string]string{
		"Accept": "application/json",
	})
	defer response.Body.Close()
	assertStatusCode(t, response, http.StatusOK)

	var stored models.PushSubscription
	if err := ctx.database.Where("endpoint = ?", "https://push.example/subscription-1").First(&stored).Error; err != nil {
		t.Fatalf("expected push subscription to be stored: %v", err)
	}
	if stored.UserID != ctx.user.ID || stored.P256dh != "client-key" || stored.Auth != "auth-secret" {
		t.Fatalf("unexpected stored subscription: %#v", stored)
	}
}

func TestPushUnsubscribeRemovesStoredSubscription(t *testing.T) {
	ctx := newSettingsSecurityTestContext(t, "push-unsubscribe@example.com")
	subscription := models.PushSubscription{
		UserID:   ctx.user.ID,
		Endpoint: "https://push.example/subscription-2",
		P256dh:   "client-key",
		Auth:     "auth-secret",
	}
	if err := ctx.database.Create(&subscription).Error; err != nil {
		t.Fatalf("seed push subscription: %v", err)
	}

	response := settingsFormRequestWithCSRF(t, ctx, http.MethodPost, "/api/settings/push/unsubscribe", url.Values{
		"endpoint": {subscription.Endpoint},
	}, map[string]string{
		"Accept": "application/json",
	})
	defer response.Body.Close()
	assertStatusCode(t, response, http.StatusOK)

	var count int64
	if err := ctx.database.Model(&models.PushSubscription{}).Where("endpoint = ?", subscription.Endpoint).Count(&count).Error; err != nil {
		t.Fatalf("count push subscriptions: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected subscription to be removed, got count=%d", count)
	}
}
