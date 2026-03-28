package security

import (
	"context"
	"net/http"
	"testing"

	"golang.org/x/oauth2"
)

func TestNewOIDCClientConfiguresBoundedHTTPClient(t *testing.T) {
	client := NewOIDCClient(OIDCConfig{Enabled: true})
	if client == nil {
		t.Fatal("expected OIDC client")
	}
	if client.httpClient == nil {
		t.Fatal("expected bounded OIDC http client")
	}
	if client.httpClient.Timeout != defaultOIDCHTTPTimeout {
		t.Fatalf("expected OIDC http timeout %s, got %s", defaultOIDCHTTPTimeout, client.httpClient.Timeout)
	}
}

func TestOIDCClientContextInjectsConfiguredHTTPClient(t *testing.T) {
	client := NewOIDCClient(OIDCConfig{Enabled: true})

	ctx := client.clientContext(context.Background())
	httpClient, ok := ctx.Value(oauth2.HTTPClient).(*http.Client)
	if !ok {
		t.Fatal("expected oauth2 http client in context")
	}
	if httpClient != client.httpClient {
		t.Fatal("expected OIDC context to reuse configured http client")
	}

	type contextKey string
	parentKey := contextKey("parent")
	parent := context.WithValue(context.Background(), parentKey, "value")
	ctx = client.clientContext(parent)
	if got := ctx.Value(parentKey); got != "value" {
		t.Fatalf("expected parent context values to be preserved, got %#v", got)
	}
}
