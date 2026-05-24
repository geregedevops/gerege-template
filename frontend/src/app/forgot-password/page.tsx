import React from 'react';
import SigninShell from '@/components/SigninShell';
import ForgotPasswordForm from './ForgotPasswordForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Нууц үг сэргээх — Gerege' };

export default function ForgotPasswordPage() {
  return (
    <SigninShell>
      <section className="signin-card signin-card--narrow" aria-labelledby="forgot-title">
        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>Нууц үг сэргээх</div>
          <h1 id="forgot-title">Нууц үгээ мартсан уу?</h1>
          <p className="signin-card__lede" style={{ marginTop: 8, fontSize: 14 }}>
            И-мэйл хаягаа оруулбал нууц үг сэргээх холбоосыг илгээнэ.
          </p>
        </div>
        <ForgotPasswordForm />
      </section>
    </SigninShell>
  );
}
