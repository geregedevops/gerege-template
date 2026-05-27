# Gerege Template

> 🌐 **English** · [Монгол](README_MN.md)

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-v3-00ACD7.svg)](https://gofiber.io/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black.svg)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A production-ready, security-hardened **full-stack template** built on Clean
Architecture. It pairs a Go (**Fiber v3 + GORM + PostgreSQL + Redis**) backend
with a Next.js (**BFF**) frontend, wired together and ready to extend into any
system.

## 📌 Origin & Open Source

The **backend** is derived from the open-source
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
(MIT, by Najib Fikri); we ported the HTTP layer **Gin → Fiber v3** and the data
layer **sqlx → GORM**, keeping the full feature set. Fiber v3 idioms were
cross-referenced against [rachmanzz/fiber-starter](https://github.com/rachmanzz/fiber-starter)
(MIT). All upstream copyright and license terms are honored — see [LICENSE](LICENSE),
[NOTICE](NOTICE), and [AUTHORS](AUTHORS). This project is **MIT-licensed**.

## Monorepo structure

```
gerege-template/
├── backend/    # Go · Fiber v3 · GORM · PostgreSQL · Redis · JWT/OTP auth
│   └── docs/   # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
└── frontend/   # Next.js BFF (server-side proxy to the backend; cookie sessions)
```

- **[backend/README.md](backend/README.md)** — Clean Architecture Go API.
- **[frontend/README.md](frontend/README.md)** — Next.js Backend-for-Frontend.

## Features

- **Clean Architecture** — `handler → usecase → repository → domain`, no back-imports; the business core never imports the web framework.
- **Auth** — JWT access + refresh (rotation), OTP-verified registration, bcrypt, login lockout.
- **Security-hardened** — strict security headers (CSP, HSTS, COOP/COEP/CORP), CORS allow-list, rate limiting, request timeouts, parameterized queries. See [SECURITY.md](SECURITY.md).
- **Observability** — OpenTelemetry tracing + Prometheus metrics + structured Zap logs.
- **Frontend BFF** — the browser talks only to same-origin Next.js routes, which proxy to the backend server-side (tokens never reach client JS).
- **Tested** — unit tests + testcontainers integration tests.

## Quick start

**Prerequisites:** Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+.

### 0) Rename the Go module (one-time)

The template ships under the placeholder module path `templatev27`. Replace it
with your own module path before doing anything else — every Go file imports
`templatev27/...`, so renaming early avoids a sed sweep later.

```bash
./scripts/rename-module.sh github.com/myorg/my-api
cd backend && go mod tidy && cd ..
```

### 1) Backend → http://localhost:8080

```bash
cd backend
cp internal/config/.env.example internal/config/.env   # set JWT_SECRET (≥32 chars), DB, Redis
make mig-up        # create schema
make serve
```

### 2) Frontend → http://localhost:3000

```bash
cd ../frontend
cp .env.example .env.local                              # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

Open **http://localhost:3000** and register / log in.

## Documentation

| Doc | What |
|-----|------|
| [backend/docs/ARCHITECTURE.md](backend/docs/ARCHITECTURE.md) | Layers, dependency flow, components |
| [backend/docs/DEVELOPMENT.md](backend/docs/DEVELOPMENT.md) | Add-a-feature guide, testing, code style |
| [backend/docs/API_CONTRACT.md](backend/docs/API_CONTRACT.md) | REST endpoints, request/response shapes |
| [backend/docs/SECURITY.md](backend/docs/SECURITY.md) | Implemented controls + ASVS roadmap |
| [SECURITY.md](SECURITY.md) | How to report a vulnerability |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |

## Contributing

Contributions are welcome — please read [CONTRIBUTING.md](CONTRIBUTING.md) and
the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE) — derivative of snykk/go-rest-boilerplate (MIT). Upstream notices
are retained in [LICENSE](LICENSE) and [NOTICE](NOTICE).

---

**Gerege Template Version 27.0** — Co-developed by the **Gerege Systems
Development Team** and **Claude AI**, 2026.
