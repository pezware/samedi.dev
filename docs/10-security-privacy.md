# Security & Privacy

## Principles

1. **Privacy by Default**: No data collection without explicit consent
2. **Local-First**: Data stays on device unless user enables sync
3. **Minimal Data**: Only collect what's necessary
4. **User Control**: Users own their data, can export/delete anytime
5. **Transparency**: Clear documentation of what's collected and why

## Threat Model

### Assets to Protect

1. **Learning Data**: Plans, sessions, notes (sensitive intellectual property)
2. **Personal Information**: Email, learning habits, progress
3. **Authentication Credentials**: JWT tokens, API keys
4. **LLM Interactions**: Prompts sent to external LLM services

### Threat Actors

1. **Malicious User**: Attempts to access others' data
2. **Network Attacker**: Intercepts sync traffic
3. **Compromised Device**: Malware attempts to steal local data
4. **Malicious LLM Provider**: Could log prompts containing sensitive data
5. **Cloud Provider Breach**: Cloudflare or D1 compromise

### Out of Scope (MVP)

- Advanced persistent threats (APT)
- Physical device theft (rely on OS encryption)
- Zero-knowledge encryption (complexity vs. benefit)

## Local Security (Phase 1)

### File Permissions

**Default Permissions**:
```bash
~/.samedi/
├── sessions.db          (chmod 600)  # User read/write only
├── config.toml          (chmod 600)
├── plans/               (chmod 700)
│   └── *.md             (chmod 600)
└── cards/               (chmod 700)
    └── *.md             (chmod 600)
```

**Implementation**:
```go
func (s *Storage) CreateFile(path string, data []byte) error {
    // Create with restrictive permissions
    if err := os.WriteFile(path, data, 0600); err != nil {
        return err
    }

    // Ensure parent directory is also protected
    dir := filepath.Dir(path)
    if err := os.Chmod(dir, 0700); err != nil {
        return err
    }

    return nil
}
```

### API Key Storage

**Never in Config Files**:
```toml
# ❌ Bad: API key in config
[llm]
api_key = "sk-..."

# ✅ Good: Reference environment variable
[llm]
api_key_env = "OPENAI_API_KEY"
```

**Implementation**:
```go
func (l *LLMProvider) GetAPIKey() (string, error) {
    envVar := l.config.APIKeyEnv
    if envVar == "" {
        return "", errors.New("no API key configured")
    }

    key := os.Getenv(envVar)
    if key == "" {
        return "", fmt.Errorf("environment variable %s not set", envVar)
    }

    return key, nil
}
```

### SQLite Encryption (Optional)

**Using SQLCipher** (opt-in):
```go
import "github.com/mutecomm/go-sqlcipher/v4"

func OpenEncryptedDB(path, passphrase string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3",
        path+"?_pragma_key="+passphrase+"&_pragma_cipher_page_size=4096")
    if err != nil {
        return nil, err
    }

    return db, nil
}
```

**User Experience**:
```bash
samedi config set storage.encrypted true
samedi config set-passphrase

Enter passphrase: ********
Confirm: ********

✓ Database encrypted. You'll need this passphrase to access your data.
⚠ IMPORTANT: Store this passphrase safely. It cannot be recovered if lost.
```

### Sensitive Data in Memory

**Clear Sensitive Strings**:
```go
type SecureString struct {
    data []byte
}

func (s *SecureString) Clear() {
    // Overwrite with zeros
    for i := range s.data {
        s.data[i] = 0
    }
    s.data = nil
}

// Usage
apiKey := NewSecureString(os.Getenv("OPENAI_API_KEY"))
defer apiKey.Clear()

// Use apiKey.String() as needed
```

### LLM Prompt Safety

**Sanitize Before Sending**:
```go
func (l *LLMProvider) SanitizePrompt(prompt string) string {
    // Remove potential PII
    re := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
    prompt = re.ReplaceAllString(prompt, "[EMAIL]")

    // Remove phone numbers
    re = regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`)
    prompt = re.ReplaceAllString(prompt, "[PHONE]")

    // Remove credit cards
    re = regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
    prompt = re.ReplaceAllString(prompt, "[REDACTED]")

    return prompt
}
```

**Warning to Users**:
```bash
samedi init "project x internal roadmap" --warn
```

```
⚠ Prompt Safety Warning:

