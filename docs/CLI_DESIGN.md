# CurseForgeServerInstaller — CLI Design (Proposed)

## Goal
Turn this repo into a single cohesive CLI tool that can:

- Resolve a CurseForge modpack (URL/name) → pack ID
- Find the latest *server pack* file (or a user-selected version)
- Fetch a direct download URL / download the ZIP
- Install a Minecraft server from that server-pack ZIP
- Update an existing server in-place while preserving critical data (world + server config)

This should unify the current:
- `mcserver-installer/get_modpack_id.py` (URL → pack ID)
- `mcserver-installer/download_server_pack.py` (pack ID → download server pack ZIP)
- `mc-update` (update server folder from a server-pack ZIP URL)

## Non-goals (for v1)
- GUI
- Mod/launcher support beyond CurseForge server packs
- Managing multiple java distributions

## Key design constraints from current repo
- Updates must preserve `world/`, `server.properties`, whitelists, ops, and JVM args.
- Updates should replace modpack-managed folders (`mods/`, `config/`, etc.).
- CurseForge API is required for resolving pack IDs and file metadata.

## Security / config note (important)
The current code hardcodes a CurseForge API key inside Python files. For a real CLI:

- Store it in a user config file (e.g. `~/.config/mcserver/config.toml`) with clear instructions.

Avoid committing keys into the repo.

## Proposed CLI name
Pick one and standardize it:

- `mcserver` (recommended, short)
- `curseforge-server` (explicit)

Examples below use `mcserver`.

## Commands (v1)
### 0) Core principle: one smart command
Primary UX is a single command that “does the right thing” in the current directory:

- `mcserver install [SOURCE]`

Where:
- If the current directory already looks like a server (e.g. `server.properties` exists), it performs an **update**.
- Otherwise, it performs an **install** into the current directory.

`SOURCE` is optional and auto-detected:
- If `SOURCE` looks like a URL (`http(s)://.../minecraft/modpacks/...`) treat it as a CurseForge modpack URL.
- If `SOURCE` is digits-only, treat it as a modpack ID.
- If `SOURCE` is omitted, read the saved modpack ID from the current folder.

Keep explicit commands (`update`, `status`, `cf ...`) for power users, but the happy-path should be the single line:

- `mcserver install 14810`

### 1) CurseForge helpers
These are building blocks and also useful standalone.

- `mcserver cf resolve <curseforge_modpack_url>`
  - Output: pack ID

- `mcserver cf search <query> [--game-version <ver>] [--limit N]`
  - Output: list of matching modpacks with IDs

- `mcserver cf files <pack_id> [--server-only] [--limit N]`
  - Output: files/versions metadata (display name, date, fileId, serverPackFileId, isServerPack)

- `mcserver cf download-url <pack_id> [--file-id <id> | --latest] [--server-pack]`
  - Output: direct download URL for the file you want (server pack preferred)
  - Notes:
    - If the latest file is *not* a server pack, use `serverPackFileId` (your notebook cell demonstrates this).

### 2) Install
Canonical (recommended):

- `mcserver install [SOURCE] [--dir <server_dir>] [--file-id <id> | --latest] [--accept-eula]`

Where:
- `--dir` defaults to `.`
- If the directory is already a server directory, `install` switches to update mode automatically.

Also acceptable (explicit, optional to implement):

- `mcserver install --pack-id <id> [--dir <server_dir>] ...`
- `mcserver install --pack-url <url> [--dir <server_dir>] ...`

Saved modpack ID behavior:

- On successful install/update, write pack metadata into `<server_dir>/.mcserver/state.json`.
- When `mcserver install` is run with no `SOURCE`, it reads `packId` from `<server_dir>/.mcserver/state.json`.
- If both are present and they differ (saved `packId` vs provided `SOURCE`), prompt:
  - “Folder is configured for packId=X but you provided Y. Which should I use?”
  - Non-interactive mode should require an explicit choice flag.

Suggested flags for mismatch/non-interactive:

- `--use-saved` (prefer the folder’s saved packId)
- `--use-arg` (prefer the command-line SOURCE)
- `--no-prompt` (fail with a clear message unless `--use-saved`/`--use-arg` is set)

Install responsibilities:

- Download server pack ZIP
- Extract into target directory
- Ensure `eula.txt` exists and set `eula=true` if `--accept-eula`
- Write a local state file so updates know what pack/version you installed

### 3) Update
The CLI may expose `mcserver update`, but it should be effectively:

- `mcserver install [SOURCE] --dir <server_dir>`

If you keep an explicit update command, it can simply call the same implementation:

- `mcserver update [SOURCE] [--dir <server_dir>] [--check-only] [--backup] [--dry-run]`

Update responsibilities (mirrors current `mc-update`):

- Verify `server.properties` exists (safety check)
- Determine pack/version to update *from local state* (default) or from explicit flags
- Download server pack ZIP to temp
- Extract
- Detect pack root (find `mods/` within extracted zip, same as `mc-update`)
- Replace these directories if present:
  - `mods/`, `config/`, `scripts/`, `kubejs/`, `libraries/`, `defaultconfigs/`
- Copy top-level `*.jar`, `*.sh`, `*.bat` into server dir (but do **not** overwrite `user_jvm_args.txt`)
- Preserve:
  - `world/`, `server.properties`, `whitelist.json`, `ops.json`, `user_jvm_args.txt`

Optional safety flags:

- `--backup` creates a timestamped backup of `world/` and maybe config snapshots.
- `--dry-run` prints what would be replaced/copied.

### 4) Status
- `mcserver status --dir <server_dir>`

Show:
- installed pack ID
- installed file ID / displayName (from state)
- last update timestamp

## Local state & folder layout
To make update deterministic, store metadata under the server directory:

- `<server_dir>/.mcserver/state.json`

Suggested schema (minimal):

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

## Module layout (Python)
Today you have `mcserver-installer/` (hyphenated) which is awkward as a Python package name.

Recommended:

- Create a proper package, e.g. `curseforgeserverinstaller/` or `mcserver/`
- Move requester + CF logic under `mcserver/curseforge.py`
- Move installer/update logic under `mcserver/install.py`, `mcserver/update.py`
- Keep CLI entry under `mcserver/cli.py`

## Packaging (pyproject.toml)
Add a console script entrypoint:

```toml
[project]
name = "mcserver"
...

[project.scripts]
mcserver = "mcserver.cli:main"
```

For CLI parsing:
- `argparse` is fine (no deps).
- If you want a nicer UX quickly, consider `typer` (adds dependency).

## Error handling (UX)
- Fail fast with actionable messages (missing API key, missing server dir markers, network issues).
- Use non-zero exit codes.
- Provide `--verbose` to print HTTP status/response excerpts.

## “MVP first” milestone checklist
1. `mcserver cf download-url` works for pack ID (matches your notebook logic)
2. `mcserver install` downloads+extracts server pack
3. `mcserver update` replicates `mc-update` behavior using saved state
