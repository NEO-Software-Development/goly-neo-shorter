# Goly

Goly started as a URL shortener and now also hosts the **Company Contact Directory** — a privacy-first, API-driven way for companies to publish every way to reach them at a short, QR-friendly URL.

Stack: Go + Fiber, GORM + SQLite, Argon2id sessions, Svelte client.

---

## Quickstart

```bash
# install Go deps
go mod download

# pick a pepper and run
PASSWORD_PEPPER=$(openssl rand -hex 32) \
PUBLIC_BASE_URL=http://localhost:3000 \
go run ./goly
```

Server listens on `:3000`. A SQLite database `goly.db` is created on first run and the schema is auto-migrated.

---

## Configuration

| Env var | Required | Purpose |
|---|---|---|
| `PUBLIC_BASE_URL` | yes in prod | Canonical web origin used in generated QR codes (e.g. `https://goly.example.com`). Defaults to `http://localhost:3000` for dev. |
| `PASSWORD_PEPPER` | yes in prod | Server-wide secret HMAC'd into passwords before Argon2 hashing. A DB-only breach is uncrackable without this value. Use ≥ 32 random bytes. |
| `EMAIL_PROVIDER` | optional | Identifier for an SMTP/transactional provider. When unset, email verification tokens are echoed inline in API responses (dev mode). |
| `SMS_PROVIDER` | optional | Same as above for SMS-based contact verification. Unset = dev mode. |

---

## Company Contact Directory

### The product

One `Directory` per company is exposed through three concentric surfaces over a single canonical JSON API:

1. **Data plane** — versioned at `/api/v1/...`, consumable by any client (Svelte, mobile, embeds, partner sites).
2. **Distribution** — a short URL `/c/:slug` and a server-generated QR encoding it. The QR always points at a URL the platform controls.
3. **Owner control** — authenticated CRUD with draft/publish, per-link visibility, audit log, account export, and right-to-erasure.

### Public endpoints (no auth, identity-stripped logging)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/c/:slug` | Published directory + public links. Sensitive values (`phone`, `email`, `whatsapp`, `sms`) are omitted unless owner sets `reveal_mode: "inline"`. |
| GET | `/api/v1/c/:slug/links/:linkId/value` | Reveal-on-tap endpoint. Only returns the value when the link is published + public. `Cache-Control: private, no-store`. Rate-limited stricter than the list. |
| GET | `/api/v1/c/:slug/qr.png?size=256` | PNG QR encoding `${PUBLIC_BASE_URL}/c/:slug`. Size clamped to 64–1024. |
| GET | `/api/v1/c/:slug/vcard` | RFC 6350 vCard download. Public links only. |

All public responses include `Referrer-Policy: no-referrer` and `Permissions-Policy: camera=(), microphone=(), geolocation=(), interest-cohort=()`. When the directory is unpublished or `is_indexable=false`, `X-Robots-Tag: noindex, nofollow` is added.

### Owner endpoints (session auth required)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/directories` | List caller's directories |
| POST | `/api/v1/directories` | Create. Body: `{slug?, name, tagline, logo_url, accent_color}`. `slug` is optional; blank or taken → random 8-char slug. |
| GET | `/api/v1/directories/:id` | Get one (owner-only; 404 on cross-owner) |
| PATCH | `/api/v1/directories/:id` | Update |
| DELETE | `/api/v1/directories/:id` | Soft delete. `?hard=true` cascades hard-delete (GDPR). |
| POST | `/api/v1/directories/:id/publish` | Body: `{is_published: bool}` |
| GET | `/api/v1/directories/:id/audit` | Owner audit log for this directory |
| GET / POST | `/api/v1/directories/:id/links` | List / add contact link |
| PATCH / DELETE | `/api/v1/directories/:id/links/:linkId` | Update / delete |
| POST | `/api/v1/directories/:id/links/reorder` | Body: `{order: [{id, position}, ...]}` |
| POST | `/api/v1/directories/:id/links/:linkId/verify/start` | Issue a verification token addressed to the listed channel (email/phone/sms/whatsapp). In dev (`EMAIL_PROVIDER` / `SMS_PROVIDER` unset) the token is echoed in the response. |
| POST | `/api/v1/directories/:id/links/:linkId/verify/confirm` | Confirm token. Flips `verified_at` and surfaces `"verified": true` in the public payload. |
| GET | `/api/v1/me/export` | Data portability — full JSON of caller's directories + links. |
| DELETE | `/api/v1/me` | Right-to-erasure — hard-deletes the user, sessions, directories, links, audit entries, backup codes, verification tokens, in one transaction. Clears the session cookie. |

