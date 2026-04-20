# Command Map

Exact mapping of every CLI command, flag, argument, and env var from the current
Python implementation (`mcserver/cli.py`) to the target Go/Cobra CLI.

---

## 1. Top-level commands

| Python command | Cobra command | Notes |
|----------------|---------------|-------|
| `mcserver install` | `cfs install` | Smart install/update |
| `mcserver update` | `cfs update` | Alias of install + `--check-only` |
| `mcserver status` | `cfs status` | Show saved pack/version |
| `mcserver cf ...` | `cfs cf ...` | CurseForge helpers |
| `mcserver config ...` | `cfs config ...` | Config management |

---

## 2. `install` command

```
mcserver install [SOURCE] [flags]
```

| Argument / Flag | Type | Default | Description |
|----------------|------|---------|-------------|
| `SOURCE` | positional, optional | — | Modpack ID (digits) or CurseForge modpack URL |
| `--dir` | string | `.` | Target server directory |
| `--file-id` | int | — | Specific CurseForge file ID |
| `--accept-eula` | bool | false | Write `eula.txt` with `eula=true` |
| `--use-saved` | bool | false | On pack ID mismatch, prefer saved ID |
| `--use-arg` | bool | false | On pack ID mismatch, prefer argument ID |
| `--no-prompt` | bool | false | Fail on ambiguity instead of prompting |

**Behavior:**
- If `server.properties` exists in `--dir`, runs in **update** mode.
- Otherwise runs in **install** mode.
- Downloads server pack ZIP, extracts, copies into target.
- On update: replaces `mods/`, `config/`, `scripts/`, `kubejs/`, `libraries/`, `defaultconfigs/`, copies `*.jar`, `*.sh`, `*.bat` (preserves `user_jvm_args.txt`, `world/`, `server.properties`).
- Saves state to `.mcserver/state.json`.

---

## 3. `update` command

```
mcserver update [SOURCE] [flags]
```

Same flags as `install` plus:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--check-only` | bool | false | Only check if an update is available (don't apply) |

Functionally identical to `install`; the `update` name is an alias.

---

## 4. `status` command

```
mcserver status [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dir` | string | `.` | Server directory to inspect |

**Output:** Reads `.mcserver/state.json` and prints `packId`, `installedFileId`, `installedDisplayName`, `lastUpdatedAt`.

---

## 5. `cf` subcommands

### 5.1 `cf resolve`

```
mcserver cf resolve <URL>
```

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `URL` | positional | yes | CurseForge modpack URL |

**Output:** numeric pack ID.

### 5.2 `cf search`

```
mcserver cf search <QUERY> [flags]
```

| Argument / Flag | Type | Default | Description |
|----------------|------|---------|-------------|
| `QUERY` | positional | yes | Search string |
| `--game-version` | string | — | Filter by Minecraft version |
| `--limit` | int | 10 | Max results |

**Output:** TSV lines `<id>\t<name>`.

### 5.3 `cf files`

```
mcserver cf files <PACK_ID> [flags]
```

| Argument / Flag | Type | Default | Description |
|----------------|------|---------|-------------|
| `PACK_ID` | positional | yes | CurseForge modpack ID |
| `--server-only` | bool | false | Only show server packs |
| `--limit` | int | 20 | Max results |

**Output:** TSV lines with file metadata.

### 5.4 `cf download-url`

```
mcserver cf download-url <PACK_ID> [flags]
```

| Argument / Flag | Type | Default | Description |
|----------------|------|---------|-------------|
| `PACK_ID` | positional | yes | CurseForge modpack ID |
| `--file-id` | int | — | Specific file to resolve |
| `--verbose` | bool | false | Print extra metadata |

**Output:** Direct download URL.

---

## 6. `config` subcommands

### 6.1 `config set-api-key`

```
mcserver config set-api-key [API_KEY]
```

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `API_KEY` | positional | no | If omitted, prompts interactively |

Saves key to `~/.config/mcserver/config.json` (respects `XDG_CONFIG_HOME`).

### 6.2 `config show`

```
mcserver config show
```

Prints config path and masked API key.

### 6.3 `config unset-api-key`

```
mcserver config unset-api-key
```

Clears the saved API key.

### 6.4 `config path`

```
mcserver config path
```

Prints the config file path.

---

## 7. Environment variables

| Variable | Used by | Description |
|----------|---------|-------------|
| `XDG_CONFIG_HOME` | `config.py` | Override default config directory |

---

## 8. Config file

- **Path:** `$XDG_CONFIG_HOME/mcserver/config.json` or `~/.config/mcserver/config.json`
- **Schema:**
  ```json
  {
    "curseforgeApiKey": "<string>"
  }
  ```
- **Permissions:** `0600` (best-effort)

---

## 9. State file

- **Path:** `<server_dir>/.mcserver/state.json`
- **Schema:**
  ```json
  {
    "provider": "curseforge",
    "packId": 1409114,
    "installedFileId": 1234567,
    "installedDisplayName": "Some Pack v1.2.3",
    "channel": "latest",
    "lastUpdatedAt": "2026-01-06T00:00:00Z"
  }
  ```

---

## 10. Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | User-facing error (missing key, bad input, API failure) |
| 130 | Interrupted (Ctrl+C) |

---

## 11. Go wrapper phase — Python passthrough

In Phase 2 the Cobra subcommands will delegate to:

```
python3 -m mcserver.cli <args...>
```

with `--python-binary` flag on the root command to override the Python interpreter.
