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
 * CSRF-ийн эсрэг defense-in-depth. State-changing POST route-ууд дээр
 * `Origin` толгойг аппын өөрийн origin-той тулгана. Browser нь fetch POST-д
 * Origin-г үргэлж тавьдаг тул same-origin хүсэлт асуудалгүй нэвтэрнэ.
 *
 * Чанд бодлого:
 *   - Origin байхгүй: 403 (anti-CSRF defense-in-depth; SameSite=Strict дээр
 *     найдсан хуучин hedge-ийг зайлуулав).
 *   - production-д APP_ORIGIN env заавал тохируулсан байна; өөрөөс host
 *     header дээр найдах нь spoofed Host-той proxy-д сул юм.
 *   - dev (NODE_ENV!=='production') орчинд APP_ORIGIN заагаагүй бол
 *     request URL-ийн origin-ыг pragmatic default-аар авна.
 */
export function checkOrigin(req: Request): NextResponse | null {
  const origin = req.headers.get('origin');
  if (!origin) {
    return NextResponse.json(
      { ok: false, status: 403, message: 'Origin толгой шаардлагатай.' },
      { status: 403 },
    );
  }

  const configured = process.env.APP_ORIGIN;
  if (!configured && process.env.NODE_ENV === 'production') {
    return NextResponse.json(
      { ok: false, status: 500, message: 'Сервер тохиргоо дутуу: APP_ORIGIN env шаардлагатай.' },
      { status: 500 },
    );
  }
  const expected = configured ?? new URL(req.url).origin;
  if (origin === expected) return null;

  return NextResponse.json(
    { ok: false, status: 403, message: 'Origin тохирохгүй байна.' },
    { status: 403 },
  );
}

/**
 * Backend ApiResult-г browser рүү буцаах client хэлбэрт хувиргана. Токен зэрэг
 * нууц талбарыг хэзээ ч client рүү гаргахгүй — зөвхөн ok/status/message/fieldErrors.
 * 5xx алдааны нарийвчилсан мэдээллийг гадаад руу гаргахгүй (info leak).
 */
export function toClientResponse(r: ApiResult<unknown>): NextResponse {
  const httpStatus = r.ok ? 200 : r.status >= 400 && r.status < 600 ? r.status : 502;
  // 5xx үед backend-ийн дотоод алдааны мессежийг хааж, ерөнхий мессеж буцаана.
  // 4xx ба амжилт үед мессеж нь хэрэглэгчийн харах user-facing string гэж
  // үздэг (validation алдаа, "имэйл буруу" гэх мэт) — тэдгээрийг дамжуулна.
  const safeMessage =
    !r.ok && r.status >= 500
      ? 'Дотоод алдаа гарлаа. Дахин оролдоно уу.'
      : r.message;
  return NextResponse.json(
    {
      ok: r.ok,
      status: r.status,
      message: safeMessage,
      ...(r.ok ? {} : { fieldErrors: r.fieldErrors }),
    },
    { status: httpStatus },
  );
}

/**
 * BFF route-ийн body-аас шаардлагатай талбаруудыг шалгана. Алдаа байвал
 * 400 NextResponse-г, тааралцвал validated payload-г буцаана.
 *
 * Backend дахин validate хийнэ — энэ нь зөвхөн "цөөн миллисекундийн өмнө"
 * хог payload-ыг хаах хэмжээний хямд шалгалт юм (zod гэх мэт runtime
 * dep шаардахгүй).
 */
export function requireFields<T extends Record<string, unknown>>(
  body: T,
  fields: ReadonlyArray<keyof T>,
): NextResponse | null {
  const missing: string[] = [];
  for (const f of fields) {
    const v = body[f];
    if (v === undefined || v === null || (typeof v === 'string' && v.trim() === '')) {
      missing.push(String(f));
    }
  }
  if (missing.length === 0) return null;
  return NextResponse.json(
    {
      ok: false,
      status: 400,
      message: 'Шаардлагатай талбар дутуу байна.',
      fieldErrors: Object.fromEntries(missing.map((m) => [m, 'required'])),
    },
    { status: 400 },
  );
}
