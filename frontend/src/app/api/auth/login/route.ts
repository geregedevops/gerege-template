import { backendFetch } from '@/lib/api';
import { setSession } from '@/lib/session';
import { readJson, toClientResponse, checkOrigin, requireFields } from '@/lib/bff';
import type { BackendUser } from '@/lib/types';

export const dynamic = 'force-dynamic';

// POST /api/auth/login — backend /auth/login руу прокси, амжилттай бол токен
// хосыг httpOnly cookie-д суулгана. Токен хэзээ ч browser руу буцахгүй.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ email?: string; password?: string }>(req);
  const missing = requireFields(body, ['email', 'password']);
  if (missing) return missing;

  const result = await backendFetch<BackendUser>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email: body.email, password: body.password }),
  });

  if (result.ok && result.data?.token && result.data?.refresh_token) {
    setSession(result.data.token, result.data.refresh_token);
  }

  return toClientResponse(result);
}
