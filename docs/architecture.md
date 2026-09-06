# Architecture

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| **No interfaces** | Public functions with `config` struct injection. Testable without interfaces, simpler to consume. |
| **No package managers** | `apt`, `brew`, `winget` require admin privileges, are platform-specific, and break unattended installs. Tarball/zip is universal and sudo-free. |
| **No shell modification** | Writing to `~/.bashrc`, `~/.zshrc`, or registry is invasive, requires restart, and breaks CI/Docker. Env is injected at `exec.Command` level via `GetEnv()`. |
| **Hardcoded default version** | Pinning `DefaultVersion` ensures reproducible installs. Consumers can override with `WithVersion()`. |
| **Single install path** | `~/.webtyp/tinygo/` as default avoids conflicts with system installs and is predictable across all tools in the ecosystem. |

## Testing Strategy

Function injection via unexported `config` fields — no interfaces, no real network, no real tinygo binary required.

| Function | Mock strategy |
|----------|---------------|
| `IsInstalled`, `GetPath` | `withLookPath` (internal option) + temp dir with fake binary |
| `GetVersion` | Shell script as fake binary; skip if `/bin/bash` unavailable |
| `Install` | `httptest.Server` serves a programmatically created tar.gz/zip |
| `EnsureInstalled` | Combines `withLookPath` mock + `httptest.Server` |
| `GetRoot`, `GetEnv` | `withLookPath` to control system vs local branch |
| `extractTarGz`, `extractZip` | Programmatically created archives; Zip Slip cases |
| `downloadURL`, `binPath` | Direct unit test on `config` struct |

Run all tests:

```bash
go test ./...
```
