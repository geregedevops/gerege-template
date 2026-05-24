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
