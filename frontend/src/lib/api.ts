import 'server-only';
import type { Envelope, BackendUser, MeData, ValidationData, SessionUser } from './types';
import { toSessionUser } from './types';
import { getAccessToken, getRefreshToken, setSession } from './session';

// Серверийн талд gerege-backend-template-v27 рүү хандах цорын ганц цэг.
// Browser энд хэзээ ч хүрэхгүй — зөвхөн route handler ба server component.

const BASE = (process.env.BACKEND_URL ?? 'http://localhost:8080').replace(/\/$/, '') + '/api/v1';

export type ApiOk<T> = { ok: true; status: number; message?: string; data?: T };
export type ApiErr = { ok: false; status: number; message: string; fieldErrors?: Record<string, string> };
export type ApiResult<T> = ApiOk<T> | ApiErr;

/** Дугтуйг тайлж, нэгдсэн ApiResult болгож буцаах суурь fetch. */
export async function backendFetch<T>(path: string, init?: RequestInit): Promise<ApiResult<T>> {
  let res: Response;
  try {
    res = await fetch(BASE + path, {
      ...init,
      cache: 'no-store',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json', ...init?.headers },
    });
  } catch {
    return {
      ok: false,
      status: 503,
      message: 'Backend-тэй холбогдож чадсангүй. gerege-backend-template-v27 ажиллаж байгаа эсэхийг шалгана уу.',
    };
  }

  let body: Envelope<T> | null = null;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    /* хариу JSON биш (жишээ нь 502) — доор статусаар шийднэ */
  }

  // 2xx-г амжилттай гэж үзнэ. Хоосон body (204 эсвэл задлагдаагүй JSON →
  // body=null) бол status талбар шаардахгүй. Зөвхөн дугтуйд `status` boolean
  // тодорхой байгаа үед л `status:false`-г алдаа гэж тооцно.
  if (res.ok && (body === null || body.status !== false)) {
    return { ok: true, status: res.status, message: body?.message, data: body?.data };
  }

  const fieldErrors = (body?.data as ValidationData | undefined)?.errors;
  return {
    ok: false,
    status: res.status,
    message: body?.message ?? `Хүсэлт амжилтгүй (${res.status})`,
    fieldErrors,
  };
}

/** Refresh токеноор шинэ access токен авах. Амжилттай бол шинэ токенг буцаана. */
async function tryRefresh(): Promise<string | null> {
  const refresh = getRefreshToken();
  if (!refresh) return null;
  const r = await backendFetch<BackendUser>('/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refresh }),
  });
  if (r.ok && r.data?.token && r.data?.refresh_token) {
    // Server component-ийн render үед cookie бичих боломжгүй (зөвхөн route
    // handler / server action). Тэр тохиолдолд алдааг залгиад, шинэ токеныг
    // зөвхөн энэ хүсэлтэд санах ойд ашиглана.
    try {
      setSession(r.data.token, r.data.refresh_token);
    } catch {
      /* RSC render — cookie бичих боломжгүй */
    }
    return r.data.token;
  }
  return null;
}

/**
 * Bearer токен хавсаргаж, 401 ирвэл нэг удаа reactive refresh хийгээд дахин
 * оролддог хамгаалагдсан дуудлага. Refresh бүтэлгүйтвэл анхны 401-г буцаана.
 */
export async function authedFetch<T>(path: string, init?: RequestInit): Promise<ApiResult<T>> {
  const withAuth = (token?: string) =>
    backendFetch<T>(path, {
      ...init,
      headers: { ...(token ? { Authorization: `Bearer ${token}` } : {}), ...init?.headers },
    });

  const res = await withAuth(getAccessToken());
  if (res.ok || res.status !== 401) return res;

  const newToken = await tryRefresh();
  if (!newToken) return res;
  return withAuth(newToken);
}

/** GET /users/me — нэвтэрсэн хэрэглэгчийн профайл, эсвэл null. */
export async function fetchMe(): Promise<SessionUser | null> {
  const r = await authedFetch<MeData>('/users/me', { method: 'GET' });
  if (r.ok && r.data?.user) return toSessionUser(r.data.user);
  return null;
}
