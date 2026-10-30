export type Locale = 'en' | 'zh-CN' | 'ja';
export type ThemeSetting = 'system' | 'light' | 'dark';
export type ResolvedTheme = 'light' | 'dark';
export type PageName = 'home' | 'favorites' | 'videos' | 'detail' | 'queue' | 'settings';
export type TaskStatus = 'queued' | 'running' | 'paused' | 'done' | 'failed';
export type MediaKind = 'video' | 'audio';

export type Favorite = {
  id: string;
  title: string;
  video_count: number;
  last_sync_at?: string;
};

export type Video = {
  id: string;
  favorite_id: string;
  title: string;
  creator: string;
  duration: number;
  cover: string;
  has_video: boolean;
  has_audio: boolean;
};

export type Task = {
  id: number;
  video_id: string;
  title?: string;
  kind: MediaKind;
  status: TaskStatus;
  downloaded_bytes: number;
  total_bytes: number;
  progress: number;
  error?: string;
  updated_at?: string;
};

export type Settings = {
  language: Locale;
  theme: ThemeSetting;
  cache_dir: string;
  download_dir: string;
  sync_minutes: number;
  auto_sync: boolean;
  proxy: string;
};

export type Account = {
  mode: 'demo' | 'live';
  provider: string;
  authenticated: boolean;
  display_name: string;
  collections: number;
};

export type VideoPage = {
  items: Video[];
  limit: number;
  offset: number;
  total: number;
  downloaded: number;
};

export type Route = {
  page: PageName;
  videoId?: string;
  favoriteId?: string;
};
