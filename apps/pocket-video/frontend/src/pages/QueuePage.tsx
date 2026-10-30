import React from 'react';
import { Download, Pause, Play, Trash2 } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import { formatBytes } from '../format.ts';
import type { Task, TaskStatus, Video } from '../types.ts';
import { EmptyState } from '../components/EmptyState.tsx';

const STATUS_KEY: Record<TaskStatus, keyof Copy> = {
  queued: 'statusQueued',
  running: 'statusRunning',
  paused: 'statusPaused',
  done: 'statusDone',
  failed: 'statusFailed',
};

export function QueuePage({
  t,
  tasks,
  videos,
  onToggle,
  onRemove,
  onBrowse,
}: {
  t: Copy;
  tasks: Task[];
  videos: Video[];
  onToggle: (task: Task) => void;
  onRemove: (task: Task) => void;
  onBrowse: () => void;
}) {
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">{t.transferEyebrow}</span>
          <h1>{t.queue}</h1>
          <p>{t.queueIntro}</p>
        </div>
      </div>
      {tasks.length === 0 ? (
        <EmptyState
          title={t.emptyTitle}
          hint={t.emptyHint}
          icon={<Download size={20} />}
          action={{ label: t.allVideos, onClick: onBrowse }}
        />
      ) : (
        <div className="queue-list">
          {tasks.map((task) => {
            const title = task.title || videos.find((video) => video.id === task.video_id)?.title || task.video_id;
            const running = task.status === 'running' || task.status === 'queued' || task.status === 'paused';
            return (
              <div className="task" key={task.id}>
                <div className={'task-icon ' + task.status}>
                  <Download size={17} />
                </div>
                <div className="task-main">
                  <div className="task-head">
                    <b>{title}</b>
                    <span className={'task-badge ' + task.status}>{t[STATUS_KEY[task.status]]}</span>
                  </div>
                  <small>
                    {task.kind === 'audio' ? t.kindAudio : t.kindVideo}
                    {task.total_bytes > 0
                      ? ` · ${formatBytes(task.downloaded_bytes)} / ${formatBytes(task.total_bytes)}`
                      : ''}
                  </small>
                  <div className="progress" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={task.progress}>
                    <span style={{ width: `${Math.min(100, Math.max(0, task.progress))}%` }} />
                  </div>
                  {task.error && <small className="task-error">{`${t.taskError}: ${task.error}`}</small>}
                </div>
                <strong>{task.progress}%</strong>
                <div className="task-actions">
                  {running && (
                    <button
                      className="icon-button"
                      aria-label={task.status === 'running' ? t.statusPaused : t.statusRunning}
                      onClick={() => onToggle(task)}
                    >
                      {task.status === 'running' ? <Pause size={15} /> : <Play size={15} />}
                    </button>
                  )}
                  {task.status !== 'running' && (
                    <button className="icon-button" aria-label={t.removeTask} title={t.removeTask} onClick={() => onRemove(task)}>
                      <Trash2 size={15} />
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}
