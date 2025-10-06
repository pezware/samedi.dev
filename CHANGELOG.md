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
- TBD

### Deprecated
- TBD

### Removed
- TBD

### Fixed
- golangci-lint configuration updated to fix deprecated options and linters
  - Replaced deprecated `run.skip-dirs` and `run.skip-files` with `issues.exclude-dirs` and `issues.exclude-files`
  - Changed deprecated `output.format` to `output.formats` (array syntax)
  - Replaced deprecated `exportloopref` linter with `copyloopvar` (Go 1.22+ compatible)
  - Replaced deprecated `gomnd` linter with `mnd` (renamed)
  - Resolves CI/CD exit code 7 error from golangci-lint

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
