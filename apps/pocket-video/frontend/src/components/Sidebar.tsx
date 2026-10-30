import React from 'react';
import { Download, HardDrive, Heart, Library, Play, Settings as SettingsIcon, Home } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import type { Account, PageName, Route, Task } from '../types.ts';

const NAV: { page: PageName; icon: LucideIcon; label: keyof Copy }[] = [
  { page: 'home', icon: Home, label: 'home' },
  { page: 'favorites', icon: Heart, label: 'favorites' },
  { page: 'videos', icon: Library, label: 'videos' },
  { page: 'queue', icon: Download, label: 'queue' },
  { page: 'settings', icon: SettingsIcon, label: 'settings' },
];

export function Sidebar({
  t,
  route,
  tasks,
  account,
  open,
  onNavigate,
  onClose,
}: {
  t: Copy;
  route: Route;
  tasks: Task[];
  account: Account | null;
  open: boolean;
  onNavigate: (page: PageName) => void;
  onClose: () => void;
}) {
  const active = route.page === 'detail' ? 'videos' : route.page;
  const syncing = tasks.some((task) => task.status === 'running');
  return (
    <aside className={'sidebar ' + (open ? 'show' : '')}>
      <div className="brand">
        <span className="brand-mark">
          <Play size={17} fill="currentColor" />
        </span>
        <span>
          Pocket<span className="brand-light">Video</span>
        </span>
        <button className="close-nav" aria-label={t.closeMenu} onClick={onClose}>
          ×
        </button>
      </div>
      <div className="nav-caption">{t.libraryCaption}</div>
      <nav aria-label={t.libraryCaption}>
        {NAV.map(({ page, icon: Icon, label }) => (
          <button
            className={'nav-item ' + (active === page ? 'active' : '')}
            key={page}
            aria-current={active === page ? 'page' : undefined}
            onClick={() => onNavigate(page)}
          >
            <Icon size={17} />
            <span>{t[label]}</span>
            {page === 'queue' && syncing && <i aria-hidden="true" />}
          </button>
        ))}
      </nav>
      <div className="sidebar-bottom">
        <div className="account-card">
          <div className="avatar">{t.monogram}</div>
          <div>
            <b>{account?.mode === 'live' ? t.accountLive : t.accountDemo}</b>
            <small>{account ? account.display_name : t.loading}</small>
          </div>
          <span className={'status-dot ' + (account?.authenticated ? 'online' : 'offline')} aria-hidden="true" />
        </div>
        <div className="side-foot" title={t.providerLabel}>
          <span>{account ? account.provider : 'PocketVideo'}</span>
          <span>
            {account?.authenticated ? t.accountSignedIn : t.accountSignedOut} · {t.version}
          </span>
        </div>
      </div>
    </aside>
  );
}

export function StatIcon({ kind }: { kind: 'library' | 'disk' | 'clock' }) {
  if (kind === 'disk') return <HardDrive size={17} />;
  if (kind === 'clock') return <Heart size={17} />;
  return <Library size={17} />;
}
