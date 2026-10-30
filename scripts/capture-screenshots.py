#!/usr/bin/env python3
"""Capture every PocketVideo screen from the live app and verify the result.

The script starts the Go service and the Vite dev server when they are not
running, seeds one video and one audio download so the queue and player states
are real, captures each state in headless Chrome, and then verifies PNG sizes,
unique hashes, and the absence of horizontal overflow at every target viewport.

Usage:
    python3 scripts/capture-screenshots.py [--verify] [--serve] [--no-serve] [--no-seed]
"""
from __future__ import annotations

import hashlib
import json
import os
import shutil
import signal
import struct
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1] / "apps" / "pocket-video"
SHOTS = ROOT / "docs" / "screenshots"
FRONTEND = ROOT / "frontend"
BACKEND = ROOT / "backend"
APP = os.environ.get("POCKETVIDEO_CAPTURE_URL", "http://127.0.0.1:5173").rstrip("/")
API = os.environ.get("POCKETVIDEO_API_URL", "http://127.0.0.1:8080").rstrip("/")

# name, css width, css height, dpi, route, theme, language
CAPTURES = [
    ("home-1280x720.png", 1280, 720, 1, "home", "light", "en"),
    ("favorites-1280x720.png", 1280, 720, 1, "favorites", "light", "en"),
    ("collection-1280x720.png", 1280, 720, 1, "favorites/fav-pjs", "light", "en"),
    ("videos-1280x720.png", 1280, 720, 1, "videos", "light", "en"),
    ("detail-1280x720.png", 1280, 720, 1, "detail/BV1DEMO002", "light", "en"),
    ("detail-audio-1280x720.png", 1280, 720, 1, "detail/BV1DEMO004", "light", "en"),
    ("queue-1280x720.png", 1280, 720, 1, "queue", "light", "en"),
    ("settings-1280x720.png", 1280, 720, 1, "settings", "light", "en"),
    ("language-zh-1280x720.png", 1280, 720, 1, "settings", "light", "zh-CN"),
    ("language-ja-1280x720.png", 1280, 720, 1, "settings", "light", "ja"),
    ("theme-dark-1280x720.png", 1280, 720, 1, "home", "dark", "en"),
    ("view-320-240.png", 320, 240, 1, "videos", "light", "en"),
    ("view-480-320.png", 480, 320, 1, "videos", "light", "en"),
    ("view-640-480.png", 640, 480, 1, "videos", "light", "en"),
    ("view-720-720.png", 720, 720, 1, "videos", "light", "en"),
    ("view-1280-720.png", 1280, 720, 1, "favorites", "light", "ja"),
    ("view-1920-1080.png", 1920, 1080, 1, "home", "light", "en"),
    ("dpi-1x.png", 1280, 720, 1, "detail/BV1DEMO002", "dark", "en"),
    ("dpi-1_5x.png", 1280, 720, 1.5, "detail/BV1DEMO002", "dark", "en"),
    ("dpi-2x.png", 1280, 720, 2, "detail/BV1DEMO002", "dark", "en"),
    ("dpi-3x.png", 1280, 720, 3, "detail/BV1DEMO002", "dark", "en"),
]

# Narrow viewports that must not introduce horizontal scrolling.
OVERFLOW_CHECKS = [(320, 240), (480, 320), (640, 480), (720, 720), (1280, 720)]
DEMO_DOWNLOADS = [("BV1DEMO002", "video"), ("BV1DEMO004", "audio")]


def find_chrome() -> str | None:
    candidates = [os.environ.get("CHROME_BIN", ""), "chromium", "chromium-browser", "google-chrome", "google-chrome-stable"]
    for name in candidates:
        if name and (path := shutil.which(name)):
            return path
    for cache in (Path.home() / ".cache" / "ms-playwright", Path.home() / ".hermes" / "tools"):
        for path in sorted(cache.glob("chromium-*/chrome-linux*/chrome"), reverse=True):
            if path.is_file():
                return str(path)
    return None


CHROME: str = find_chrome() or ""


