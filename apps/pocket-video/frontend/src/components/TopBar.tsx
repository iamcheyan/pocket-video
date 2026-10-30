import React, { type RefObject } from 'react';
import { ChevronRight, Globe, Menu, Search } from 'lucide-react';
import type { Copy, Locale } from '../i18n.ts';
import { localeNames } from '../i18n.ts';

export function TopBar({
  t,
  locale,
  title,
  query,
  searchRef,
  onQuery,
  onMenu,
  onToggleLanguage,
}: {
  t: Copy;
  locale: Locale;
  title: string;
  query: string;
  searchRef: RefObject<HTMLInputElement | null>;
  onQuery: (value: string) => void;
  onMenu: () => void;
  onToggleLanguage: () => void;
}) {
  return (
    <header className="topbar">
      <button className="mobile-menu" aria-label={t.menu} onClick={onMenu}>
        <Menu size={18} />
      </button>
      <div className="breadcrumb">
        <span className="breadcrumb-brand">{t.brand}</span>
        <ChevronRight size={13} />
        <span>{title}</span>
      </div>
      <div className="top-actions">
        <label className="search">
          <Search size={16} />
          <input
            ref={searchRef}
            type="search"
            value={query}
            aria-label={t.search}
            onChange={(event) => onQuery(event.target.value)}
            placeholder={t.search}
          />
          <kbd>{t.searchShortcut}</kbd>
        </label>
        <span className="top-divider" />
        <button
          className="icon-button"
          aria-label={`${t.language}: ${localeNames[locale]}`}
          title={localeNames[locale]}
          onClick={onToggleLanguage}
        >
          <Globe size={17} />
        </button>
        <button className="profile-avatar" aria-label={t.profile}>
          {t.monogram.slice(0, 1)}
        </button>
      </div>
    </header>
  );
}
