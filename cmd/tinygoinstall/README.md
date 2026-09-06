# tinygoinstall

CLI command for automated TinyGo installation following the official instructions at
https://tinygo.org/getting-started/install/

## Installation

```bash
go install webtyp.com/tinygo/cmd/tinygoinstall@latest
```

## Permissions

| Platform | Default dir | Requires |
|----------|-------------|----------|
| Linux / macOS | `/usr/local` | `sudo` |
| Windows | Scoop-managed | none (user install) |

On Linux/macOS, run with `sudo` or use `-dir` to point to a user-writable directory:

```bash
# Option A — sudo (installs to /usr/local, matches official docs)
sudo tinygoinstall

# Option B — no sudo (installs to a user directory)
tinygoinstall -dir ~/.local
```

## Usage

```bash
# Install default version (sudo required on Linux/macOS)
sudo tinygoinstall

# Install specific version
sudo tinygoinstall -version 0.35.0

# Install to user directory (no sudo)
tinygoinstall -dir ~/.local

# Verbose output
sudo tinygoinstall -v

# Windows (no sudo needed — uses Scoop)
tinygoinstall
```

## Options

- `-version string` — TinyGo version to install, e.g. `0.35.0` (Linux/macOS only, default: `0.40.1`)
- `-dir string` — Installation directory (Linux/macOS only, default: `/usr/local`)
- `-v` — Verbose output
- `-h` — Show help

> `-version` and `-dir` are ignored on Windows; Scoop manages version and directory automatically.

## How it Works

1. **Detection**: Checks if `tinygo` is available in system `PATH`
2. **Version check**: If found, verifies the version matches the required one
3. **Install / update**: Downloads official release from GitHub (Linux/macOS) or uses Scoop (Windows)
4. **Verification**: Confirms binary is functional after install

## Cross-Platform

| Platform | Method | Admin required |
|----------|--------|----------------|
| Linux amd64/arm64 | tarball → `/usr/local` | `sudo` |
| macOS amd64/arm64 | tarball → `/usr/local` | `sudo` |
| Windows amd64 | Scoop | no |

## Exit Codes

- `0` — Success
- `1` — Installation failed
