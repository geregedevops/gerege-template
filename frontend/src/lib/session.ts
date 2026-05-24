import 'server-only';
import { cookies } from 'next/headers';
import {
  ACCESS_COOKIE, REFRESH_COOKIE,
  ACCESS_MAX_AGE, REFRESH_MAX_AGE, cookieOptions,
} from './cookies';

// Серверийн талд токен cookie-г унших / бичих / цэвэрлэх туслахууд.
// Зөвхөн route handler болон server component-аас дуудагдана.

export function getAccessToken(): string | undefined {
  return cookies().get(ACCESS_COOKIE)?.value;
}

export function getRefreshToken(): string | undefined {
  return cookies().get(REFRESH_COOKIE)?.value;
}

/** Нэвтрэлт / refresh-ийн дараа токен хосыг cookie-д суулгана. */
export function setSession(accessToken: string, refreshToken: string): void {
  const jar = cookies();
  jar.set(ACCESS_COOKIE, accessToken, cookieOptions(ACCESS_MAX_AGE));
  jar.set(REFRESH_COOKIE, refreshToken, cookieOptions(REFRESH_MAX_AGE));
}

/** Зөвхөн access токенг шинэчилнэ (refresh урсгалын дараа). */
export function setAccessToken(accessToken: string): void {
  cookies().set(ACCESS_COOKIE, accessToken, cookieOptions(ACCESS_MAX_AGE));
}

/** Гарах үед хоёр cookie-г устгана. */
export function clearSession(): void {
  const jar = cookies();
  jar.delete(ACCESS_COOKIE);
  jar.delete(REFRESH_COOKIE);
}

/** Refresh токен байгаа эсэх — "нэвтэрсэн" гэж тооцох durable сигнал. */
export function hasSession(): boolean {
  return !!getRefreshToken();
}
