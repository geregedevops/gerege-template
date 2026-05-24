import { backendFetch } from '@/lib/api';
import { readJson, toClientResponse, checkOrigin } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/forgot-password — нууц үг сэргээх холбоос/токен и-мэйлдэнэ.
// Backend нь enumeration-аас сэргийлж бүртгэлтэй эсэхээс үл хамааран ижил
// мессеж буцаадаг.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const { email } = await readJson<{ email?: string }>(req);
  const result = await backendFetch('/auth/password/forgot', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
  return toClientResponse(result);
}
