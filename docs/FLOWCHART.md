# Flowcharts (Proposed)

These flowcharts describe the intended CLI behavior.

## Install flow

```mermaid
flowchart TD
  A[User runs: mcserver install [SOURCE]] --> B[Set target dir (default: .)]
  B --> C{Is this already a server dir?\n(server.properties exists)}
  C -->|Yes| C1[Mode = UPDATE]
  C -->|No| C2[Mode = INSTALL]

  C1 --> D{SOURCE provided?}
  C2 --> D
  D -->|Yes| E[Parse SOURCE]
  D -->|No| F[Read packId from .mcserver/state.json]

  E --> E1{SOURCE looks like URL?}
  E1 -->|Yes| G[Resolve packId from URL]
  E1 -->|No, digits-only| H[Use SOURCE as packId]

  F --> I{Saved packId found?}
  I -->|No| I1[Fail: need SOURCE or initialized state] --> Z[Exit non-zero]
  I -->|Yes| J[Use saved packId]

  G --> K{Saved packId also exists and differs?}
  H --> K
  J --> L[Fetch modpack files list]

  K -->|Yes| K1[Prompt: use saved vs arg\n(or require --use-saved/--use-arg)]
  K -->|No| L
  K1 --> L

  L --> M{CurseForge API key available?}
  M -->|No| M1[Fail: ask user to save API key via mcserver config] --> Z
  M -->|Yes| N[Choose server pack file]

  N --> O[Resolve server pack ZIP download URL]
  O --> P[Download server pack ZIP to temp]
  P --> Q[Extract ZIP]
  Q --> R[Detect pack root (find mods/)]

  R --> S{Mode?}
  S -->|INSTALL| T[Copy extracted contents into server dir]
  S -->|UPDATE| U[Replace folders + copy executables\n(preserve world/config files)]

  T --> V{--accept-eula?}
  U --> V
  V -->|Yes| W[Write eula.txt with eula=true]
  V -->|No| X[Leave eula.txt unchanged]

  W --> Y[Write/update .mcserver/state.json]
  X --> Y
  Y --> AA[Print completion message]
  AA --> AB[Exit 0]
```

## Update flow

Note: Update is primarily handled by `mcserver install` automatically when run inside an existing server directory.

```mermaid
flowchart TD
  A[User runs: mcserver update [SOURCE]] --> B{Is server.properties present?}
  B -->|No| B1[Fail: not a server directory] --> Z[Exit non-zero]
  B -->|Yes| C{Load .mcserver/state.json exists?}

  C -->|No| C1[Fail or require --pack-id/--pack-url] --> Z
  C -->|Yes| D[Read packId + installed file info]

  D --> E{--check-only?}
  E -->|Yes| F[Fetch latest server pack metadata]
  F --> G[Compare installed vs latest]
  G --> H{Update available?}
  H -->|No| I[Print up-to-date] --> J[Exit 0]
  H -->|Yes| K[Print update available] --> J

  E -->|No| L{--backup?}
  L -->|Yes| M[Backup world/ (+ optional snapshot)]
  L -->|No| N[Skip backup]

  M --> O[Resolve new server pack download URL]
  N --> O

  O --> P[Download ZIP to temp]
  P --> Q[Extract ZIP]
  Q --> R[Detect pack root (find mods/)]

  R --> S[Replace folders: mods/config/scripts/kubejs/libraries/defaultconfigs]
  S --> T[Copy top-level *.jar/*.sh/*.bat except user_jvm_args.txt]
  T --> U[Update .mcserver/state.json timestamps + installed version]
  U --> V[Print completion message]
  V --> W[Exit 0]
```

## Notes on current repo mapping
- Folder replacement + executable copying is modeled directly on the existing `mc-update` script.
- Pack resolution + file discovery matches the logic in `mcserver-installer/get_modpack_id.py` and the notebook.
