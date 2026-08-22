# The GitHub Action

This repo ships a composite action at its root, so a workflow installs TinyGo
in one line:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: 'go.mod'
- uses: tinywasm/tinygo@v0
```

`tinygo` is on `PATH` for every later step.

## What it costs: nothing

Publishing a GitHub Action is free. Two things get confused here:

| | Cost | Needed for `uses:`? |
|---|---|---|
| `uses: owner/repo@ref` from a **public** repo | free | this is all there is |
| Listing on GitHub **Marketplace** | free | no — only discoverability |
| Actions minutes on public repos | free | — |

An Action is **an `action.yml` in a public repo plus a git tag**. No registry to
push to, no account tier, no build artifact to upload.

It is a *composite* action — a YAML file running shell steps — because the
installer is already a Go program. The alternatives would each drag a toolchain
into a Go repo for nothing: a JavaScript action needs Node and a committed
bundled `dist/`, a Docker action needs an image pull on every run.

## Inputs and outputs

| Input | Default | Notes |
|---|---|---|
| `version` | `''` | Empty uses the pinned `DefaultVersion`. See the warning below. |
| `cache` | `'true'` | Caches the install tree between runs. |

| Output | |
|---|---|
| `bindir` | directory holding the binary, already appended to `PATH` |
| `version` | what `tinygo version` reports |

### Do not pass `version` inside this ecosystem

A project that also runs `sitec` or `goflare` must leave `version` empty.
`sitec` calls `EnsureInstalled` again during its build, and if the version it
wants differs from the one the workflow installed, it takes the mismatch branch:
it uninstalls what the workflow put there — on Linux via
`sudo apt-get remove -y tinygo` when the existing install came from a package —
and downloads its own. Slower CI, and a deletion nobody asked for.

Leaving it empty resolves to `DefaultVersion`, which is by construction the
version the rest of the ecosystem agrees on.

That leaves two legitimate sources for the version, one per kind of consumer:

| Consumer | Source of truth | How |
|---|---|---|
| Uses `sitec`/`goflare` | its own `go.mod` | `go run github.com/tinywasm/tinygo/cmd/tinygoinstall -print-bindir` |
| Any other TinyGo project | the action ref | `uses: tinywasm/tinygo@v0` |

Both are a single source. What must never happen is a version written in the
workflow *and* another resolved from the module — the two drift, and the
mismatch branch above is the result.

## Two implementation details worth knowing

**The action runs its own copy of the installer.** The step invokes
`go run ./cmd/tinygoinstall` from `${{ github.action_path }}` — the checkout of
this repo at the ref the consumer pinned — not
`go run github.com/tinywasm/tinygo/...`, which would resolve through the
*consumer's* `go.mod`. An ordinary TinyGo project has no reason to require this
module, and a project that does could pin a different version than the `uses:`
ref. Running from `action_path` makes the ref the only thing that decides.

**The bin directory is discovered, not assumed.** `defaultInstallDir` uses
`/usr/local` when writable and falls back to `~/.local` otherwise, and Scoop
picks its own location on Windows — so no fixed path can be appended to
`$GITHUB_PATH`. `cmd/tinygoinstall -print-bindir` prints the resolved directory
and nothing else to stdout, with every log line on stderr, so a shell can
capture it directly:

```bash
echo "$(tinygoinstall -print-bindir)" >> "$GITHUB_PATH"
```

It derives from the path `EnsureInstalled` returns, which covers all three
branches `getPath` can take: a hit on `PATH`, a Scoop shim, and the local
tarball install.

## The caller supplies Go

The action does **not** call `actions/setup-go` itself: that would silently
override the Go version the caller chose. Every consumer of this package already
sets Go up to build with it, so the prerequisite costs them nothing, and
guessing a version on their behalf would cost them a debugging session.

## Releasing

```bash
git tag v0.0.12 && git push origin main --tags
```

`uses: tinywasm/tinygo@v0.0.12` works the moment the tag lands. Then move the
floating major tag, which is what consumers actually pin:

```bash
git tag -f v0 v0.0.12 && git push -f origin v0
```

Repoint `v0` on every release, or consumers pinning it stay frozen.

### Optional: Marketplace

Repo page → *Releases* → *Draft a new release* → check **Publish this Action to
the GitHub Marketplace**. It needs `action.yml` at the root with `name` and
`description`, a name unique across the Marketplace, and a README — all of which
exist. It changes nothing about how `uses:` resolves; it only makes the action
searchable.

## How it is tested

[`.github/workflows/action.yml`](../.github/workflows/action.yml) consumes the
action with `uses: ./`, so what runs is the working tree of the commit under
review — a broken `action.yml` fails its own PR instead of the first user who
pins the tag. It runs on Linux, macOS and Windows, and asserts:

- a bare `tinygo version` resolves — the only real proof `$GITHUB_PATH` was
  written, which checking the step output alone would not give;
- `bindir` and `version` outputs are non-empty;
- invoking the action a second time in the same job still leaves `tinygo` on
  `PATH`, exercising the idempotent branch;
- `version: '0.40.0'` actually installs 0.40.0, so an explicit pin is proven to
  beat `DefaultVersion`.
