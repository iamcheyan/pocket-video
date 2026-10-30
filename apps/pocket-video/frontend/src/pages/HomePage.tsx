import React from 'react';
import { ChevronRight, Clock, HardDrive, Library, Play } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import type { Account, Favorite, Video } from '../types.ts';
import { VideoGrid } from '../components/VideoGrid.tsx';
import { EmptyState } from '../components/EmptyState.tsx';

export function HomePage({
  t,
  videos,
  total,
  downloaded,
  favorites,
  account,
  onBrowse,
  onOpen,
  onDownload,
  onSeeAll,
}: {
  t: Copy;
  videos: Video[];
  total: number;
  downloaded: number;
  favorites: Favorite[];
  account: Account | null;
  onBrowse: () => void;
  onOpen: (video: Video) => void;
  onDownload: (video: Video) => void;
  onSeeAll: () => void;
}) {
  return (
    <>
      <div className="hero">
        <div className="hero-copy">
          <div className="eyebrow">
            <span />
            {t.personalEyebrow}
          </div>
          <h1>
            {t.hero.split('\n').map((row, index) => (
              <React.Fragment key={row}>
                {index > 0 && <br />}
                {row}
              </React.Fragment>
            ))}
          </h1>
          <p>{t.subtitle}</p>
          <button className="primary" onClick={onBrowse}>
            {t.browse}
            <ChevronRight size={16} />
          </button>
          <div className="hero-note">
            <span className="avatar mini">{t.monogram}</span>
            <span>{account ? `${account.display_name} · ${account.provider}` : t.loading}</span>
          </div>
        </div>
        <div className="hero-art">
          <div className="art-orbit orbit-one" />
          <div className="art-orbit orbit-two" />
          <div className="art-card back-card">
            <div className="art-sun" />
            <div className="art-mountain" />
            <div className="art-water" />
          </div>
          <div className="art-card front-card">
            <div className="art-sun small" />
            <div className="art-mountain" />
            <div className="art-water" />
            <span className="art-play">
              <Play size={16} fill="currentColor" />
            </span>
          </div>
          <div className="art-label">
            <span className="green-dot" /> {t.offline}
          </div>
        </div>
        <div className="hero-grain" />
      </div>

      <div className="stats">
        <div>
          <div className="stat-icon">
            <Library size={17} />
          </div>
          <small>{t.allVideos}</small>
          <strong>
            {String(total).padStart(2, '0')} <em>{t.videosUnit}</em>
          </strong>
        </div>
        <div>
          <div className="stat-icon">
            <HardDrive size={17} />
          </div>
          <small>{t.savedOffline}</small>
          <strong>
            {String(downloaded).padStart(2, '0')} <em>{t.available}</em>
          </strong>
        </div>
        <div>
          <div className="stat-icon">
            <Clock size={17} />
          </div>
          <small>{t.favorites}</small>
          <strong>
            {String(favorites.length).padStart(2, '0')} <em>{t.foldersUnit}</em>
          </strong>
        </div>
      </div>

      <div className="section-heading">
        <div>
          <span className="eyebrow">{t.collectionEyebrow}</span>
          <h2>{t.recent}</h2>
        </div>
        <button className="text-button" onClick={onSeeAll}>
          {t.allVideos} <ChevronRight size={15} />
        </button>
      </div>
      {videos.length > 0 ? (
        <VideoGrid videos={videos.slice(0, 6)} t={t} onOpen={onOpen} onDownload={onDownload} />
      ) : (
        <EmptyState title={t.favoritesEmptyTitle} hint={t.favoritesEmptyHint} />
      )}
    </>
  );
}
