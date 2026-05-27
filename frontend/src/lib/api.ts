import 'server-only';
import type { Envelope, BackendUser, MeData, ValidationData, SessionUser } from './types';
import { toSessionUser } from './types';
import { getAccessToken, getRefreshToken, setSession } from './session';

// Серверийн талд gerege-backend-template-v27 рүү хандах цорын ганц цэг.
// Browser энд хэзээ ч хүрэхгүй — зөвхөн route handler ба server component.

// BACKEND_URL-ийг startup үед нэг удаа resolve хийж, scheme-ийг хатуу
// шалгана. Production-д private IP / loopback зөвшөөрөхгүй (SSRF guard) —
// compose дотор service нэрээр (жнь "api") холбогддог тул IP literal-ыг
// тусгай DISABLE_BACKEND_URL_GUARD=true env-ээр л давах боломжтой.
function resolveBackendBase(): string {
  const raw = process.env.BACKEND_URL ?? 'http://localhost:8080';
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    throw new Error(`Invalid BACKEND_URL: ${raw}`);
  }
  if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
    throw new Error(`BACKEND_URL must use http(s) scheme, got ${parsed.protocol}`);
  }
  const isProd = process.env.NODE_ENV === 'production';
  const allowGuardBypass = process.env.DISABLE_BACKEND_URL_GUARD === 'true';
  if (isProd && !allowGuardBypass && isPrivateHost(parsed.hostname)) {
    throw new Error(
      `BACKEND_URL points to a private/loopback host (${parsed.hostname}) in production. ` +
        `Use a service hostname or set DISABLE_BACKEND_URL_GUARD=true if this is intentional.`,
    );
  }
  return parsed.toString().replace(/\/$/, '') + '/api/v1';
}

function isPrivateHost(host: string): boolean {
  if (host === 'localhost') return true;
  // IPv4 literal-ыг шалгана (compose service нэр шиг "api"-г private гэж тооцохгүй).
  const v4 = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/.exec(host);
  if (v4) {
    const [a, b] = [Number(v4[1]), Number(v4[2])];
    if (a === 10) return true;
    if (a === 127) return true;
    if (a === 169 && b === 254) return true; // link-local
    if (a === 172 && b >= 16 && b <= 31) return true;
    if (a === 192 && b === 168) return true;
    if (a === 0) return true;
  }
  // IPv6 loopback / unique-local
  if (host === '::1' || host.startsWith('[::1') || host.startsWith('fc') || host.startsWith('fd'))
    return true;
  return false;
}

// BASE-ийг функц байдлаар экспозлоно — lazy resolution-ийг хадгалахын тулд.
const baseURL = () => resolveBackendBase();

// Backend дуудлагын timeout — backend санамсаргүй удааширсан тохиолдолд
// BFF unbounded hang болохоос сэргийлнэ. Refresh дуудлагад тусдаа богино
// timeout ашиглана.
const DEFAULT_TIMEOUT_MS = 15_000;
const REFRESH_TIMEOUT_MS = 8_000;

export type ApiOk<T> = { ok: true; status: number; message?: string; data?: T };
export type ApiErr = { ok: false; status: number; message: string; fieldErrors?: Record<string, string> };
export type ApiResult<T> = ApiOk<T> | ApiErr;

/** Дугтуйг тайлж, нэгдсэн ApiResult болгож буцаах суурь fetch. */
export async function backendFetch<T>(
  path: string,
  init?: RequestInit,
  opts?: { timeoutMs?: number },
): Promise<ApiResult<T>> {
  const controller = new AbortController();
  const timeoutMs = opts?.timeoutMs ?? DEFAULT_TIMEOUT_MS;
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  let res: Response;
  try {
    res = await fetch(baseURL() + path, {
      ...init,
      signal: controller.signal,
      cache: 'no-store',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json', ...init?.headers },
    });
  } catch (e: unknown) {
    const aborted = (e as { name?: string } | null)?.name === 'AbortError';
    return {
      ok: false,
      status: aborted ? 504 : 503,
      message: aborted
        ? 'Backend хариу хэт удаашрав.'
        : 'Backend-тэй холбогдож чадсангүй. gerege-backend-template-v27 ажиллаж байгаа эсэхийг шалгана уу.',
    };
  } finally {
    clearTimeout(timer);
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

// Concurrent refresh race-аас сэргийлэх module-scope lock. Зэрэг олон
// authedFetch 401 авбал зөвхөн нэг refresh дуудлага явуулна; үлдсэн нь
// ижил promise-ийг хүлээнэ. Энэ нь backend-ийн refresh reuse-detection-ийг
// false-positive trigger хийхгүй болгоно.
let refreshInFlight: Promise<string | null> | null = null;

/** Refresh токеноор шинэ access токен авах. Амжилттай бол шинэ токенг буцаана. */
async function tryRefresh(): Promise<string | null> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    try {
      const refresh = getRefreshToken();
      if (!refresh) return null;
      const r = await backendFetch<BackendUser>(
        '/auth/refresh',
        {
          method: 'POST',
          body: JSON.stringify({ refresh_token: refresh }),
        },
        { timeoutMs: REFRESH_TIMEOUT_MS },
      );
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
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
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
