// Browser → BFF (Next.js route handler) рүү хандах нимгэн туслах. Browser
// хэзээ ч Go backend руу шууд хандахгүй — зөвхөн адил origin дахь /api/auth/*.

export interface ClientResult {
  ok: boolean;
  status: number;
  message?: string;
  /** 422 үед backend-ийн талбар бүрийн validation алдаа. */
  fieldErrors?: Record<string, string>;
}

/** JSON body-тэй POST хийгээд нэгдсэн ClientResult буцаана. */
export async function postJSON(path: string, body: unknown): Promise<ClientResult> {
  try {
    const res = await fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    let data: ClientResult | null = null;
    try {
      data = (await res.json()) as ClientResult;
    } catch {
      /* body хоосон байж болно */
    }
    return {
      ok: data?.ok ?? res.ok,
      status: data?.status ?? res.status,
      message: data?.message,
      fieldErrors: data?.fieldErrors,
    };
  } catch {
    return { ok: false, status: 0, message: 'Сүлжээний алдаа. Дахин оролдоно уу.' };
  }
}
