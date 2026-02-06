# Release workflow: automatic binaries + build-time versioning

This repository uses **Git tags** + **GitHub Actions** to build release binaries and attach them to a GitHub Release.

The CLI version shown by `--version` / `-v` is **injected into the binary at build time** (no “bump version” commits, no file edits).

---

## What you get

- Release assets (binaries) automatically uploaded to GitHub Releases
- `--version` / `-v` prints the **release tag**, **commit hash**, and **UTC build time**
- Clean binaries (no local filesystem paths embedded) via `-trimpath`

Example:

```bash
./demo_linux_amd64 --version
./demo_linux_amd64 -v
```

Output:

```text
v0.1.1-alpha (commit a0abd66c5eb4..., built 2026-02-06T17:11:32Z)
```

---

## How it works (high-level)

1. You create and push a Git tag like `v0.1.1-alpha`.
2. GitHub Actions runs a workflow on that tag push (`on: push: tags: "v*"`).
3. The workflow:
   - checks out the code
   - sets up Go
   - reads your Go module path from `go.mod`
   - builds the binaries
   - injects `Version`, `Commit`, and `Date` into the binary using `-ldflags -X`
   - uploads binaries in `dist/` as GitHub Release assets

---

## Why build-time injection (and not editing a version file)

- No extra commits just to change a version string
- The binary is self-identifying (ideal for issue reports)
- Works offline (no runtime calls to GitHub)
- Deterministic: version comes from the tag that triggered the build

---

## Files involved

- Workflow: `.github/workflows/release.yml`
- Version variables: `internal/version/version.go` (or equivalent)
- CLI entrypoint: e.g. `./app` or `./cmd/<tool>`

---

## Version variables in code (required)

`-ldflags -X` can only set **package-level `var` strings** (not `const`).

Example `internal/version/version.go`:

```go
package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
```

Your CLI should print these for `--version` / `-v`.

---

## GitHub Actions workflow behaviour (detailed)

### Trigger

Runs only when a tag starting with `v` is pushed:

```yaml
on:
  push:
    tags:
      - "v*"
```

So pushing `v0.1.1-alpha` triggers a release build, while normal branch pushes do not.

### Permissions

Required to upload binaries to GitHub Releases:

```yaml
permissions:
  contents: write
```

### Build step: environment variables and metadata

GitHub provides:

- `${{ github.ref_name }}` → the tag name, e.g. `v0.1.1-alpha`
- `${{ github.sha }}` → the commit SHA for that tag

In the workflow we map these to:

- `VERSION` → `${{ github.ref_name }}`
- `COMMIT` → `${{ github.sha }}`

The build script also creates:

- `DATE` → UTC timestamp of the build (e.g. `2026-02-06T17:11:32Z`)

### Build step: module path + package to build

The workflow reads the module path from `go.mod`:

```bash
MOD_PATH="$(awk '/^module /{print $2}' go.mod)"
```

Example result:

```text
github.com/THD-Spatial/demo-repository
```

Then it defines what to build, e.g.:

```bash
MAIN_PKG="./app"
```

### Build step: injecting the version into the binary

We use Go linker flags:

- `-ldflags "-X '<package>.<var>=<value>'"`

Example:

```bash
LDFLAGS="-s -w \
  -X '${MOD_PATH}/internal/version.Version=${VERSION}' \
  -X '${MOD_PATH}/internal/version.Commit=${COMMIT}' \
  -X '${MOD_PATH}/internal/version.Date=${DATE}'"
```

Notes:
- `-X` needs the fully qualified import path of the package containing the variables
- variables must be `var`, not `const`
- `-s -w` strips debug symbols to reduce binary size (optional)

We also use:

- `-trimpath` to remove local filesystem paths from the binary

### Uploading release assets

Any file in `dist/*` is uploaded to the GitHub Release for that tag:

```yaml
- name: Upload assets to Github release
  uses: softprops/action-gh-release@v2
  with:
    files: dist/*
```

---

## How to create a release

1. Commit and push your changes to the default branch.
2. Create a tag and push it:

```bash
git tag v0.1.1-alpha
git push origin v0.1.1-alpha
```

3. Go to GitHub:
   - **Actions**: confirm the workflow ran successfully
   - **Releases**: you should see the tag release with attached binaries

---

## Common pitfalls

- **Wrong module path**: `go mod init` must be `github.com/<org>/<repo>` (no `https://`).
- **Wrong `-X` path**: must match the actual package path where the vars live.
- **Using `const`**: `-ldflags -X` cannot overwrite constants.
- **Local builds show `dev`**: expected when you build without tag/CI injection.
- **Shell variables**: use `$MAIN_PKG` in bash, not `"MAIN_PKG"` or `%MAIN_PKG%`.

---

## Minimal example workflow (single target)

This is a minimal working workflow (Linux amd64 only). Expand it with more `GOOS/GOARCH` lines if needed.

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  build-and-release:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22.x"

      - name: Build binaries
        env:
          VERSION: ${{ github.ref_name }}
          COMMIT: ${{ github.sha }}
        run: |
          set -euo pipefail
          DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
          mkdir -p dist

          MOD_PATH="$(awk '/^module /{print $2}' go.mod)"
          MAIN_PKG="./app"

          LDFLAGS="-s -w \
            -X '${MOD_PATH}/internal/version.Version=${VERSION}' \
            -X '${MOD_PATH}/internal/version.Commit=${COMMIT}' \
            -X '${MOD_PATH}/internal/version.Date=${DATE}'"

          GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$LDFLAGS" -o dist/demo_linux_amd64 "$MAIN_PKG"

      - name: Upload assets to Github release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
```
