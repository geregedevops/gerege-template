import React from 'react';
import Link from 'next/link';
import { redirect } from 'next/navigation';
import { User, Mail, ShieldCheck, Clock, Hash, RefreshCw } from 'lucide-react';
import AppShell from '@/components/AppShell';
import { fetchMe } from '@/lib/api';
import { roleLabel } from '@/lib/types';
import { formatTS, initialsOf } from '@/lib/format';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Профайл — Gerege' };

export default async function ProfilePage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/profile');

  const initials = initialsOf(me.username);

  return (
    <AppShell user={{ username: me.username, email: me.email, initials }}>
      <div className="page-head">
        <span className="page-head__eyebrow">Хувь хүн</span>
        <h1>Хувь хүний профайл</h1>
        <p className="page-head__sub">Нэвтэрсэн бүртгэлийн мэдээлэл. Эх сурвалж: <span className="mono">GET /users/me</span>.</p>
      </div>

      <section className="card" aria-label="Профайлын тойм">
        <div className="profile-card">
          <div className="profile-card__avatar" aria-hidden="true">{initials}</div>
          <div className="profile-card__body">
            <div className="profile-card__name">
              <span className="profile-card__name-text">{me.username}</span>
              <span className="badge badge--primary">{roleLabel(me.roleId)}</span>
            </div>
            <div className="profile-card__sub">
              <span className="mono">{me.email}</span>
            </div>
          </div>
          <div className="profile-card__action">
            <Link className="btn btn--secondary" href="/settings">Нууц үг солих</Link>
          </div>
        </div>
      </section>

      <section className="card" aria-label="Бүртгэлийн талбарууд">
        <div className="card__head card__head--with-sub">
          <div className="card__title"><h2>Бүртгэлийн талбарууд</h2></div>
          <span className="card__sub">Эдгээр талбарыг backend удирдана.</span>
        </div>

        <div>
          <div className="defrow">
            <span className="defrow__label"><Hash size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />ID</span>
            <span className="defrow__value mono">{me.id}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><User size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />Нэвтрэх нэр</span>
            <span className="defrow__value">{me.username}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><Mail size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />И-мэйл</span>
            <span className="defrow__value mono">{me.email}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><ShieldCheck size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />Эрх</span>
            <span className="defrow__value">
              <span className="chip chip--neutral">role_id {me.roleId}</span> {roleLabel(me.roleId)}
            </span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><Clock size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />Бүртгэгдсэн</span>
            <span className="defrow__value mono">{formatTS(me.createdAt)}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><RefreshCw size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />Шинэчлэгдсэн</span>
            <span className="defrow__value mono">{me.updatedAt ? formatTS(me.updatedAt) : '—'}</span>
          </div>
        </div>
      </section>
    </AppShell>
  );
}
