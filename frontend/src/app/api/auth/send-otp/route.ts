import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/send-otp — идэвхгүй бүртгэлд 6 оронтой баталгаажуулах код илгээнэ.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const { email } = await readJson<{ email?: string }>(req);
  const result = await backendFetch('/auth/send-otp', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
  return toClientResponse(result);
}
