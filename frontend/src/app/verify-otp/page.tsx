import React from 'react';
import SigninShell from '@/components/SigninShell';
import VerifyOtpForm from './VerifyOtpForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Баталгаажуулах — Gerege' };

export default function VerifyOtpPage({
  searchParams,
}: {
  searchParams: { email?: string; autosend?: string };
}) {
  const email = typeof searchParams.email === 'string' ? searchParams.email : '';
  const autosend = searchParams.autosend === '1';

  return (
    <SigninShell>
      <section className="signin-card signin-card--narrow" aria-labelledby="otp-title">
        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>Бүртгэл идэвхжүүлэх</div>
          <h1 id="otp-title">Баталгаажуулах</h1>
          <p className="signin-card__lede" style={{ marginTop: 8, fontSize: 14 }}>
            И-мэйлээр ирсэн 6 оронтой кодыг оруулж бүртгэлээ идэвхжүүлнэ үү.
          </p>
        </div>
        <VerifyOtpForm initialEmail={email} autosend={autosend} />
      </section>
    </SigninShell>
  );
}
