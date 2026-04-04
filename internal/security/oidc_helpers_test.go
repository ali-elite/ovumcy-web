package security

import "testing"

func TestSanitizeOIDCConfigNormalizesDefaultsAndDomains(t *testing.T) {
	t.Parallel()

	config := sanitizeOIDCConfig(OIDCConfig{
		IssuerURL:                   " https://id.example.com ",
		ClientID:                    " ovumcy ",
		ClientSecret:                " secret ",
		RedirectURL:                 " https://ovumcy.example.com/auth/oidc/callback ",
		PostLogoutRedirectURL:       " https://ovumcy.example.com/logout-done ",
		AutoProvisionAllowedDomains: []string{" Example.com ", "@staff.example.com", "example.com", ""},
	})

	if config.LoginMode != OIDCLoginModeHybrid {
		t.Fatalf("expected default login mode hybrid, got %q", config.LoginMode)
	}
	if config.LogoutMode != OIDCLogoutModeLocal {
		t.Fatalf("expected default logout mode local, got %q", config.LogoutMode)
	}
	if config.IssuerURL != "https://id.example.com" || config.ClientID != "ovumcy" || config.ClientSecret != "secret" {
		t.Fatalf("expected trimmed OIDC config, got %#v", config)
	}
	if len(config.AutoProvisionAllowedDomains) != 2 || config.AutoProvisionAllowedDomains[0] != "example.com" || config.AutoProvisionAllowedDomains[1] != "staff.example.com" {
		t.Fatalf("expected normalized and deduplicated domains, got %#v", config.AutoProvisionAllowedDomains)
	}
}

func TestValidateOIDCHTTPSURLRejectsUnsafeInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rawURL string
	}{
		{name: "relative", rawURL: "/callback"},
		{name: "http", rawURL: "http://id.example.com"},
		{name: "query", rawURL: "https://id.example.com/callback?foo=bar"},
		{name: "fragment", rawURL: "https://id.example.com/callback#frag"},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if _, err := validateOIDCHTTPSURL(testCase.rawURL, "OIDC_TEST_URL"); err == nil {
				t.Fatalf("expected validateOIDCHTTPSURL(%q) to fail", testCase.rawURL)
			}
		})
	}
}

func TestValidateOIDCHTTPSURLAcceptsAbsoluteHTTPSURL(t *testing.T) {
	t.Parallel()

	parsed, err := validateOIDCHTTPSURL("https://id.example.com/path", "OIDC_TEST_URL")
	if err != nil {
		t.Fatalf("expected valid HTTPS URL, got %v", err)
	}
	if parsed.String() != "https://id.example.com/path" {
		t.Fatalf("unexpected parsed URL %q", parsed.String())
	}
}

func TestSanitizeOIDCEndSessionEndpoint(t *testing.T) {
	t.Parallel()

	if got := sanitizeOIDCEndSessionEndpoint("https://id.example.com/logout?client=ovumcy"); got != "https://id.example.com/logout?client=ovumcy" {
		t.Fatalf("expected https endpoint with query to remain intact, got %q", got)
	}
	if got := sanitizeOIDCEndSessionEndpoint("http://id.example.com/logout"); got != "" {
		t.Fatalf("expected insecure endpoint to be rejected, got %q", got)
	}
	if got := sanitizeOIDCEndSessionEndpoint("https://id.example.com/logout#frag"); got != "" {
		t.Fatalf("expected fragment endpoint to be rejected, got %q", got)
	}
}

func TestSameOriginURLAndEffectivePort(t *testing.T) {
	t.Parallel()

	left, err := validateOIDCHTTPSURL("https://ovumcy.example.com/auth/oidc/callback", "LEFT")
	if err != nil {
		t.Fatalf("left URL: %v", err)
	}
	rightDefault, err := validateOIDCHTTPSURL("https://ovumcy.example.com/login", "RIGHT")
	if err != nil {
		t.Fatalf("right default URL: %v", err)
	}
	rightExplicit, err := validateOIDCHTTPSURL("https://ovumcy.example.com:443/logout", "RIGHT")
	if err != nil {
		t.Fatalf("right explicit URL: %v", err)
	}
	otherHost, err := validateOIDCHTTPSURL("https://other.example.com/login", "OTHER")
	if err != nil {
		t.Fatalf("other host URL: %v", err)
	}

	if !sameOriginURL(left, rightDefault) || !sameOriginURL(left, rightExplicit) {
		t.Fatal("expected sameOriginURL to accept identical https origins with implicit and explicit ports")
	}
	if sameOriginURL(left, otherHost) {
		t.Fatal("did not expect sameOriginURL to accept a different host")
	}
	if sameOriginURL(nil, rightDefault) {
		t.Fatal("did not expect sameOriginURL to accept nil URLs")
	}
}

func TestOIDCConfigBehaviorFlagsAndAutoProvisionEdges(t *testing.T) {
	t.Parallel()

	config := OIDCConfig{
		Enabled:       true,
		AutoProvision: true,
		LoginMode:     OIDCLoginModeOIDCOnly,
		LogoutMode:    OIDCLogoutModeProvider,
	}

	if config.LocalPublicAuthEnabled() {
		t.Fatal("expected oidc_only mode to disable local public auth")
	}
	if !config.ProviderLogoutEnabled() {
		t.Fatal("expected provider logout mode to enable provider logout")
	}
	if config.AllowsAutoProvision(" ") {
		t.Fatal("did not expect blank email to pass auto-provision")
	}
	if !config.AllowsAutoProvision("owner@example.com") {
		t.Fatal("expected auto-provision to allow valid emails when no allowlist is configured")
	}

	config.AutoProvisionAllowedDomains = []string{"example.com"}
	if config.AllowsAutoProvision("missing-at-sign") {
		t.Fatal("did not expect malformed email to pass auto-provision with an allowlist")
	}
	if config.AllowsAutoProvision("owner@other.example.com") {
		t.Fatal("did not expect a non-allowlisted domain to pass auto-provision")
	}
	if !config.AllowsAutoProvision("owner@example.com") {
		t.Fatal("expected allowlisted domain to pass auto-provision")
	}

	config.Enabled = false
	if !config.LocalPublicAuthEnabled() {
		t.Fatal("expected disabled OIDC to allow local public auth")
	}
}

func TestOIDCConfigResolvedPostLogoutRedirectURLReturnsEmptyForInvalidRedirect(t *testing.T) {
	t.Parallel()

	config := OIDCConfig{RedirectURL: "not-a-url"}
	if got := config.ResolvedPostLogoutRedirectURL(); got != "" {
		t.Fatalf("expected invalid redirect URL to resolve to empty string, got %q", got)
	}
}
