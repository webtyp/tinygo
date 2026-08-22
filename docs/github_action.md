# Shipping this as a GitHub Action

How to turn `tinygoinstall` into a one-line CI step:

```yaml
- uses: tinywasm/tinygo@v0.0.12
```

instead of the multi-line `go run` + `$GITHUB_PATH` block consumers write today.

## The short answers

**No, it does not require a paid account.** Publishing a GitHub Action is free.
Two separate things get confused here:

| | Cost | Needed for the one-liner? |
|---|---|---|
| `uses: owner/repo@ref` from a **public** repo | free | this is all you need |
| Listing on GitHub **Marketplace** | free | no — only for discoverability |
| Actions **minutes** on public repos | free | — |

A published Action is just **a file named `action.yml` in a public repo, plus a
git tag**. There is no registry to push to, no build to publish, no account
tier. `uses:` resolves a repo and a ref directly.

**No, it is not complex.** For this package it is a *composite* action — a YAML
file that runs shell steps. No JavaScript, no bundler, no Docker image, no
`node_modules`. The three kinds:

| Kind | What it takes | Fit here |
|---|---|---|
| **Composite** | one `action.yml`, shell steps | ✅ the installer is already a Go binary |
| JavaScript | Node project, bundled `dist/` committed | ❌ adds a JS toolchain to a Go repo |
| Docker | Dockerfile, image pull per run | ❌ slowest, Linux-only |

## Why the repo root

Placing `action.yml` at the **repo root** is what buys `uses: tinywasm/tinygo@v0.0.12`.
In a subdirectory it becomes `uses: tinywasm/tinygo/action@v0.0.12`, and a
separate `tinywasm/setup-tinygo` repo would split the version of the installer
from the version of the thing that installs it — the exact drift this package
exists to avoid.

A root `action.yml` is inert for Go tooling: `go build`, `go test` and the
module graph never look at it.

## One prerequisite: a code change

The action must put TinyGo's `bin` directory on `$GITHUB_PATH`, and that
directory is **not fixed** — `defaultInstallDir` picks `/usr/local` when
writable and falls back to `~/.local` otherwise (and Scoop manages its own path
on Windows). Today the only way to learn it is scraping the human-readable
`  Binary: /path/to/tinygo` line, which is not a contract.

Add a flag to `cmd/tinygoinstall` that prints the bin directory alone, nothing
else, so the action can consume it:

```go
printBinDir = flag.Bool("print-bindir", false, "Print only the bin directory to stdout and exit")
```

```go
binPath, err := tinygo.EnsureInstalled(opts...)
if err != nil {
    log.Fatalf("error: %v\n", err)
}

if *printBinDir {
    fmt.Println(filepath.Dir(binPath))
    return
}
```

Two rules make it usable from a shell: the path goes to **stdout alone** (every
log line already goes through `WithLogger`, so keep those on stderr or off), and
it **exits 0**. That is the same stdout/stderr split the ecosystem's CLI contract
requires.

## `action.yml`

```yaml
name: 'Setup TinyGo'
description: 'Installs TinyGo without admin privileges and puts it on PATH'
branding:
  icon: 'package'
  color: 'blue'

inputs:
  version:
    description: 'TinyGo version. Defaults to this package''s DefaultVersion.'
    required: false
    default: ''

outputs:
  bindir:
    description: 'Directory containing the tinygo binary'
    value: ${{ steps.install.outputs.bindir }}

runs:
  using: 'composite'
  steps:
    # Cache the install directory, not the tarball: a cache hit then costs no
    # download AND no extraction. Keyed by version and runner so a matrix build
    # never restores a macOS tree onto Linux.
    - name: Restore TinyGo cache
      uses: actions/cache@v4
      with:
        path: |
          /usr/local/tinygo
          ~/.local/tinygo
        key: tinygo-${{ runner.os }}-${{ runner.arch }}-${{ inputs.version || 'default' }}

    - name: Install TinyGo
      id: install
      shell: bash
      run: |
        set -euo pipefail
        args=()
        if [ -n "${{ inputs.version }}" ]; then
          args+=(-version "${{ inputs.version }}")
        fi
        bindir="$(go run github.com/tinywasm/tinygo/cmd/tinygoinstall "${args[@]}" -print-bindir)"
        echo "$bindir" >> "$GITHUB_PATH"
        echo "bindir=$bindir" >> "$GITHUB_OUTPUT"
```

`shell: bash` is required — composite steps have no default shell, and bash is
present on all three hosted runners, Windows included.

### The caller still needs Go

The action runs `go run`, so a consumer must set Go up first:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: 'go.mod'
- uses: tinywasm/tinygo@v0.0.12
```

The action deliberately does **not** call `actions/setup-go` itself: that would
silently override the Go version the caller chose, and every consumer of this
package already sets Go up to build with it. Document the prerequisite instead
of guessing a version.

## Which version of TinyGo it installs

There are two defensible sources, and they suit different consumers. Say which
one applies rather than letting a second number appear:

| Consumer | Source | How |
|---|---|---|
| Inside this ecosystem (uses `sitec`/`goflare`) | **`go.mod`** | keep `go run github.com/tinywasm/tinygo/cmd/tinygoinstall`; the module version pins `DefaultVersion` |
| Outside it (any TinyGo project) | **the action ref** | `uses: tinywasm/tinygo@v0.0.12` |

The distinction matters because `sitec` calls `EnsureInstalled` again during a
build. If the workflow installed a *different* version than the one `sitec`
wants, `EnsureInstalled` takes its version-mismatch branch: it uninstalls what
the workflow put there — on Linux via `sudo apt-get remove -y tinygo` when the
existing install came from a package — and downloads its own. Slower CI and a
deletion nobody asked for.

So an ecosystem consumer must not pass `version:`. Leaving it empty resolves to
`DefaultVersion` from the module already in `go.mod`, which is by construction
the version `sitec` agrees with.

## Publishing

```bash
git add action.yml
git commit -m "feat: composite action for one-line CI setup"
git tag v0.0.12 && git push origin main --tags
```

That is the whole publish. `uses: tinywasm/tinygo@v0.0.12` works the moment the
tag is pushed.

**Also move a floating major tag**, because that is what consumers pin to:

```bash
git tag -f v0 v0.0.12 && git push -f origin v0
```

`uses: tinywasm/tinygo@v0` then picks up patches without every consumer editing
their workflow. Repoint it on each release.

### Optional: Marketplace

Repo page → *Releases* → *Draft a new release* → check **Publish this Action to
the GitHub Marketplace**. It requires `action.yml` at the root with `name` and
`description`, a unique name across the Marketplace, and a README. It is free,
and it changes nothing about how `uses:` resolves — it only makes the action
searchable.

## Verifying it

A tag is not proof. Add a workflow that consumes the action the way a stranger
would, on all three runners:

```yaml
name: Action
on: [push, pull_request]
jobs:
  consume:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - uses: ./                      # the action in this very commit
      - run: tinygo version           # fails unless $GITHUB_PATH worked
        shell: bash
```

`uses: ./` tests the working tree, so a broken `action.yml` fails the PR that
introduced it instead of the first consumer who tries the tag. The bare `tinygo
version` is the assertion that matters: it only resolves if the step actually
exported the bin directory.
