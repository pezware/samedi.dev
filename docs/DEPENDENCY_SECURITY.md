# Dependency Security Policy

This document explains our approach to dependency security vulnerabilities.

## Current Status

### Known Vulnerabilities in Dev Dependencies

**golangci-lint v1.64.8 → mapstructure v2.2.1**

- **Severity**: Moderate
- **CVE**: GHSA-fv92-fjc5-jj9h, GHSA-2464-8j7c-4cjm
- **Issue**: May leak sensitive information in logs when processing malformed data
- **Status**: Accepted risk (see rationale below)

## Risk Assessment

### Why This Vulnerability is Acceptable

**1. Development Tool Only**
- golangci-lint only runs during development and CI
- Never deployed to production
- Doesn't process user data or production traffic

**2. Limited Exposure**
- Vulnerability requires malformed data input to trigger
- Linter processes Go source code (controlled input)
- No external/untrusted data processed

**3. Upgrade Path Blocked**
- golangci-lint v1.64.8 is the last v1 release
- golangci-lint v2.x has breaking configuration changes
- Migration requires significant effort (see docs/GOLANGCI_LINT_V2_ISSUE.md)

**4. Mitigation in Place**
- Logs are only visible to authorized developers and CI systems
- No sensitive credentials passed to linter
- Pre-commit hooks prevent committing secrets

## Dependency Review Configuration

We configure `actions/dependency-review-action` with:

```yaml
fail-on-severity: high        # Only fail on HIGH/CRITICAL
fail-on-scopes: runtime       # Only check runtime dependencies
comment-summary-in-pr: on-failure
```

**Rationale:**
- **fail-on-severity: high** - Accept MODERATE vulnerabilities in dev tools
- **fail-on-scopes: runtime** - Ignore development-only dependencies
- Dev tool vulnerabilities have lower risk than production dependencies

## Production vs Development Dependencies

### Runtime/Production Dependencies
- ✅ Must have NO known vulnerabilities (HIGH or CRITICAL)
- ✅ Updated immediately when vulnerabilities found
- ✅ These affect actual application security

### Development Dependencies (tools.go)
- ⚠️ MODERATE vulnerabilities accepted case-by-case
- ✅ HIGH/CRITICAL must still be addressed
- ✅ Risk assessment required

## When to Upgrade/Fix

### Immediate Action Required
- **HIGH or CRITICAL** severity in ANY dependency
- **MODERATE** severity in runtime/production dependencies
- Actively exploited vulnerabilities
- Affects data security or user privacy

### Can Be Deferred
- **LOW** severity in any dependency
- **MODERATE** severity in dev-only tools with limited exposure
- No known exploits or public PoCs
- Mitigation controls in place

## Monitoring

### Regular Reviews
- **Weekly**: Check for HIGH/CRITICAL vulnerabilities
- **Monthly**: Review all MODERATE vulnerabilities in dev tools
- **Quarterly**: Evaluate upgrade paths for pinned dependencies

### Tools
1. **GitHub Dependabot** - Automated alerts
2. **dependency-review-action** - PR checks
3. **govulncheck** - Go-specific vulnerability scanner

```bash
# Run locally
make vuln
# Or directly
govulncheck ./...
```

## Current Pinned Dependencies

| Dependency | Version | Reason | Review Date |
|------------|---------|--------|-------------|
| golangci-lint | v1.64.8 | v2.x breaking changes | 2025-10-06 |

## Upgrade Paths

### golangci-lint v1.64.8 → v2.x

**When to upgrade:**
- v2.x configuration stabilizes
- Security vulnerability becomes HIGH/CRITICAL
- Dedicated time for config migration (estimated 2-4 hours)

**Upgrade checklist:**
- [ ] Read migration guide: https://golangci-lint.run/docs/product/migration-guide
- [ ] Rewrite `.golangci.yml` for v2 schema
- [ ] Test all linters still work
- [ ] Update CI/CD workflows
- [ ] Update team documentation
- [ ] Create PR with migration

## False Positives

If dependency-review reports false positives:

1. **Verify it's actually a false positive**
   ```bash
   govulncheck ./...
   go mod why <vulnerable-package>
   ```

2. **Document the exception** in this file

3. **Use allow-list if needed**
   ```yaml
   # .github/workflows/ci.yml
   with:
     allow-ghsas: GHSA-xxxx-xxxx-xxxx
   ```

## Useful Commands

```bash
# Check direct dependencies
go list -m all

# Check why a package is included
go mod why github.com/some/package

# Find vulnerabilities
govulncheck ./...

# Update dependencies
go get -u ./...
go mod tidy

# Check for outdated packages
go list -u -m all
```

## References

- [GitHub Advisory Database](https://github.com/advisories)
- [Go Vulnerability Database](https://vuln.go.dev/)
- [OWASP Dependency-Check](https://owasp.org/www-project-dependency-check/)
- [dependency-review-action docs](https://github.com/actions/dependency-review-action)

## Questions?

- **Security issues**: Contact @arbeitandy immediately
- **Vulnerability found**: Open issue with `security` label
- **False positive**: Document in this file and create PR

---

**Last Updated**: 2025-10-06
**Next Review**: 2025-11-06
