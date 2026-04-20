# Cobra Migration Plan

Goal: Replace the current Python CLI with a Go-based CLI using `spf13/cobra`, producing a single, cross-platform binary and a clear command surface that maps to existing functionality.

**Summary recommendation**
- Recommended approach: Iterative migration (start with a Cobra-based CLI that wraps the existing Python implementation, then incrementally port core modules to Go). This minimizes risk and lets us deliver a Go CLI quickly while enabling a full port over time.

---

## 1. Constraints & assumptions
- The current project is Python-first (see `main.py` and `mcserver/cli.py`).
- We will keep the repository but add Go sources under a `cmd/` and `internal/` layout.
- Two migration options: full-port (rewrite everything in Go) or hybrid (Cobra CLI that initially shells out to Python). Hybrid is recommended.

## 2. High-level strategy options
- Full port to Go
  - Pros: single-language, better performance, single binary distribution.
  - Cons: large upfront effort, higher risk, longer time to parity.
  - Completeness: 9/10 (once done)
- Incremental wrapper -> port (recommended)
  - Start: Cobra CLI subcommands call the existing Python entrypoints via `exec.Command("python3", ...)`.
  - Then: Port modules in prioritized order (http client, downloader, fs ops, state) into Go packages and swap implementations behind the CLI.
  - Pros: fast initial delivery, low risk, easier testing / rollback.
  - Completeness: 8/10 for staged delivery, 10/10 final once ported.

## 3. Phase-by-phase plan

### Phase 0 — Analysis (1-2 days)
- Inventory current CLI commands and flags by reading `main.py`, `mcserver/cli.py`, and related config.
- Produce a Command Map: existing command names, flags, env vars, and config files used.
- Identify high-complexity modules to port (networking, downloads, file system operations, state management).

Deliverable: `COMMAND_MAP.md` with exact flag and behavior mapping.

### Phase 1 — Scaffold Cobra app (half day)
- Create `cmd/cfs/` (or `cmd/curseforge-installer/`) containing `main.go` with a root Cobra command.
- Initialize Go module: `go mod init github.com/<org>/curseforge-server-installer` (replace `<org>`).
- Add dependency: `go get github.com/spf13/cobra`.
- Scaffold root command and top-level subcommands that match the Command Map.

Deliverable: `cmd/cfs/main.go`, `cmd/cfs/cmd/*.go` scaffold.

### Phase 2 — Wrapper implementation (1-2 days)
- Implement subcommands that execute the existing Python entrypoints using `os/exec` (preserve stdout/stderr/exit codes).
- Add flags and env var passthrough so behavior is identical.
- Provide a `--python-binary` option and detect `python3` by default.

Deliverable: working Go binary that is a drop-in CLI for most users.

### Phase 3 — Port core modules (1-3 weeks, incremental)
- Prioritize by complexity and value: `http_client.py` → `internal/http`, `download.py` → `internal/download`, `fs_ops.py` → `internal/fs`, `state.py` → `internal/state`.
- For each module:
  - Port logic to Go package with unit tests.
  - Swap wrapper command to call Go implementation.
  - Iterate until parity validated by integration tests.

Deliverable: native Go implementations with test coverage.

### Phase 4 — Tests, CI, and quality (3-5 days)
- Add unit tests for Go packages and small integration tests for CLI behavior.
- Add GitHub Actions workflows:
  - `go test` + `golangci-lint` + `go vet` on PRs.
  - Build and cross-compile via `goreleaser` or `xgo` for releases.

Deliverable: passing CI, linters configured.

### Phase 5 — Docs, packaging, and release (1-2 days)
- Update README and `docs/CLI_DESIGN.md` to include Go installation and usage.
- Add `goreleaser` config for producing binaries and checksums.
- Decide on versioning and release process.

Deliverable: release artifacts and updated docs.

### Phase 6 — Cutover and cleanup (1-2 days)
- When all functionality ports are complete and verified, either archive the Python implementation or move it to `legacy/` with a clear deprecation note.
- Final clean-up PRs and docs.

Deliverable: single-language repo (optional) and deprecation notes.

## 4. Suggested repository layout (target)
```
cmd/cfs/                # cobra root command + subcommands
internal/http/          # http client implementation
internal/download/      # downloader logic
internal/fs/            # filesystem helpers
internal/curseforge/    # curseforge-specific logic
pkg/shared/             # shared helpers (if needed)
scripts/                # dev scripts (build, lint, goreleaser)
```

## 5. Module mapping (quick)
- `mcserver/http_client.py` -> `internal/http`
- `mcserver/download.py` -> `internal/download`
- `mcserver/fs_ops.py` -> `internal/fs`
- `mcserver/state.py` -> `internal/state`
- `mcserver/curseforge.py` -> `internal/curseforge`

## 6. Risks & mitigations
- Risk: Large port introduces regressions. Mitigation: hybrid wrapper to deliver functionality early and port modules one-by-one.
- Risk: Missing Python-only libraries (complex parsing). Mitigation: keep the Python implementation as fallback until ported.
- Risk: Unexpected OS behaviors. Mitigation: cross-platform tests and CI on linux/macos/windows.

## 7. Timeline (rough)
- Quick prototype (scaffold + wrapper): 1–3 days
- Incremental port of critical modules: 1–3 weeks (depending on scope)
- Full port + CI + releases: 3–6 weeks

## 8. Next immediate steps (what I can do now)
- Produce the `COMMAND_MAP.md` by scanning `main.py` and `mcserver/cli.py` and listing all commands/flags.
- Scaffold the Cobra app under `cmd/` and create the first PR with wrapper commands.

---

If you want, I can start by generating the `COMMAND_MAP.md` and scaffolding the Cobra app (create `cmd/` with `main.go` and root command). Which would you like me to do next?
