// Гарах: BFF-ийн logout route-г дуудаж (refresh токенг backend-ийн blacklist
// руу илгээж, амжилттай үед cookie-г цэвэрлэнэ). Backend амжилтгүй (5xx) бол
// route 502/503 буцаах ба cookie үлдэнэ — энэ үед /login руу шилжүүлэхгүй,
// caller-д алдааг тэмдэглэх боломж олгоно.
//
// Олон удаа дарагдахаас сэргийлж module-scope-д промис түгжих.
let inFlight: Promise<boolean> | null = null;

export async function signOut(): Promise<boolean> {
  if (inFlight) return inFlight;
  inFlight = (async () => {
    try {
      const res = await fetch('/api/auth/logout', { method: 'POST' });
      if (res.ok) {
        window.location.href = '/login';
        return true;
      }
      return false;
    } catch {
      return false;
    } finally {
      inFlight = null;
    }
  })();
  return inFlight;
}
