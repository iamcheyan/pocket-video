import React from 'react';
import { createRoot, type Root } from 'react-dom/client';
import App from './App.tsx';
import { applyTheme, isThemeSetting, resolveTheme, systemPrefersDark } from './theme.ts';
import './style.css';

// Paint the right theme before React mounts so handhelds never flash white.
const params = new URLSearchParams(location.search);
const stored = localStorage.getItem('pocketvideo.theme');
const initial = isThemeSetting(params.get('theme'))
  ? (params.get('theme') as 'system' | 'light' | 'dark')
  : isThemeSetting(stored)
    ? stored
    : 'system';
document.documentElement.dataset.theme = resolveTheme(initial, systemPrefersDark());
applyTheme(initial);

const host = window as typeof window & { __pocketVideoRoot?: Root };
const container = document.getElementById('root');
if (!container) throw new Error('PocketVideo root element missing');
const root = host.__pocketVideoRoot ?? (host.__pocketVideoRoot = createRoot(container));
root.render(<App />);