### Contact-link `kind` whitelist

`phone`, `email`, `website`, `whatsapp`, `telegram`, `signal`, `sms`, `linkedin`, `instagram`, `x`, `facebook`, `youtube`, `tiktok`, `github`, `address`, `custom`.

Per-kind validation:
- `phone`/`whatsapp`/`sms`: E.164 (`+` then 7–15 digits).
- `email`: RFC 5322 (`net/mail.ParseAddress`).
- URL-shaped kinds: must parse, must use `http` or `https` (rejects `javascript:`, `data:`, `file:`).
- `address` / `custom`: HTML-stripped via bluemonday, max 500 chars.

### Slug rules

3–40 chars from `[a-z0-9-]`, no leading/trailing hyphen, lowercase only. Reserved: `admin api auth c r goly login register logout me static assets health qr vcard well-known robots.txt favicon.ico settings directories directory`. Non-ASCII rejected by the regex, which thwarts confusables (`acmе` with Cyrillic 'е' ≠ `acme`).

---

## Auth

Session-based with HTTP-only, `Secure`, `SameSite=Lax` cookies. Tokens are 32 random bytes, 24-hour TTL.

### Endpoints

| Method | Path | Notes |
|---|---|---|
| POST | `/auth/register` | `{username, password, email}`. Email is required and validated via `net/mail.ParseAddress`. Rate limited 10/min/IP. |
| POST | `/auth/login` | `{username, password, otp?}`. Failures return a single `invalid_credentials` shape regardless of cause (unknown user vs wrong password vs missing OTP for a 2FA-enabled user is the only branch surfaced separately, via `totp_required: true`). |
| POST | `/auth/logout` | |
| POST | `/auth/2fa/enroll` | Generates a fresh TOTP secret + provisioning URL. |
| POST | `/auth/2fa/verify` | Confirms the secret and returns 10 single-use **backup codes** (shown once; only hashes stored). |
| POST | `/auth/2fa/disable` | Requires password + valid OTP/backup code. |
| POST | `/auth/email/start` | `{email}`. Issues a 30-minute verification token. Dev mode echoes the token. |
| POST | `/auth/email/confirm` | `{token}`. Sets `email_verified_at`. |
| GET | `/auth/sessions/` | List your active sessions with a coarse device label. |
| DELETE | `/auth/sessions/:id` | Revoke a single session. |

### What's hardened

- **Argon2id with random salt + server-wide pepper.** `PASSWORD_PEPPER` is HMAC-SHA256'd into the password before Argon2. A DB-only leak is unusable without the pepper.
- **Constant-time, ambiguous login failures.** Both "unknown user" and "wrong password" return the same status, body, and approximate latency (the unknown-user branch still performs an Argon2 hash against a dummy hash to flatten timing).
- **TOTP 2FA + 10 single-use backup codes** (hashed with SHA-256 at rest, constant-time scan on consume).
- **Coarse new-device detection** on login. UA-only fingerprint (no IP), HMAC'd with the pepper. New-device events emit an audit log entry / server log line; SMTP delivery is the next hook.
- **Password pepper, peppered device hash, peppered backup-code hash** all share the same secret rotation surface.

---

## Privacy

### Visitor side (anyone who hits `/c/:slug` or scans a QR)