You're about to send this to an external LLM (claude):
"project x internal roadmap"

This prompt may contain sensitive information:
- Company project names
- Internal terminology

The LLM provider (Anthropic) may log this prompt.

Continue? (y/N):
```

**Config Option**:
```toml
[llm]
warn_on_sensitive = true  # Prompt before sending potentially sensitive data
```

## Cloud Security (Phase 2)

### Transport Security

**TLS for All API Calls**:
```go
func NewSyncClient(apiURL string) *SyncClient {
    // Enforce HTTPS
    if !strings.HasPrefix(apiURL, "https://") {
        panic("API URL must use HTTPS")
    }

    return &SyncClient{
        httpClient: &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    MinVersion: tls.VersionTLS12,
                },
            },
        },
        apiURL: apiURL,
    }
}
```

**Certificate Pinning** (optional, high-security):
```go
import "crypto/x509"

func (c *SyncClient) PinCertificate(pemCert []byte) error {
    cert, err := x509.ParseCertificate(pemCert)
    if err != nil {
        return err
    }

    c.pinnedCert = cert
    c.httpClient.Transport.(*http.Transport).TLSClientConfig.VerifyPeerCertificate =
        func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
            for _, rawCert := range rawCerts {
                if bytes.Equal(rawCert, c.pinnedCert.Raw) {
                    return nil
                }
            }
            return errors.New("certificate not pinned")
        }

    return nil
}
```

### Authentication

**JWT Structure**:
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "user-123",
    "email": "user@example.com",
    "iat": 1705843200,
    "exp": 1705846800  // 1 hour expiry
  }
}
```

**JWT Validation** (Worker):
```typescript
import { jwtVerify } from 'jose';

export async function authenticateRequest(
  req: Request,
  env: Env
): Promise<User> {
  const authHeader = req.headers.get('Authorization');
  if (!authHeader?.startsWith('Bearer ')) {
    throw new Error('Missing or invalid Authorization header');
  }

  const token = authHeader.substring(7);

  try {
    const { payload } = await jwtVerify(
      token,
      new TextEncoder().encode(env.JWT_SECRET),
      { algorithms: ['HS256'] }
    );

    // Check expiry
    if (payload.exp && payload.exp < Date.now() / 1000) {
      throw new Error('Token expired');
    }

    return {
      id: payload.user_id as string,
      email: payload.email as string,
    };
  } catch (e) {
    throw new Error('Invalid token');
  }
}
```

**Token Refresh**:
```go
func (c *AuthClient) RefreshToken() error {
    // Check if token expires soon (< 5 min)
    if !c.TokenExpiresSoon() {
        return nil
    }

    // Request new token
    resp, err := c.httpClient.Post(
        c.apiURL + "/auth/refresh",
        "application/json",
        strings.NewReader(fmt.Sprintf(`{"refresh_token":"%s"}`, c.refreshToken)),
    )
    if err != nil {
        return err
    }

    var newAuth struct {
        JWT string `json:"jwt"`
    }
    json.NewDecoder(resp.Body).Decode(&newAuth)

    c.SaveToken(newAuth.JWT)
    return nil
}
```

### Authorization

**Row-Level Security**:
```typescript
// All queries MUST filter by user_id
export async function getPlan(
  db: D1Database,
  userId: string,
  planId: string
): Promise<Plan | null> {
  const result = await db
    .prepare('SELECT * FROM plans WHERE user_id = ? AND id = ?')
    .bind(userId, planId)
    .first();

  return result as Plan | null;
}

// Never trust client-provided user_id
// ❌ BAD: const userId = req.params.userId;
// ✅ GOOD: const userId = authenticatedUser.id;
```

**Access Control Middleware**:
```typescript
export async function requireAuth(
  req: Request,
  env: Env,
  next: () => Promise<Response>
): Promise<Response> {
  try {
    const user = await authenticateRequest(req, env);
    // Attach user to request context
    req.user = user;
    return await next();
  } catch (e) {
    return new Response('Unauthorized', { status: 401 });
  }
}
```

