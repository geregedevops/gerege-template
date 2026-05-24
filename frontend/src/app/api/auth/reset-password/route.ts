import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/reset-password — и-мэйлээр ирсэн токен + шинэ нууц үгээр сэргээнэ.
export async function POST(req: Request) {
  const { token, new_password } = await readJson<{ token?: string; new_password?: string }>(req);
  const result = await backendFetch('/auth/password/reset', {
    method: 'POST',
    body: JSON.stringify({ token, new_password }),
  });
  return toClientResponse(result);
}
