import React from 'react';
import SigninShell from '@/components/SigninShell';
import RegisterForm from './RegisterForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Бүртгүүлэх — Gerege' };

export default function RegisterPage() {
  return (
    <SigninShell>
      <section className="signin-card signin-card--narrow" aria-labelledby="register-title">
        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>Хэрэглэгчийн булан</div>
          <h1 id="register-title">Шинээр бүртгүүлэх</h1>
          <p className="signin-card__lede" style={{ marginTop: 8, fontSize: 14 }}>
            Бүртгүүлсний дараа и-мэйлээр ирэх 6 оронтой кодоор бүртгэлээ баталгаажуулна.
          </p>
        </div>
        <RegisterForm />
      </section>
    </SigninShell>
  );
}
