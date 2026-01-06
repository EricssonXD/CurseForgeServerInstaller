# CurseForgeServerInstaller

This repo is evolving into a single CLI tool to install and update CurseForge-based Minecraft servers.

## Current state
- `mc-update` (bash): updates an existing server folder from a *server pack ZIP URL* by replacing modpack-managed folders while preserving world + key server files.
- `mcserver-installer/*.py`: small Python scripts to interact with the CurseForge API (resolve modpack IDs, list files, download server pack ZIP).
- `test.ipynb`: experimentation (includes working logic for getting a download URL from a pack ID).

## Planning docs
- CLI design proposal: [docs/CLI_DESIGN.md](docs/CLI_DESIGN.md)
- Flowcharts: [docs/FLOWCHART.md](docs/FLOWCHART.md)

## Intended basic usage (planned)
- `mcserver install 14810` (auto-detects whether to install or update in the current directory)
- `mcserver install <curseforge_modpack_url>` (same, but uses URL)
- `mcserver install` (uses saved packId from `.mcserver/state.json` in the current folder)

## Next implementation steps (suggested)
- Convert `mcserver-installer/` into a real Python package (hyphen-free name)
- Add a `pyproject.toml` console script entrypoint (e.g. `mcserver`)
- Replace hardcoded API key with `CURSEFORGE_API_KEY` env var or user config
- Port `mc-update` logic into Python (or call the script from the CLI initially)

## Install the CLI (local)
- Run: `./install.sh`
- Ensure `~/.local/bin` is on your `PATH` (the script will tell you if it isn't)
- Set API key: `export CURSEFORGE_API_KEY=...`
- Or save it permanently: `mcserver config set-api-key '...'`
- Then you can run: `mcserver --help`
