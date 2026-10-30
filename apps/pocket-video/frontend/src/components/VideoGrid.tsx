import React from 'react';
import { Check, Download, Play } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import { formatDuration } from '../format.ts';
import type { Video } from '../types.ts';

export function VideoGrid({
  videos,
  t,
  onOpen,
  onDownload,
}: {
  videos: Video[];
  t: Copy;
  onOpen: (video: Video) => void;
  onDownload: (video: Video) => void;
}) {
  return (
    <div className="video-grid">
      {videos.map((video, index) => (
        <article className="video-card" key={video.id}>
          <button className="thumbnail" onClick={() => onOpen(video)} aria-label={video.title}>
            <div className={'landscape landscape-' + (index % 3)}>
              <span className="sun" />
              <div className="mountains" />
              <div className="lake" />
            </div>
            {video.cover && <img className="cover-img" src={video.cover} alt="" loading="lazy" />}
            <span className="duration">{formatDuration(video.duration)}</span>
            <span className="thumb-play">
              <Play size={14} fill="currentColor" />
            </span>
            {video.has_video && (
              <span className="offline-flag">
                <Check size={11} />
                {t.offline}
              </span>
            )}
          </button>
          <div className="video-meta">
            <div>
              <span className="eyebrow">{video.has_video ? t.statusDone : t.kindVideo}</span>
              <h3>
                <button onClick={() => onOpen(video)}>{video.title}</button>
              </h3>
              <p>
                {video.creator} · {formatDuration(video.duration)}
              </p>
            </div>
            <button className="download-icon" aria-label={t.downloadVideo} onClick={() => onDownload(video)}>
              {video.has_video ? <Check size={16} /> : <Download size={16} />}
            </button>
          </div>
        </article>
      ))}
    </div>
  );
}
