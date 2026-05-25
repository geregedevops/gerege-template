# Security Posture — Gerege Backend Template v27

> 🌐 **English** · Монгол тайлбарыг кодын комментуудаас үзнэ үү. Эмзэг байдлыг
> мэдээлэх журмыг [`/SECURITY.md`](../../SECURITY.md)-аас үз.

This document maps the backend's implemented controls to the project security
standard — based on **OWASP ASVS / API Top 10, NIST SP 800-63B / 800-218, and
CIS Controls**. It records what is enforced in code, what was hardened, and
what remains for later phases. To report a vulnerability, see the repository
[security policy](../../SECURITY.md).

## Implemented controls (in code)

| Area | Control | Where | Guide § |
|------|---------|-------|---------|
| Auth | JWT access+refresh, rotation, `kind`-claim guard | `pkg/jwt`, `usecases/auth` | §1.3–1.4 |
| Auth | bcrypt (cost ≥12), password ≥12 + `strongpassword` | `domain.users.go`, `pkg/validators` | §1.1 |
| Auth | OTP-verified registration | `usecases/auth` (send/verify) | §1.5 |
| Auth | Login lockout + per-account rate limit | `usecases/auth`, `middleware.ratelimit` | §1.5 |
| Auth | Enumeration mitigation (timing-safe, generic msgs) | `usecases/auth.login`, `forgot_password` | §1.5 |
| Crypto | `crypto/rand` everywhere; OTP rejection-sampled (no modulo bias) | `pkg/helpers/helper.otp_code_generator.go` | §13.2 |
| AuthZ | Role check in domain (`IsAdmin`), per-request `CurrentUser` | `domain.users.go`, `http/auth` | §2 |
| DB | Parameterized queries only (GORM) | `datasources/repositories/postgres` | §3.1 |
| DB | `INSERT … RETURNING` single round-trip; `TranslateError` | `users.store.go`, `driver.gorm.go` | §3 |
| DB | Row-Level Security on `users` (ENABLE + **FORCE**): self/admin/service policies driven by `app.user_id`/`app.user_role` GUCs set per-transaction with `SET LOCAL` | `migrations/6_enable_rls_users.up.sql`, `datasources/rls`, `repositories/postgres/users` | §2.4/§3.3 |
| API | Mass-assignment safe (explicit request DTOs) | `http/datatransfers/requests` | API3 §5.1 |
| API | Body size limit (global + 4 KiB on `/auth`) | `middleware.bodysizelimit`, `routes` | §5.3 |
| Web | Security headers: CSP `default-src 'none'`, HSTS (prod), nosniff, X-Frame DENY, Referrer-Policy, Permissions-Policy | `middleware.security.go` | §4.7 |
| Web | CORS strict origin list, never `*`+credentials | `middleware.cors.go` | §4.8 |
| Obs | Structured Zap logs w/ request-id; no secrets logged | `pkg/logger`, `handler.base_response.go` | §9.1–9.2 |
| Obs | OpenTelemetry tracing + Prometheus metrics | `pkg/observability`, `driver.gorm.go` | §9.4 |
| Ops | Graceful shutdown (drain HTTP, mailer, DB, Redis, tracer) | `cmd/api/server` | §7 |

## Hardening applied (this pass — against the guide)

1. **Cross-origin isolation headers** — added `Cross-Origin-Opener-Policy: same-origin`,
   `Cross-Origin-Resource-Policy: same-site`, `Cross-Origin-Embedder-Policy: require-corp`
   to `middleware.security.go` (guide §4.6/4.7). *Verified live in the running server.*
2. **Production DB TLS guard** — config validation now rejects a production
   `DB_POSTGRE_URL` unless `sslmode=verify-full` (or `verify-ca`); `.env.example`
   documents it (`internal/config/config.go`, guide §3.5).
3. **Per-request timeout** — `middleware.TimeoutMiddleware` sets a 30s context
   deadline that propagates to GORM queries, bounding stuck handlers
   (`middleware.timeout.go`, guide §5.3 / API4).
4. **Swagger served Fiber-v3-natively** — replaced the Fiber-v2-only
   `gofiber/swagger` handler (which panicked at runtime under Fiber v3) with a
   native `/swagger/doc.json` OpenAPI endpoint.

## ASVS roadmap status (guide §14)

- **Phase 1 (ASVS L1):** ✅ HTTPS-ready + HSTS, bcrypt, parameterized queries,
  security headers, strict CORS, input validation, structured logging, `.gitignore`
  + no committed secrets. ⏳ container scan / `govulncheck` wired in CI (`.github/`).
- **Phase 2 (ASVS L2):** ✅ rate limiting, refresh-token rotation, OTP MFA-style
  verification, request timeout. ⏳ leaked-password (HIBP k-anonymity, §1.1),
  WAF, centralized SIEM, encrypted-backup restore test, IR plan.
- **Phase 3 (ASVS L3):** ◻ WebAuthn/passkeys, field-level PII encryption (KMS),
  mTLS, SLSA L3 provenance, external pentest. (Out of template scope.)

## Known gaps / follow-ups

- **Interactive Swagger UI** — currently serves the raw spec at `/swagger/doc.json`
  (load it in Swagger Editor / Postman). A Fiber-v3-compatible UI handler can be
  added later.
- **Leaked-password check (HIBP)** — guide §1.1; not yet wired (needs outbound
  call, config-gated, fail-open). Password story already meets the OWASP baseline
  (bcrypt cost 12 + ≥12 chars + complexity).
- **Postgres RLS** (guide §2.4/§3.3) — ✅ enabled **and FORCED** on `users`
  with self/admin/service policies driven by the `app.user_id` / `app.user_role`
  session GUCs (set per-transaction via `SET LOCAL` in
  `repositories/postgres/users/users.postgres.go`). This is defense-in-depth on
  top of the WHERE clauses the repository already writes. A request with no
  identity is denied by default (fail-closed). To go **multi-tenant**, add a
  `tenant_id` column + a tenant policy to each new table and carry the tenant in
  the same `rls.Identity`.
- **Secrets manager / KMS** (guide §7.3) — use a real secret store in production;
  `.env` is local-dev only and gitignored.
- **DB role separation** (guide §3.4) — ✅ **wired into the compose stack** (it
  is required: RLS, even `FORCE`d, is bypassed by **superusers** / `BYPASSRLS`
  roles, and the postgres image makes `POSTGRES_USER` a superuser). On first DB
  init, `backend/deploy/initdb/10-create-app-user.sh` creates a **non-superuser**
  role `APP_DB_USER` (`NOSUPERUSER NOBYPASSRLS`) and grants it DML via default
  privileges. The **api** connects as that role (compose overrides
  `DB_POSTGRE_DSN`/`URL`), so the RLS policies are enforced; the **migrate**
  container keeps using `POSTGRES_USER` because it needs superuser for
  `CREATE EXTENSION "uuid-ossp"`.

  Sanity check from the api's connection:
  `SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user;`
  — both must be `false`. Verified: with the non-superuser role the
  self/admin/service policies enforce correctly and a no-identity request is
  fail-closed; a superuser would see all rows (bypass). If `APP_DB_USER` is left
  unset the app falls back to the superuser and RLS is *not* enforced.

---

**Gerege Template Version 27.0** — Co-developed by the **Gerege Systems
Development Team** and **Claude AI**, 2026.
