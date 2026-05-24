import 'server-only';
import { NextResponse } from 'next/server';
import type { ApiResult } from './api';

// BFF route handler-уудын хуваалцсан туслахууд.

/** Request body-г аюулгүйгээр JSON болгож уншина. */
export async function readJson<T = Record<string, unknown>>(req: Request): Promise<T> {
  try {
    return (await req.json()) as T;
  } catch {
    return {} as T;
  }
}

/**
 * CSRF-ийн эсрэг хямд хамгаалалт (defense-in-depth). State-changing POST
 * route-ууд дээр `Origin` толгойг аппын өөрийн origin-той тулгана. Browser нь
 * fetch POST-д Origin-г үргэлж тавьдаг тул same-origin хүсэлт асуудалгүй
 * нэвтэрнэ. Зөрвөл 403 буцаах NextResponse-г, тааралцвал `null`-г буцаана.
 *
 * Аппын origin-г `APP_ORIGIN` env-ээс, эсвэл байхгүй бол хүсэлтийн өөрийн
 * URL-ээс гаргана. Origin толгойгүй (зарим non-browser хэрэгсэл) тохиолдолд
 * SameSite=Lax cookie-д найдаж зөвшөөрнө.
 */
export function checkOrigin(req: Request): NextResponse | null {
  const origin = req.headers.get('origin');
  if (!origin) return null; // Origin байхгүй — SameSite=Lax cookie хамгаална.

  const expected = process.env.APP_ORIGIN ?? new URL(req.url).origin;
  if (origin === expected) return null;

  return NextResponse.json(
    { ok: false, status: 403, message: 'Origin тохирохгүй байна.' },
    { status: 403 },
  );
}

/**
 * backend ApiResult-г browser рүү буцаах client хэлбэрт хувиргана. Токен зэрэг
 * нууц талбарыг хэзээ ч client рүү гаргахгүй — зөвхөн ok/status/message/fieldErrors.
 */
export function toClientResponse(r: ApiResult<unknown>): NextResponse {
  const httpStatus = r.ok ? 200 : r.status >= 400 && r.status < 600 ? r.status : 502;
  return NextResponse.json(
    {
      ok: r.ok,
      status: r.status,
      message: r.message,
      ...(r.ok ? {} : { fieldErrors: r.fieldErrors }),
    },
    { status: httpStatus },
  );
}
