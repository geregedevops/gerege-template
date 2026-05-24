"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { LogIn } from 'lucide-react';
import Alert from '@/components/Alert';
import PasswordField from '@/components/PasswordField';
import { postJSON } from '@/lib/client';
import { safeNext } from '@/lib/navigation';

export default function LoginForm({ next, notice }: { next: string; notice?: string }) {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [needsActivation, setNeedsActivation] = useState(false);

  const noticeText =
    notice === 'verified' ? 'Бүртгэл баталгаажлаа. Одоо нэвтэрнэ үү.'
    : notice === 'registered' ? 'Бүртгэл үүслээ. Кодоо баталгаажуулсны дараа нэвтэрнэ үү.'
    : notice === 'reset' ? 'Нууц үг шинэчлэгдлээ. Шинэ нууц үгээрээ нэвтэрнэ үү.'
    : '';

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    setFieldErrors({});
    setNeedsActivation(false);

    const res = await postJSON('/api/auth/login', { email, password });
    setBusy(false);

    if (res.ok) {
      router.push(safeNext(next));
      router.refresh();
      return;
    }
    if (res.status === 422 && res.fieldErrors) {
      setFieldErrors(res.fieldErrors);
      return;
    }
    // Backend: идэвхжээгүй бүртгэл → 403 "account is not activated".
    if (res.status === 403 && /activat/i.test(res.message ?? '')) {
      setNeedsActivation(true);
      return;
    }
    setError(res.message ?? 'Нэвтрэх амжилтгүй боллоо.');
  };

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {noticeText && <Alert kind="success">{noticeText}</Alert>}
      {error && <Alert kind="danger">{error}</Alert>}
      {needsActivation && (
        <Alert kind="info">
          Энэ бүртгэл идэвхжээгүй байна.{' '}
          <Link href={`/verify-otp?email=${encodeURIComponent(email)}`}>Кодоор баталгаажуулах →</Link>
        </Alert>
      )}

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
          aria-invalid={fieldErrors.email ? true : undefined}
          required
        />
        {fieldErrors.email && <span className="field__error">{fieldErrors.email}</span>}
      </div>

      <PasswordField
        label="Нууц үг"
        value={password}
        onChange={setPassword}
        autoComplete="current-password"
        error={fieldErrors.password}
        placeholder="••••••••••••"
      />

      <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: -6 }}>
        <Link href="/forgot-password" style={{ fontSize: 13 }}>Нууц үгээ мартсан уу?</Link>
      </div>

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy}>
        <LogIn size={18} strokeWidth={2} />
        <span>{busy ? 'Нэвтэрч байна…' : 'Нэвтрэх'}</span>
      </button>

      <p className="signin-card__alt">
        Бүртгэлгүй юу? <Link href="/register">Шинээр бүртгүүлэх</Link>
      </p>
    </form>
  );
}
