import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin, requireFields } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/verify-otp — код таарвал backend бүртгэлийг идэвхжүүлнэ.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ email?: string; code?: string }>(req);
  const missing = requireFields(body, ['email', 'code']);
  if (missing) return missing;

  const result = await backendFetch('/auth/verify-otp', {
    method: 'POST',
    body: JSON.stringify({ email: body.email, code: body.code }),
  });
  return toClientResponse(result);
}
