import React from 'react';
import SigninShell from '@/components/SigninShell';
import ResetPasswordForm from './ResetPasswordForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Нууц үг шинэчлэх — Gerege' };

export default function ResetPasswordPage({
  searchParams,
}: {
  searchParams: { token?: string };
}) {
  const token = typeof searchParams.token === 'string' ? searchParams.token : '';

  return (
    <SigninShell>
      <section className="signin-card signin-card--narrow" aria-labelledby="reset-title">
        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>Нууц үг шинэчлэх</div>
          <h1 id="reset-title">Шинэ нууц үг тохируулах</h1>
          <p className="signin-card__lede" style={{ marginTop: 8, fontSize: 14 }}>
            Шинэ нууц үгээ оруулна уу. Дор хаяж 12 тэмдэгт, том/жижиг үсэг, тоо, тусгай тэмдэгт агуулсан байх ёстой.
          </p>
        </div>
        <ResetPasswordForm initialToken={token} />
      </section>
    </SigninShell>
  );
}
