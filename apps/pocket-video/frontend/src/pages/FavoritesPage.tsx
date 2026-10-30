import React from 'react';
import { ArrowLeft, FolderOpen, Heart, RefreshCw } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import type { Favorite, Video } from '../types.ts';
import { VideoGrid } from '../components/VideoGrid.tsx';
import { EmptyState } from '../components/EmptyState.tsx';
import { formatCount, formatSyncTime } from '../format.ts';

export function FavoritesPage({
  t,
  favorites,
  videos,
  locale,
  activeFavorite,
  busy,
  query,
  onOpenFavorite,
  onAllCollections,
  onOpenVideo,
  onDownload,
  onSync,
}: {
  t: Copy;
  favorites: Favorite[];
  videos: Video[];
  locale: string;
  activeFavorite: Favorite | null;
  busy: boolean;
  query: string;
  onOpenFavorite: (id: string) => void;
  onAllCollections: () => void;
  onOpenVideo: (video: Video) => void;
  onDownload: (video: Video) => void;
  onSync: () => void;
}) {
  if (activeFavorite) {
    return (
      <>
        <button className="back-button" onClick={onAllCollections}>
          <ArrowLeft size={15} />
          {t.allCollections}
        </button>
        <div className="page-heading">
          <div>
            <span className="eyebrow">{t.favoriteEyebrow}</span>
            <h1>{activeFavorite.title}</h1>
            <p>{formatCount(t.collectionCount, videos.length)}</p>
          </div>
          <button className="outline" onClick={onSync} disabled={busy}>
            <RefreshCw size={15} className={busy ? 'spin' : ''} />
            {busy ? t.syncing : t.sync}
          </button>
        </div>
        {videos.length > 0 ? (
          <VideoGrid videos={videos} t={t} onOpen={onOpenVideo} onDownload={onDownload} />
        ) : (
          <EmptyState
            title={query ? t.noResultsTitle : t.emptyTitle}
            hint={query ? t.noResultsHint : t.emptyHint}
            icon={<Heart size={20} />}
          />
        )}
      </>
    );
  }

  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">{t.collectionEyebrow}</span>
          <h1>{t.favorites}</h1>
          <p>{t.favoritesIntro}</p>
        </div>
        <button className="outline" onClick={onSync} disabled={busy}>
          <RefreshCw size={15} className={busy ? 'spin' : ''} />
          {busy ? t.syncing : t.sync}
        </button>
      </div>
      {favorites.length === 0 ? (
        <EmptyState
          title={t.favoritesEmptyTitle}
          hint={t.favoritesEmptyHint}
          icon={<Heart size={20} />}
          action={{ label: t.sync, onClick: onSync }}
        />
      ) : (
        <div className="favorite-grid">
          {favorites.map((favorite, index) => (
            <button
              className={'favorite-card tone-' + (index % 4)}
              key={favorite.id}
              onClick={() => onOpenFavorite(favorite.id)}
            >
              <span className="favorite-icon">
                <FolderOpen size={18} />
              </span>
              <span className="favorite-copy">
                <b>{favorite.title}</b>
                <small>{formatCount(t.collectionCount, favorite.video_count)}</small>
                <small className="favorite-sync">
                  {favorite.last_sync_at
                    ? formatSyncTime(favorite.last_sync_at, locale)
                    : t.neverSynced}
                </small>
              </span>
              <span className="favorite-open">{t.openCollection}</span>
            </button>
          ))}
        </div>
      )}
    </>
  );
}
