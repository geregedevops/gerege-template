import React from 'react';
import { redirect } from 'next/navigation';
import { KeyRound, LogOut } from 'lucide-react';
import AppShell from '@/components/AppShell';
import ChangePasswordForm from './ChangePasswordForm';
import SignOutButton from '@/components/SignOutButton';
import { fetchMe } from '@/lib/api';
import { initialsOf } from '@/lib/format';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Аюулгүй байдал — Gerege' };

export default async function SettingsPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/settings');

  const initials = initialsOf(me.username);

  return (
    <AppShell user={{ username: me.username, email: me.email, initials }}>
      <div className="page-head">
        <span className="page-head__eyebrow">Аюулгүй байдал</span>
        <h1>Аюулгүй байдлын тохиргоо</h1>
        <p className="page-head__sub">Нууц үгээ солих, идэвхтэй сессиэ удирдах.</p>
      </div>

      <section className="card" aria-label="Нууц үг солих">
        <div className="card__head card__head--with-sub">
          <div className="card__title">
            <KeyRound size={18} strokeWidth={2} style={{ color: 'var(--dan-blue-text)' }} />
            <h2>Нууц үг солих</h2>
          </div>
          <span className="card__sub">
            Шинэ нууц үг нь дор хаяж 12 тэмдэгт, том/жижиг үсэг, тоо, тусгай тэмдэгт агуулсан байх ёстой.
          </span>
        </div>
        <ChangePasswordForm />
      </section>

      <section className="card" aria-label="Сесси">
        <div className="card__head card__head--with-sub">
          <div className="card__title">
            <LogOut size={18} strokeWidth={2} style={{ color: 'var(--danger)' }} />
            <h2>Сесси</h2>
          </div>
          <span className="card__sub">
            Энэ төхөөрөмж дээрх refresh токенг хүчингүй болгож, дахин нэвтрэхийг шаардана.
          </span>
        </div>
        <SignOutButton />
      </section>
    </AppShell>
  );
}
