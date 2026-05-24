# gerege-web

[`gerege-backend-template-v27`](../gerege-backend-template-v27)-д тохирсон **Next.js** frontend.
Дизайныг [`me.gerege.mn`](../me.gerege.mn)-ээс хуулбарласан **gerege theme** дээр суурилсан.

## Архитектур — BFF (Backend-for-Frontend)

```
Browser ──(адил origin)──► Next.js (route handlers /api/auth/*) ──(server→server)──► Go API /api/v1
   ▲                              │
   └── httpOnly cookie (токен) ◄──┘
```

- **Токен browser-т хэзээ ч ил гарахгүй.** Access/refresh JWT-г `httpOnly` cookie-д
  (`gerege_access`, `gerege_refresh`) хадгална → XSS-д тэсвэртэй.
- **Browser↔Go хооронд CORS хэрэггүй.** Browser зөвхөн Next.js рүү (адил origin)
  хандана; Go API руу зөвхөн Next.js server прокси хийнэ.
- **Reactive refresh.** Хамгаалагдсан дуудлага `401` авбал refresh токеноор нэг
  удаа автоматаар шинэчилж, дахин оролдоно (`src/lib/api.ts`).

## Хуудаснууд

| Зам | Тайлбар | Backend endpoint |
|-----|---------|------------------|
| `/` | Landing (анон) / Хяналтын самбар (нэвтэрсэн) | `GET /users/me` |
| `/login` | Нэвтрэх | `POST /auth/login` |
| `/register` | Бүртгүүлэх | `POST /auth/register` |
| `/verify-otp` | OTP-аар бүртгэл идэвхжүүлэх | `POST /auth/send-otp`, `/auth/verify-otp` |
| `/forgot-password` | Нууц үг сэргээх хүсэлт | `POST /auth/password/forgot` |
| `/reset-password` | Токеноор нууц үг шинэчлэх | `POST /auth/password/reset` |
| `/profile` 🔒 | Профайл (read-only) | `GET /users/me` |
| `/settings` 🔒 | Нууц үг солих + гарах | `PUT /auth/password/change`, `POST /auth/logout` |

🔒 = `src/middleware.ts`-аар хамгаалагдсан (refresh cookie байхгүй бол `/login` руу).

## Идэвхжүүлэлтийн урсгал

Backend нь бүртгүүлсэн хэрэглэгчийг **идэвхгүй** үүсгэдэг:
`register` → `send-otp` → `verify-otp` (идэвхжүүлнэ) → `login`.
`/register` амжилттай бол `/verify-otp` руу шилжиж кодыг автоматаар илгээнэ.

## Ажиллуулах

```bash
# 1) Backend-ийг асаа (өөр терминал дээр)
cd ../gerege-backend-template-v27 && make run    # http://localhost:8080

# 2) Орчны хувьсагч
cp .env.example .env.local       # шаардлагатай бол BACKEND_URL-ийг засна

# 3) Frontend
npm install
npm run dev                      # http://localhost:3001
```

| Хувьсагч | Анхдагч | Тайлбар |
|----------|---------|---------|
| `BACKEND_URL` | `http://localhost:8080` | Go API-ийн суурь (api/v1 угтваргүй). Зөвхөн server тал уншина. |
| `COOKIE_SECURE` | `false` | Production (HTTPS) дээр `true` болго. |

## gerege theme

Дизайн систем `src/app/globals.css` дотор — OKLCH токен (DAN blue `#1767E7`),
гэгээн/харанхуй загвар, Inter + JetBrains Mono фонт. Загвар/хэлийн сонголт
`localStorage` (`gerege.theme` / `gerege.lang`)-д хадгалагдаж, FOUC-аас сэргийлэх
`public/theme-bootstrap.js`-ээр render-ийн өмнө тусгагдана.

## Бүтэц

```
src/
  app/
    api/auth/*/route.ts   # BFF прокси (login, register, otp, logout, …)
    (pages)/page.tsx      # хуудас бүр server component + client form
    layout.tsx, globals.css
  components/             # AppShell, SigninShell, UserMenu, PasswordField, …
  lib/
    api.ts                # server→Go fetch + reactive refresh
    session.ts, cookies.ts# httpOnly cookie менежмент
    client.ts             # browser→BFF fetch
    password.ts           # нууц үгийн хүч (тохируулж болно)
    format.ts, preferences.ts, types.ts
  middleware.ts           # route хамгаалалт
```
