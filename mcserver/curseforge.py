from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Tuple
from urllib.error import HTTPError

from .errors import InvalidApiKeyError, MissingApiKeyError, UserFacingError
from .config import AppConfig
from .http_client import http_get_json


@dataclass(frozen=True)
class ModFile:
    id: int
    display_name: str
    file_date: str
    is_server_pack: bool
    server_pack_file_id: Optional[int]
    download_url: Optional[str]


class CurseForgeClient:
    BASE_URL = "https://api.curseforge.com"

    def __init__(self, api_key: Optional[str] = None):
        cfg = AppConfig.load()
        self.api_key = api_key or cfg.curseforge_api_key
        if not self.api_key:
            raise MissingApiKeyError(
                "Missing CurseForge API key. Run: mcserver config set-api-key"
            )

    def _wrap_http_errors(self, fn, *args, **kwargs):
        try:
            return fn(*args, **kwargs)
        except HTTPError as e:
            if e.code == 403:
                raise InvalidApiKeyError(
                    "CurseForge API returned 403 Forbidden (API key invalid). "
                    "Update it with: mcserver config set-api-key"
                )
            raise UserFacingError(
                f"CurseForge API request failed: HTTP {e.code} {e.reason}"
            )

    def _headers(self) -> Dict[str, str]:
        return {"Accept": "application/json", "x-api-key": self.api_key}

    def search_modpacks(
        self,
        *,
        query: str,
        game_version: Optional[str] = None,
        index: int = 0,
        page_size: int = 10,
        sort_field: int = 2,
        sort_order: str = "desc",
    ) -> List[Dict[str, Any]]:
        params: Dict[str, Any] = {
            "gameId": 432,
            "classId": 4471,
            "index": index,
            "pageSize": page_size,
            "searchFilter": query,
            "sortField": sort_field,
            "sortOrder": sort_order,
        }
        if game_version:
            params["gameVersion"] = game_version
        payload = self._wrap_http_errors(
            http_get_json,
            f"{self.BASE_URL}/v1/mods/search",
            headers=self._headers(),
            params=params,
        )
        return payload.get("data", [])

    def resolve_pack_id_from_url(self, url: str) -> int:
        match = re.search(r"/modpacks/([^/?#]+)", url)
        if not match:
            raise UserFacingError(
                "Invalid CurseForge modpack URL (expected /minecraft/modpacks/<slug>)."
            )
        slug = match.group(1)
        query = slug.replace("-", " ")
        results = self.search_modpacks(query=query, page_size=5)
        if not results:
            raise UserFacingError("No modpack found for the given URL.")
        return int(results[0]["id"])

    def list_files(self, pack_id: int) -> List[ModFile]:
        payload = self._wrap_http_errors(
            http_get_json,
            f"{self.BASE_URL}/v1/mods/{pack_id}/files",
            headers=self._headers(),
        )
        files: List[ModFile] = []
        for item in payload.get("data", []):
            files.append(
                ModFile(
                    id=int(item.get("id")),
                    display_name=str(item.get("displayName", "")),
                    file_date=str(item.get("fileDate", "")),
                    is_server_pack=bool(item.get("isServerPack", False)),
                    server_pack_file_id=item.get("serverPackFileId"),
                    download_url=item.get("downloadUrl"),
                )
            )
        return files

    def get_download_url(self, pack_id: int, file_id: int) -> str:
        payload = self._wrap_http_errors(
            http_get_json,
            f"{self.BASE_URL}/v1/mods/{pack_id}/files/{file_id}/download-url",
            headers=self._headers(),
        )
        data = payload.get("data")
        if not data:
            raise UserFacingError("Could not resolve download URL from CurseForge API.")
        return str(data)

    def choose_latest_server_pack(self, pack_id: int) -> Tuple[int, str, str]:
        """Returns (server_pack_file_id, display_name, file_date)."""
        files = self.list_files(pack_id)
        if not files:
            raise UserFacingError("No files found for this modpack.")

        # Prefer explicit server packs; otherwise use newest file that points to serverPackFileId.
        server_candidates = [f for f in files if f.is_server_pack]
        if server_candidates:
            server_candidates.sort(key=lambda f: f.file_date, reverse=True)
            f = server_candidates[0]
            return f.id, f.display_name, f.file_date

        # Fallback: newest file overall that has serverPackFileId
        files.sort(key=lambda f: f.file_date, reverse=True)
        newest = files[0]
        if newest.server_pack_file_id:
            return (
                int(newest.server_pack_file_id),
                newest.display_name,
                newest.file_date,
            )

        raise UserFacingError("No server pack found for this modpack.")

    def resolve_server_pack_download(
        self, pack_id: int, *, file_id: Optional[int] = None
    ) -> Tuple[str, int, str]:
        """Returns (download_url, server_pack_file_id, display_name)."""
        if file_id is None:
            server_file_id, display_name, _file_date = self.choose_latest_server_pack(
                pack_id
            )
        else:
            # User provided a file id. It might be a server pack or a normal file.
            # We resolve it by scanning for an exact id match; if it's not a server pack
            # but has serverPackFileId, use that.
            files = self.list_files(pack_id)
            match = next((f for f in files if f.id == file_id), None)
            if not match:
                raise UserFacingError(
                    f"File id {file_id} not found for pack {pack_id}."
                )
            display_name = match.display_name
            if match.is_server_pack:
                server_file_id = match.id
            elif match.server_pack_file_id:
                server_file_id = int(match.server_pack_file_id)
            else:
                raise UserFacingError(
                    "Selected file does not have an associated server pack."
                )

        url = self.get_download_url(pack_id, server_file_id)
        return url, int(server_file_id), str(display_name)