### Rate Limiting

**Per-User Limits**:
```typescript
export async function checkRateLimit(
  userId: string,
  endpoint: string,
  env: Env
): Promise<void> {
  const key = `ratelimit:${userId}:${endpoint}:${Math.floor(Date.now() / 60000)}`;

  const current = await env.KV.get(key);
  const count = current ? parseInt(current) : 0;

  const limit = RATE_LIMITS[endpoint] || 100; // Default: 100 req/min

  if (count >= limit) {
    throw new Error('Rate limit exceeded');
  }

  await env.KV.put(key, (count + 1).toString(), {
    expirationTtl: 60,
  });
}

const RATE_LIMITS = {
  '/sync': 10,        // 10 syncs/min
  '/plans': 20,       // 20 plan operations/min
  '/cards': 50,       // 50 card operations/min
};
```

**Global Rate Limit** (DDoS protection):
```typescript
// Cloudflare built-in: 1000 req/min per IP
// Configure in wrangler.toml or dashboard
```

### Data Encryption at Rest

**D1 Data** (Cloudflare-managed):
- Encrypted at rest by default (AES-256)
- Encrypted in transit (TLS 1.2+)

**R2 Backups**:
```typescript
import { subtle } from 'crypto';

export async function encryptBackup(
  data: ArrayBuffer,
  key: CryptoKey
): Promise<ArrayBuffer> {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const encrypted = await subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    data
  );

  // Prepend IV to encrypted data
  const result = new Uint8Array(iv.length + encrypted.byteLength);
  result.set(iv);
  result.set(new Uint8Array(encrypted), iv.length);

  return result.buffer;
}

export async function uploadEncryptedBackup(
  userId: string,
  data: ArrayBuffer,
  env: Env
): Promise<void> {
  // Derive key from user's password (stored in KV, hashed)
  const key = await deriveKey(userId, env);

  const encrypted = await encryptBackup(data, key);

  await env.R2.put(`${userId}/backup-${Date.now()}.enc`, encrypted);
}
```

## Privacy

### Data Collection

**What We Collect** (Phase 1 - Local Only):
- ❌ None. All data stays local.

**What We Collect** (Phase 2 - Cloud Sync):
- ✅ Email address (for auth)
- ✅ Learning data (plans, sessions, cards) - **only if user enables sync**
- ✅ Anonymous usage stats (opt-in):
  - Command frequency (which commands are used)
  - Error rates (to fix bugs)
  - Performance metrics (response times)

**What We DON'T Collect**:
- ❌ LLM prompts/responses (never logged)
- ❌ Learning content (notes, code, etc.)
- ❌ IP addresses (beyond temporary rate limiting)
- ❌ Device fingerprints
- ❌ Third-party tracking (no Google Analytics, no ads)

### Telemetry (Opt-In)

**User Control**:
```bash
# Disabled by default
samedi config set telemetry.enabled false

# Opt-in
samedi config set telemetry.enabled true
```

**What's Sent**:
```json
{
  "event": "command_executed",
  "command": "start",
  "duration_ms": 150,
  "success": true,
  "user_id_hash": "sha256(user_id)",  // Anonymized
  "version": "1.0.0",
  "platform": "darwin"
}
```

**Privacy-Preserving Analytics**:
```typescript
// Hash user_id before sending
function hashUserId(userId: string): string {
  return sha256(userId + SALT);  // One-way hash, can't reverse
}

// Aggregate only, no individual tracking
// Example query: "How many users ran 'samedi start' today?"
// NOT: "What did user X do today?"
```

### GDPR Compliance

**Right to Access**:
```bash
samedi export all > my-data.json
```

Exports everything in portable JSON format.

**Right to Erasure**:
```bash
samedi account delete

⚠ This will permanently delete:
- All cloud data (plans, sessions, cards)
- Your account and email
- Backups in R2

Local data will be preserved.

Type 'DELETE' to confirm: DELETE

✓ Account deleted. All cloud data erased.
```

