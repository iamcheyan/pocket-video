import React, { useEffect, useRef, useState } from 'react';
import { Music, Pause, Play } from 'lucide-react';
import type { Copy } from '../i18n.ts';
import { formatDuration } from '../format.ts';

// Native audio controls are tiny and inconsistent on handhelds, so the app owns
// the transport: one large play target, a seek bar, and a readable clock.
export function AudioPlayer({ src, title, t }: { src: string; title: string; t: Copy }) {
  const ref = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState(false);
  const [time, setTime] = useState(0);
  const [duration, setDuration] = useState(0);

  useEffect(() => {
    setPlaying(false);
    setTime(0);
    setDuration(0);
  }, [src]);

  const toggle = () => {
    const audio = ref.current;
    if (!audio) return;
    if (audio.paused) void audio.play().catch(() => setPlaying(false));
    else audio.pause();
  };

  return (
    <div className="audio-stage">
      <Music size={30} />
      <b className="audio-title">{title}</b>
      <audio
        ref={ref}
        src={src}
        preload="metadata"
        onPlay={() => setPlaying(true)}
        onPause={() => setPlaying(false)}
        onEnded={() => setPlaying(false)}
        onTimeUpdate={(event) => setTime(event.currentTarget.currentTime)}
        onLoadedMetadata={(event) => setDuration(event.currentTarget.duration || 0)}
      />
      <div className="audio-controls">
        <button type="button" aria-label={playing ? t.pause : t.play} onClick={toggle}>
          {playing ? <Pause size={18} /> : <Play size={18} fill="currentColor" />}
        </button>
        <input
          type="range"
          min={0}
          max={duration || 1}
          step={0.1}
          value={time}
          aria-label={t.status}
          onChange={(event) => {
            const next = Number(event.target.value);
            setTime(next);
            if (ref.current) ref.current.currentTime = next;
          }}
        />
        <span>
          {formatDuration(time)} / {formatDuration(duration)}
        </span>
      </div>
    </div>
  );
}
