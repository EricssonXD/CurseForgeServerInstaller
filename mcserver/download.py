from __future__ import annotations

from pathlib import Path
import sys
from urllib.request import urlopen


def _format_bytes(n: int) -> str:
    value = float(n)
    for unit in ("B", "KB", "MB", "GB"):
        if value < 1024 or unit == "GB":
            if unit == "B":
                return f"{int(value)}{unit}"
            return f"{value:.1f}{unit}"
        value /= 1024
    return f"{int(value)}B"


def download_to(
    url: str,
    dest: Path,
    *,
    chunk_size: int = 1024 * 256,
    label: str = "Downloading",
) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    with urlopen(url) as resp, open(dest, "wb") as f:
        total = resp.headers.get("Content-Length")
        total_bytes = int(total) if total and total.isdigit() else None

        downloaded = 0
        last_pct = -1
        last_bytes_print = 0
        while True:
            chunk = resp.read(chunk_size)
            if not chunk:
                break
            f.write(chunk)
            downloaded += len(chunk)

            if total_bytes:
                pct = int(downloaded * 100 / total_bytes)
                if pct != last_pct and (pct % 2 == 0 or pct == 100):
                    sys.stderr.write(
                        f"\r{label}: {pct:3d}% ({_format_bytes(downloaded)} / {_format_bytes(total_bytes)})"
                    )
                    sys.stderr.flush()
                    last_pct = pct
            else:
                if downloaded - last_bytes_print >= 5 * 1024 * 1024:
                    sys.stderr.write(f"\r{label}: {_format_bytes(downloaded)}")
                    sys.stderr.flush()
                    last_bytes_print = downloaded

    sys.stderr.write("\n")