- **No third-party assets.** The public surface intentionally has no Google Fonts / analytics / CDN scripts.
- **No cookies are set** on public reads. No IPs, user agents, or referers are persisted — the public route group uses a custom logger that records method/path/status/duration only.
- **`Referrer-Policy: no-referrer`** on the public page so a visitor's previous URL doesn't leak when they tap "Visit website".
- **`Permissions-Policy`** disables camera/microphone/geolocation and opts out of FLoC / Topics.
- **Reveal-on-tap by default** for `phone`, `email`, `whatsapp`, `sms`. The list response omits the value; a separate, rate-limited endpoint returns the value only after the visitor explicitly asks. Defeats bulk harvesters in one query. Owner can override per-link to `reveal_mode: "inline"`.
- **DNT / Sec-GPC honored.** View counter is skipped when either header is present. The counter itself stores only an aggregate `uint64` — no per-visitor row.
- **View debounce is in-memory.** SHA-256(slug|ip) is held for 5 minutes in a `sync.Map`, never persisted, so reloads don't inflate counts. The hash never reaches disk.

### Owner side

- **Public payload strips owner identity.** `owner_id`, owner username, and timestamps are never returned on public endpoints.
- **Email is required at signup** so the account has a recovery channel from day one; the verification flow flips `email_verified_at` once the owner confirms the token.
- **Right-to-erasure** via `DELETE /api/v1/me` cascades hard-deletes across directories, links, audit entries, sessions, backup codes, verification tokens, and the user record itself, in one transaction.
- **Data portability** via `GET /api/v1/me/export`.

---

## Defense in depth

| Layer | Mechanism |
|---|---|
| Schema | Per-kind validators; URL scheme whitelist; reserved-slug denylist; non-ASCII slug rejection (confusables defense). |
| Sanitization | bluemonday strict policy on every name/label/address/custom field. |
| Authn | Argon2id + pepper. 2FA via TOTP + single-use backup codes. |
| Authz | Every owner write resolves the resource and asserts `owner_id == session.user_id`. Cross-owner access returns 404, not 403, so slug existence doesn't leak. |
| Session | HttpOnly + Secure + SameSite=Lax. 24h TTL. Revocable from `/auth/sessions/`. |
| Transport | HSTS (1y, includeSubDomains). |
| Headers | CSP locked to `'self'`, frame-ancestors none, base-uri self, X-Content-Type-Options nosniff, X-Frame-Options DENY. |
| Rate limits | Public read 60/min, public reveal 15/min, owner writes 30/min/session, `/auth/login` + `/auth/register` 10/min/IP. |
| Bodies | Global 64 KiB body limit (Fiber config). |
| Input | All JSON parsing rejects malformed payloads. URL parsing rejects `javascript:`, `data:`, `file:`. |
| Logging | Public route group never logs IP, UA, or referer. |
| Storage | All passwords + backup codes + verification tokens stored only as hashes. The TOTP secret is the only reversible secret and is needed by the algorithm. |
| Audit | Every owner write produces an `audit_entries` row visible at `GET /api/v1/directories/:id/audit`. |

---

## Production checklist (to do before going live)

- [ ] Set `PASSWORD_PEPPER` (≥ 32 bytes, random) and back it up alongside, but separately from, the database. Rotating it invalidates all existing passwords.
- [ ] Set `PUBLIC_BASE_URL` to the canonical https origin.
- [ ] Replace SQLite with Postgres for concurrent-write durability. Swap the GORM dialect in `database/database.go`.
- [ ] Replace `AutoMigrate` with versioned migrations (`goose` or `golang-migrate`).
- [ ] Wire `EMAIL_PROVIDER` to a real transactional sender so email + contact verification stop echoing tokens.
- [ ] Wire `SMS_PROVIDER` for phone-channel verification.
- [ ] Add a CAPTCHA gate (Cloudflare Turnstile is privacy-respecting) on `/auth/register` and on `/api/v1/c/:slug/links/:linkId/value` if scraping becomes an issue.
- [ ] Add WebAuthn / passkeys as a stronger alternative to TOTP.
- [ ] DNS TXT domain-verification flow for premium "Verified Owner" badges.
- [ ] Backups encrypted at rest. Document the retention policy.
- [ ] Bug bounty / responsible disclosure address.

---

## Build

```bash
make build        # produces ./goly-app
make test         # current test (goly/model/goly_test.go) has a pre-existing build error unrelated to this branch
make docker-build # builds the container image from Dockerfile
```

---

:zap: Happy Coding!