**Implementation**:
```typescript
export async function deleteUser(
  userId: string,
  env: Env
): Promise<void> {
  // 1. Delete from D1
  await env.DB.prepare('DELETE FROM plans WHERE user_id = ?').bind(userId).run();
  await env.DB.prepare('DELETE FROM sessions WHERE user_id = ?').bind(userId).run();
  await env.DB.prepare('DELETE FROM cards WHERE user_id = ?').bind(userId).run();
  await env.DB.prepare('DELETE FROM users WHERE id = ?').bind(userId).run();

  // 2. Delete R2 backups
  const objects = await env.R2.list({ prefix: `${userId}/` });
  for (const obj of objects.objects) {
    await env.R2.delete(obj.key);
  }

  // 3. Delete KV entries
  await env.KV.delete(`sync:${userId}`);
  await env.KV.delete(`auth:${userId}`);
}
```

**Right to Portability**:
```bash
samedi export anki > anki-deck.txt
samedi export markdown > learning-report.md
samedi export json > data-export.json
```

### Cookie Policy

**Web Dashboard**:
- ✅ Session cookies (auth only, HttpOnly, Secure, SameSite=Strict)
- ❌ No tracking cookies
- ❌ No third-party cookies

```typescript
// Set secure session cookie
export function setSessionCookie(res: Response, token: string): Response {
  res.headers.set('Set-Cookie', [
    `session=${token}`,
    'HttpOnly',
    'Secure',
    'SameSite=Strict',
    'Max-Age=3600',  // 1 hour
    'Path=/',
  ].join('; '));

  return res;
}
```

## Incident Response

### Security Issue Reporting

**Public Process**:
```markdown
# Security Policy

## Reporting a Vulnerability

Email: security@samedi.dev
PGP Key: [public key]

We'll respond within 24 hours.

## Disclosure Timeline

- Day 0: Report received
- Day 1: Confirmation sent
- Day 7: Fix developed
- Day 14: Fix deployed
- Day 30: Public disclosure (CVE assigned if applicable)
```

### Breach Response Plan

**If Cloud Data Compromised**:

1. **Immediate** (Hour 0):
   - Revoke all JWT tokens
   - Rotate JWT secret
   - Force all users to re-login

2. **Short-term** (Day 1):
   - Email all users about breach
   - Provide steps to secure accounts
   - Offer data deletion

3. **Long-term** (Week 1):
   - Publish incident report
   - Implement additional security measures
   - Third-party security audit

**User Notification**:
```
Subject: Samedi Security Incident Notification

We detected unauthorized access to our cloud database on [DATE].

What happened:
- An attacker gained access to user email addresses and encrypted learning data
- No passwords were exposed (we don't store them)
- Local data (on your device) was NOT affected

What we did:
- Immediately revoked all access tokens
- Rotated all encryption keys
- Fixed the vulnerability

What you should do:
1. Log in again: samedi login [email]
2. Review your data: samedi export all
3. Enable 2FA (when available): samedi config set auth.2fa true

We're deeply sorry. Full incident report: https://samedi.dev/security/incident-2024-01
```

## Compliance

### Data Storage Locations

**Phase 1 (Local)**:
- User's device only
- Subject to device's jurisdiction

**Phase 2 (Cloudflare)**:
- D1/R2/KV: Cloudflare's global network
- Users can specify region (future):
  ```toml
  [sync]
  region = "eu"  # EU data residency
  ```

### Data Retention

**Local**:
- Forever (until user deletes)

**Cloud**:
- Active data: Forever (until user deletes)
- Deleted data: 30-day soft delete, then permanent
- Backups: 90 days, then auto-delete

**Implementation**:
```typescript
// Soft delete (mark as deleted)
export async function softDeletePlan(
  userId: string,
  planId: string,
  env: Env
): Promise<void> {
  await env.DB
    .prepare('UPDATE plans SET deleted_at = ? WHERE user_id = ? AND id = ?')
    .bind(Date.now(), userId, planId)
    .run();
}

// Cron job: Permanently delete after 30 days
export async function purgeDeletedData(env: Env): Promise<void> {
  const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000;

  await env.DB
    .prepare('DELETE FROM plans WHERE deleted_at < ?')
    .bind(thirtyDaysAgo)
    .run();
}
```

