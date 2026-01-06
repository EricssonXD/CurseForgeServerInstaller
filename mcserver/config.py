from __future__ import annotations

import json
import os
from dataclasses import dataclass
from pathlib import Path
from typing import Optional


def _config_dir() -> Path:
    xdg = os.environ.get("XDG_CONFIG_HOME")
    if xdg:
        return Path(xdg) / "mcserver"
    return Path.home() / ".config" / "mcserver"


def config_path() -> Path:
    return _config_dir() / "config.json"


@dataclass
class AppConfig:
    curseforge_api_key: Optional[str] = None

    @staticmethod
    def load() -> "AppConfig":
        path = config_path()
        if not path.exists():
            return AppConfig()
        data = json.loads(path.read_text(encoding="utf-8"))
        return AppConfig(curseforge_api_key=data.get("curseforgeApiKey"))

    def save(self) -> None:
        path = config_path()
        path.parent.mkdir(parents=True, exist_ok=True)
        payload = {"curseforgeApiKey": self.curseforge_api_key}
        path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
        try:
            os.chmod(path, 0o600)
        except Exception:
            # Best-effort on platforms that may not support chmod
            pass


def mask_secret(value: Optional[str]) -> str:
    if not value:
        return "(not set)"
    if len(value) <= 6:
        return "***"
    return value[:2] + "***" + value[-2:]
