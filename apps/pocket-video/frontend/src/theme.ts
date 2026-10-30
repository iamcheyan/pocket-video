import type { ResolvedTheme, ThemeSetting } from './types.ts';

export function isThemeSetting(value: unknown): value is ThemeSetting {
  return value === 'system' || value === 'light' || value === 'dark';
}

export function resolveTheme(setting: ThemeSetting, prefersDark: boolean): ResolvedTheme {
  if (setting === 'system') return prefersDark ? 'dark' : 'light';
  return setting;
}

export function systemPrefersDark(): boolean {
  return typeof matchMedia === 'function' && matchMedia('(prefers-color-scheme: dark)').matches;
}

export function applyTheme(
  setting: ThemeSetting,
  root: HTMLElement = document.documentElement,
  prefersDark: boolean = systemPrefersDark(),
): ResolvedTheme {
  const resolved = resolveTheme(setting, prefersDark);
  root.dataset.theme = resolved;
  root.style.colorScheme = resolved;
  return resolved;
}

// System mode must follow the device live, not only at startup.
export function watchSystemTheme(onChange: (prefersDark: boolean) => void): () => void {
  if (typeof matchMedia !== 'function') return () => {};
  const query = matchMedia('(prefers-color-scheme: dark)');
  const listener = (event: MediaQueryListEvent) => onChange(event.matches);
  query.addEventListener('change', listener);
  return () => query.removeEventListener('change', listener);
}