### Terms of Service (Key Points)

```markdown
# Samedi Terms of Service

## Your Data
- You own all learning data
- We never sell your data
- You can export/delete anytime

## Our Responsibilities
- Keep your data secure
- Notify you of breaches within 72 hours
- Maintain 99.9% uptime (cloud services)

## Your Responsibilities
- Keep your login credentials secure
- Don't share accounts
- Don't abuse the service (rate limits)

## Data Processing
- Local data: No processing by Samedi
- Cloud data: Only for sync, stats, backups
- LLM data: Sent to third-party providers (Anthropic, OpenAI, etc.)
  - Subject to their terms
  - We don't control their data handling
```

## Security Checklist

### Before Launch (MVP)

- [ ] All local files are chmod 600/700
- [ ] API keys only via environment variables
- [ ] LLM prompts sanitized (PII removed)
- [ ] HTTPS enforced for all API calls
- [ ] Input validation on all user inputs
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (escape all outputs)
- [ ] CSRF protection (SameSite cookies)
- [ ] Rate limiting implemented
- [ ] Security.txt published
- [ ] Incident response plan documented

### Before Cloud Sync (Phase 2)

- [ ] JWT validation with expiry
- [ ] Row-level security in all queries
- [ ] Encrypted backups in R2
- [ ] TLS 1.2+ enforced
- [ ] GDPR data export/deletion
- [ ] Privacy policy published
- [ ] Cookie consent (if needed in EU)
- [ ] Security audit completed
- [ ] Penetration testing done

## Security Tools & Practices

### Dependency Scanning

```bash
# Go dependencies
go list -json -m all | nancy sleuth

# NPM dependencies (Workers)
npm audit

# Automated (GitHub Actions)
- uses: github/super-linter@v4
```

### Secret Scanning

```bash
# Prevent secrets in git
git secrets --install
git secrets --register-aws

# Pre-commit hook
# .git/hooks/pre-commit
#!/bin/bash
git secrets --pre_commit_hook -- "$@"
```

### Static Analysis

```bash
# Go
golangci-lint run

# TypeScript
npm run lint

# Security-specific
gosec ./...
```

### Fuzz Testing

```go
// Fuzz test LLM prompt sanitization
func FuzzSanitizePrompt(f *testing.F) {
    f.Add("test@example.com")
    f.Add("123-456-7890")

    f.Fuzz(func(t *testing.T, input string) {
        result := SanitizePrompt(input)
        if strings.Contains(result, "@") && strings.Contains(input, "@") {
            t.Error("Email not sanitized")
        }
    })
}
```

## User Education

### Security Best Practices (Docs)

```markdown
# Security Best Practices

## Protect Your Data

1. **Use Strong Passphrases** (if encrypting local DB)
   - 12+ characters
   - Mix of letters, numbers, symbols
   - Unique to Samedi

2. **Enable Disk Encryption** (FileVault, BitLocker)
   - Protects data if device stolen

3. **Be Careful with LLM Prompts**
   - Don't include passwords, API keys
   - Sanitize company-sensitive info
   - Remember: LLM providers may log prompts

4. **Regular Backups**
   - `samedi backup` weekly
   - Store in secure location (encrypted USB, cloud)

5. **Keep Software Updated**
   - `brew upgrade samedi` regularly
   - Check for security patches

## Privacy Tips

1. **Review Sync Settings**
   - `samedi config get sync.enabled`
   - Disable if you don't need multi-device

2. **Audit Your Data**
   - `samedi export all` periodically
   - Check what's being synced

3. **Use Separate Accounts** (Phase 2+)
   - Work vs. personal learning
   - Different email addresses
```

### In-App Warnings

```bash
samedi init "company project X roadmap" --hours 100
```

```
⚠ LLM Privacy Notice:

This plan will be generated by claude (Anthropic).
Your prompt will be sent to Anthropic's servers.

Prompt: "company project X roadmap"

Potential risks:
- Anthropic may log this prompt
- Internal project names could be exposed

Suggestions:
1. Use generic terms: "software project roadmap"
2. Review/edit plan after generation
3. Use local LLM (ollama) for sensitive data

Continue? (y/N):
```
