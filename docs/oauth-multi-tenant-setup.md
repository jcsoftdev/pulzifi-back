# OAuth — Multi-Tenant Setup

How Google / GitHub OAuth works in a subdomain-per-tenant deployment (e.g. `acme.pulzifi.com`, `foo.pulzifi.com`), and how to configure the OAuth providers ONCE for the whole platform.

---

## Table of Contents

1. [Architecture](#architecture)
2. [The BFF nonce flow](#the-bff-nonce-flow)
3. [Google OAuth setup](#google-oauth-setup)
4. [GitHub OAuth setup](#github-oauth-setup)
5. [Integration providers (Sheets / Teams / Slack / Discord)](#integration-providers)
6. [Environment variables](#environment-variables)
7. [Why no wildcards](#why-no-wildcards)
8. [Code references](#code-references)

---

## Architecture

Login lives on the ROOT domain (`pulzifi.com`), not on tenant subdomains. After successful OAuth, the backend issues a one-time `nonce` and redirects the user to their tenant subdomain (`<sub>.pulzifi.com/?nonce=...`). The frontend on the tenant subdomain exchanges the nonce for cookies via the BFF endpoint.

```
                ┌─────────────────────┐
                │  pulzifi.com/login  │  ← single login surface
                └──────────┬──────────┘
                           │
                           ▼
        ┌────────────────────────────────────┐
        │ pulzifi.com/api/v1/auth/oauth/...  │  ← single redirect URI
        └──────────┬─────────────────────────┘
                   │
          ┌────────▼────────┐
          │ Google / GitHub │
          └────────┬────────┘
                   │ callback
                   ▼
        ┌────────────────────────────────────┐
        │ pulzifi.com/api/v1/auth/oauth/.../ │
        │ callback                            │
        └──────────┬─────────────────────────┘
                   │ resolve org → mint nonce
                   ▼
        ┌────────────────────────────────────┐
        │  acme.pulzifi.com/?nonce=xyz       │  ← tenant
        └──────────┬─────────────────────────┘
                   │ BFF exchanges nonce → cookies
                   ▼
                logged in
```

Net effect: ONE OAuth client per provider in the platform, regardless of how many tenants exist.

---

## The BFF nonce flow

| Step | Endpoint | What happens |
|---|---|---|
| 1 | `pulzifi.com/login` | User clicks "Sign in with Google" |
| 2 | `GET pulzifi.com/api/v1/auth/oauth/google` | Backend sets `oauth_state` cookie with `Domain=.pulzifi.com` and redirects to Google with `redirect_uri=https://pulzifi.com/api/v1/auth/oauth/google/callback` |
| 3 | Google consent | User approves |
| 4 | `GET pulzifi.com/api/v1/auth/oauth/google/callback?code&state` | Backend validates state, exchanges code, upserts user, looks up org's subdomain |
| 5 | New user, no org | Redirect to `pulzifi.com/onboarding` (no nonce) |
| 6 | Existing user with org | Mint nonce, redirect to `<sub>.pulzifi.com/?nonce=...` |
| 7 | `<sub>.pulzifi.com` | Frontend calls `POST /api/auth/set-base-session` with the nonce → backend sets cookies scoped `Domain=.pulzifi.com` |

`COOKIE_DOMAIN=.pulzifi.com` is REQUIRED — without it, the `oauth_state` cookie set on root is not readable on the callback (bug history: see `handleOAuthRedirect` + `handleOAuthCallback`). Same env enables session cookies to span tenant subdomains.

---

## Google OAuth setup

Single OAuth client in Google Cloud Console, used for ALL tenants.

### 1. Create the OAuth client

[Google Cloud Console](https://console.cloud.google.com) → APIs & Services → Credentials → Create credentials → OAuth client ID:

- Application type: **Web application**
- Name: `Pulzifi`
- Authorized JavaScript origins:
  - `https://pulzifi.com`
  - (dev) `http://pulzifi.lvh.me:3002`
- Authorized redirect URIs:
  - `https://pulzifi.com/api/v1/auth/oauth/google/callback`
  - (dev) `http://pulzifi.lvh.me:3002/api/v1/auth/oauth/google/callback`

Tenant subdomains (`acme.pulzifi.com`, etc.) DO NOT belong in either list — Google never sees them.

### 2. OAuth consent screen

APIs & Services → OAuth consent screen:

- User type: **External**
- App name: `Pulzifi`
- Support email, logo
- App domain: `pulzifi.com`
- Authorized domain: `pulzifi.com`
- Privacy policy URL: `https://pulzifi.com/privacy`
- Terms of service URL: `https://pulzifi.com/terms`
- Scopes: `openid`, `email`, `profile` (no Drive/Sheets here — that's the Integrations OAuth app)
- **Publish app** (move out of Testing). Otherwise only 100 test users can log in.

### 3. Env vars

```bash
GOOGLE_CLIENT_ID=...apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-...
OAUTH_REDIRECT_BASE_URL=https://pulzifi.com
FRONTEND_URL=https://pulzifi.com
COOKIE_DOMAIN=.pulzifi.com
```

---

## GitHub OAuth setup

GitHub → Settings → Developer settings → OAuth Apps → New OAuth App:

- Homepage URL: `https://pulzifi.com`
- Authorization callback URL: `https://pulzifi.com/api/v1/auth/oauth/github/callback`

GitHub only accepts ONE callback URL per app — multi-tenant works because we never hit subdomains during the OAuth dance.

### Env vars

```bash
GITHUB_CLIENT_ID=Iv1.xxxxxxxx
GITHUB_CLIENT_SECRET=...
# OAUTH_REDIRECT_BASE_URL shared with Google
```

---

## Integration providers

Slack / Discord / Google Sheets / Microsoft Teams use a SEPARATE OAuth surface — `modules/integration` — but follow the same root-domain pattern:

```
INTEGRATION_OAUTH_REDIRECT_BASE=https://pulzifi.com
```

Callback URIs registered in each provider:

| Provider | Redirect URI |
|---|---|
| Slack | `https://pulzifi.com/api/v1/integrations/oauth/slack/callback` |
| Discord | `https://pulzifi.com/api/v1/integrations/oauth/discord/callback` |
| Google Sheets | `https://pulzifi.com/api/v1/integrations/oauth/sheets/callback` |
| Microsoft Teams | `https://pulzifi.com/api/v1/integrations/oauth/teams/callback` |

Tokens are stored encrypted (AES-256-GCM via `shared/crypto/`) per-tenant in `integrations.provider_meta`. Tenant scope is enforced server-side after OAuth completes — not via per-tenant OAuth apps.

Full provider walkthroughs: `docs/integrations-setup.md`.

---

## Environment variables

```bash
# Login OAuth (modules/auth)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
OAUTH_REDIRECT_BASE_URL=https://pulzifi.com

# Integrations OAuth (modules/integration)
INTEGRATION_OAUTH_REDIRECT_BASE=https://pulzifi.com
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
SHEETS_CLIENT_ID=
SHEETS_CLIENT_SECRET=
TEAMS_CLIENT_ID=
TEAMS_CLIENT_SECRET=

# Cross-subdomain auth
FRONTEND_URL=https://pulzifi.com
COOKIE_DOMAIN=.pulzifi.com
NEXT_PUBLIC_APP_DOMAIN=pulzifi.com
NEXT_PUBLIC_APP_BASE_URL=https://pulzifi.com
```

---

## Why no wildcards

Neither Google nor GitHub supports wildcard redirect URIs (`*.pulzifi.com/...`). The platform works around this by:

1. Hosting the login surface on a single root domain.
2. Using the BFF nonce flow to hand off the session to the correct tenant subdomain AFTER the OAuth handshake completes.

Adding a tenant requires zero changes in Google Cloud Console or GitHub.

---

## Code references

| Area | File |
|---|---|
| OAuth redirect entry | `modules/auth/infrastructure/http/module.go` → `handleOAuthRedirect` |
| OAuth callback + nonce mint | `modules/auth/infrastructure/http/module.go` → `handleOAuthCallback`, `buildTenantRedirectURL` |
| Google provider | `modules/auth/infrastructure/oauth/google_provider.go` |
| GitHub provider | `modules/auth/infrastructure/oauth/github_provider.go` |
| Nonce store | `shared/noncestore/` |
| BFF cookie exchange | `shared/bff/handler.go` → `set-base-session` |
| Public path allowlist | `shared/middleware/tenant.go` → `publicPaths` |
| Integrations OAuth | `modules/integration/infrastructure/http/` |
