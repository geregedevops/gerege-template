import { backendFetch } from '@/lib/api';
import { getRefreshToken, clearSession } from '@/lib/session';
import { toClientResponse, checkOrigin } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/logout — refresh токенг backend-ийн blacklist руу илгээж,
// амжилттай үед л client cookie-г цэвэрлэнэ. Backend амжилтгүй ч client
// "гарсан" гэж буруу мэдэгдвэл хуучин refresh token 7 хоног backend-д
// амьд үлдэх эрсдэлтэй — иймд 5xx үед cookie үлдээж, хэрэглэгчийг дахин
// оролдохыг шаардана.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const refresh = getRefreshToken();
  if (!refresh) {
    // Аль хэдийн нэвтрэхгүй — idempotent амжилт.
    clearSession();
    return toClientResponse({ ok: true, status: 200, message: 'Гарлаа' });
  }

  const r = await backendFetch('/auth/logout', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refresh }),
  });

  // Backend амжилттай хариу өгсөн (2xx) эсвэл 401 (token аль хэдийн
  // хүчингүй) тохиолдолд хоёуланд нь cookie-г цэвэрлэнэ — backend талд
  // session аль хэдийн байхгүй.
  if (r.ok || r.status === 401) {
    clearSession();
    return toClientResponse({ ok: true, status: 200, message: 'Гарлаа' });
  }

  // Backend хариу өгөөгүй / 5xx — cookie үлдээнэ. Client дахин оролдох
  // боломжтой; ингэснээр backend session leak хаагдана.
  return toClientResponse(r);
}
