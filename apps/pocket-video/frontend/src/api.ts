// The only module allowed to talk to the Go service. PocketJS/PJOS can replace
// the base URL or the transport here without touching any page component.
import type { Account, Favorite, MediaKind, Settings, Task, Video, VideoPage } from './types.ts';

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, message: string, status = 0) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

export type ApiClientOptions = {
  base?: string;
  fetchImpl?: typeof fetch;
};

export function apiBase(): string {
  const injected = (globalThis as { __POCKETVIDEO_API__?: string }).__POCKETVIDEO_API__;
  return (injected ?? '/api').replace(/\/$/, '');
}

export function createApiClient(options: ApiClientOptions = {}) {
  const base = (options.base ?? apiBase()).replace(/\/$/, '');
  const doFetch = options.fetchImpl ?? ((...args: Parameters<typeof fetch>) => fetch(...args));

  async function request<T>(path: string, init?: RequestInit): Promise<T> {
    let response: Response;
    try {
      response = await doFetch(base + path, {
        ...init,
        headers: init?.body ? { 'Content-Type': 'application/json', ...(init?.headers ?? {}) } : init?.headers,
      });
    } catch {
      throw new ApiError('network', 'PocketVideo service is unavailable');
    }
    if (response.status === 204) {
      return undefined as T;
    }
    let payload: unknown = null;
    try {
      payload = await response.json();
    } catch {
      payload = null;
    }
    if (!response.ok) {
      const error = (payload as { error?: { code?: string; message?: string } } | null)?.error;
      throw new ApiError(error?.code ?? 'http_error', error?.message ?? `HTTP ${response.status}`, response.status);
    }
    return payload as T;
  }

  return {
    health: () => request<{ ok: boolean; provider: string; mode: string }>('/health'),
    account: () => request<Account>('/account'),
    favorites: () => request<Favorite[]>('/favorites'),
    videos: (params: { q?: string; favoriteId?: string; limit?: number; offset?: number } = {}) => {
      const search = new URLSearchParams();
      if (params.q) search.set('q', params.q);
      if (params.favoriteId) search.set('favorite_id', params.favoriteId);
      if (params.limit) search.set('limit', String(params.limit));
      if (params.offset) search.set('offset', String(params.offset));
      const suffix = search.toString();
      return request<VideoPage>('/videos' + (suffix ? `?${suffix}` : ''));
    },
    video: (id: string) => request<Video>(`/videos/${encodeURIComponent(id)}`),
    tasks: () => request<Task[]>('/task'),
    task: (id: number) => request<Task>(`/task/${id}`),
    download: (videoId: string, kind: MediaKind) =>
      request<{ id: number; status: string; existing?: boolean }>('/download', {
        method: 'POST',
        body: JSON.stringify({ video_id: videoId, kind }),
      }),
    pause: (id: number) => request<{ status: string }>(`/task/${id}/pause`, { method: 'POST' }),
    resume: (id: number) => request<{ status: string }>(`/task/${id}/resume`, { method: 'POST' }),
    removeTask: (id: number) => request<void>(`/task/${id}`, { method: 'DELETE' }),
    sync: () => request<{ favorites: number; videos: number; synced_at: string }>('/sync', { method: 'POST' }),
    settings: () => request<Settings>('/settings'),
    saveSettings: (settings: Settings) =>
      request<Settings>('/settings', { method: 'PUT', body: JSON.stringify(settings) }),
    mediaUrl: (id: string, kind: MediaKind = 'video') =>
      `${base}/media/${encodeURIComponent(id)}${kind === 'audio' ? '?kind=audio' : ''}`,
  };
}

export type ApiClient = ReturnType<typeof createApiClient>;
export const api = createApiClient();
