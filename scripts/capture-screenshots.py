#!/usr/bin/env python3
"""Capture distinct live PocketVideo states and verify PNG uniqueness/dimensions."""
import hashlib
import os
import shutil
import struct
import sys
from pathlib import Path
import subprocess
ROOT = Path(__file__).resolve().parents[1] / "apps" / "pocket-video"
SHOTS = ROOT / "docs" / "screenshots"
APP = os.environ.get("POCKETVIDEO_CAPTURE_URL", "http://127.0.0.1:5173").rstrip("/")
CHROME = shutil.which("chromium") or shutil.which("chromium-browser")

# name, css width, css height, dpi, route, theme, language
CAPTURES = [
    ("home-1280x720.png", 1280, 720, 1, "home", "light", "en"),
    ("favorites-1280x720.png", 1280, 720, 1, "favorites", "light", "en"),
    ("videos-1280x720.png", 1280, 720, 1, "videos", "light", "en"),
    ("detail-1280x720.png", 1280, 720, 1, "detail", "light", "en"),
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
    ("dpi-1x.png", 1280, 720, 1, "detail", "dark", "en"),
    ("dpi-1_5x.png", 1280, 720, 1.5, "detail", "dark", "en"),
    ("dpi-2x.png", 1280, 720, 2, "detail", "dark", "en"),
    ("dpi-3x.png", 1280, 720, 3, "detail", "dark", "en"),
]


def png_size(path):
    data = path.read_bytes()[:24]
    if data[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit(f"{path} is not a PNG")
    return struct.unpack(">II", data[16:24])


def capture(name, width, height, dpi, route, theme, lang):
    out = SHOTS / name
    url = f"{APP}/?theme={theme}&lang={lang}#{route}"
    command = [
        CHROME, "--headless=new", "--disable-gpu", "--no-sandbox", "--hide-scrollbars",
        "--virtual-time-budget=2500",
        f"--force-device-scale-factor={dpi}", f"--window-size={width},{height}",
        f"--screenshot={out}", url,
    ]
    result = subprocess.run(command, capture_output=True, text=True, timeout=90)
    if result.returncode != 0 or not out.exists():
        raise SystemExit(f"capture failed for {name}: {result.stderr.strip()}")
    print(f"captured {name} <- {url}")

def verify():
    expected = {name: (round(width * dpi), round(height * dpi)) for name, width, height, dpi, *_ in CAPTURES}
    files = sorted(SHOTS.glob("*.png"))
    names = {file.name for file in files}
    missing = set(expected) - names
    extra = names - set(expected)
    if missing or extra:
        raise SystemExit(f"screenshot set mismatch; missing={sorted(missing)} extra={sorted(extra)}")
    seen = {}
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
        return False
    return True


def main():
    if "--verify" in sys.argv:
        if not verify():
            raise SystemExit("screenshot uniqueness check failed")
        return
    if not CHROME:
        raise SystemExit("chromium is required to capture screenshots")
    SHOTS.mkdir(parents=True, exist_ok=True)
    for entry in CAPTURES:
        capture(*entry)
    if not verify():
        raise SystemExit("screenshot uniqueness check failed")


if __name__ == "__main__":
    main()
