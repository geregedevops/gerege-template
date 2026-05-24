"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { KeyRound } from 'lucide-react';
import Alert from '@/components/Alert';
import PasswordField from '@/components/PasswordField';
import { postJSON } from '@/lib/client';

export default function ResetPasswordForm({ initialToken }: { initialToken: string }) {
  const router = useRouter();
  const [token, setToken] = useState(initialToken);
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    setFieldErrors({});
    const res = await postJSON('/api/auth/reset-password', { token, new_password: password });
    setBusy(false);
    if (res.ok) {
      router.push('/login?notice=reset');
      return;
    }
    if (res.status === 422 && res.fieldErrors) {
      setFieldErrors(res.fieldErrors);
      return;
    }
    setError(res.message ?? 'Нууц үг шинэчилж чадсангүй. Холбоос хүчингүй эсвэл хугацаа дууссан байж магадгүй.');
  };

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {error && <Alert kind="danger">{error}</Alert>}

      {!initialToken && (
        <div className="field">
          <label className="field__label" htmlFor="token">Сэргээх токен</label>
          <input
            id="token"
            name="token"
            className="input mono"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="И-мэйлээр ирсэн токен"
            aria-invalid={fieldErrors.token ? true : undefined}
            required
          />
          {fieldErrors.token && <span className="field__error">{fieldErrors.token}</span>}
        </div>
      )}

      <PasswordField
        label="Шинэ нууц үг"
        value={password}
        onChange={setPassword}
        showStrength
        autoComplete="new-password"
        error={fieldErrors.new_password}
        placeholder="Шинэ хүчтэй нууц үг"
      />

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy || !token}>
        <KeyRound size={18} strokeWidth={2} />
        <span>{busy ? 'Шинэчилж байна…' : 'Нууц үг шинэчлэх'}</span>
      </button>

      <p className="signin-card__alt">
        <Link href="/login">← Нэвтрэх рүү буцах</Link>
      </p>
    </form>
  );
}
