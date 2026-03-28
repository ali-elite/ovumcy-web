package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/models"
	"github.com/ovumcy/ovumcy-web/internal/security"
)

var (
	ErrOIDCDisabled              = errors.New("oidc disabled")
	ErrOIDCUnavailable           = errors.New("oidc unavailable")
	ErrOIDCCallbackInvalid       = errors.New("oidc callback invalid")
	ErrOIDCAuthenticationFailed  = errors.New("oidc authentication failed")
	ErrOIDCAccountUnavailable    = errors.New("oidc account unavailable")
	ErrOIDCIdentityResolveFailed = errors.New("oidc identity resolve failed")
	ErrOIDCLinkFailed            = errors.New("oidc identity link failed")
)

type OIDCProviderClient interface {
	Enabled() bool
	AuthCodeURL(ctx context.Context, state string, nonce string, codeVerifier string) (string, error)
	ExchangeCode(ctx context.Context, code string, codeVerifier string, expectedNonce string) (security.OIDCClaims, error)
}

type OIDCIdentityStore interface {
	FindByIssuerSubject(issuer string, subject string) (models.OIDCIdentity, bool, error)
	Create(identity *models.OIDCIdentity) error
	TouchLastUsed(identityID uint, usedAt time.Time) error
}

type OIDCUserStore interface {
	FindByID(userID uint) (models.User, error)
	FindByNormalizedEmailOptional(email string) (models.User, bool, error)
}

type OIDCLoginResult struct {
	User        models.User
	NewlyLinked bool
}

type OIDCLoginService struct {
	client     OIDCProviderClient
	identities OIDCIdentityStore
	users      OIDCUserStore
}

func NewOIDCLoginService(client OIDCProviderClient, identities OIDCIdentityStore, users OIDCUserStore) *OIDCLoginService {
	return &OIDCLoginService{
		client:     client,
		identities: identities,
		users:      users,
	}
}

func (service *OIDCLoginService) Enabled() bool {
	return service != nil && service.client != nil && service.client.Enabled()
}

func (service *OIDCLoginService) StartAuth(ctx context.Context, state string, nonce string, codeVerifier string) (string, error) {
	if !service.Enabled() {
		return "", ErrOIDCDisabled
	}
	if strings.TrimSpace(state) == "" || strings.TrimSpace(nonce) == "" || strings.TrimSpace(codeVerifier) == "" {
		return "", ErrOIDCCallbackInvalid
	}
	url, err := service.client.AuthCodeURL(ctx, state, nonce, codeVerifier)
	if err != nil {
		return "", ErrOIDCUnavailable
	}
	return url, nil
}

func (service *OIDCLoginService) Authenticate(ctx context.Context, code string, codeVerifier string, expectedNonce string, now time.Time) (OIDCLoginResult, error) {
	if !service.Enabled() {
		return OIDCLoginResult{}, ErrOIDCDisabled
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(codeVerifier) == "" || strings.TrimSpace(expectedNonce) == "" {
		return OIDCLoginResult{}, ErrOIDCCallbackInvalid
	}

	claims, err := service.client.ExchangeCode(ctx, code, codeVerifier, expectedNonce)
	if err != nil {
		return OIDCLoginResult{}, ErrOIDCAuthenticationFailed
	}

	identity, found, err := service.identities.FindByIssuerSubject(claims.Issuer, claims.Subject)
	if err != nil {
		return OIDCLoginResult{}, ErrOIDCIdentityResolveFailed
	}
	if found {
		user, err := service.users.FindByID(identity.UserID)
		if err != nil {
			return OIDCLoginResult{}, ErrOIDCIdentityResolveFailed
		}
		_ = service.identities.TouchLastUsed(identity.ID, effectiveOIDCLoginTime(now))
		return OIDCLoginResult{User: user}, nil
	}

	normalizedEmail := NormalizeAuthEmail(claims.Email)
	if !claims.EmailVerified || normalizedEmail == "" {
		return OIDCLoginResult{}, ErrOIDCAccountUnavailable
	}

	user, found, err := service.users.FindByNormalizedEmailOptional(normalizedEmail)
	if err != nil {
		return OIDCLoginResult{}, ErrOIDCIdentityResolveFailed
	}
	if !found {
		return OIDCLoginResult{}, ErrOIDCAccountUnavailable
	}

	linkTime := effectiveOIDCLoginTime(now)
	identity = models.OIDCIdentity{
		UserID:     user.ID,
		Issuer:     strings.TrimSpace(claims.Issuer),
		Subject:    strings.TrimSpace(claims.Subject),
		CreatedAt:  linkTime,
		LastUsedAt: &linkTime,
	}
	if err := service.identities.Create(&identity); err != nil {
		return OIDCLoginResult{}, ErrOIDCLinkFailed
	}

	return OIDCLoginResult{
		User:        user,
		NewlyLinked: true,
	}, nil
}

func effectiveOIDCLoginTime(now time.Time) time.Time {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return now.UTC()
}
