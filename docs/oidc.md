# OIDC / SSO Guide

This guide documents Ovumcy's optional OpenID Connect sign-in for self-hosters who already run a central identity provider.

Use this page together with [README.md](../README.md) and [docs/self-hosted.md](self-hosted.md):

- `README.md` stays the short product and configuration overview;
- `docs/self-hosted.md` defines the supported deployment contract;
- this page explains the OIDC-specific operator setup, provider recipes, and troubleshooting.

## Current Contract

Ovumcy's OIDC support is intentionally narrow in the current release:

- it is optional and keeps the existing local username/password flow in place;
- it links only existing local accounts;
- the first successful OIDC sign-in must match an existing Ovumcy account by verified email;
- later sign-ins use the stored `(issuer, subject)` identity link;
- `OIDC_AUTO_PROVISION` must stay `false` because automatic account creation is not supported yet.

This means OIDC is currently a hybrid sign-in method, not a full replacement for account provisioning.

## Required Environment

Use the following variables together:

```env
COOKIE_SECURE=true
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://id.example.com
OIDC_CLIENT_ID=ovumcy
OIDC_CLIENT_SECRET=replace_with_a_client_secret
OIDC_REDIRECT_URL=https://ovumcy.example.com/auth/oidc/callback
OIDC_AUTO_PROVISION=false
```

Notes:

- `COOKIE_SECURE=true` is mandatory when `OIDC_ENABLED=true`.
- `OIDC_REDIRECT_URL` must be an absolute `https://` URL and its path must be exactly `/auth/oidc/callback`.
- `OIDC_ISSUER_URL` must be the issuer URL itself, not a URL with query parameters, fragments, or a copied browser login page.
- When your provider publishes a discovery document, copy the issuer value from that document instead of guessing.
- Ovumcy does not accept `OIDC_AUTO_PROVISION=true` yet.

## How Sign-In Works

1. The login page shows a `Sign in with SSO` button when `OIDC_ENABLED=true`.
2. When the user clicks it, Ovumcy starts a server-side Authorization Code flow with PKCE.
3. Ovumcy stores a sealed one-time state cookie containing the OIDC `state`, `nonce`, PKCE verifier, and expiry timestamp.
4. The identity provider authenticates the user and returns the browser to `/auth/oidc/callback` with `response_mode=form_post`.
5. Ovumcy validates the sealed state, exchanges the authorization code for tokens, and verifies the ID token plus `nonce`.
6. On the first successful OIDC login, Ovumcy matches the user to an existing local account by verified email and then stores the `(issuer, subject)` link.
7. Ovumcy finishes the login by issuing the normal local `ovumcy_auth` session cookie, so the rest of the app keeps using the existing cookie-based session model.

Provider and auth errors are intentionally kept out of query strings and fragments. Browser-facing failures return through the existing flash-based login UX instead.

## Provider Recipes

The exact UI labels may differ a little by provider version, but the required values are stable:

- callback URL: `https://ovumcy.example.com/auth/oidc/callback`
- client type: confidential web application
- response type / flow: authorization code
- scopes: at least `openid` and `email`

### Keycloak

Recommended setup:

1. Create or choose a realm for Ovumcy.
2. Create a new OpenID Connect client for a web application.
3. Keep the standard authorization code flow enabled and use a confidential client with a client secret.
4. Set `Valid Redirect URIs` to the exact Ovumcy callback URL:
   `https://ovumcy.example.com/auth/oidc/callback`
5. Use the realm issuer URL for `OIDC_ISSUER_URL`, for example:
   `https://keycloak.example.com/realms/ovumcy`
6. Copy the client ID and client secret into Ovumcy.
7. Make sure the user accounts that should sign in expose a verified email address.

Recommended Ovumcy mapping:

```env
OIDC_ISSUER_URL=https://keycloak.example.com/realms/ovumcy
OIDC_CLIENT_ID=ovumcy
OIDC_CLIENT_SECRET=replace_with_a_client_secret
OIDC_REDIRECT_URL=https://ovumcy.example.com/auth/oidc/callback
```

### Authentik

Recommended setup:

1. Create an Application and an OAuth2/OpenID Provider for Ovumcy.
2. Register the exact Ovumcy callback URL in the provider redirect URI list instead of relying on the first-visit auto-save behavior.
3. Use a confidential client with a generated secret.
4. Make sure the provider exposes the email scope and a verified email claim for the users who should sign in.
5. Use the application-specific issuer URL for `OIDC_ISSUER_URL`, usually:
   `https://authentik.example.com/application/o/<application-slug>/`
6. Copy the client ID and client secret into Ovumcy.

Recommended Ovumcy mapping:

```env
OIDC_ISSUER_URL=https://authentik.example.com/application/o/ovumcy/
OIDC_CLIENT_ID=ovumcy
OIDC_CLIENT_SECRET=replace_with_a_client_secret
OIDC_REDIRECT_URL=https://ovumcy.example.com/auth/oidc/callback
```

If discovery fails against the Authentik root domain, double-check that you used the application-specific issuer URL rather than the generic site URL.

### Authelia

Authelia configures OIDC clients in its own configuration file. A typical confidential-client entry looks like this:

```yaml
identity_providers:
  oidc:
    clients:
      - client_id: "ovumcy"
        client_name: "Ovumcy"
        client_secret: "digest-of-the-raw-client-secret"
        public: false
        redirect_uris:
          - "https://ovumcy.example.com/auth/oidc/callback"
        scopes:
          - "openid"
          - "email"
          - "profile"
        response_types:
          - "code"
        grant_types:
          - "authorization_code"
```

