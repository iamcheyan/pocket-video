import React from 'react';
import { Library, Search, X } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import type { Video } from '../types.ts';
import { VideoGrid } from '../components/VideoGrid.tsx';
import { EmptyState } from '../components/EmptyState.tsx';
import { formatCount } from '../format.ts';

export function VideosPage({
  t,
  videos,
  total,
  query,
  onOpen,
  onDownload,
  onClearSearch,
  onBrowseFavorites,
}: {
  t: Copy;
  videos: Video[];
  total: number;
  query: string;
  onOpen: (video: Video) => void;
  onDownload: (video: Video) => void;
  onClearSearch: () => void;
  onBrowseFavorites: () => void;
}) {
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">{t.collectionEyebrow}</span>
          <h1>{t.allVideos}</h1>
          <p>{query ? formatCount(t.searchResults, total) : t.libraryIntro}</p>
        </div>
        {query && (
          <button className="outline" onClick={onClearSearch}>
            <X size={15} />
            {t.clearSearch}
          </button>
        )}
      </div>
      {videos.length > 0 ? (
        <VideoGrid videos={videos} t={t} onOpen={onOpen} onDownload={onDownload} />
      ) : query ? (
        <EmptyState
          title={t.noResultsTitle}
          hint={t.noResultsHint}
          icon={<Search size={20} />}
          action={{ label: t.clearSearch, onClick: onClearSearch }}
        />
      ) : (
        <EmptyState
          title={t.favoritesEmptyTitle}
          hint={t.favoritesEmptyHint}
          icon={<Library size={20} />}
          action={{ label: t.browse, onClick: onBrowseFavorites }}
        />
      )}
    </>
  );
}
