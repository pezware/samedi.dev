# Contributing to Samedi

Thank you for your interest in contributing to Samedi! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Make
- Pre-commit (optional but recommended)

### Development Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/pezware/samedi.dev.git
   cd samedi
   ```

2. **Install development tools**

   ```bash
   make install-tools
   ```

3. **Install dependencies**

   ```bash
   make deps
   ```

4. **Run tests to verify setup**

   ```bash
   make test
   ```

5. **Build the binary**

   ```bash
   make build
   ```

## Development Workflow

**Read [CLAUDE.md](./CLAUDE.md) for detailed development guidelines.**

### Quick Summary

1. **Create a branch**

   ```bash
   git checkout -b feat/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Write tests first** (TDD approach)

   ```bash
   # Create test file
   vim internal/yourmodule/yourfile_test.go

   # Run tests (should fail)
   go test ./internal/yourmodule/...
   ```

3. **Implement the feature**

   ```bash
   vim internal/yourmodule/yourfile.go

   # Run tests (should pass)
   go test ./internal/yourmodule/...
   ```

4. **Run all checks**

   ```bash
   make check
   ```

5. **Commit with conventional commit message**

   ```bash
   git add .
   git commit -m "feat(module): add awesome feature"
   ```

6. **Push and create PR**

   ```bash
   git push origin feat/your-feature-name
   gh pr create
   ```

## Code Standards

### Go Style

- Follow [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- Use `gofmt` for formatting
- Add godoc comments for all exported functions/types
- See [CLAUDE.md](./CLAUDE.md) for detailed examples

### Testing

- Write tests for all new code
- Aim for 80%+ coverage on core logic
- Use table-driven tests for multiple scenarios
- See [CLAUDE.md](./CLAUDE.md#testing-standards) for testing guidelines

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Build process, dependencies, tooling

**Examples:**

```
feat(flashcard): implement SM-2 spaced repetition algorithm

Add SM-2 algorithm for calculating next review intervals based on
user ratings (1-4). Includes unit tests for all edge cases.

Closes #42
```

```
fix(session): prevent duplicate active sessions

Check for existing active session before creating new one.
Return clear error with suggestion to stop current session.

Fixes #89
```

## Pull Request Process

1. **Update documentation**
   - Update README.md if adding new commands
   - Update relevant docs/ files if changing architecture
   - Update CHANGELOG.md with your changes

2. **Ensure all checks pass**
   - All tests pass (`make test`)
   - Linters pass (`make lint`)
   - Coverage doesn't decrease significantly

3. **PR Description**
   - Use the PR template (auto-generated)
   - Link to related issues
   - Add screenshots/demos for UI changes
   - Describe testing performed

4. **Code Review**
   - Address reviewer comments
   - Keep discussion focused and constructive
   - Update PR based on feedback

5. **Merge**
   - PRs require 1 approval from maintainer
   - Squash commits on merge (unless told otherwise)
   - Delete branch after merge

## Types of Contributions

### Bug Reports

- Use issue template
- Include minimal reproduction steps
- Include environment details (OS, Go version, samedi version)
- Include relevant logs/errors

### Feature Requests

- Use issue template
- Describe the problem you're solving
- Provide use cases and examples
- Consider creating a design doc for large features

### Documentation

- Fix typos, improve clarity
- Add examples and use cases
- Update outdated information
- Translate to other languages (future)

### Code Contributions

- Start with good first issues
- Discuss major changes in issues first
- Follow the development workflow above
- Write tests and documentation

## Project Structure

```
samedi/
├── cmd/samedi/           # Main application entry point
├── internal/             # Private application code
│   ├── cli/             # CLI commands
│   ├── tui/             # TUI components
│   ├── plan/            # Plan management
│   ├── session/         # Session tracking
│   ├── flashcard/       # Flashcard system
│   ├── llm/             # LLM integration
│   ├── storage/         # Data persistence
│   └── ...
├── pkg/                  # Public packages
├── docs/                 # Documentation
├── tests/                # Integration and E2E tests
└── templates/            # LLM prompt templates
```

See [docs/04-architecture.md](./docs/04-architecture.md) for detailed architecture.

## Testing

### Run Tests

```bash
# Unit tests only (fast)
make test

# Integration tests
make test-integration

# E2E tests
make test-e2e

# All tests
make test-all

# With coverage
make coverage
```

### Write Tests

See [CLAUDE.md](./CLAUDE.md#testing-standards) for detailed testing guidelines.

**Quick example:**

```go
func TestCreatePlan_ValidTopic_ReturnsPlan(t *testing.T) {
    // Arrange
    mockLLM := new(MockLLMProvider)
    mockLLM.On("Call", mock.Anything).Return("# Plan...", nil)
    svc := NewPlanService(mockLLM)

    // Act
    plan, err := svc.Create("rust-async", 40)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "rust-async", plan.ID)
    mockLLM.AssertExpectations(t)
}
```

## Documentation

### Code Documentation

- Add godoc comments to all exported functions/types
- Explain *why*, not *what* for complex logic
- Include examples in godoc when helpful

### User Documentation

- Update README.md for user-facing changes
- Update docs/ for architecture/design changes
- Add examples to docs/01-user-journeys.md

## Getting Help

- **Questions?** Open a [Discussion](https://github.com/pezware/samedi.dev/discussions)
- **Bug?** Open an [Issue](https://github.com/pezware/samedi.dev/issues)

## Code of Conduct

Be respectful, inclusive, and constructive. See [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md).

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT).

## Recognition

Contributors will be:

- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Invited to the contributors team (after significant contributions)

Thank you for contributing to Samedi! 🎓
