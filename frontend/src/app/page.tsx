import React from 'react';
import Link from 'next/link';
import {
  ChevronRight, User, ShieldCheck, KeyRound, Info, LogIn,
  Mail, Clock, type LucideIcon,
} from 'lucide-react';
import AppShell from '@/components/AppShell';
import SigninShell from '@/components/SigninShell';
import { hasSession } from '@/lib/session';
import { fetchMe } from '@/lib/api';
import { roleLabel } from '@/lib/types';
import { formatDateMN, formatTS, formatWeekdayMN, initialsOf } from '@/lib/format';

export const dynamic = 'force-dynamic';

interface CardLink {
  href: string;
  eyebrow: string;
  title: string;
  desc: string;
  icon: LucideIcon;
}

const SECTION_CARDS: CardLink[] = [
  { href: '/profile',  eyebrow: 'Хувь хүн',    title: 'Хувь хүний профайл', desc: 'Нэвтрэх нэр, холбоо барих и-мэйл, эрхийн төвшин болон бүртгэлийн мэдээлэл.', icon: User },
  { href: '/settings', eyebrow: 'Аюулгүй байдал', title: 'Аюулгүй байдал', desc: 'Нууц үгээ солих, идэвхтэй сессиэ удирдах.', icon: ShieldCheck },
];

export default async function Home() {
  if (!hasSession()) return <Landing />;

  const me = await fetchMe();
  if (!me) return <Landing />;

  const today = new Date();
  const initials = initialsOf(me.username);

  return (
    <AppShell user={{ username: me.username, email: me.email, initials }}>
      <div className="page-head">
        <span className="page-head__eyebrow">Хяналтын самбар</span>
        <h1>Сайн байна уу, {me.username}</h1>
        <p className="page-head__sub">
          {formatDateMN(today)}, {formatWeekdayMN(today)} · <span className="mono">UTC+08</span>
        </p>
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
              <span className="dot" />
              <span>Бүртгүүлсэн <span className="mono">{formatTS(me.createdAt)}</span></span>
            </div>
          </div>
          <div className="profile-card__action">
            <Link className="btn btn--secondary" href="/profile">Профайл харах</Link>
          </div>
        </div>
      </section>

      <div className="section-divider">Булангийн хэсгүүд</div>

      <div className="grid-2">
        {SECTION_CARDS.map((c) => {
          const Icon = c.icon;
          return (
            <Link
              key={c.href}
              href={c.href}
              className="card"
              style={{ textDecoration: 'none', color: 'inherit', display: 'flex', flexDirection: 'column', gap: 10 }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ width: 36, height: 36, borderRadius: 10, display: 'grid', placeItems: 'center', background: 'var(--dan-blue-soft)', color: 'var(--dan-blue-text)' }}>
                  <Icon size={18} strokeWidth={2} />
                </div>
                <span className="page-head__eyebrow">{c.eyebrow}</span>
              </div>
              <h3 style={{ fontSize: 16, fontWeight: 600 }}>{c.title}</h3>
              <p style={{ fontSize: 13, color: 'var(--muted)', lineHeight: 1.55 }}>{c.desc}</p>
              <span style={{ marginTop: 'auto', display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 12, fontWeight: 500, color: 'var(--dan-blue-text)' }}>
                Нээх <ChevronRight size={12} strokeWidth={2} />
              </span>
            </Link>
          );
        })}
      </div>

      <section className="card" aria-label="Бүртгэлийн дэлгэрэнгүй" style={{ marginTop: 16 }}>
        <div className="card__head card__head--with-sub">
          <div className="card__title">
            <Info size={18} strokeWidth={2} style={{ color: 'var(--dan-blue-text)' }} />
            <h2>Бүртгэлийн дэлгэрэнгүй</h2>
          </div>
          <span className="card__sub">gerege-backend-template-v27 · GET /users/me</span>
        </div>
        <div>
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
            <span className="defrow__value"><span className="chip chip--neutral">role_id {me.roleId}</span> {roleLabel(me.roleId)}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><Clock size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />Бүртгэгдсэн</span>
            <span className="defrow__value mono">{formatTS(me.createdAt)}</span>
          </div>
        </div>
      </section>

      <div className="trust-strip">
        <span className="trust-strip__item">
          <KeyRound size={12} strokeWidth={2.5} style={{ color: 'var(--dan-blue)' }} />
          JWT access + refresh
        </span>
        <span className="trust-strip__dot">·</span>
        <span className="trust-strip__item">bcrypt</span>
        <span className="trust-strip__dot">·</span>
        <span className="trust-strip__item">Fiber v3 + GORM</span>
        <span className="trust-strip__dot">·</span>
        <span className="trust-strip__item mono">TLS 1.3</span>
      </div>

      <footer className="footer">
        <span>© 2026 Gerege Systems · <span className="mono">gerege-backend-template-v27</span></span>
        <span className="footer__links">
          <a href="https://gerege.mn/privacy">Нууцлал</a>
          <a href="https://gerege.mn/terms">Нөхцөл</a>
        </span>
      </footer>
    </AppShell>
  );
}

/** Нийтийн landing — нэвтрээгүй зочдод харагдах нүүр. */
function Landing() {
  return (
    <SigninShell>
      <section className="signin-card" aria-labelledby="landing-title">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img className="signin-card__crest" src="/brand.webp" alt="" aria-hidden="true" />

        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>Хэрэглэгчийн булан</div>
          <h1 id="landing-title">Gerege</h1>
          <p className="signin-card__lede" style={{ marginTop: 12 }}>
            <strong style={{ color: 'var(--fg)', fontWeight: 600 }}>gerege-backend-template-v27</strong>{' '}
            дээр суурилсан хэрэглэгчийн булан. И-мэйл хаягаараа{' '}
            <strong style={{ color: 'var(--fg)', fontWeight: 600 }}>бүртгүүлж</strong>,{' '}
            нэг удаагийн кодоор баталгаажуулаад, профайл болон аюулгүй байдлын тохиргоогоо нэг дороос удирдана.
          </p>
        </div>

        <Link className="btn btn--primary btn--lg btn--block" href="/login" aria-label="И-мэйлээр нэвтрэх">
          <LogIn size={18} strokeWidth={2} />
          <span>Нэвтрэх</span>
        </Link>

        <p className="signin-card__alt">
          Бүртгэлгүй юу? <Link href="/register">Шинээр бүртгүүлэх</Link>
        </p>

        <p className="signin-card__helper">
          <Info size={14} strokeWidth={2} />
          <span>
            Нууц үг нь <span className="mono" style={{ color: 'var(--fg)' }}>bcrypt</span>-ээр хэшлэгдэж хадгалагдана.
            Нэвтрэлт нь богино TTL-тэй access болон урт TTL-тэй refresh JWT хослолоор хийгдэнэ.
          </span>
        </p>

        <div className="signin-card__trust" aria-label="Аюулгүй байдлын тэмдэг">
          <span className="badge"><KeyRound size={11} strokeWidth={2} /> JWT</span>
          <span className="badge">bcrypt</span>
          <span className="badge">Fiber v3 + GORM</span>
          <span className="badge"><span className="mono" style={{ fontSize: 11 }}>TLS 1.3</span></span>
        </div>
      </section>
    </SigninShell>
  );
}
