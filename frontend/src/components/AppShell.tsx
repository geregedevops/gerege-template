"use client";

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard, User, ShieldCheck, Settings, HelpCircle, LogOut,
} from 'lucide-react';
import UserMenu from './UserMenu';
import { signOut } from '@/lib/signout';

export interface AppUser {
  username: string;
  email: string;
  initials: string;
}

interface Props {
  user: AppUser;
  children: React.ReactNode;
}

const NAV_PRIMARY = [
  { href: '/',        label: 'Хяналтын самбар',     icon: LayoutDashboard },
  { href: '/profile', label: 'Хувь хүний профайл',  icon: User },
];

const NAV_SECURITY = [
  { href: '/settings', label: 'Аюулгүй байдал',      icon: ShieldCheck },
];

/**
 * Нэвтэрсэн хуудас бүрийн эргэн тойронд topbar + rail бүрхүүл. usePathname-аар
 * идэвхтэй холбоосыг тодотгоно.
 */
export default function AppShell({ user, children }: Props) {
  const pathname = usePathname() ?? '/';
  const isActive = (href: string) => (href === '/' ? pathname === '/' : pathname.startsWith(href));

  const renderLink = (item: { href: string; label: string; icon: typeof User }) => {
    const Icon = item.icon;
    const active = isActive(item.href);
    return (
      <Link
        key={item.href}
        className={`rail__link${active ? ' is-active' : ''}`}
        href={item.href}
        aria-current={active ? 'page' : undefined}
      >
        <Icon size={18} strokeWidth={2} />
        <span>{item.label}</span>
      </Link>
    );
  };

  return (
    <div className="app-shell">
      <header className="topbar">
        <Link href="/" className="topbar__brand">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img className="topbar__brand-mark" src="/brand.webp" alt="Gerege" />
          <div className="topbar__brand-text">
            <span className="topbar__brand-name">Gerege</span>
            <span className="topbar__brand-tag">Бүртгэлийн булан</span>
          </div>
        </Link>
        <div className="topbar__divider" aria-hidden="true" />
        <div className="topbar__session" role="status" aria-live="polite">
          <span className="topbar__session-dot" aria-hidden="true" />
          <span>
            <span className="topbar__session-strong">JWT</span>{' '}
            <span>· идэвхтэй сесси</span>
          </span>
        </div>
        <div className="topbar__spacer" />
        <div className="topbar__actions">
          <UserMenu username={user.username} email={user.email} initials={user.initials} />
        </div>
      </header>

      <div className="i18n-banner">
        Англи хувилбар бүрэн бус — агуулга монголоор үлдсэн.{' '}
        <strong>Partial English UI</strong> — page content remains in Mongolian.
      </div>

      <div className="shell-body">
        <aside className="rail" aria-label="Үндсэн цэс">
          <div className="rail__group">
            <span className="rail__label">Профайл</span>
            {NAV_PRIMARY.map(renderLink)}
          </div>

          <div className="rail__group">
            <span className="rail__label">Аюулгүй байдал</span>
            {NAV_SECURITY.map(renderLink)}
          </div>

          <div className="rail__group">
            <span className="rail__label">Тусламж</span>
            <a className="rail__link" href="https://gerege.mn/help" target="_blank" rel="noreferrer">
              <HelpCircle size={18} strokeWidth={2} />
              <span>Тусламж</span>
            </a>
            <button className="rail__signout" type="button" onClick={() => signOut()}>
              <LogOut size={18} strokeWidth={2} />
              <span>Гарах</span>
            </button>
          </div>
        </aside>

        <main className="main">
          <div className="main__inner">{children}</div>
        </main>
      </div>
    </div>
  );
}
