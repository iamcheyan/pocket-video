import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ApiError, api } from './api.ts';
import { Sidebar } from './components/Sidebar.tsx';
import { TopBar } from './components/TopBar.tsx';
import { installNavigation } from './navigation.ts';
import { parseRoute, navigate } from './router.ts';
import { applyTheme, isThemeSetting, watchSystemTheme } from './theme.ts';
import { detectLocale, isLocale, t as dictionaries, type Locale } from './i18n.ts';
import { HomePage } from './pages/HomePage.tsx';
import { FavoritesPage } from './pages/FavoritesPage.tsx';
import { VideosPage } from './pages/VideosPage.tsx';
import { DetailPage } from './pages/DetailPage.tsx';
import { QueuePage } from './pages/QueuePage.tsx';
import { SettingsPage } from './pages/SettingsPage.tsx';
import type {
  Account,
  Favorite,
  MediaKind,
  PageName,
  Route,
  Settings,
  Task,
  ThemeSetting,
  Video,
} from './types.ts';

const initialCapture = () => new URLSearchParams(location.search);

function initialLocale(capture: URLSearchParams): Locale {
  const forced = capture.get('lang');
  if (isLocale(forced)) return forced;
  const stored = localStorage.getItem('pocketvideo.lang');
  if (isLocale(stored)) return stored;
  return detectLocale(navigator.languages ?? [navigator.language]);
}

function initialTheme(capture: URLSearchParams): ThemeSetting {
  const forced = capture.get('theme');
  if (isThemeSetting(forced)) return forced;
  const stored = localStorage.getItem('pocketvideo.theme');
  return isThemeSetting(stored) ? stored : 'system';
}

