// Video client: offline Bilibili archive over videod.
// Download/playback providers are mock on this host; the device has no yt-dlp.
import { createPortClient } from "./client-kit";
import type { AppConfig, Metric, Section, SettingRow } from "./model";

const sections: Section[] = [
  {
    id: "videos", title: "OFFLINE", jp: "オフライン", items: [
      { id: "BV1DEMO001", title: "PocketJS in 12 Minutes", subtitle: "Tech · 12:04 · 1080p", meta: "downloaded", badge: "MP4", progress: 0.24, body: "A tour of the PocketJS runtime, the native contract and the device build.", lines: [{ label: "size", value: "184 MB" }, { label: "codec", value: "h264" }, { label: "audio", value: "aac" }] },
      { id: "BV1DEMO002", title: "Night Harbour Walk", subtitle: "Travel · 28:31 · 720p", meta: "downloaded", badge: "MP4", progress: 0, body: "A slow walk along the harbour with ambient audio." },
      { id: "BV1DEMO003", title: "Synth Repair Log", subtitle: "DIY · 19:12 · 1080p", meta: "queued", badge: "QUE", progress: 0.07, body: "Repairing a 1980s polysynth, part three." },
    ],
  },
  {
    id: "queue", title: "QUEUE", jp: "キュー", items: [
      { id: "q1", title: "Download queue", subtitle: "2 running · 1 paused", meta: "yt-dlp", badge: "DWN", body: "Two active yt-dlp jobs; the adapter is mock on this machine." },
      { id: "q2", title: "Favourites", subtitle: "42 videos · 18 offline", meta: "synced", badge: "FAV", body: "Bilibili favourite folders mirrored to the archive." },
      { id: "q3", title: "Storage", subtitle: "12.4 GB used · 58 GB free", meta: "internal", badge: "DSK", body: "Archive volume usage." },
    ],
  },
];

const metrics: Metric[] = [
  { label: "VIDEOS", value: "42" },
  { label: "OFFLINE", value: "18" },
  { label: "QUEUE", value: "2" },
];

const settings: SettingRow[] = [
  { id: "quality", label: "Download quality", jp: "画質", value: "1080P", options: ["1080P", "720P", "480P"] },
  { id: "codec", label: "Codec", jp: "コーデック", value: "H264", options: ["H264", "H265", "AV1"] },
  { id: "wifi", label: "Wi-Fi only", jp: "Wi-Fiのみ", value: "ON", options: ["ON", "OFF"] },
];

export const APP: AppConfig = {
  title: "VIDEO",
  jp: "動画",
  kind: "video",
  page: "VIDEO",
  tabs: ["動画", "再生", "設定"],
  actions: ["PREV", "PLAY", "NEXT", "QUEUE"],
  settingsTitle: "ARCHIVE",
  settingsJp: "保存設定",
  deviceGap: "videod + yt-dlp missing.",
  sourceNote: "/api/videos · /api/task",
};

export const client = createPortClient({
  resource: "video",
  fixture: { sections, metrics, settings },
  endpoints: {
    list: "/api/videos",
    get: (id) => `/api/videos/${encodeURIComponent(id)}`,
    action: (name) => `/api/${name === "QUEUE" ? "download" : "task"}/${encodeURIComponent(name.toLowerCase())}`,
  },
  act(name, payload, bundle) {
    const items = bundle.sections.flatMap((section) => section.items);
    const item = items.find((entry) => entry.id === payload?.id) ?? items[0];
    if (!item) return { ok: false, message: "no video selected" };
    if (name === "PLAY") { item.state = item.state === "playing" ? "paused" : "playing"; item.progress = Math.min(1, (item.progress ?? 0) + 0.05); return { ok: true, message: `${item.title} ${item.state} · mock player` }; }
    if (name === "NEXT" || name === "PREV") {
      const at = items.findIndex((entry) => entry.id === item.id);
      const next = items[(at + (name === "NEXT" ? 1 : items.length - 1)) % items.length]!;
      return { ok: true, message: `${name}: ${next.title}` };
    }
    if (name === "QUEUE") { item.badge = "QUE"; return { ok: true, message: `${item.title} queued for download` }; }
    return { ok: false, message: `${name} unavailable` };
  },
});
