import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/verify-otp — код таарвал backend бүртгэлийг идэвхжүүлнэ.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const { email, code } = await readJson<{ email?: string; code?: string }>(req);
  const result = await backendFetch('/auth/verify-otp', {
    method: 'POST',
    body: JSON.stringify({ email, code }),
  });
  return toClientResponse(result);
}
