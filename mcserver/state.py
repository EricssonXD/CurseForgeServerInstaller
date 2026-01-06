from __future__ import annotations

import json
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional


STATE_DIRNAME = ".mcserver"
STATE_FILENAME = "state.json"


def utc_now_iso() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


@dataclass
class ServerState:
    provider: str = "curseforge"
    pack_id: Optional[int] = None
    installed_file_id: Optional[int] = None
    installed_display_name: Optional[str] = None
    channel: str = "latest"
    last_updated_at: Optional[str] = None

    @staticmethod
    def load(server_dir: Path) -> "ServerState | None":
        path = server_dir / STATE_DIRNAME / STATE_FILENAME
        if not path.exists():
            return None
        data = json.loads(path.read_text(encoding="utf-8"))
        return ServerState(
            provider=str(data.get("provider", "curseforge")),
            pack_id=data.get("packId"),
            installed_file_id=data.get("installedFileId"),
            installed_display_name=data.get("installedDisplayName"),
            channel=str(data.get("channel", "latest")),
            last_updated_at=data.get("lastUpdatedAt"),
        )

    def save(self, server_dir: Path) -> None:
        path = server_dir / STATE_DIRNAME / STATE_FILENAME
        path.parent.mkdir(parents=True, exist_ok=True)
        payload = {
            "provider": self.provider,
            "packId": self.pack_id,
            "installedFileId": self.installed_file_id,
            "installedDisplayName": self.installed_display_name,
            "channel": self.channel,
            "lastUpdatedAt": self.last_updated_at or utc_now_iso(),
        }
        path.write_text(json.dumps(payload, indent=2, sort_keys=False) + "\n", encoding="utf-8")
