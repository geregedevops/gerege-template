// Cookie тогтмолууд ба сонголтууд. BFF загварт токенуудыг httpOnly cookie-д
// хадгалдаг тул browser-ийн JS тэдгээрийг хэзээ ч уншихгүй (XSS-д тэсвэртэй).

export const ACCESS_COOKIE = 'gerege_access';
export const REFRESH_COOKIE = 'gerege_refresh';

// Cookie-ийн насжилт. Backend-ийн анхдагч: JWT_EXPIRED=5 цаг, JWT_REFRESH_EXPIRED=7 хоног.
// Эдгээрийг backend-ийн тохиргоотой ойролцоо барина — хэтэрсэн access cookie-г
// refresh урсгал шинэчилнэ.
export const ACCESS_MAX_AGE = 60 * 60 * 5; // 5 цаг (секундээр)
export const REFRESH_MAX_AGE = 60 * 60 * 24 * 7; // 7 хоног (секундээр)

// COOKIE_SECURE парс — fail-closed:
//   - яг "false" гэж бичсэн үед л Secure-гүй (зөвхөн дотоод http dev).
//   - өөр аливаа утга (хоосон, "0", "no", "False") нь NODE_ENV-ээс хамаарна:
//     production бол Secure, бусад үед Secure биш.
//   - "true" ба бусад truthy утгууд Secure-г сонгоно (production биш ч).
// Энэ нь "COOKIE_SECURE=0 гэж буруу бичсэн → silently insecure" эрсдэлээс
// сэргийлнэ.
function isSecureCookie(): boolean {
  const v = process.env.COOKIE_SECURE;
  if (v === 'false') return false;
  if (v === 'true') return true;
  return process.env.NODE_ENV === 'production';
}

/** Токен cookie-д хэрэглэх стандарт httpOnly сонголтууд. */
export function cookieOptions(maxAge: number) {
  return {
    httpOnly: true,
    secure: isSecureCookie(),
    // SameSite=Strict — токен cookie-уудыг гадаад navigation-аас (top-level
    // POST, link click) огт илгээхгүй. Энэ нь Origin check + nonce CSP-тэй
    // хослуулсан defense-in-depth. Strict дотор anchor-аас орж ирсэн
    // хэрэглэгчийг "нэвтэрсэн" эсэх хайдаг middleware байхгүй учир UX
    // ердийн (re-сайтаас редирект болж буцах flow огт байхгүй).
    sameSite: 'strict' as const,
    path: '/',
    maxAge,
  };
}
