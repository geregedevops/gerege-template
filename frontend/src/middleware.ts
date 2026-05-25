import { NextResponse, type NextRequest } from 'next/server';
import { REFRESH_COOKIE } from '@/lib/cookies';

// Хамгаалагдсан хуудаснууд — refresh токен (durable session) байхгүй бол /login руу.
const PROTECTED = ['/profile', '/settings'];

// Тэмдэглэл: "нэвтэрсэн" хэрэглэгчийг /login, /register хуудаснаас буцаах
// AUTH_ONLY redirect-ийг зориудаар авч хаясан. Шалтгаан: refresh cookie нь
// browser-д 7 хоног үлддэг ч backend талаас (logout, password rotation,
// token cutoff) хүчингүйжсэн байж болно — энэ үед middleware "нэвтэрсэн"
// гэж буруу тооцож /login руу буцалт хийдгийг хаасан тул хэрэглэгч 307 loop-д
// гацаж нэвтэрч чаддаггүй байв. Нэвтэрсэн хэрэглэгч /login руу орвол ердөө
// маягт харагдана; нэвтэрмэгц setSession нь хуучин cookie-г шинээр дарж бичнэ.

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const signedIn = !!req.cookies.get(REFRESH_COOKIE)?.value;

  if (PROTECTED.some((p) => pathname.startsWith(p)) && !signedIn) {
    const url = req.nextUrl.clone();
    url.pathname = '/login';
    url.searchParams.set('next', pathname);
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  // Статик хөрөнгө, API route, Next дотоод замуудаас бусдыг л шалгана.
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico|brand.webp|theme-bootstrap.js).*)'],
};
