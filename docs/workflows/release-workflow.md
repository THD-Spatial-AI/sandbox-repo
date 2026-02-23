# Release workflow: automatic binaries and build-time versioning

This repository uses **Git tags** and **GitHub Actions** to build release binaries and attach them to a GitHub Release.

The Go application exposes a `--version` / `-v` flag and prints build metadata from an `internal/version` package (or equivalent).

On every tagged release, the CI workflow at [`.github/workflows/release.yml`](.github/workflows/release.yml):

- builds release binaries
- injects the release tag, commit SHA, and build timestamp into the binary (via `-ldflags -X`)
- uploads the binaries as GitHub Release assets

---

## What users get

Users can check which build they are running using `--version` or `-v`.

### Example

**Command**

```bash
./demo_linux_amd64 -v
```

**Output**

```text
v0.1.1-alpha (commit a0abd66c5eb4d501c93211bafa89ade1d07a2c99, built 2026-02-06T17:11:32Z)
```

This helps with bug reports and support because the binary identifies itself without needing any runtime network calls.

---

## What this workflow provides

- Release binaries automatically uploaded to GitHub Releases
- `--version` / `-v` prints the **release tag**, **commit hash**, and **UTC build time**
- Clean binaries built with `-trimpath` (no local filesystem paths embedded)
- No version-bump commits required

### Example commands

```bash
./demo_linux_amd64 --version
./demo_linux_amd64 -v
```

---

## Reusing this pattern in other Go projects

You can reuse this workflow pattern in your own Go projects to get CI-driven releases and reliable version output.

The version shown by `--version` / `-v` is **injected at build time**. This means:

- no manual version edits in source files for releases
- no “bump version” commit just to change a string
- version metadata always matches the tag that triggered the release build

---

## How it works (high level)

1. You create and push a Git tag (for example `v0.1.1-alpha`).
2. GitHub Actions runs the release workflow on tag push (`on.push.tags: "v*"`).
3. The workflow:
   - checks out the code
   - sets up Go
   - reads the Go module path from `go.mod`
   - builds binaries
   - injects `Version`, `Commit`, and `Date` into the binary via `-ldflags -X`
   - uploads files from `dist/` as GitHub Release assets

---

## Why build-time injection (instead of editing a version file)

- No extra commits just to change a version string
- The binary is self-identifying (ideal for issue reports)
- Works offline (no runtime calls to GitHub)
- Deterministic: version comes from the tag that triggered the build

---

## Files involved

Adjust these paths to match your project structure:

- Workflow: `.github/workflows/release.yml`
- Version variables: `internal/version/version.go` (or equivalent)
- CLI entrypoint: e.g. `./app` or `./cmd/<tool>`

---

## Version variables in code (required)

`-ldflags -X` can only set **package-level string variables** (`var`), not constants (`const`).

Example `internal/version/version.go`:

```go
package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
```

Your CLI should print these values for `--version` / `-v`.

---

## GitHub Actions workflow behaviour (detailed)

### Trigger

The workflow runs only when a tag starting with `v` is pushed:

```yaml
on:
  push:
    tags:
      - "v*"
```

Example:
- `v0.1.1-alpha` triggers a release build
- normal branch pushes do not trigger this workflow

### Permissions

To upload binaries to GitHub Releases, the workflow needs:

```yaml
permissions:
  contents: write
```

### Build metadata from GitHub Actions

GitHub provides metadata you can use during the build:

- `${{ github.ref_name }}` → tag name (e.g. `v0.1.1-alpha`)
- `${{ github.sha }}` → commit SHA for that tag

In the workflow, these are typically mapped to:

- `VERSION` → `${{ github.ref_name }}`
- `COMMIT` → `${{ github.sha }}`

The build step also generates:

- `DATE` → UTC build timestamp (e.g. `2026-02-06T17:11:32Z`)

### Build step: module path and package to build

The workflow reads the module path from `go.mod`:

```bash
MOD_PATH="$(awk '/^module /{print $2}' go.mod)"
```

Example result:

```text
github.com/THD-Spatial/demo-repository
```

Then define the package to build (adjust to your project):

```bash
MAIN_PKG="./app"
```

> If your Go example is stored in a subdirectory (for example `examples/go/`), run the build step in that directory or read the correct `go.mod` path.

### Build step: injecting version metadata into the binary

Go linker flags are used to set variables at build time:

- `-ldflags "-X '<package>.<var>=<value>'"`

Example:

```bash
LDFLAGS="-s -w \
  -X '${MOD_PATH}/internal/version.Version=${VERSION}' \
  -X '${MOD_PATH}/internal/version.Commit=${COMMIT}' \
  -X '${MOD_PATH}/internal/version.Date=${DATE}'"
```

Notes:
- `-X` requires the fully qualified import path of the package containing the variables
- variables must be `var`, not `const`
- `-s -w` strips debug symbols to reduce binary size (optional)
- `-trimpath` removes local filesystem paths from the binary

### Uploading release assets

Any file matching `dist/*` can be uploaded to the GitHub Release:

```yaml
- name: Upload assets to GitHub Release
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

3. In GitHub:
   - check **Actions** to confirm the workflow completed successfully
   - check **Releases** to confirm binaries were attached to the tag release

---

## Common pitfalls

- **Wrong module path**: `go mod init` should use `github.com/<org>/<repo>` (not a URL with `https://`)
- **Wrong `-X` import path**: must match the actual package path containing the variables
- **Using `const` instead of `var`**: `-ldflags -X` cannot overwrite constants
- **Local builds show `dev`**: expected when building without CI/tag injection
- **Wrong shell variable syntax**: in bash use `$MAIN_PKG`, not `"MAIN_PKG"` or `%MAIN_PKG%`
- **Nested Go example directories**: ensure the workflow runs in the directory containing the relevant `go.mod`

---

## Minimal example workflow (single target)

This is a minimal working example for **Linux amd64** only. Extend it with more `GOOS` / `GOARCH` targets as needed.

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

      - name: Upload assets to GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
```

---

## Template placeholders to customise

If you reuse this document in another repository, update:

- binary name in examples (e.g. `demo_linux_amd64`)
- `MAIN_PKG` path (e.g. `./cmd/<tool>`)
- version package path (if not `internal/version`)
- Go version in the workflow
- target platforms (`GOOS` / `GOARCH`)
