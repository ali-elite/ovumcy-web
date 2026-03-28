package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/security"
)

type stubOIDCProviderClient struct {
	enabled     bool
	authURL     string
	claims      security.OIDCClaims
	authErr     error
	exchangeErr error
}

func (stub *stubOIDCProviderClient) Enabled() bool {
	return stub.enabled
}

func (stub *stubOIDCProviderClient) AuthCodeURL(context.Context, string, string, string) (string, error) {
	if stub.authErr != nil {
		return "", stub.authErr
	}
	return stub.authURL, nil
}

func (stub *stubOIDCProviderClient) ExchangeCode(context.Context, string, string, string) (security.OIDCClaims, error) {
	if stub.exchangeErr != nil {
		return security.OIDCClaims{}, stub.exchangeErr
	}
	return stub.claims, nil
}

type stubOIDCIdentityStore struct {
	identity       models.OIDCIdentity
	found          bool
	findErr        error
	createErr      error
	touchedID      uint
	touchedAt      time.Time
	created        models.OIDCIdentity
	createCallSeen bool
}

func (stub *stubOIDCIdentityStore) FindByIssuerSubject(string, string) (models.OIDCIdentity, bool, error) {
	if stub.findErr != nil {
		return models.OIDCIdentity{}, false, stub.findErr
	}
	if !stub.found {
		return models.OIDCIdentity{}, false, nil
	}
	return stub.identity, true, nil
}

func (stub *stubOIDCIdentityStore) Create(identity *models.OIDCIdentity) error {
	stub.createCallSeen = true
	if identity != nil {
		stub.created = *identity
	}
	return stub.createErr
}

func (stub *stubOIDCIdentityStore) TouchLastUsed(identityID uint, usedAt time.Time) error {
	stub.touchedID = identityID
	stub.touchedAt = usedAt
	return nil
}

type stubOIDCUserStore struct {
	byID            models.User
	byIDErr         error
	byEmail         models.User
	byEmailFound    bool
	byEmailErr      error
	lastLookupEmail string
}

func (stub *stubOIDCUserStore) FindByID(uint) (models.User, error) {
	if stub.byIDErr != nil {
		return models.User{}, stub.byIDErr
	}
	return stub.byID, nil
}

func (stub *stubOIDCUserStore) FindByNormalizedEmailOptional(email string) (models.User, bool, error) {
	stub.lastLookupEmail = email
	if stub.byEmailErr != nil {
		return models.User{}, false, stub.byEmailErr
	}
	if !stub.byEmailFound {
		return models.User{}, false, nil
	}
	return stub.byEmail, true, nil
}

func TestOIDCLoginServiceStartAuthRequiresEnabledProvider(t *testing.T) {
	t.Parallel()

	service := NewOIDCLoginService(&stubOIDCProviderClient{}, &stubOIDCIdentityStore{}, &stubOIDCUserStore{})

	if _, err := service.StartAuth(context.Background(), "state", "nonce", "verifier"); !errors.Is(err, ErrOIDCDisabled) {
		t.Fatalf("expected ErrOIDCDisabled, got %v", err)
	}
}

func TestOIDCLoginServiceAuthenticateUsesExistingIdentityLink(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 11, 30, 0, 0, time.UTC)
	identities := &stubOIDCIdentityStore{
		found: true,
		identity: models.OIDCIdentity{
			ID:      44,
			UserID:  7,
			Issuer:  "https://id.example.com",
			Subject: "owner-subject",
		},
	}
	users := &stubOIDCUserStore{
		byID: models.User{
			ID:                  7,
			Role:                models.RoleOwner,
			OnboardingCompleted: true,
		},
	}
	service := NewOIDCLoginService(&stubOIDCProviderClient{
		enabled: true,
		claims: security.OIDCClaims{
			Issuer:        "https://id.example.com",
			Subject:       "owner-subject",
			Email:         "owner@example.com",
			EmailVerified: true,
		},
	}, identities, users)

	result, err := service.Authenticate(context.Background(), "code", "verifier", "nonce", now)
	if err != nil {
		t.Fatalf("Authenticate() unexpected error: %v", err)
	}
	if result.NewlyLinked {
		t.Fatal("did not expect existing identity to be linked again")
	}
	if result.User.ID != 7 {
		t.Fatalf("expected linked user id 7, got %d", result.User.ID)
	}
	if identities.touchedID != 44 {
		t.Fatalf("expected last-used touch for identity 44, got %d", identities.touchedID)
	}
	if !identities.touchedAt.Equal(now) {
		t.Fatalf("expected last-used timestamp %s, got %s", now, identities.touchedAt)
	}
	if identities.createCallSeen {
		t.Fatal("did not expect Create() for an existing identity link")
	}
}

