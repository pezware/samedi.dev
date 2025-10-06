# Setup Summary

This document summarizes the development setup for the Samedi project.

## Repository Information

- **Repository**: https://github.com/pezware/samedi.git
- **Code Owner**: @arbeitandy
- **License**: MIT

## Files Created/Updated

### Security Files ⚠️

1. **SECURITY.md** - Security policy and vulnerability reporting
2. **.env.example** - Template for environment variables (never commit real .env!)
3. **.gitignore** - Updated with comprehensive secret patterns
4. **.secrets.baseline** - Baseline for detect-secrets scanner
5. **.github/CODEOWNERS** - Sets @arbeitandy as code owner

### Development Guidelines

1. **CLAUDE.md** - **Updated with critical security reminders**
   - ⚠️ Added prominent warning about never committing secrets
   - Security checklist before every commit
   - Incident response procedures

2. **CONTRIBUTING.md** - Contribution guidelines
3. **README.md** - Project overview and quick start
4. **CHANGELOG.md** - Version history template

### License Files

1. **LICENSE** - MIT License (2025)
2. **.license-header.txt** - Template for Go file headers
3. All Go files should include:
   ```go
   // Copyright (c) 2025 Samedi Contributors
   // SPDX-License-Identifier: MIT
   ```

### Development Tools

1. **Makefile** - Common development tasks
   ```bash
   make help          # See all commands
   make install-tools # Install dev tools
   make check         # Run all quality checks
   make test          # Run tests
   ```

2. **.pre-commit-config.yaml** - Pre-commit hooks including:
   - Go formatting (gofmt, goimports)
   - Linting (golangci-lint)
   - Security scanning (detect-secrets)
   - Commit message validation

3. **.golangci.yml** - Linter configuration (20+ linters enabled)
4. **.editorconfig** - Editor consistency
5. **go.mod** - Go module initialization

### CI/CD

1. **.github/workflows/ci.yml** - GitHub Actions pipeline
   - Multi-platform builds (Linux, macOS, Windows)
   - Unit and integration tests
   - Security scanning (gosec)
   - Code coverage (Codecov)

### IDE Configuration

1. **.vscode/settings.json** - VS Code Go development settings

## Security Highlights 🔒

### Critical Rules (see CLAUDE.md for full details)

**NEVER commit:**
- Passwords or API keys
- Private keys or certificates
- .env files with real credentials
- Any secrets in code or config

**ALWAYS:**
- Use environment variables for secrets
- Check `git diff` before committing
- Use .env.example as template
- Let pre-commit hooks scan for secrets

### Pre-commit Hooks Will Catch:

✅ Common secret patterns (API keys, passwords)
✅ AWS keys, private keys, JWT tokens
✅ Formatting issues
✅ Linting errors
✅ Failing tests

**But you are the final line of defense!**

### If You Accidentally Commit a Secret:

1. **Immediately** rotate/revoke the credential
2. **Do NOT** just delete it (it's in git history!)
3. Use BFG Repo-Cleaner to remove from history
4. Contact @arbeitandy

## Getting Started

### Initial Setup

```bash
# Clone repository
git clone https://github.com/pezware/samedi.git
cd samedi

# Install development tools
make install-tools

# Create .env file from template
cp .env.example .env
# Edit .env with your API keys (NEVER commit this file!)

# Verify setup
make check
```

### Environment Variables

Create `.env` file (already in .gitignore):

```bash
# Required for LLM integration
ANTHROPIC_API_KEY=sk-ant-...your-key-here
OPENAI_API_KEY=sk-...your-key-here

# Optional development settings
SAMEDI_DATA_DIR=~/.samedi
SAMEDI_LOG_LEVEL=info
```

### Daily Development

```bash
# Create feature branch
git checkout -b feat/your-feature

# Write tests (TDD)
make test

# Implement feature
# ... code ...

# Run all checks (formatting, linting, tests)
make check

# Commit (pre-commit hooks will run automatically)
git add .
git commit -m "feat(scope): add awesome feature"

# Push
git push origin feat/your-feature
```

### Pre-commit Hooks

Installed automatically with `make install-tools`.

To run manually:
```bash
pre-commit run --all-files
```

To skip hooks (NOT RECOMMENDED):
```bash
git commit --no-verify  # Don't do this!
```

## Testing Framework

**Unit Tests:**
- `testing` (stdlib)
- `testify` for assertions and mocking

**Integration Tests:**
- `dockertest` for real dependencies
- Run with: `make test-integration`

**E2E Tests:**
- Standard `testing` with real CLI execution
- Run with: `make test-e2e`

**All Tests:**
```bash
make test-all
```

## Quality Standards

Every commit must:
- ✅ Pass all tests
- ✅ Pass golangci-lint
- ✅ Be formatted with gofmt
- ✅ Have no security issues
- ✅ Follow conventional commit format

## Documentation

- **Development Guide**: [CLAUDE.md](./CLAUDE.md) - **READ THIS FIRST**
- **Contributing**: [CONTRIBUTING.md](./CONTRIBUTING.md)
- **Security Policy**: [SECURITY.md](./SECURITY.md)
- **Architecture**: [docs/04-architecture.md](./docs/04-architecture.md)
- **Full Specs**: [docs/](./docs/)

## Key Makefile Commands

```bash
make help          # Show all available commands
make install-tools # Install development tools
make deps          # Download and tidy dependencies
make build         # Build binary
make test          # Run unit tests
make test-all      # Run all tests (unit + integration + e2e)
make coverage      # Generate coverage report
make lint          # Run linters
make fmt           # Format code
make check         # Run all checks (fmt + lint + test)
make clean         # Remove build artifacts
make run           # Build and run
make security      # Run security checks
make ci            # Run full CI pipeline locally
```

## VS Code Integration

Recommended extensions (install from Extensions marketplace):
- **Go** (golang.go) - Official Go extension
- **GitLens** - Git supercharged
- **EditorConfig** - EditorConfig support

Settings are pre-configured in `.vscode/settings.json`.

## Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:** feat, fix, docs, refactor, test, chore

**Examples:**
```
feat(flashcard): implement SM-2 spaced repetition
fix(session): prevent duplicate active sessions
docs: update installation instructions
test(plan): add integration tests
```

Pre-commit hooks will validate format.

## Next Steps

1. **Read CLAUDE.md** - Development guidelines and security rules
2. **Read SECURITY.md** - Security policy
3. **Set up environment** - Create .env file with your API keys
4. **Install tools** - Run `make install-tools`
5. **Verify setup** - Run `make check`
6. **Start coding** - Follow TDD: test → implement → refactor

## Questions?

- **Security Issues**: Contact @arbeitandy immediately
- **General Questions**: Open a [Discussion](https://github.com/pezware/samedi/discussions)
- **Bugs**: Open an [Issue](https://github.com/pezware/samedi/issues)

---

**Remember: Security is everyone's responsibility. Never commit secrets!** 🔒
