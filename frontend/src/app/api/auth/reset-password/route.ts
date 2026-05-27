import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin, requireFields } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/reset-password — и-мэйлээр ирсэн токен + шинэ нууц үгээр сэргээнэ.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ token?: string; new_password?: string }>(req);
  const missing = requireFields(body, ['token', 'new_password']);
  if (missing) return missing;

  const result = await backendFetch('/auth/password/reset', {
    method: 'POST',
    body: JSON.stringify({ token: body.token, new_password: body.new_password }),
  });
  return toClientResponse(result);
}