func TestOIDCLoginServiceAuthenticateLinksVerifiedEmailOnFirstLogin(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 12, 0, 0, 0, time.UTC)
	identities := &stubOIDCIdentityStore{}
	users := &stubOIDCUserStore{
		byEmailFound: true,
		byEmail: models.User{
			ID:                  9,
			Email:               "owner@example.com",
			Role:                models.RoleOwner,
			OnboardingCompleted: true,
		},
	}
	service := NewOIDCLoginService(&stubOIDCProviderClient{
		enabled: true,
		claims: security.OIDCClaims{
			Issuer:        "https://id.example.com",
			Subject:       "first-login-sub",
			Email:         " Owner@Example.com ",
			EmailVerified: true,
		},
	}, identities, users)

	result, err := service.Authenticate(context.Background(), "code", "verifier", "nonce", now)
	if err != nil {
		t.Fatalf("Authenticate() unexpected error: %v", err)
	}
	if !result.NewlyLinked {
		t.Fatal("expected first verified login to create an identity link")
	}
	if result.User.ID != 9 {
		t.Fatalf("expected user id 9, got %d", result.User.ID)
	}
	if users.lastLookupEmail != "owner@example.com" {
		t.Fatalf("expected normalized email lookup, got %q", users.lastLookupEmail)
	}
	if !identities.createCallSeen {
		t.Fatal("expected Create() to persist new identity link")
	}
	if identities.created.UserID != 9 {
		t.Fatalf("expected linked user id 9, got %d", identities.created.UserID)
	}
	if identities.created.Issuer != "https://id.example.com" || identities.created.Subject != "first-login-sub" {
		t.Fatalf("expected issuer/subject link to be persisted, got %+v", identities.created)
	}
	if identities.created.LastUsedAt == nil || !identities.created.LastUsedAt.Equal(now) {
		t.Fatalf("expected LastUsedAt=%s, got %+v", now, identities.created.LastUsedAt)
	}
}

func TestOIDCLoginServiceAuthenticateRejectsUnverifiedEmail(t *testing.T) {
	t.Parallel()

	service := NewOIDCLoginService(&stubOIDCProviderClient{
		enabled: true,
		claims: security.OIDCClaims{
			Issuer:        "https://id.example.com",
			Subject:       "no-verified-email",
			Email:         "owner@example.com",
			EmailVerified: false,
		},
	}, &stubOIDCIdentityStore{}, &stubOIDCUserStore{})

	if _, err := service.Authenticate(context.Background(), "code", "verifier", "nonce", time.Time{}); !errors.Is(err, ErrOIDCAccountUnavailable) {
		t.Fatalf("expected ErrOIDCAccountUnavailable, got %v", err)
	}
}

func TestOIDCLoginServiceAuthenticateMapsLinkPersistenceFailure(t *testing.T) {
	t.Parallel()

	service := NewOIDCLoginService(&stubOIDCProviderClient{
		enabled: true,
		claims: security.OIDCClaims{
			Issuer:        "https://id.example.com",
			Subject:       "duplicate-link",
			Email:         "owner@example.com",
			EmailVerified: true,
		},
	}, &stubOIDCIdentityStore{
		createErr: errors.New("duplicate key"),
	}, &stubOIDCUserStore{
		byEmailFound: true,
		byEmail:      models.User{ID: 5, Email: "owner@example.com"},
	})

	if _, err := service.Authenticate(context.Background(), "code", "verifier", "nonce", time.Time{}); !errors.Is(err, ErrOIDCLinkFailed) {
		t.Fatalf("expected ErrOIDCLinkFailed, got %v", err)
	}
}
