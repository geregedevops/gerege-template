import { authedFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin, requireFields } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/change-password — нэвтэрсэн хэрэглэгчийн нууц үг солих. Backend
// тал PUT /auth/password/change (JWT шаардана) руу прокси; access токенг
// cookie-оос авч 401 дээр автоматаар refresh хийнэ.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{
    current_password?: string;
    new_password?: string;
  }>(req);
  const missing = requireFields(body, ['current_password', 'new_password']);
  if (missing) return missing;

  const result = await authedFetch('/auth/password/change', {
    method: 'PUT',
    body: JSON.stringify({
      current_password: body.current_password,
      new_password: body.new_password,
    }),
  });

  return toClientResponse(result);
}
