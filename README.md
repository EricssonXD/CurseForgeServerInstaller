# CurseForge Server Installer (`cfs`)

A CLI tool to install and update CurseForge-based Minecraft modpack servers.
Written in Go with [Cobra](https://github.com/spf13/cobra). Single binary, no dependencies.

## Install

### From source

```bash
go install github.com/ericsson/curseforge-server-installer/cmd/cfs@latest
```

### From release

Download the binary for your platform from [Releases](../../releases), extract, and place `cfs` on your `PATH`.

### Build locally

```bash
go build -o cfs ./cmd/cfs
```

## Quick start

```bash
# Save your CurseForge API key (get one at https://console.curseforge.com)
cfs config set-api-key

# Install a modpack server by pack ID
cfs install 14810 --dir ./my-server --accept-eula

# Install by CurseForge URL
cfs install https://www.curseforge.com/minecraft/modpacks/all-the-mods-10 --dir ./my-server

# Update an existing server
cfs update --dir ./my-server

# Check for updates without installing
cfs update --dir ./my-server --check-only
```

## Commands

| Command | Description |
|---------|-------------|
| `cfs install <PACK>` | Install a modpack server (ID, URL, or saved) |
| `cfs update` | Update an existing server to the latest server pack |
| `cfs status` | Show installation state for a server directory |
| `cfs config set-api-key` | Save CurseForge API key |
| `cfs config show` | Show current configuration |
| `cfs config unset-api-key` | Remove saved API key |
| `cfs config path` | Print config file path |
| `cfs cf search <QUERY>` | Search CurseForge modpacks |
| `cfs cf resolve <URL>` | Resolve a CurseForge URL to a pack ID |
| `cfs cf files <PACK_ID>` | List available files for a modpack |
| `cfs cf download-url <PACK_ID>` | Get direct download URL for a server pack |

Run `cfs --help` or `cfs <command> --help` for full flag details.

## Configuration

Config is stored at `~/.config/mcserver/config.json` (respects `$XDG_CONFIG_HOME`).
Per-server state is stored in `<server-dir>/.mcserver/state.json`.

## Docs

- [CLI design](docs/CLI_DESIGN.md)
- [Command map](docs/COMMAND_MAP.md)
- [Migration plan](docs/COBRA_MIGRATION_PLAN.md)
- [Flowcharts](docs/FLOWCHART.md)

## Cross-platform releases

This project uses [GoReleaser](https://goreleaser.com/) for automated builds.
Tag a release to build binaries for Linux, macOS, and Windows (amd64 + arm64):

```bash
git tag v0.1.0
git push origin v0.1.0
```
