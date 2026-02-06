![Auto Assign](https://github.com/THD-Spatial/demo-repository/actions/workflows/auto-assign.yml/badge.svg)

# Demo repository

This repository is a minimal demo for a Go CLI release workflow on GitHub.

## Go release workflow with build-time versioning

The `./app` directory contains a minimal Go program. It prints a version string via the `internal/version` package.

On every tagged release, the CI workflow [.github/workflows/release.yml](.github/workflows/release.yml):

- builds the release binaries
- injects the release tag, commit SHA, and build timestamp into the binary (via `-ldflags -X`)
- uploads the binaries as GitHub Release assets

A full explanation of the workflow is in:

- [docs/release-workflow.md](docs/release-workflow.md)

### Checking the version

Users can check which build they are running using `-v` or `--version`.

Example:

**Command:**

```bash
./demo_linux_amd64 -v
```

**Output:**

```bash
v0.1.1-alpha (commit a0abd66c5eb4d501c93211bafa89ade1d07a2c99, built 2026-02-06T17:11:32Z)
```

### Reusing this in other projects

You can adapt this template in your own Go program to set up CI-driven release builds and reliable `--version` output, without making version-bump commits.
