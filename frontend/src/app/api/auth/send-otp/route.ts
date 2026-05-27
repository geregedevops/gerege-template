import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin, requireFields } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/send-otp — идэвхгүй бүртгэлд 6 оронтой баталгаажуулах код илгээнэ.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ email?: string }>(req);
  const missing = requireFields(body, ['email']);
  if (missing) return missing;

  const result = await backendFetch('/auth/send-otp', {
    method: 'POST',
    body: JSON.stringify({ email: body.email }),
  });
  return toClientResponse(result);
}
