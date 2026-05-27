import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin, requireFields } from '@/lib/bff';
import type { BackendUser } from '@/lib/types';

export const dynamic = 'force-dynamic';

// POST /api/auth/register — backend /auth/register руу прокси. Хэрэглэгч идэвхгүй
// үүснэ; cookie суулгахгүй. Дараа нь OTP баталгаажуулалт шаардана.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{
    username?: string;
    email?: string;
    password?: string;
  }>(req);
  const missing = requireFields(body, ['username', 'email', 'password']);
  if (missing) return missing;

  const result = await backendFetch<BackendUser>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({
      username: body.username,
      email: body.email,
      password: body.password,
    }),
  });

  return toClientResponse(result);
}