def http_json(path: str, method: str = "GET", payload: dict | None = None) -> Any:
    data = json.dumps(payload).encode() if payload is not None else None
    request = urllib.request.Request(
        API + path,
        data=data,
        method=method,
        headers={"Content-Type": "application/json"} if data else {},
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        body = response.read()
        return json.loads(body) if body else None


def wait_for(url: str, timeout: float = 90.0) -> bool:
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=2) as response:
                if response.status < 500:
                    return True
        except (urllib.error.URLError, TimeoutError, ConnectionError):
            time.sleep(0.4)
    return False


def app_ready() -> bool:
    return wait_for(APP + "/", 1) and wait_for(API + "/api/health", 1)


def spawn(command: list[str], cwd: Path, env: dict[str, str] | None = None) -> subprocess.Popen:
    log = tempfile.NamedTemporaryFile(prefix="pocketvideo-", suffix=".log", delete=False)
    return subprocess.Popen(
        command,
        cwd=str(cwd),
        env=env,
        stdout=log,
        stderr=subprocess.STDOUT,
        start_new_session=True,
    )


class Servers:
    """Starts only what is missing and only stops what it started."""

    def __init__(self, serve: bool) -> None:
        self.serve = serve
        self.processes: list[subprocess.Popen] = []
        self.data_dir = tempfile.mkdtemp(prefix="pocketvideo-capture-")

    def __enter__(self) -> "Servers":
        if not self.serve:
            return self
        if not wait_for(API + "/api/health", 2):
            go = os.environ.get("GO_BIN") or shutil.which("go")
            if not go:
                raise SystemExit("go is required to start the backend; set GO_BIN or run `make backend`")
            env = os.environ.copy()
            env.update({"POCKETVIDEO_DATA": self.data_dir, "POCKETVIDEO_ADDR": "127.0.0.1:8080"})
            self.processes.append(spawn([go, "run", "."], BACKEND, env))
            if not wait_for(API + "/api/health"):
                raise SystemExit("backend did not become healthy")
        if not wait_for(APP + "/", 2):
            npm = shutil.which("npm")
            if not npm:
                raise SystemExit("npm is required; install Node.js first")
            if not (FRONTEND / "node_modules").exists():
                subprocess.run([npm, "install", "--no-audit", "--no-fund"], cwd=FRONTEND, check=True)
            self.processes.append(spawn([npm, "run", "dev", "--", "--host", "127.0.0.1", "--port", "5173"], FRONTEND))
            if not wait_for(APP + "/"):
                raise SystemExit("frontend did not become ready")
        return self

    def __exit__(self, *_exc: object) -> None:
        for process in self.processes:
            if process.poll() is None:
                try:
                    os.killpg(os.getpgid(process.pid), signal.SIGTERM)
                except ProcessLookupError:
                    pass
        shutil.rmtree(self.data_dir, ignore_errors=True)


def seed_demo_media() -> None:
    """Make the queue and the offline player show real downloads.

    A dev server that is still proxying to a restarting backend can answer 500
    for a moment, so the seed step retries before giving up.
    """
    for attempt in range(3):
        try:
            seed_demo_media_once()
            return
        except (urllib.error.URLError, urllib.error.HTTPError):
            if attempt == 2:
                raise
            time.sleep(1.0)


def seed_demo_media_once() -> None:
    tasks = http_json("/api/task") or []
    done = {f"{task['video_id']}:{task['kind']}" for task in tasks if task["status"] == "done"}
    for video_id, kind in DEMO_DOWNLOADS:
        if f"{video_id}:{kind}" not in done:
            http_json("/api/download", "POST", {"video_id": video_id, "kind": kind})
    wanted = {f"{video_id}:{kind}" for video_id, kind in DEMO_DOWNLOADS}
    deadline = time.time() + 120
    while time.time() < deadline:
        tasks = http_json("/api/task") or []
        done = {f"{task['video_id']}:{task['kind']}" for task in tasks if task["status"] == "done"}
        failed = [task for task in tasks if task["status"] == "failed"]
        if failed:
            raise SystemExit(f"demo download failed: {failed[0].get('error')}")
        if wanted <= done:
            return
        time.sleep(0.3)
    raise SystemExit("demo downloads did not finish in time")


