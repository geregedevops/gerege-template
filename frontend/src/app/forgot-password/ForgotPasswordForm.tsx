"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { Mail } from 'lucide-react';
import Alert from '@/components/Alert';
import { postJSON } from '@/lib/client';

export default function ForgotPasswordForm() {
  const [email, setEmail] = useState('');
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState('');

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    const res = await postJSON('/api/auth/forgot-password', { email });
    setBusy(false);
    // Backend нь enumeration-аас сэргийлж бүртгэлтэй эсэхээс үл хамааран ижил
    // хариу буцаадаг — тиймээс амжилттай үед үргэлж ерөнхий мессеж харуулна.
    if (res.ok || res.status === 200) {
      setDone(true);
    } else if (res.status === 422 && res.fieldErrors?.email) {
      setError(res.fieldErrors.email);
    } else {
      setError(res.message ?? 'Хүсэлт амжилтгүй боллоо.');
    }
  };

  if (done) {
    return (
      <div className="form-grid">
        <Alert kind="success">
          Хэрэв энэ и-мэйл бүртгэлтэй бол нууц үг сэргээх холбоосыг илгээлээ. Шуудангаа шалгана уу.
        </Alert>
        <p className="signin-card__alt">
          <Link href="/login">← Нэвтрэх рүү буцах</Link>
        </p>
      </div>
    );
  }

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {error && <Alert kind="danger">{error}</Alert>}

      <div className="field">
        <label className="field__label" htmlFor="email">И-мэйл</label>
        <input
          id="email"
          name="email"
          type="email"
          className="input"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoComplete="email"
          placeholder="tani@gerege.mn"
          required
        />
        <span className="field__help">Бүртгэлтэй и-мэйл хаягаа оруулна уу.</span>
      </div>

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy}>
        <Mail size={18} strokeWidth={2} />
        <span>{busy ? 'Илгээж байна…' : 'Сэргээх холбоос илгээх'}</span>
      </button>

      <p className="signin-card__alt">
        <Link href="/login">← Нэвтрэх рүү буцах</Link>
      </p>
    </form>
  );
}
