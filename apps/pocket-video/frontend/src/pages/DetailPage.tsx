import React, { useEffect, useState } from 'react';
import { ArrowLeft, Check, Download, Film, Music, Play } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import { formatBytes, formatDuration } from '../format.ts';
import type { Favorite, MediaKind, Task, Video } from '../types.ts';
import { AudioPlayer } from '../components/AudioPlayer.tsx';

const STATUS_KEY = {
  queued: 'statusQueued',
  running: 'statusRunning',
  paused: 'statusPaused',
  done: 'statusDone',
  failed: 'statusFailed',
} as const;

export function DetailPage({
  t,
  video,
  favorite,
  tasks,
  mediaUrl,
  onBack,
  onDownload,
}: {
  t: Copy;
  video: Video | null;
  favorite: Favorite | null;
  tasks: Task[];
  mediaUrl: (id: string, kind: MediaKind) => string;
  onBack: () => void;
  onDownload: (kind: MediaKind) => void;
}) {
  const [preview, setPreview] = useState<MediaKind>('video');
  const [playing, setPlaying] = useState(false);

  useEffect(() => {
    if (!video) return;
    setPreview(video.has_video ? 'video' : 'audio');
    setPlaying(false);
  }, [video?.id, video?.has_video, video?.has_audio]);

  if (!video) {
    return (
      <>
        <button className="back-button" onClick={onBack}>
          <ArrowLeft size={15} />
          {t.back}
        </button>
        <p className="muted-note">{t.loading}</p>
      </>
    );
  }

  const both = video.has_video && video.has_audio;
  const ready = preview === 'video' ? video.has_video : video.has_audio;
  const related = tasks.filter((task) => task.video_id === video.id);
  const latest = related[0];
  const active = related.find((task) => task.status === 'running' || task.status === 'queued');
  const failed = related.find((task) => task.status === 'failed');

  return (
    <>
      <button className="back-button" onClick={onBack}>
        <ArrowLeft size={15} />
        {t.back}
      </button>
      <div className="detail-layout">
        <div className={'detail-cover ' + (ready ? 'has-media' : '')}>
          {ready ? (
            preview === 'video' ? (
              playing ? (
                <video controls autoPlay playsInline preload="auto" poster={video.cover || undefined} src={mediaUrl(video.id, 'video')} />
              ) : (
                <button className="poster-stage" onClick={() => setPlaying(true)} aria-label={`${t.playOffline}: ${video.title}`}>
                  {video.cover ? (
                    <img src={video.cover} alt="" />
                  ) : (
                    <div className="cover-landscape">
                      <span className="sun" />
                      <div className="mountains" />
                      <div className="lake" />
                    </div>
                  )}
                  <span className="poster-play">
                    <Play size={22} fill="currentColor" />
                  </span>
                  <span className="poster-label">{t.playOffline}</span>
                </button>
              )
            ) : (
              <AudioPlayer src={mediaUrl(video.id, 'audio')} title={video.title} t={t} />
            )
          ) : (
            <div className="cover-landscape">
              <span className="sun" />
              <div className="mountains" />
              <div className="lake" />
            </div>
          )}
          {!ready && (
            <div className="cover-overlay">
              <button className="play-large" onClick={() => onDownload(preview)} aria-label={t.downloadVideo}>
                <Download />
              </button>
              <span>{t.notDownloaded}</span>
            </div>
          )}
        </div>
        <div className="detail-info">
          <span className="eyebrow">{t.favoriteEyebrow}</span>
          <h1>{video.title}</h1>
          <p>
            {video.creator} · {formatDuration(video.duration)}
            {favorite ? ` · ${favorite.title}` : ''}
          </p>
          {both && (
            <div className="player-switch" role="group" aria-label={t.status}>
              <button
                className={preview === 'video' ? 'chosen' : ''}
                onClick={() => {
                  setPreview('video');
                  setPlaying(false);
                }}
              >
                <Film size={14} />
                {t.kindVideo}
              </button>
              <button className={preview === 'audio' ? 'chosen' : ''} onClick={() => setPreview('audio')}>
                <Music size={14} />
                {t.kindAudio}
              </button>
            </div>
          )}
          <div className="detail-rule" />
          <div className="detail-actions">
            <button className="primary" onClick={() => onDownload('video')} disabled={Boolean(active) || video.has_video}>
              {video.has_video ? <Check size={16} /> : <Download size={16} />}
              {video.has_video ? t.downloadedVideo : t.downloadVideo}
            </button>
            <button className="outline" onClick={() => onDownload('audio')} disabled={Boolean(active) || video.has_audio}>
              {video.has_audio ? <Check size={16} /> : <Download size={16} />}
              {video.has_audio ? t.downloadedAudio : t.downloadAudio}
            </button>
          </div>
          <div className="detail-meta">
            <span>{t.status}</span>
            <b>{ready ? (preview === 'video' ? t.downloadedVideo : t.downloadedAudio) : t.notDownloaded}</b>
            {latest && (
              <small className={'task-state ' + latest.status}>
                {t[STATUS_KEY[latest.status]]}
                {latest.kind === 'audio' ? ` · ${t.kindAudio}` : ` · ${t.kindVideo}`}
                {latest.total_bytes > 0 ? ` · ${formatBytes(latest.downloaded_bytes)} / ${formatBytes(latest.total_bytes)}` : ''}
              </small>
            )}
            {failed && <small className="task-error">{`${t.taskError}: ${failed.error ?? t.statusFailed}`}</small>}
          </div>
        </div>
      </div>
    </>
  );
}
