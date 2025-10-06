# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure
- Comprehensive specification documents
- Development tooling (Makefile, pre-commit hooks, CI/CD)
- Testing framework setup
- Go ecosystem guide (docs/GO_ECOSYSTEM.md) with complete tooling reference
- `.tool-versions` file for asdf version management (golang 1.23.5)
- `tools/tools.go` for pinning development tool versions
- `make vuln` target for vulnerability scanning with govulncheck
- Pinned versions for all development tools in Makefile

### Changed
- Updated to Go 1.25.1 (from 1.21)
- Updated CI/CD to read Go version from `.go-version` (GitHub Actions format)
- Changed tool installation to use `@latest` versions for Go 1.25+ compatibility
- Removed Go version matrix from CI (now uses single version from `.go-version`)
- Dependency review now only fails on HIGH severity (accepts MODERATE in dev tools)

### Deprecated
- TBD

### Removed
- TBD

### Fixed
- golangci-lint configuration updated to fix deprecated options and linters
  - Replaced deprecated `run.skip-dirs` and `run.skip-files` with `issues.exclude-dirs` and `issues.exclude-files`
  - Replaced deprecated `exportloopref` linter with `copyloopvar` (Go 1.22+ compatible)
  - Replaced deprecated `gomnd` linter with `mnd` (renamed)
  - Removed `output` section incompatible with v1.64.8
  - Pinned golangci-lint to v1.64.8 to avoid v2.x breaking changes
  - Resolves CI/CD exit code 7 error from golangci-lint
  - See docs/GOLANGCI_LINT_V2_ISSUE.md for details
- Tool installation errors with Go 1.25.1
  - Fixed incompatible tool versions by using `@latest` for most tools
  - Pinned golangci-lint to v1.64.8 (v2.x has breaking config changes)
  - Fixed import ordering in `tools/tools.go` to satisfy gofmt
  - Updated Makefile to use GOPATH binaries explicitly
  - See docs/GO_1_25_UPGRADE.md for details

### Security
- TBD

## [0.1.0] - TBD

### Added
- Initial release
- Core CLI commands (init, start, stop, status)
- Plan generation with LLM integration
- Session tracking
- Basic TUI dashboard
- SQLite storage
- Markdown plan format

[Unreleased]: https://github.com/pezware/samedi.dev/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/pezware/samedi.dev/releases/tag/v0.1.0
