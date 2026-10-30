import type { PageName, Route } from './types.ts';

const PAGES: PageName[] = ['home', 'favorites', 'videos', 'detail', 'queue', 'settings'];

export function isPageName(value: string): value is PageName {
  return (PAGES as string[]).includes(value);
}

// Hash routing keeps the app portable: a PocketJS shell can supply its own
// navigation later without changing page components.
export function parseRoute(hash: string): Route {
  const clean = hash.replace(/^#\/?/, '').replace(/\/+$/, '');
  if (!clean) return { page: 'home' };
  const [pageName, id] = clean.split('/');
  if (!isPageName(pageName)) return { page: 'home' };
  if (pageName === 'detail') {
    return id ? { page: 'detail', videoId: decodeURIComponent(id) } : { page: 'videos' };
  }
  if (pageName === 'favorites' && id) {
    return { page: 'favorites', favoriteId: decodeURIComponent(id) };
  }
  return { page: pageName };
}

export function routeToHash(route: Route): string {
  switch (route.page) {
    case 'detail':
      return route.videoId ? `#detail/${encodeURIComponent(route.videoId)}` : '#videos';
    case 'favorites':
      return route.favoriteId ? `#favorites/${encodeURIComponent(route.favoriteId)}` : '#favorites';
    default:
      return `#${route.page}`;
  }
}

export function navigate(route: Route): void {
  const hash = routeToHash(route);
  if (location.hash === hash) {
    window.dispatchEvent(new HashChangeEvent('hashchange'));
    return;
  }
  location.hash = hash;
}
