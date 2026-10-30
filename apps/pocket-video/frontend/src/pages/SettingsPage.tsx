import React, { useEffect, useState } from 'react';
import { Check, Monitor, Moon, Save, Settings as SettingsIcon, Sun } from 'lucide-react';
import type { Copy, Locale } from '../i18n.ts';
import { localeNames, locales } from '../i18n.ts';
import type { Settings, ThemeSetting } from '../types.ts';

const THEMES: { value: ThemeSetting; icon: typeof Sun; label: keyof Copy }[] = [
  { value: 'system', icon: Monitor, label: 'system' },
  { value: 'light', icon: Sun, label: 'light' },
  { value: 'dark', icon: Moon, label: 'dark' },
];

export function SettingsPage({
  t,
  locale,
  settings,
  onChangeLanguage,
  onSelectTheme,
  onSave,
}: {
  t: Copy;
  locale: Locale;
  settings: Settings | null;
  onChangeLanguage: (locale: Locale) => void;
  onSelectTheme: (theme: ThemeSetting) => void;
  onSave: (draft: Settings) => void;
}) {
  const [draft, setDraft] = useState<Settings | null>(settings);
  useEffect(() => setDraft(settings), [settings]);

  if (!draft) {
    return <p className="muted-note">{t.loading}</p>;
  }

  const update = <K extends keyof Settings>(key: K, value: Settings[K]) => setDraft({ ...draft, [key]: value });
  const absolutePaths = draft.cache_dir.startsWith('/') && draft.download_dir.startsWith('/');

  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">{t.makeItYoursEyebrow}</span>
          <h1>{t.settings}</h1>
          <p>{t.preferencesHint}</p>
        </div>
      </div>
      <form
        className="settings-card"
        onSubmit={(event) => {
          event.preventDefault();
          onSave(draft);
        }}
      >
        <div className="settings-title">
          <div className="setting-symbol">
            <SettingsIcon size={18} />
          </div>
          <div>
            <h3>{t.preferences}</h3>
            <p>{t.shortcuts}</p>
          </div>
        </div>

        <div className="setting-row">
          <label htmlFor="settings-language">
            <b>{t.language}</b>
            <small>{t.languageHint}</small>
          </label>
          <select
            id="settings-language"
            value={locale}
            onChange={(event) => {
              const next = event.target.value as Locale;
              update('language', next);
              onChangeLanguage(next);
            }}
          >
            {locales.map((value) => (
              <option value={value} key={value}>
                {localeNames[value]}
              </option>
            ))}
          </select>
        </div>

        <div className="setting-row">
          <div>
            <b>{t.theme}</b>
            <small>{t.themeHint}</small>
          </div>
          <div className="theme-options" role="group" aria-label={t.theme}>
            {THEMES.map(({ value, icon: Icon, label }) => (
              <button
                type="button"
                className={draft.theme === value ? 'chosen' : ''}
                aria-pressed={draft.theme === value}
                key={value}
                onClick={() => {
                  update('theme', value);
                  onSelectTheme(value);
                }}
              >
                <Icon size={15} />
                {t[label]}
              </button>
            ))}
          </div>
        </div>

        <div className="setting-row">
          <label htmlFor="settings-cache">
            <b>{t.cache}</b>
            <small>{t.cacheHint}</small>
          </label>
          <input
            id="settings-cache"
            value={draft.cache_dir}
            spellCheck={false}
            onChange={(event) => update('cache_dir', event.target.value)}
          />
        </div>

        <div className="setting-row">
          <label htmlFor="settings-download">
            <b>{t.folder}</b>
            <small>{t.folderHint}</small>
          </label>
          <input
            id="settings-download"
            value={draft.download_dir}
            spellCheck={false}
            onChange={(event) => update('download_dir', event.target.value)}
          />
        </div>

        <div className="setting-row">
          <label htmlFor="settings-frequency">
            <b>{t.frequency}</b>
            <small>{t.frequencyHint}</small>
          </label>
          <input
            id="settings-frequency"
            type="number"
            min={1}
            max={10080}
            value={draft.sync_minutes}
            onChange={(event) => update('sync_minutes', Number(event.target.value))}
          />
        </div>

        <div className="setting-row">
          <div>
            <b>{t.automatic}</b>
            <small>{t.automaticHint}</small>
          </div>
          <button
            type="button"
            className={'toggle ' + (draft.auto_sync ? 'on' : '')}
            aria-label={t.automatic}
            aria-pressed={draft.auto_sync}
            onClick={() => update('auto_sync', !draft.auto_sync)}
          >
            <span />
          </button>
        </div>

        <div className="setting-row">
          <label htmlFor="settings-proxy">
            <b>{t.proxy}</b>
            <small>{t.proxyHint}</small>
          </label>
          <input
            id="settings-proxy"
            value={draft.proxy}
            placeholder={t.proxyPlaceholder}
            spellCheck={false}
            onChange={(event) => update('proxy', event.target.value)}
          />
        </div>

        <div className="save-row">
          <span>{!absolutePaths && <small className="field-error">{t.absolutePath}</small>}</span>
          <button className="primary" type="submit">
            {draft === settings ? <Check size={16} /> : <Save size={16} />}
            {t.save}
          </button>
        </div>
      </form>
    </>
  );
}
