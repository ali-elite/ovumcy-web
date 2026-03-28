package security

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const OIDCCallbackPath = "/auth/oidc/callback"
const defaultOIDCHTTPTimeout = 10 * time.Second

type OIDCConfig struct {
	Enabled       bool
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	AutoProvision bool
}

type OIDCClaims struct {
	Issuer        string
	Subject       string
	Email         string
	EmailVerified bool
}

type OIDCClient struct {
	config OIDCConfig

	mu          sync.Mutex
	httpClient  *http.Client
	provider    *oidc.Provider
	oauthConfig *oauth2.Config
	verifier    *oidc.IDTokenVerifier
}

func NewOIDCClient(config OIDCConfig) *OIDCClient {
	return &OIDCClient{
		config:     sanitizeOIDCConfig(config),
		httpClient: &http.Client{Timeout: defaultOIDCHTTPTimeout},
	}
}

func (config OIDCConfig) Validate(cookieSecure bool) error {
	config = sanitizeOIDCConfig(config)
	if !config.Enabled {
		return nil
	}
	if !cookieSecure {
		return errors.New("OIDC_ENABLED=true requires COOKIE_SECURE=true")
	}
	if config.AutoProvision {
		return errors.New("OIDC_AUTO_PROVISION=true is not supported yet")
	}
	if config.IssuerURL == "" {
		return errors.New("OIDC_ISSUER_URL is required when OIDC_ENABLED=true")
	}
	if config.ClientID == "" {
		return errors.New("OIDC_CLIENT_ID is required when OIDC_ENABLED=true")
	}
	if config.ClientSecret == "" {
		return errors.New("OIDC_CLIENT_SECRET is required when OIDC_ENABLED=true")
	}
	if config.RedirectURL == "" {
		return errors.New("OIDC_REDIRECT_URL is required when OIDC_ENABLED=true")
	}

	issuerURL, err := url.Parse(config.IssuerURL)
	if err != nil || !issuerURL.IsAbs() {
		return errors.New("OIDC_ISSUER_URL must be an absolute URL")
	}
	if !strings.EqualFold(issuerURL.Scheme, "https") {
		return errors.New("OIDC_ISSUER_URL must use https")
	}
	if issuerURL.RawQuery != "" || issuerURL.Fragment != "" {
		return errors.New("OIDC_ISSUER_URL must not include query or fragment")
	}

	redirectURL, err := url.Parse(config.RedirectURL)
	if err != nil || !redirectURL.IsAbs() {
		return errors.New("OIDC_REDIRECT_URL must be an absolute URL")
	}
	if !strings.EqualFold(redirectURL.Scheme, "https") {
		return errors.New("OIDC_REDIRECT_URL must use https")
	}
	if redirectURL.RawQuery != "" || redirectURL.Fragment != "" {
		return errors.New("OIDC_REDIRECT_URL must not include query or fragment")
	}
	if path.Clean(strings.TrimSpace(redirectURL.Path)) != OIDCCallbackPath {
		return fmt.Errorf("OIDC_REDIRECT_URL path must be %s", OIDCCallbackPath)
	}

	return nil
}

func (client *OIDCClient) Enabled() bool {
	return client != nil && client.config.Enabled
}

func (client *OIDCClient) AuthCodeURL(ctx context.Context, state string, nonce string, codeVerifier string) (string, error) {
	if !client.Enabled() {
		return "", errors.New("oidc is disabled")
	}
	oauthConfig, _, err := client.loadProvider(client.clientContext(ctx))
	if err != nil {
		return "", err
	}
	return oauthConfig.AuthCodeURL(
		strings.TrimSpace(state),
		oidc.Nonce(strings.TrimSpace(nonce)),
		oauth2.S256ChallengeOption(strings.TrimSpace(codeVerifier)),
		oauth2.SetAuthURLParam("response_mode", "form_post"),
	), nil
}

func (client *OIDCClient) ExchangeCode(ctx context.Context, code string, codeVerifier string, expectedNonce string) (OIDCClaims, error) {
	if !client.Enabled() {
		return OIDCClaims{}, errors.New("oidc is disabled")
	}
	ctx = client.clientContext(ctx)
	oauthConfig, verifier, err := client.loadProvider(ctx)
	if err != nil {
		return OIDCClaims{}, err
	}

	token, err := oauthConfig.Exchange(ctx, strings.TrimSpace(code), oauth2.VerifierOption(strings.TrimSpace(codeVerifier)))
	if err != nil {
		return OIDCClaims{}, fmt.Errorf("exchange oidc authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || strings.TrimSpace(rawIDToken) == "" {
		return OIDCClaims{}, errors.New("oidc token response is missing id_token")
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return OIDCClaims{}, fmt.Errorf("verify oidc id_token: %w", err)
	}
	if strings.TrimSpace(idToken.Nonce) != strings.TrimSpace(expectedNonce) {
		return OIDCClaims{}, errors.New("oidc nonce mismatch")
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return OIDCClaims{}, fmt.Errorf("decode oidc id_token claims: %w", err)
	}

	return OIDCClaims{
		Issuer:        strings.TrimSpace(idToken.Issuer),
		Subject:       strings.TrimSpace(idToken.Subject),
		Email:         strings.TrimSpace(claims.Email),
		EmailVerified: claims.EmailVerified,
	}, nil
}

func (client *OIDCClient) clientContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil || client.httpClient == nil {
		return ctx
	}
	return context.WithValue(ctx, oauth2.HTTPClient, client.httpClient)
}

func (client *OIDCClient) loadProvider(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.oauthConfig != nil && client.verifier != nil {
		return client.oauthConfig, client.verifier, nil
	}

	provider, err := oidc.NewProvider(ctx, client.config.IssuerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("discover oidc provider: %w", err)
	}

	client.provider = provider
	client.oauthConfig = &oauth2.Config{
		ClientID:     client.config.ClientID,
		ClientSecret: client.config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  client.config.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
	client.verifier = provider.Verifier(&oidc.Config{
		ClientID: client.config.ClientID,
	})

	return client.oauthConfig, client.verifier, nil
}

func sanitizeOIDCConfig(config OIDCConfig) OIDCConfig {
	config.IssuerURL = strings.TrimSpace(config.IssuerURL)
	config.ClientID = strings.TrimSpace(config.ClientID)
	config.ClientSecret = strings.TrimSpace(config.ClientSecret)
	config.RedirectURL = strings.TrimSpace(config.RedirectURL)
	return config
}
