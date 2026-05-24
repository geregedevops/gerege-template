import { backendFetch } from '@/lib/api';
import { getRefreshToken, clearSession } from '@/lib/session';
import { toClientResponse } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/auth/logout — refresh токенг backend-ийн blacklist руу илгээж
// (boj болсон ч) cookie-г заавал цэвэрлэнэ. Backend амжилтгүй ч client тал
// нэвтрэлтгүй болж, дахин нэвтрэхийг шаардана.
export async function POST() {
  const refresh = getRefreshToken();
  if (refresh) {
    await backendFetch('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refresh }),
    });
  }
  clearSession();
  return toClientResponse({ ok: true, status: 200, message: 'Гарлаа' });
}
