from __future__ import annotations

import shutil
from pathlib import Path
from urllib.request import urlopen


def download_to(url: str, dest: Path, *, chunk_size: int = 1024 * 256) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    with urlopen(url) as resp, open(dest, "wb") as f:
        shutil.copyfileobj(resp, f, length=chunk_size)
