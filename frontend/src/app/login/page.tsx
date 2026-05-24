import React from 'react';
import SigninShell from '@/components/SigninShell';
import { safeNext } from '@/lib/navigation';
import LoginForm from './LoginForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Нэвтрэх — Gerege' };

export default function LoginPage({
  searchParams,
}: {
  searchParams: { next?: string; notice?: string };
}) {
  const next = safeNext(searchParams.next);

  return (
    <SigninShell>
      <section className="signin-card signin-card--narrow" aria-labelledby="login-title">
        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>Хэрэглэгчийн булан</div>
          <h1 id="login-title">Нэвтрэх</h1>
          <p className="signin-card__lede" style={{ marginTop: 8, fontSize: 14 }}>
            И-мэйл болон нууц үгээрээ нэвтэрнэ үү.
          </p>
        </div>
        <LoginForm next={next} notice={searchParams.notice} />
      </section>
    </SigninShell>
  );
}
