import { NextResponse, type NextRequest } from 'next/server';
import { REFRESH_COOKIE } from '@/lib/cookies';

// Хамгаалагдсан хуудаснууд — refresh токен (durable session) байхгүй бол /login руу.
const PROTECTED = ['/profile', '/settings'];

// CSP-г per-request nonce-тэй middleware-аас өгнө. `next.config.mjs`-ийн
// static `headers()` per-request nonce өгөх боломжгүй учир CSP-г тэндээс
// зөөв. Гол өөрчлөлт: script-src дотроос 'unsafe-inline' хасагдан зөвхөн
// nonce-той inline script зөвшөөрөгдөнө. Next.js нь "x-nonce" request
// header-аас уншиж өөрийн inline bootstrap script-уудад nonce
// автоматаар тавьдаг (RSC + app router).
//
// style-src-д 'unsafe-inline' хадгалагдсан хэвээр — апп `style=`
// attribute-ыг ашигладаг тул nonce-style боломжгүй; style hijack нь
// бодит XSS payload-аас хол хохирол багатай хязгаарлалттай.

// Тэмдэглэл: "нэвтэрсэн" хэрэглэгчийг /login, /register хуудаснаас буцаах
// AUTH_ONLY redirect-ийг зориудаар авч хаясан. Шалтгаан: refresh cookie нь
// browser-д 7 хоног үлддэг ч backend талаас (logout, password rotation,
// token cutoff) хүчингүйжсэн байж болно — энэ үед middleware "нэвтэрсэн"
// гэж буруу тооцож /login руу буцалт хийдгийг хаасан тул хэрэглэгч 307 loop-д
// гацаж нэвтэрч чаддаггүй байв. Нэвтэрсэн хэрэглэгч /login руу орвол ердөө
// маягт харагдана; нэвтэрмэгц setSession нь хуучин cookie-г шинээр дарж бичнэ.

function buildCSP(nonce: string): string {
  // strict-dynamic нь nonce-той скриптээс ачаалагдсан скриптүүдэд итгэхийг
  // зөвшөөрөх боловч одоохондоо Next.js бүх RSC chunk-уудыг nonce-р
  // тэмдэглэдэг тул strict-dynamic-гүйгээр л ажиллана. Ирээдүйд аль нэг
  // 3rd-party скрипт нэмэгдэх үед strict-dynamic нэмэхэд хялбар.
  return [
    "default-src 'self'",
    `script-src 'self' 'nonce-${nonce}'`,
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data:",
    "font-src 'self' data:",
    "connect-src 'self'",
    "frame-ancestors 'none'",
    "base-uri 'self'",
    "form-action 'self'",
  ].join('; ');
}

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const signedIn = !!req.cookies.get(REFRESH_COOKIE)?.value;

  if (PROTECTED.some((p) => pathname.startsWith(p)) && !signedIn) {
    const url = req.nextUrl.clone();
    url.pathname = '/login';
    url.searchParams.set('next', pathname);
    return NextResponse.redirect(url);
  }

  // Nonce-ийг crypto.randomUUID-ээр үүсгэнэ — Edge runtime support, 128 бит
  // санамсаргүй, бичигдсэн request бүрд цоо шинэ. Base64-урт байх албагүй,
  // зөвхөн CSP-д хүлээн зөвшөөрөгдөх character set-д багтах ёстой.
  const nonce = crypto.randomUUID().replace(/-/g, '');
  const csp = buildCSP(nonce);

  // Next.js нь "x-nonce" request header-ыг үндсэн скрипт chunk-ууддаа
  // автоматаар хэрэглэдэг. layout.tsx нь нэмэлт inline скрипт (theme-bootstrap)
  // дээр өөрөө унших боломжтой.
  const reqHeaders = new Headers(req.headers);
  reqHeaders.set('x-nonce', nonce);
  reqHeaders.set('content-security-policy', csp);

  const res = NextResponse.next({ request: { headers: reqHeaders } });
  res.headers.set('content-security-policy', csp);
  return res;
}

export const config = {
  // Статик хөрөнгө, API route, Next дотоод замуудаас бусдыг л шалгана.
  // API route нь өөрийн JSON хариу гаргадаг тул CSP хэрэггүй; статик файлууд
  // CSP-гүй ч аюулгүй.
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico|brand.webp|theme-bootstrap.js).*)'],
};
