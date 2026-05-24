# Gerege Template

> 🌐 [English](README.md) · **Монгол**

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-v3-00ACD7.svg)](https://gofiber.io/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black.svg)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Clean Architecture зарчмаар бүтээгдсэн, аюулгүй байдлыг хатууруулсан,
production-д бэлэн **full-stack template**. Go (**Fiber v3 + GORM + PostgreSQL +
Redis**) backend болон Next.js (**BFF**) frontend-ийг хослуулсан — хооронд нь
холбож, ямар ч систем рүү өргөтгөхөд бэлэн.

## 📌 Эх сурвалж ба нээлттэй эх

**Backend** нь нээлттэй эх
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
(MIT, Najib Fikri)-аас гаралтай; HTTP давхаргыг **Gin → Fiber v3**, өгөгдлийн
давхаргыг **sqlx → GORM** болгож хөрвүүлсэн, бүх фичерийг хадгалсан. Fiber v3-ийн
хэрэглээг [rachmanzz/fiber-starter](https://github.com/rachmanzz/fiber-starter)
(MIT)-ээс лавласан. Бүх эх төслийн зохиогчийн эрх, лицензийг хүндэтгэн хадгалсан
([LICENSE](LICENSE), [NOTICE](NOTICE), [AUTHORS](AUTHORS)). Энэ төсөл **MIT
лицензтэй**.

## Monorepo бүтэц

```
gerege-template/
├── backend/    # Go · Fiber v3 · GORM · PostgreSQL · Redis · JWT/OTP танилт
│   └── docs/   # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
└── frontend/   # Next.js BFF (backend руу server талаас прокси; cookie session)
```

- **[backend/README_MN.md](backend/README_MN.md)** — Clean Architecture Go API.
- **[frontend/README.md](frontend/README.md)** — Next.js Backend-for-Frontend.

## Онцлог

- **Clean Architecture** — `handler → usecase → repository → domain`, back-import байхгүй; business core нь web framework-ийг import хийдэггүй.
- **Танилт** — JWT access + refresh (rotation), OTP-баталгаажуулсан бүртгэл, bcrypt, login lockout.
- **Аюулгүй хатууруулсан** — security headers (CSP, HSTS, COOP/COEP/CORP), CORS allow-list, rate limiting, request timeout, parameterized query. [SECURITY.md](SECURITY.md)-г үз.
- **Observability** — OpenTelemetry trace + Prometheus metrics + Zap structured log.
- **Frontend BFF** — браузер зөвхөн ижил-origin Next.js route рүү залгаж, тэр нь server талаас backend руу проксиолдог (токен client JS-д хүрэхгүй).
- **Тесттэй** — unit + testcontainers integration тест.

## Түргэн эхлүүлэх

**Шаардлага:** Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+.

```bash
# 1) Backend  →  http://localhost:8080
cd backend
cp internal/config/.env.example internal/config/.env   # JWT_SECRET (≥32), DB, Redis тохируул
make mig-up
make serve

# 2) Frontend →  http://localhost:3000
cd ../frontend
cp .env.example .env.local                              # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

**http://localhost:3000** нээж бүртгүүлэх / нэвтрэх.

## Баримтжуулалт

| Doc | Юу |
|-----|------|
| [backend/docs/ARCHITECTURE_MN.md](backend/docs/ARCHITECTURE_MN.md) | Давхаргууд, dependency flow |
| [backend/docs/DEVELOPMENT_MN.md](backend/docs/DEVELOPMENT_MN.md) | Фичер нэмэх заавар, тест, code style |
| [backend/docs/API_CONTRACT_MN.md](backend/docs/API_CONTRACT_MN.md) | REST endpoint, request/response |
| [backend/docs/SECURITY.md](backend/docs/SECURITY.md) | Хэрэгжсэн хяналт + ASVS roadmap |
| [SECURITY.md](SECURITY.md) | Эмзэг байдлыг хэрхэн мэдээлэх |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Хэрхэн хувь нэмэр оруулах |

## Хувь нэмэр

Хувь нэмэр оруулахыг урьж байна — [CONTRIBUTING.md](CONTRIBUTING.md) болон
[Code of Conduct](CODE_OF_CONDUCT.md)-ийг уншина уу.

## Лиценз

[MIT](LICENSE) — snykk/go-rest-boilerplate (MIT)-ийн derivative. Эх төслийн
мэдэгдлийг [LICENSE](LICENSE), [NOTICE](NOTICE)-д хадгалсан.

---

**Gerege Template Version 27.0** — **Gerege Systems Development Team** болон
**Claude AI** хамтран бүтээв, 2026.
