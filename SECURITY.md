# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Samedi, please report it by emailing:

**<security@pezware.com>** (or contact @arbeitandy directly)

Please do **NOT** open a public issue for security vulnerabilities.

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **24 hours**: Initial acknowledgment
- **7 days**: Preliminary assessment
- **30 days**: Fix and disclosure (or explanation if not applicable)

## Security Best Practices

### For Contributors

**NEVER commit:**

- ❌ Passwords or passphrases
- ❌ API keys (OpenAI, Anthropic, Cloudflare, etc.)
- ❌ Private keys or certificates
- ❌ Database credentials
- ❌ OAuth secrets
- ❌ Real `.env` files

**ALWAYS:**

- ✅ Use `.env.example` with placeholder values
- ✅ Reference environment variables in code (never hardcode)
- ✅ Add sensitive patterns to `.gitignore`
- ✅ Review `git diff` before committing
- ✅ Let pre-commit hooks scan for secrets

### If You Accidentally Commit a Secret

1. **STOP** - Don't just delete it in a new commit (it's in git history!)
2. **Rotate** - Immediately revoke/rotate the compromised credential
3. **Clean History** - Use BFG Repo-Cleaner or `git filter-branch`

   ```bash
   # Using BFG (recommended)
   brew install bfg
   bfg --replace-text passwords.txt
   git reflog expire --expire=now --all
   git gc --prune=now --aggressive
   ```

4. **Force Push** - Coordinate with team before force pushing
5. **Report** - Contact @arbeitandy immediately

### Pre-commit Security Checks

Automated checks that run before every commit:

- **detect-secrets**: Scans for accidentally committed secrets
- **gosec**: Go security scanner
- **Pattern matching**: Checks for common secret patterns

To run manually:

```bash
pre-commit run detect-secrets --all-files
```

### Environment Variables

Store secrets in environment variables, never in code:

**❌ Bad:**

```go
apiKey := "sk-ant-1234567890abcdef"  // NEVER DO THIS
```

**✅ Good:**

```go
apiKey := os.Getenv("ANTHROPIC_API_KEY")
if apiKey == "" {
    return errors.New("ANTHROPIC_API_KEY not set")
}
```

**✅ Even Better:**

```toml
# config.toml
[llm]
api_key_env = "ANTHROPIC_API_KEY"  # Reference to env var name
```

### Testing with Secrets

Use mock/fake credentials in tests:

```go
func TestLLMProvider(t *testing.T) {
    // Use mock provider, not real API keys
    mockLLM := &MockLLMProvider{
        Response: "test response",
    }
    // ... test logic
}
```

For integration tests that need real credentials:

```go
// +build integration

func TestRealAPI(t *testing.T) {
    apiKey := os.Getenv("TEST_API_KEY")
    if apiKey == "" {
        t.Skip("TEST_API_KEY not set, skipping integration test")
    }
    // ... test with real API
}
```

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Security Features

### Local Data Protection

- File permissions: `chmod 600` for sensitive files
- SQLite encryption: Optional with SQLCipher
- No plaintext secrets in config

### Cloud Security (Phase 2)

- TLS 1.2+ for all API calls
- JWT authentication with 1-hour expiry
- Row-level security in D1 database
- Rate limiting per user

See [docs/10-security-privacy.md](./docs/10-security-privacy.md) for detailed security architecture.

## Dependency Security

We scan dependencies regularly:

```bash
# Check Go dependencies
go list -json -m all | nancy sleuth

# Check for outdated/vulnerable packages
go list -u -m all
```

Automated via GitHub Dependabot and CI/CD pipeline.

## Contact

- **Security Issues**: <security@pezware.com>
- **Code Owner**: @arbeitandy
- **General Issues**: <https://github.com/pezware/samedi.dev/issues>