def chrome_args(width: int, height: int, dpi: float, profile: str, extra: list[str]) -> list[str]:
    return [
        CHROME,
        "--headless=new",
        "--disable-gpu",
        "--no-sandbox",
        "--hide-scrollbars",
        "--no-first-run",
        "--disable-dev-shm-usage",
        f"--user-data-dir={profile}",
        "--virtual-time-budget=4000",
        f"--force-device-scale-factor={dpi}",
        f"--window-size={width},{height}",
        *extra,
    ]


def capture(name: str, width: int, height: int, dpi: float, route: str, theme: str, lang: str) -> None:
    out = SHOTS / name
    url = f"{APP}/?theme={theme}&lang={lang}#{route}"
    with tempfile.TemporaryDirectory(prefix="pocketvideo-chrome-") as profile:
        command = chrome_args(width, height, dpi, profile, [f"--screenshot={out}", url])
        result = subprocess.run(command, capture_output=True, text=True, timeout=90)
    if result.returncode != 0 or not out.exists():
        raise SystemExit(f"capture failed for {name}: {result.stderr.strip()}")
    print(f"captured {name} <- {url}")


def overflow_free(width: int, height: int) -> bool:
    url = f"{APP}/?theme=light&lang=en#videos"
    with tempfile.TemporaryDirectory(prefix="pocketvideo-chrome-") as profile:
        command = chrome_args(width, height, 1, profile, ["--dump-dom", url])
        result = subprocess.run(command, capture_output=True, text=True, timeout=90)
    if result.returncode != 0:
        raise SystemExit(f"overflow probe failed at {width}x{height}: {result.stderr.strip()}")
    return 'data-overflow="true"' not in result.stdout


def png_size(path: Path) -> tuple[int, int]:
    data = path.read_bytes()[:24]
    if data[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit(f"{path} is not a PNG")
    return struct.unpack(">II", data[16:24])


def check_images() -> None:
    expected = {name: (round(width * dpi), round(height * dpi)) for name, width, height, dpi, *_ in CAPTURES}
    files = sorted(SHOTS.glob("*.png"))
    names = {file.name for file in files}
    missing = set(expected) - names
    extra = names - set(expected)
    if missing or extra:
        raise SystemExit(f"screenshot set mismatch; missing={sorted(missing)} extra={sorted(extra)}")
    seen: dict[str, list[str]] = {}
    for file in files:
        actual = png_size(file)
        if actual != expected[file.name]:
            raise SystemExit(f"{file.name}: PNG size {actual}, expected {expected[file.name]}")
        digest = hashlib.sha256(file.read_bytes()).hexdigest()
        seen.setdefault(digest, []).append(file.name)
    duplicates = {digest: group for digest, group in seen.items() if len(group) > 1}
    print(f"png files={len(files)} unique_sha256={len(seen)}")
    if duplicates:
        for digest, group in duplicates.items():
            print(f"DUPLICATE {digest[:16]}: {', '.join(group)}")
        raise SystemExit("duplicate screenshots")


def check_layouts() -> None:
    for width, height in OVERFLOW_CHECKS:
        if not overflow_free(width, height):
            raise SystemExit(f"horizontal overflow at {width}x{height}")
        print(f"no horizontal overflow at {width}x{height}")


def main() -> None:
    if not CHROME:
        raise SystemExit("headless Chrome is required (set CHROME_BIN or install chromium)")
    verifying = "--verify" in sys.argv
    if not verifying:
        SHOTS.mkdir(parents=True, exist_ok=True)
    serve = "--no-serve" not in sys.argv and ("--serve" in sys.argv or not verifying or not app_ready())
    with Servers(serve):
        if not verifying and "--no-seed" not in sys.argv:
            seed_demo_media()
        if not verifying:
            for entry in CAPTURES:
                capture(*entry)
        check_images()
        if app_ready():
            check_layouts()
        else:
            print("layout checks skipped: start the app or pass --serve")


if __name__ == "__main__":
    main()
