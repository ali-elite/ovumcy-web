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
	ErrOIDCProvisionFailed       = errors.New("oidc account provision failed")
)

type OIDCProviderClient interface {
	Enabled() bool
	LocalPublicAuthEnabled() bool
	Config() security.OIDCConfig
	AuthCodeURL(ctx context.Context, state string, nonce string, codeVerifier string) (string, error)
	ExchangeCode(ctx context.Context, code string, codeVerifier string, expectedNonce string) (security.OIDCExchangeResult, error)
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

type OIDCAutoProvisioner interface {
	AutoProvisionOwnerAccount(email string, createdAt time.Time) (models.User, error)
}

type OIDCLogoutState struct {
	EndSessionEndpoint    string
	IDTokenHint           string
	PostLogoutRedirectURL string
}

type OIDCLoginResult struct {
	User            models.User
	NewlyLinked     bool
	AutoProvisioned bool
	Logout          *OIDCLogoutState
}

type OIDCLoginService struct {
	client      OIDCProviderClient
	identities  OIDCIdentityStore
	users       OIDCUserStore
	provisioner OIDCAutoProvisioner
	config      security.OIDCConfig
}

func NewOIDCLoginService(client OIDCProviderClient, identities OIDCIdentityStore, users OIDCUserStore, provisioner OIDCAutoProvisioner) *OIDCLoginService {
	config := security.OIDCConfig{}
	if client != nil {
		config = client.Config()
	}
	return &OIDCLoginService{
		client:      client,
		identities:  identities,
		users:       users,
		provisioner: provisioner,
		config:      config,
	}
}

func (service *OIDCLoginService) Enabled() bool {
	return service != nil && service.client != nil && service.client.Enabled()
}

func (service *OIDCLoginService) LocalPublicAuthEnabled() bool {
	if service == nil {
		return true
	}
	return service.config.LocalPublicAuthEnabled()
}

func (service *OIDCLoginService) OIDCOnly() bool {
	return service.Enabled() && !service.LocalPublicAuthEnabled()
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

	exchange, err := service.client.ExchangeCode(ctx, code, codeVerifier, expectedNonce)
	if err != nil {
		return OIDCLoginResult{}, ErrOIDCAuthenticationFailed
	}

	logoutState := service.buildLogoutState(exchange.Session)

	identity, found, err := service.identities.FindByIssuerSubject(exchange.Claims.Issuer, exchange.Claims.Subject)
	if err != nil {
		return OIDCLoginResult{}, ErrOIDCIdentityResolveFailed
	}
	if found {
		user, err := service.users.FindByID(identity.UserID)
		if err != nil {
			return OIDCLoginResult{}, ErrOIDCIdentityResolveFailed
		}
		_ = service.identities.TouchLastUsed(identity.ID, effectiveOIDCLoginTime(now))
		return OIDCLoginResult{User: user, Logout: logoutState}, nil
	}

	normalizedEmail := NormalizeAuthEmail(exchange.Claims.Email)
	if !exchange.Claims.EmailVerified || normalizedEmail == "" {
		return OIDCLoginResult{}, ErrOIDCAccountUnavailable
	}

	user, found, err := service.users.FindByNormalizedEmailOptional(normalizedEmail)
	if err != nil {
		return OIDCLoginResult{}, ErrOIDCIdentityResolveFailed
	}
	autoProvisioned := false
	if !found {
		if !service.config.AllowsAutoProvision(normalizedEmail) || service.provisioner == nil {
			return OIDCLoginResult{}, ErrOIDCAccountUnavailable
		}
		user, err = service.provisioner.AutoProvisionOwnerAccount(normalizedEmail, effectiveOIDCLoginTime(now))
		if err != nil {
			if errors.Is(err, ErrAuthEmailExists) {
				user, found, err = service.users.FindByNormalizedEmailOptional(normalizedEmail)
				if err != nil {
					return OIDCLoginResult{}, ErrOIDCIdentityResolveFailed
				}
				if !found {
					return OIDCLoginResult{}, ErrOIDCProvisionFailed
				}
			} else {
				return OIDCLoginResult{}, ErrOIDCProvisionFailed
			}
		} else {
			found = true
			autoProvisioned = true
		}
	}
	if !found {
		return OIDCLoginResult{}, ErrOIDCAccountUnavailable
	}

	linkTime := effectiveOIDCLoginTime(now)
	identity = models.OIDCIdentity{
		UserID:     user.ID,
		Issuer:     strings.TrimSpace(exchange.Claims.Issuer),
		Subject:    strings.TrimSpace(exchange.Claims.Subject),
		CreatedAt:  linkTime,
		LastUsedAt: &linkTime,
	}
	if err := service.identities.Create(&identity); err != nil {
		return OIDCLoginResult{}, ErrOIDCLinkFailed
	}

	return OIDCLoginResult{
		User:            user,
		NewlyLinked:     true,
		AutoProvisioned: autoProvisioned,
		Logout:          logoutState,
	}, nil
}

func effectiveOIDCLoginTime(now time.Time) time.Time {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return now.UTC()
}

func (service *OIDCLoginService) buildLogoutState(session security.OIDCSession) *OIDCLogoutState {
	if !service.config.ProviderLogoutEnabled() {
		return nil
	}

	endSessionEndpoint := strings.TrimSpace(session.EndSessionEndpoint)
	idTokenHint := strings.TrimSpace(session.IDTokenHint)
	postLogoutRedirectURL := strings.TrimSpace(service.config.ResolvedPostLogoutRedirectURL())
	if endSessionEndpoint == "" || idTokenHint == "" || postLogoutRedirectURL == "" {
		return nil
	}

	return &OIDCLogoutState{
		EndSessionEndpoint:    endSessionEndpoint,
		IDTokenHint:           idTokenHint,
		PostLogoutRedirectURL: postLogoutRedirectURL,
	}
}