Recommended setup:

1. Enable the Authelia OIDC provider and set the provider root URL you want to expose publicly.
2. Add a confidential client for Ovumcy with the exact callback URL in `redirect_uris`.
3. Keep the client on the authorization code flow.
4. Use the public Authelia root URL as `OIDC_ISSUER_URL`, for example:
   `https://auth.example.com`
5. Keep the raw client secret for Ovumcy and the digest in the Authelia config. Do not paste the digest into `OIDC_CLIENT_SECRET`.

Recommended Ovumcy mapping:

```env
OIDC_ISSUER_URL=https://auth.example.com
OIDC_CLIENT_ID=ovumcy
OIDC_CLIENT_SECRET=the_raw_client_secret
OIDC_REDIRECT_URL=https://ovumcy.example.com/auth/oidc/callback
```

### ZITADEL

Recommended setup:

1. Create a new Web Application in your ZITADEL project.
2. Choose the authorization code flow with a client secret for a server-side web app.
3. Register the exact Ovumcy callback URL:
   `https://ovumcy.example.com/auth/oidc/callback`
4. Copy the client ID and generated secret.
5. Use your ZITADEL issuer root URL for `OIDC_ISSUER_URL`, for example:
   `https://auth.example.com`
6. Ensure the users signing in have verified email addresses that match existing Ovumcy accounts.

Recommended Ovumcy mapping:

```env
OIDC_ISSUER_URL=https://auth.example.com
OIDC_CLIENT_ID=replace_with_zitadel_client_id
OIDC_CLIENT_SECRET=replace_with_zitadel_client_secret
OIDC_REDIRECT_URL=https://ovumcy.example.com/auth/oidc/callback
```

## Rollout Checklist

Before enabling OIDC for real users:

1. Create or confirm an existing local Ovumcy account for your test user.
2. Confirm the provider sends the same email address and that the email is marked verified.
3. Keep local login enabled while testing the first OIDC login.
4. Test the callback URL over the same public hostname that users will actually open.
5. Test a fresh incognito login and a repeat login after the `(issuer, subject)` link has been created.
6. Confirm that logout still clears the Ovumcy session cookie as expected.

## Troubleshooting

### The SSO button does not appear

Check:

- `OIDC_ENABLED=true` is actually present in the running environment;
- startup did not reject another OIDC variable;
- the instance was restarted after changing env;
- you are looking at the same deployment profile you edited.

### Startup fails with `OIDC_ENABLED=true requires COOKIE_SECURE=true`

This is expected. OIDC currently requires secure cookies, so serve Ovumcy over HTTPS and set:

```env
COOKIE_SECURE=true
```

### Startup fails with `OIDC_REDIRECT_URL path must be /auth/oidc/callback`

The callback path is fixed in Ovumcy. Use the full public callback URL and do not invent a custom path:

```env
OIDC_REDIRECT_URL=https://ovumcy.example.com/auth/oidc/callback
```

### The provider rejects the callback URL

Typical causes:

- the provider client registration uses a different hostname;
- the provider has `http://` registered but Ovumcy is configured with `https://`;
- the provider has a trailing slash mismatch or a different callback path.

Fix it by making the provider's registered redirect URI exactly match `OIDC_REDIRECT_URL`.

### Clicking SSO returns to `/login` with a generic error

This usually means Ovumcy could not complete provider discovery or token exchange.

Check:

- `OIDC_ISSUER_URL` points to the real issuer, not to a login form URL;
- the provider is reachable from the Ovumcy host or container;
- the provider certificate chain is trusted by the Ovumcy runtime;
- reverse-proxy DNS and firewall rules allow Ovumcy to reach the provider.

### The account exists locally but OIDC still fails

The first OIDC login requires a verified email match.

Check:

- the provider actually sends an `email` claim;
- the provider marks that email as verified;
- the email string exactly matches the existing Ovumcy account email after normalization.

If the provider does not supply a verified email for that user, Ovumcy will not auto-link the account in the current release.

### Authentik fails with an issuer or discovery mismatch

For Authentik, the most common mistake is using the generic site URL instead of the application-specific issuer URL.

Prefer copying the issuer from the application's OpenID Configuration endpoint and use the issuer URL itself, not the full `/.well-known/openid-configuration` path.

### Authelia authentication fails even though the client exists

Double-check the client secret handling:

- Authelia stores a digest in its config;
- Ovumcy needs the raw secret value;
- `redirect_uris` in Authelia are case-sensitive and must exactly match the Ovumcy callback URL.

### A local lab provider uses a self-signed or private CA certificate

Ovumcy must trust the provider certificate chain for discovery and token exchange.

In containerized deployments, install or mount the internal CA into the container trust store instead of disabling TLS verification. If the CA is not trusted, the browser may reach the IdP while Ovumcy still fails server-to-server calls.

## Official Provider Documentation

For provider-specific UI details, use the current official docs:

- Keycloak: https://www.keycloak.org/docs/latest/server_admin/
- Keycloak OIDC redirect guidance: https://www.keycloak.org/securing-apps/oidc-layers
- Authentik OAuth2 / OIDC provider docs: https://docs.goauthentik.io/add-secure-apps/providers/oauth2/
- Authelia OIDC client config: https://www.authelia.com/configuration/identity-providers/openid-connect/clients/
- ZITADEL web app / OIDC examples: https://zitadel.com/docs/examples/identity-proxy/oauth2-proxy
