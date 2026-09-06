# tinygo
<img src="docs/img/badges.svg">

Automated TinyGo installer for the `webtyp` ecosystem. This package provides a standalone, cross-platform (Linux, macOS, Windows) solution to manage TinyGo installations without requiring administrator privileges or user interaction.

## Features

- **Zero-Admin**: Installs into `~/.webtyp/tinygo/`, avoiding the need for `sudo`.
- **Cross-Platform**: Full support for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64).
- **Automated**: Handles downloading, extracting (tar.gz/zip), and binary verification.
- **Unattended**: Designed for CI/CD and developer environment bootstrap.
- **Idempotent**: Won't re-download if the specified version is already present.

## Installation

```go
import "webtyp.com/tinygo"
```

## Usage

### Ensure TinyGo is Installed

The most common use case is ensuring TinyGo is available before running a build.

```go
binPath, err := tinygo.EnsureInstalled(
    tinygo.WithLogger(func(msg string) {
        fmt.Println(msg)
    }),
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("TinyGo is ready at: %s\n", binPath)
```

### Detection

Check if TinyGo is already in the system PATH or locally installed.

```go
if tinygo.IsInstalled() {
    path, _ := tinygo.GetPath()
    version, _ := tinygo.GetVersion()
    fmt.Printf("Found TinyGo %s at %s\n", version, path)
}
```

### Manual Installation

Explicitly trigger an installation of a specific version.

```go
err := tinygo.Install(
    tinygo.WithVersion("0.41.1"),
    tinygo.WithInstallDir("/custom/path"),
)
```

### Environment for Subprocesses

When TinyGo is installed locally, consumers must inject `TINYGOROOT` into any subprocess that invokes `tinygo build`:

```go
binPath, _ := tinygo.EnsureInstalled()
cmd := exec.Command(binPath, "build", "-o", "out.wasm", ".")
cmd.Env = tinygo.GetEnv()  // injects TINYGOROOT + PATH
out, err := cmd.CombinedOutput()
```

## How it Works

See the [Install Flow diagram](docs/diagrams/install_flow.md) for the full decision tree.

1. **Detection**: Checks if `tinygo` is available in the system `PATH`.
2. **Fallback**: If not found, checks for a local install at `~/.webtyp/tinygo/`.
3. **Download**: If missing, downloads the official release from GitHub.
4. **Extraction**: `.tar.gz` for Linux/macOS, `.zip` for Windows.
5. **Verification**: Runs `tinygo version` to confirm the binary is functional.

## Testing

```bash
go test ./...
```

See [docs/architecture.md](docs/architecture.md) for the full testing strategy and design decisions.

## GitHub Actions

This repo is also a composite action, so a workflow installs TinyGo in one line
and gets it on `PATH`:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: 'go.mod'
- uses: webtyp/tinygo@v0
```

| Input | Default | |
|---|---|---|
| `version` | `''` | empty uses `DefaultVersion` — leave it empty if the project also runs `sitec`/`goflare` |
| `cache` | `'true'` | caches the install tree between runs |

Outputs `bindir` and `version`. See
[docs/github_action.md](docs/github_action.md) for why the version input is a
trap inside the ecosystem, and how releasing works.

## License

MIT