export default function App() {
  const capture = useMemo(initialCapture, []);
  const [route, setRoute] = useState<Route>(() => parseRoute(location.hash));
  const [locale, setLocale] = useState<Locale>(() => initialLocale(capture));
  const [theme, setTheme] = useState<ThemeSetting>(() => initialTheme(capture));
  const [settings, setSettings] = useState<Settings | null>(null);
  const [account, setAccount] = useState<Account | null>(null);
  const [favorites, setFavorites] = useState<Favorite[]>([]);
  const [videos, setVideos] = useState<Video[]>([]);
  const [total, setTotal] = useState(0);
  const [downloaded, setDownloaded] = useState(0);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<Video | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [toast, setToast] = useState<'saved' | 'failed' | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const doneCount = useRef(0);
  const searchRef = useRef<HTMLInputElement>(null);
  const mainRef = useRef<HTMLElement>(null);
  const copy = dictionaries[locale];

  const favoriteFilter = route.page === 'favorites' ? route.favoriteId : undefined;

  // Boot: the service is the source of truth for language, theme, and account.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const [remoteSettings, remoteAccount, remoteFavorites] = await Promise.all([
          api.settings(),
          api.account(),
          api.favorites(),
        ]);
        if (cancelled) return;
        setSettings(remoteSettings);
        setAccount(remoteAccount);
        setFavorites(remoteFavorites);
        if (!isLocale(capture.get('lang')) && !isLocale(localStorage.getItem('pocketvideo.lang'))) {
          setLocale(remoteSettings.language);
        }
        if (!isThemeSetting(capture.get('theme')) && !isThemeSetting(localStorage.getItem('pocketvideo.theme'))) {
          setTheme(remoteSettings.theme);
        }
        setError(null);
      } catch (err) {
        if (!cancelled) setError(err instanceof ApiError ? err.message : String(err));
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [capture]);

  // Routing stays hash-based so the same build runs inside PocketJS or PJOS.
  useEffect(() => {
    const onHash = () => {
      setRoute(parseRoute(location.hash));
      setSidebarOpen(false);
    };
    window.addEventListener('hashchange', onHash);
    return () => window.removeEventListener('hashchange', onHash);
  }, []);

  useEffect(() => {
    applyTheme(theme);
    if (theme !== 'system') return;
    return watchSystemTheme(() => applyTheme('system'));
  }, [theme]);

  useEffect(() => {
    let cancelled = false;
    api
      .videos({ q: query, favoriteId: favoriteFilter })
      .then((page) => {
        if (cancelled) return;
        setVideos(page.items);
        setTotal(page.total);
        setDownloaded(page.downloaded ?? 0);
        setError(null);
      })
      .catch((err) => !cancelled && setError(err instanceof ApiError ? err.message : String(err)));
    return () => {
      cancelled = true;
    };
  }, [query, favoriteFilter, refreshKey]);

  // A detail route is deep-linkable even when the video is not on the list page.
  useEffect(() => {
    if (route.page !== 'detail' || !route.videoId) {
      setSelected(null);
      return;
    }
    if (selected?.id === route.videoId) return;
    const cached = videos.find((video) => video.id === route.videoId);
    if (cached) {
      setSelected(cached);
      return;
    }
    let cancelled = false;
    api
      .video(route.videoId)
      .then((video) => !cancelled && setSelected(video))
      .catch((err) => !cancelled && setError(err instanceof ApiError ? err.message : String(err)));
    return () => {
      cancelled = true;
    };
  }, [route.page, route.videoId, videos, selected?.id]);

  // Task polling is the only place that learns about finished downloads.
  useEffect(() => {
    let stopped = false;
    const tick = async () => {
      try {
        const list = await api.tasks();
        if (stopped) return;
        setTasks(list);
        const finished = list.filter((task) => task.status === 'done').length;
        if (finished !== doneCount.current) {
          doneCount.current = finished;
          setRefreshKey((key) => key + 1);
        }
      } catch {
        /* the banner already reports service failures */
      }
    };
    void tick();
    const timer = setInterval(tick, 1000);
    return () => {
      stopped = true;
      clearInterval(timer);
    };
  }, []);

  useEffect(() => {
    if (toast === null) return;
    const timer = setTimeout(() => setToast(null), 2200);
    return () => clearTimeout(timer);
  }, [toast]);

  useEffect(() => {
    mainRef.current?.focus({ preventScroll: true });
  }, [route.page, route.videoId, route.favoriteId]);

  // Layout diagnostic used by the screenshot pipeline and by device QA:
  // data-overflow="true" means a page needs horizontal scrolling.
  useEffect(() => {
    const measure = () => {
      const doc = document.documentElement;
      doc.dataset.overflow = String(doc.scrollWidth > window.innerWidth + 1);
    };
    const frame = requestAnimationFrame(measure);
    window.addEventListener('resize', measure);
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener('resize', measure);
    };
  }, [route.page, route.videoId, route.favoriteId, videos.length, tasks.length]);

  const persistSettings = useCallback(
    async (next: Settings) => {
      try {
        const saved = await api.saveSettings(next);
        setSettings(saved);
        setError(null);
        return true;
      } catch (err) {
        setError(err instanceof ApiError ? err.message : String(err));
        return false;
      }
    },
    [],
  );

  const changeLanguage = useCallback(
    (next: Locale) => {
      setLocale(next);
      localStorage.setItem('pocketvideo.lang', next);
      if (settings) void persistSettings({ ...settings, language: next });
    },
    [settings, persistSettings],
  );

  const chooseTheme = useCallback(
    (next: ThemeSetting) => {
      setTheme(next);
      localStorage.setItem('pocketvideo.theme', next);
      if (settings) void persistSettings({ ...settings, theme: next });
    },
    [settings, persistSettings],
  );

  const openPage = useCallback((page: PageName) => {
    setSidebarOpen(false);
    navigate({ page });
  }, []);

  const openVideo = useCallback((video: Video) => {
    setSelected(video);
    navigate({ page: 'detail', videoId: video.id });
  }, []);

  const openFavorite = useCallback((favoriteId: string) => {
    navigate({ page: 'favorites', favoriteId });
  }, []);

  const startDownload = useCallback(async (video: Video, kind: MediaKind) => {
    try {
      await api.download(video.id, kind);
      setTasks(await api.tasks());
      navigate({ page: 'queue' });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    }
  }, []);

  const toggleTask = useCallback(async (task: Task) => {
    try {
      if (task.status === 'running') await api.pause(task.id);
      else await api.resume(task.id);
      setTasks(await api.tasks());
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    }
  }, []);

  const removeTask = useCallback(async (task: Task) => {
    try {
      await api.removeTask(task.id);
      setTasks(await api.tasks());
      setRefreshKey((key) => key + 1);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    }
  }, []);

  const syncNow = useCallback(async () => {
    setBusy(true);
    try {
      await api.sync();
      const [remoteFavorites, page] = await Promise.all([
        api.favorites(),
        api.videos({ q: query, favoriteId: favoriteFilter }),
      ]);
      setFavorites(remoteFavorites);
      setVideos(page.items);
      setTotal(page.total);
      setDownloaded(page.downloaded ?? 0);
      setError(null);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }, [query, favoriteFilter]);

  const saveSettings = useCallback(
    async (draft: Settings) => {
      const ok = await persistSettings(draft);
      setToast(ok ? 'saved' : 'failed');
      if (ok) {
        setLocale(draft.language);
        setTheme(draft.theme);
        localStorage.setItem('pocketvideo.lang', draft.language);
        localStorage.setItem('pocketvideo.theme', draft.theme);
      }
    },
    [persistSettings],
  );

  useEffect(
    () =>
      installNavigation(
        () => {
          if (sidebarOpen) {
            setSidebarOpen(false);
            return;
          }
          if (route.page === 'detail') {
            navigate({ page: 'videos' });
            return;
          }
          if (route.page !== 'home') navigate({ page: 'home' });
        },
        () => {
          const focused = document.activeElement;
          if (focused instanceof HTMLElement && focused !== document.body && focused !== mainRef.current) {
            focused.click();
          } else {
            document.querySelector<HTMLElement>('.nav-item.active')?.focus();
          }
        },
      ),
    [route.page, sidebarOpen],
  );

  // "/" and Cmd/Ctrl+K focus search; both stay inside the browser shell.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const typing = event.target instanceof HTMLInputElement || event.target instanceof HTMLSelectElement;
      if (event.key === '/' && !typing) {
        event.preventDefault();
        searchRef.current?.focus();
      }
      if (event.key.toLowerCase() === 'k' && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const currentFavorite = route.favoriteId
    ? favorites.find((favorite) => favorite.id === route.favoriteId) ?? null
    : null;
  const pageLabels: Record<PageName, string> = {
    home: copy.home,
    favorites: copy.favorites,
    videos: copy.videos,
    detail: copy.detailFallback,
    queue: copy.queue,
    settings: copy.settings,
  };
  const pageTitle =
    route.page === 'detail'
      ? selected?.title ?? copy.detailFallback
      : route.page === 'favorites' && currentFavorite
        ? currentFavorite.title
        : pageLabels[route.page];

  const errorCode = error !== null;

  return (
    <div className="app-shell">
      <Sidebar
        t={copy}
        route={route}
        tasks={tasks}
        account={account}
        open={sidebarOpen}
        onNavigate={openPage}
        onClose={() => setSidebarOpen(false)}
      />
      <main className="main" ref={mainRef} tabIndex={-1} aria-label={copy.brand}>
        <TopBar
          t={copy}
          locale={locale}
          title={pageTitle}
          query={query}
          searchRef={searchRef}
          onQuery={setQuery}
          onMenu={() => setSidebarOpen((open) => !open)}
          onToggleLanguage={() => changeLanguage(locale === 'en' ? 'zh-CN' : locale === 'zh-CN' ? 'ja' : 'en')}
        />
        <section className="content">
          {errorCode && (
            <div className="error-banner" role="alert">
              <div>
                <b>{copy.serviceError}</b>
                <small>{copy.serviceErrorHint}</small>
              </div>
              <button className="outline" onClick={() => window.location.reload()}>
                {copy.retry}
              </button>
            </div>
          )}
          {route.page === 'home' && (
            <HomePage
              t={copy}
              videos={videos}
              total={total}
              downloaded={downloaded}
              favorites={favorites}
              account={account}
              onBrowse={() => openPage('favorites')}
              onOpen={openVideo}
              onDownload={(video) => void startDownload(video, 'video')}
              onSeeAll={() => openPage('videos')}
            />
          )}
          {route.page === 'favorites' && (
            <FavoritesPage
              t={copy}
              favorites={favorites}
              videos={videos}
              locale={locale}
              activeFavorite={currentFavorite}
              busy={busy}
              query={query}
              onOpenFavorite={openFavorite}
              onAllCollections={() => openPage('favorites')}
              onOpenVideo={openVideo}
              onDownload={(video) => void startDownload(video, 'video')}
              onSync={() => void syncNow()}
            />
          )}
          {route.page === 'videos' && (
            <VideosPage
              t={copy}
              videos={videos}
              total={total}
              query={query}
              onOpen={openVideo}
              onDownload={(video) => void startDownload(video, 'video')}
              onClearSearch={() => setQuery('')}
              onBrowseFavorites={() => openPage('favorites')}
            />
          )}
          {route.page === 'detail' && (
            <DetailPage
              t={copy}
              video={selected}
              favorite={favorites.find((favorite) => favorite.id === selected?.favorite_id) ?? null}
              tasks={tasks}
              mediaUrl={api.mediaUrl}
              onBack={() => openPage('videos')}
              onDownload={(kind) => selected && void startDownload(selected, kind)}
            />
          )}
          {route.page === 'queue' && (
            <QueuePage
              t={copy}
              tasks={tasks}
              videos={videos}
              onToggle={(task) => void toggleTask(task)}
              onRemove={(task) => void removeTask(task)}
              onBrowse={() => openPage('videos')}
            />
          )}
          {route.page === 'settings' && (
            <SettingsPage
              t={copy}
              locale={locale}
              settings={settings}
              onChangeLanguage={changeLanguage}
              onSelectTheme={chooseTheme}
              onSave={saveSettings}
            />
          )}
        </section>
      </main>
      <div className={'toast ' + (toast ? 'show ' + toast : '')} role="status" aria-live="polite">
        {toast === 'saved' && copy.saveDone}
        {toast === 'failed' && copy.saveFailed}
      </div>
    </div>
  );
}
