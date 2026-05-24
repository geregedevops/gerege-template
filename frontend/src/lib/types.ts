// gerege-backend-template-v27-ийн REST API-тай нийцсэн хуваалцсан типүүд.
// Эх сурвалж: docs/API_CONTRACT_MN.md болон responses.users.go.

/** Бүх backend хариу ороодог дугтуй (envelope). */
export interface Envelope<T = unknown> {
  status: boolean;
  message?: string;
  data?: T;
  request_id?: string;
}

/** responses.UserResponse — /login ба /refresh дээр токентой, /users/me дээр токенгүй. */
export interface BackendUser {
  id: string;
  username: string;
  email: string;
  role_id: number;
  token?: string;
  refresh_token?: string;
  created_at: string;
  updated_at: string | null;
}

/** GET /users/me нь өгөгдлийг { user: ... } дотор савладаг. */
export interface MeData {
  user: BackendUser;
}

/** 422 validation алдааны үед data.errors талбар бүрээр ирнэ. */
export interface ValidationData {
  errors: Record<string, string>;
}

/** Frontend-д ашиглах цэвэрлэсэн хэрэглэгчийн дүрс. */
export interface SessionUser {
  id: string;
  username: string;
  email: string;
  roleId: number;
  createdAt: string;
  updatedAt: string | null;
}

export function toSessionUser(u: BackendUser): SessionUser {
  return {
    id: u.id,
    username: u.username,
    email: u.email,
    roleId: u.role_id,
    createdAt: u.created_at,
    updatedAt: u.updated_at,
  };
}

/** role_id → хүний унших нэр. Backend-д 2 = энгийн хэрэглэгч. */
export function roleLabel(roleId: number, lang: 'mn' | 'en' = 'mn'): string {
  const mn: Record<number, string> = { 1: 'Админ', 2: 'Хэрэглэгч' };
  const en: Record<number, string> = { 1: 'Admin', 2: 'User' };
  return (lang === 'en' ? en : mn)[roleId] ?? (lang === 'en' ? 'User' : 'Хэрэглэгч');
}
