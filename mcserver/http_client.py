from __future__ import annotations

import json
import time
from dataclasses import dataclass
from typing import Any, Dict, Optional
from urllib.parse import urlencode
from urllib.request import Request, urlopen


@dataclass
class HttpResponse:
    status: int
    headers: Any
    content: bytes

    @property
    def text(self) -> str:
        charset = None
        try:
            charset = self.headers.get_content_charset()
        except Exception:
            charset = None
        if not charset:
            charset = "utf-8"
        return self.content.decode(charset, errors="replace")

    def json(self) -> Any:
        return json.loads(self.content)


def http_get_json(
    url: str,
    *,
    headers: Dict[str, str],
    params: Optional[Dict[str, Any]] = None,
    retries: int = 3,
    retry_sleep_s: float = 0.5,
    timeout_s: int = 60,
) -> Any:
    if params:
        url = url + "?" + urlencode(params)

    request = Request(url, headers=headers, method="GET")

    last_exc: Exception | None = None
    for attempt in range(1, retries + 1):
        try:
            with urlopen(request, timeout=timeout_s) as resp:
                content = resp.read()
                http_resp = HttpResponse(
                    status=getattr(resp, "status", 200),
                    headers=resp.headers,
                    content=content,
                )
            return http_resp.json()
        except Exception as exc:  # pragma: no cover
            last_exc = exc
            if attempt >= retries:
                raise
            time.sleep(retry_sleep_s)

    raise last_exc  # type: ignore[misc]
