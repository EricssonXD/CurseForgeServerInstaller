# CurseForge Server Installer (`cfs`)

A CLI tool to install, update, and manage CurseForge-based Minecraft modpack servers.  
Written in Go. Single binary, no runtime dependencies.

## Install

### One-liner (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/EricssonXD/CurseForgeServerInstaller/master/install-cfs.sh | sh
```

Shell completion (bash/zsh/fish) is configured automatically.  
To install into a custom directory: `CFS_INSTALL_DIR=~/.local/bin curl -fsSL ... | sh`

### Build from source

```bash
git clone https://github.com/EricssonXD/CurseForgeServerInstaller.git
cd CurseForgeServerInstaller
./setup-dev.sh          # installs deps, builds, runs tests, sets up completion
```

### Download a pre-built binary

Grab the binary for your platform from [Releases](../../releases), extract, and place `cfs` on your `PATH`.

## Quick start

```bash
# 1. Save your CurseForge API key (get one at https://console.curseforge.com)
cfs config set-api-key

# 2. Install a modpack server
cfs install 14810 --dir ./my-server --accept-eula          # by pack ID
cfs install https://www.curseforge.com/minecraft/modpacks/all-the-mods-10 --dir ./my-server

# 3. Update later
cfs update --dir ./my-server

# 4. See what an update would change first
cfs update --dir ./my-server --dry-run
```

## Commands

### Core

| Command | Description |
|---------|-------------|
| `cfs install <PACK>` | Install a modpack server (pack ID, CurseForge URL, or saved state) |
| `cfs update [PACK]` | Update an existing server to the latest server pack |
| `cfs apply <URL>` | Apply a server pack ZIP from a direct download URL (no API key needed) |
| `cfs status [--dir DIR]` | Show installed pack and version info for a server directory |
| `cfs info <PACK>` | Show modpack name, summary, download count, and game versions |
| `cfs self-update` | Update `cfs` itself to the latest GitHub release |

### Config

| Command | Description |
|---------|-------------|
| `cfs config set-api-key` | Save your CurseForge API key |
| `cfs config show` | Show current configuration |
| `cfs config unset-api-key` | Remove saved API key |
| `cfs config path` | Print config file path |

### CurseForge helpers (debugging/scripting)

| Command | Description |
|---------|-------------|
| `cfs cf search <QUERY>` | Search CurseForge modpacks |
| `cfs cf resolve <URL>` | Resolve a CurseForge URL to a numeric pack ID |
| `cfs cf files <PACK_ID>` | List available files for a modpack |
| `cfs cf download-url <PACK_ID>` | Get the direct server pack download URL |

### Shell completion

```bash
# Add once to your shell profile, or let install-cfs.sh do it automatically:
echo 'eval "$(cfs completion bash)"' >> ~/.bashrc   # bash
echo 'eval "$(cfs completion zsh)"'  >> ~/.zshrc    # zsh
cfs completion fish > ~/.config/fish/completions/cfs.fish  # fish
```

## Flags reference

### `install` / `update`

| Flag | Description |
|------|-------------|
| `--dir DIR` | Target server directory (default: `.`) |
| `--file-id N` | Pin to a specific CurseForge file ID |
| `--accept-eula` | Write `eula.txt` with `eula=true` |
| `--dry-run` | Show what would be done without making changes |
| `--check-only` | *(update only)* Report if an update is available, don't apply |
| `--use-saved` | On pack ID mismatch, prefer the saved ID |
| `--use-arg` | On pack ID mismatch, prefer the argument ID |
| `--no-prompt` | Fail on any ambiguity instead of prompting |

### `apply`

| Flag | Description |
|------|-------------|
| `--dir DIR` | Target server directory (default: `.`) |
| `--accept-eula` | Write `eula.txt` with `eula=true` |
| `--dry-run` | Show what would be done without making changes |

## How it works

1. Resolves the modpack: accepts a numeric pack ID, a CurseForge URL, or reads the saved state.
2. Downloads the latest server pack ZIP from CurseForge.
3. Extracts and detects the pack root.
4. **Install mode**: copies the pack into the target directory.
5. **Update mode**: backs up `mods/`, `config/`, `defaultconfigs/`, and `kubejs/` to `.mcserver/backups/`, then applies the new pack. Automatically restores the backup on failure.
6. Saves state to `<server-dir>/.mcserver/state.json`.

## Configuration & state files

| File | Description |
|------|-------------|
| `~/.config/mcserver/config.json` | API key (respects `$XDG_CONFIG_HOME`) |
| `<server-dir>/.mcserver/state.json` | Installed pack ID, file ID, display name, last-updated timestamp |
| `<server-dir>/.mcserver/backups/` | Timestamped backups created before each update |

## Development

```bash
# One-time setup
./setup-dev.sh

# Everyday
go test ./...                         # run tests
go build -o cfs ./cmd/cfs            # build
./release.sh [patch|minor|major]     # bump version, tag, and push
```

CI runs on Go 1.22 and 1.23.  
Releases are built for Linux, macOS, and Windows (amd64 + arm64) via [GoReleaser](https://goreleaser.com/).

