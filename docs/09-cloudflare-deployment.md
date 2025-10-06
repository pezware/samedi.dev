# Cloudflare Deployment (Phase 2)

## Overview

Phase 2 adds **cloud sync** and **web dashboard** using Cloudflare's edge platform:
- **Workers**: API endpoints for sync
- **D1**: SQLite database at the edge
- **R2**: Object storage for backups
- **Pages**: Web dashboard (read-only, mobile-friendly)
- **KV**: Auth tokens and session state

**Philosophy**: Local-first with optional cloud sync. CLI works offline, syncs when online.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Tier                            │
│  ┌──────────────┐              ┌────────────────────┐      │
│  │  Samedi CLI  │◄────────────►│  Web Dashboard     │      │
│  │   (Local)    │   Sync API   │ (Cloudflare Pages)│      │
│  └──────────────┘              └────────────────────┘      │
└────────────┬─────────────────────────────────┬─────────────┘
             │                                  │
             ▼                                  ▼
┌─────────────────────────────────────────────────────────────┐
│              Cloudflare Workers (Edge API)                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │   Auth   │  │   Sync   │  │   Stats  │  │  Backup  │  │
│  │          │  │  Engine  │  │   API    │  │   API    │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└────────────┬─────────────────────────────────┬─────────────┘
             │                                  │
             ▼                                  ▼
┌─────────────────────────────────────────────────────────────┐
│                  Cloudflare Storage                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │      D1      │  │      R2      │  │   KV Store   │     │
│  │  (Sessions,  │  │  (Backups,   │  │ (Auth, Cache)│     │
│  │   Plans)     │  │   Exports)   │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow

### Sync Strategy: Last-Write-Wins

**Assumptions**:
- Single user across multiple devices
- No concurrent edits (rare in learning workflow)
- Conflicts resolved by timestamp

**Sync Model**:
```typescript
interface SyncState {
  device_id: string;
  last_sync: number;        // Unix timestamp
  local_version: number;
  remote_version: number;
}

interface SyncPayload {
  plans: PlanDelta[];
  sessions: SessionDelta[];
  cards: CardDelta[];
  deleted_ids: string[];
}

type DeltaAction = 'create' | 'update' | 'delete';

interface Delta {
  action: DeltaAction;
  id: string;
  data?: object;
  updated_at: number;
}
```

### Sync Flow

**1. CLI Initiates Sync**:
```bash
samedi sync
# Or automatic every 30min (configurable)
```

**2. Calculate Deltas**:
```go
func (c *SyncClient) CalculateDeltas() (*SyncPayload, error) {
    lastSync := c.GetLastSyncTime()

    // Find local changes since last sync
    plans := c.db.GetPlansModifiedSince(lastSync)
    sessions := c.db.GetSessionsModifiedSince(lastSync)
    cards := c.db.GetCardsModifiedSince(lastSync)

    // Build deltas
    payload := &SyncPayload{
        Plans:    convertToDeltas(plans),
        Sessions: convertToDeltas(sessions),
        Cards:    convertToDeltas(cards),
    }

    return payload, nil
}
```

**3. Send to API**:
```go
func (c *SyncClient) Push(payload *SyncPayload) error {
    resp, err := c.httpClient.Post(
        c.apiURL + "/sync",
        "application/json",
        marshalPayload(payload),
    )
    if err != nil {
        return err
    }

    // API returns remote deltas
    var remotePayload SyncPayload
    json.NewDecoder(resp.Body).Decode(&remotePayload)

    // Apply remote changes locally
    c.ApplyDeltas(&remotePayload)

    return nil
}
```

**4. Worker Processes Request**:
```typescript
// workers/src/routes/sync.ts
export async function handleSync(
  req: Request,
  env: Env
): Promise<Response> {
  const payload: SyncPayload = await req.json();
  const userId = req.headers.get('X-User-ID');
  const deviceId = req.headers.get('X-Device-ID');

  // Apply incoming deltas to D1
  await applyDeltas(env.DB, userId, payload);

  // Calculate remote deltas since client's last sync
  const lastSync = await getLastSync(env.KV, userId, deviceId);
  const remoteDeltas = await getRemoteDeltas(env.DB, userId, lastSync);

  // Update sync state
  await updateSyncState(env.KV, userId, deviceId, Date.now());

  return Response.json(remoteDeltas);
}
```

**5. Apply Remote Deltas**:
```go
func (c *SyncClient) ApplyDeltas(deltas *SyncPayload) error {
    tx := c.db.Begin()
    defer tx.Rollback()

    for _, delta := range deltas.Plans {
        switch delta.Action {
        case "create":
            tx.CreatePlan(delta.Data)
        case "update":
            tx.UpdatePlan(delta.ID, delta.Data)
        case "delete":
            tx.DeletePlan(delta.ID)
        }
    }

    // Same for sessions, cards...

    tx.Commit()
    c.SetLastSyncTime(time.Now())

    return nil
}
```

## Authentication

### Email Magic Link

**Flow**:
```
1. User: samedi login email@example.com
2. CLI: POST /auth/magic-link → Worker sends email
3. User: Clicks link in email
4. Browser: Opens /auth/verify?token=xyz
5. Worker: Validates token, returns JWT
6. Browser: Shows "Copy this token and paste in CLI"
7. CLI: Prompts for token, saves to ~/.samedi/auth.json
```

**Implementation**:

```typescript
// workers/src/routes/auth.ts
export async function sendMagicLink(
  email: string,
  env: Env
): Promise<void> {
  const token = generateToken();

  // Store token in KV (expires in 15min)
  await env.KV.put(`magic:${token}`, email, {
    expirationTtl: 900,
  });

  // Send email via Cloudflare Email Routing or Mailgun
  await sendEmail({
    to: email,
    subject: 'Login to Samedi',
    body: `Click here to login: https://samedi.dev/auth/verify?token=${token}`,
  });
}

export async function verifyMagicLink(
  token: string,
  env: Env
): Promise<string> {
  const email = await env.KV.get(`magic:${token}`);
  if (!email) {
    throw new Error('Invalid or expired token');
  }

  // Get or create user
  const user = await getOrCreateUser(env.DB, email);

  // Generate JWT
  const jwt = await generateJWT({
    user_id: user.id,
    email: user.email,
  }, env.JWT_SECRET);

  // Clean up magic link
  await env.KV.delete(`magic:${token}`);

  return jwt;
}
```

**CLI Auth**:
```go
func (c *AuthClient) Login(email string) error {
    // 1. Request magic link
    _, err := c.httpClient.Post(
        c.apiURL + "/auth/magic-link",
        "application/json",
        strings.NewReader(fmt.Sprintf(`{"email":"%s"}`, email)),
    )
    if err != nil {
        return err
    }

    fmt.Println("✉️  Magic link sent to", email)
    fmt.Println("Check your email and paste the token here:")

    // 2. Prompt for token
    var token string
    fmt.Scanln(&token)

    // 3. Verify token
    resp, err := c.httpClient.Get(
        c.apiURL + "/auth/verify?token=" + token,
    )
    if err != nil {
        return err
    }

    var authResp struct {
        JWT string `json:"jwt"`
    }
    json.NewDecoder(resp.Body).Decode(&authResp)

    // 4. Save JWT
    c.SaveToken(authResp.JWT)

    fmt.Println("✓ Logged in successfully")
    return nil
}
```

### JWT Validation

**Worker Middleware**:
```typescript
export async function authenticateRequest(
  req: Request,
  env: Env
): Promise<User> {
  const authHeader = req.headers.get('Authorization');
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    throw new Error('Unauthorized');
  }

  const token = authHeader.substring(7);

  try {
    const payload = await verifyJWT(token, env.JWT_SECRET);
    return {
      id: payload.user_id,
      email: payload.email,
    };
  } catch (e) {
    throw new Error('Invalid token');
  }
}
```

## D1 Database Schema

**Plans Table**:
```sql
CREATE TABLE plans (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    total_hours REAL,
    status TEXT NOT NULL,
    tags TEXT,                        -- JSON array
    content TEXT,                     -- Full markdown content

    UNIQUE(user_id, id)
);

CREATE INDEX idx_plans_user ON plans(user_id);
CREATE INDEX idx_plans_updated ON plans(user_id, updated_at);
```

**Sessions Table**:
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    chunk_id TEXT,
    start_time INTEGER NOT NULL,
    end_time INTEGER,
    duration_minutes INTEGER,
    notes TEXT,
    artifacts TEXT,                   -- JSON array
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_plan ON sessions(user_id, plan_id);
CREATE INDEX idx_sessions_updated ON sessions(user_id, updated_at);
```

**Cards Table**:
```sql
CREATE TABLE cards (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    chunk_id TEXT,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    tags TEXT,
    ease_factor REAL DEFAULT 2.5,
    interval_days INTEGER DEFAULT 1,
    repetitions INTEGER DEFAULT 0,
    next_review INTEGER NOT NULL,
    last_review INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    FOREIGN KEY (plan_id) REFERENCES plans(id)
);

CREATE INDEX idx_cards_user ON cards(user_id);
CREATE INDEX idx_cards_review ON cards(user_id, next_review);
CREATE INDEX idx_cards_updated ON cards(user_id, updated_at);
```

**Users Table**:
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX idx_users_email ON users(email);
```

## R2 Storage

### Backup Storage

**Structure**:
```
samedi-backups/
├── {user_id}/
│   ├── daily/
│   │   ├── 2024-01-20.tar.gz
│   │   ├── 2024-01-21.tar.gz
│   │   └── ...
│   ├── weekly/
│   │   ├── 2024-W03.tar.gz
│   │   └── ...
│   └── exports/
│       ├── plans-2024-01-20.md
│       └── cards-2024-01-20.json
```

**Upload Backup**:
```typescript
export async function uploadBackup(
  userId: string,
  data: ArrayBuffer,
  env: Env
): Promise<void> {
  const key = `${userId}/daily/${new Date().toISOString().split('T')[0]}.tar.gz`;

  await env.R2.put(key, data, {
    httpMetadata: {
      contentType: 'application/gzip',
    },
    customMetadata: {
      userId,
      createdAt: Date.now().toString(),
    },
  });
}
```

**Download Backup**:
```go
func (c *BackupClient) Download(backupID string) ([]byte, error) {
    resp, err := c.httpClient.Get(
        c.apiURL + "/backup/" + backupID,
    )
    if err != nil {
        return nil, err
    }

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return data, nil
}
```

## KV Store

**Use Cases**:
1. **Auth Tokens**: Magic link tokens (TTL: 15min)
2. **Sync State**: Per-device last sync timestamp
3. **Rate Limiting**: Per-user request counts
4. **Cache**: Frequently accessed stats

**Example**:
```typescript
// Store sync state
await env.KV.put(
  `sync:${userId}:${deviceId}`,
  JSON.stringify({
    last_sync: Date.now(),
    version: 123,
  }),
  { expirationTtl: 86400 * 30 } // 30 days
);

// Get sync state
const state = await env.KV.get(`sync:${userId}:${deviceId}`, 'json');
```

## Web Dashboard

### Technology Stack

- **Framework**: Hono (lightweight, works on Workers)
- **Templating**: HTMX (server-rendered, minimal JS)
- **Styling**: Tailwind CSS
- **Deployment**: Cloudflare Pages

### Pages Structure

**1. Dashboard (/)** - Overview stats
```html
<div class="dashboard">
  <h1>Learning Dashboard</h1>

  <div class="stats">
    <div class="stat">
      <span class="value">127.5h</span>
      <span class="label">Total Time</span>
    </div>
    <div class="stat">
      <span class="value">12🔥</span>
      <span class="label">Streak</span>
    </div>
  </div>

  <div class="plans">
    <!-- Plan cards -->
  </div>
</div>
```

**2. Plan Details (/plans/:id)** - Plan progress
```html
<div class="plan-detail">
  <h1>French B1 Mastery</h1>
  <progress value="24" max="100"></progress>

  <div class="chunks">
    <!-- Chunk list with status -->
  </div>

  <div class="sessions">
    <!-- Recent sessions -->
  </div>
</div>
```

**3. Stats (/stats)** - Analytics
```html
<div class="stats-page">
  <canvas id="hours-chart"></canvas>
  <div class="heatmap"><!-- Calendar heatmap --></div>
</div>
```

### Hono API Routes

```typescript
// web/src/index.ts
import { Hono } from 'hono';

const app = new Hono<{ Bindings: Env }>();

app.get('/', async (c) => {
  const user = await authenticateRequest(c.req, c.env);
  const stats = await getStats(c.env.DB, user.id);

  return c.html(renderDashboard(stats));
});

app.get('/plans/:id', async (c) => {
  const user = await authenticateRequest(c.req, c.env);
  const planId = c.req.param('id');
  const plan = await getPlan(c.env.DB, user.id, planId);

  return c.html(renderPlanDetail(plan));
});

app.get('/api/stats', async (c) => {
  const user = await authenticateRequest(c.req, c.env);
  const stats = await getStats(c.env.DB, user.id);

  return c.json(stats);
});

export default app;
```

### HTMX Interactivity

**Live Stats Update**:
```html
<div hx-get="/api/stats"
     hx-trigger="every 30s"
     hx-swap="outerHTML">
  <!-- Stats content, auto-refreshes -->
</div>
```

**Infinite Scroll (Session History)**:
```html
<div id="sessions">
  <!-- Session items -->
</div>
<div hx-get="/api/sessions?offset=20"
     hx-trigger="revealed"
     hx-swap="beforeend"
     hx-target="#sessions">
</div>
```

## Testing Strategy

### Local Development

**Miniflare** for local Workers dev:
```bash
# Install
npm install -D miniflare

# Run locally
npx miniflare --watch
```

**Vitest** for Workers tests:
```typescript
// workers/test/sync.test.ts
import { describe, it, expect } from 'vitest';
import { handleSync } from '../src/routes/sync';

describe('Sync API', () => {
  it('should merge local and remote deltas', async () => {
    const payload = {
      plans: [{ action: 'create', id: 'plan-1', data: {...} }],
      sessions: [],
      cards: [],
    };

    const req = new Request('http://localhost/sync', {
      method: 'POST',
      body: JSON.stringify(payload),
      headers: { 'X-User-ID': 'user-1' },
    });

    const resp = await handleSync(req, mockEnv);
    const data = await resp.json();

    expect(data.plans).toHaveLength(1);
  });
});
```

### Integration Tests

**CLI → Worker → D1**:
```go
func TestE2E_Sync(t *testing.T) {
    // 1. Create local session
    cli := NewCLI(testConfig)
    cli.Start("french-b1", "chunk-001")
    time.Sleep(1 * time.Second)
    cli.Stop()

    // 2. Sync to cloud
    err := cli.Sync()
    assert.NoError(t, err)

    // 3. Verify in D1 (via API)
    resp, _ := http.Get(testAPIURL + "/sessions?plan=french-b1")
    var sessions []Session
    json.NewDecoder(resp.Body).Decode(&sessions)

    assert.Len(t, sessions, 1)
    assert.Equal(t, "chunk-001", sessions[0].ChunkID)
}
```

## Deployment

### Workers Deployment

**wrangler.toml**:
```toml
name = "samedi-api"
main = "src/index.ts"
compatibility_date = "2024-01-01"

[env.production]
vars = { ENVIRONMENT = "production" }

[[env.production.d1_databases]]
binding = "DB"
database_name = "samedi-prod"
database_id = "xxx"

[[env.production.r2_buckets]]
binding = "R2"
bucket_name = "samedi-backups"

[[env.production.kv_namespaces]]
binding = "KV"
id = "xxx"
```

**Deploy**:
```bash
npx wrangler deploy
```

### Pages Deployment

**Build**:
```bash
cd web
npm run build
```

**Deploy**:
```bash
npx wrangler pages deploy dist
```

### CI/CD (GitHub Actions)

```yaml
# .github/workflows/deploy.yml
name: Deploy to Cloudflare

on:
  push:
    branches: [main]

jobs:
  deploy-workers:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npx wrangler deploy
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CF_API_TOKEN }}

  deploy-pages:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: cd web && npm ci && npm run build
      - run: npx wrangler pages deploy web/dist
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CF_API_TOKEN }}
```

## Cost Estimation

**Cloudflare Free Tier** (generous):
- Workers: 100k requests/day
- D1: 5GB storage, 5M reads/day
- R2: 10GB storage, 1M reads/month
- Pages: Unlimited requests
- KV: 100k reads/day, 1k writes/day

**Projected Usage** (100 users):
- Sync requests: ~5k/day (50 syncs/user/day)
- D1 queries: ~50k/day
- R2 storage: ~5GB (backups)
- KV ops: ~10k/day

**Cost**: $0/month (within free tier) 🎉

**Scaling** (10k users):
- Workers: $5/month (10M requests)
- D1: $5/month (500M reads)
- R2: $0.15/GB ($0.75 for 5GB)
- Total: ~$11/month

## Migration Path

### Phase 1 → Phase 2

**User Migration**:
```bash
# One-time setup
samedi login email@example.com
samedi sync push --all   # Upload all local data

# Future: Auto-sync enabled
samedi config set sync.enabled true
```

**Data Migration**:
```go
func (c *MigrationClient) UploadAll() error {
    // 1. Get all local data
    plans := c.db.GetAllPlans()
    sessions := c.db.GetAllSessions()
    cards := c.db.GetAllCards()

    // 2. Batch upload
    for _, plan := range plans {
        c.api.CreatePlan(plan)
    }

    // Same for sessions, cards...

    // 3. Mark as synced
    c.db.SetLastSyncTime(time.Now())

    return nil
}
```

## Monitoring & Observability

### Worker Analytics

**Built-in Cloudflare Analytics**:
- Request count, duration
- Error rates
- CPU time

**Custom Metrics** (via Durable Objects or external):
```typescript
export async function trackMetric(
  name: string,
  value: number,
  env: Env
): Promise<void> {
  await env.ANALYTICS.writeDataPoint({
    blobs: [name],
    doubles: [value],
    indexes: [env.USER_ID],
  });
}

// Usage
await trackMetric('sync.duration_ms', 150, env);
```

### Logging

**LogPush** to external service:
```typescript
console.log(JSON.stringify({
  event: 'sync.completed',
  user_id: user.id,
  deltas: payload.plans.length + payload.sessions.length,
  duration_ms: Date.now() - startTime,
}));
```

## Security

### Rate Limiting

```typescript
export async function checkRateLimit(
  userId: string,
  env: Env
): Promise<boolean> {
  const key = `ratelimit:${userId}:${Math.floor(Date.now() / 60000)}`;
  const count = await env.KV.get(key);

  if (count && parseInt(count) >= 100) {
    throw new Error('Rate limit exceeded');
  }

  await env.KV.put(key, (parseInt(count || '0') + 1).toString(), {
    expirationTtl: 60,
  });

  return true;
}
```

### Data Isolation

**Row-level security**:
```sql
-- All queries filter by user_id
SELECT * FROM plans WHERE user_id = ? AND id = ?;
```

**Worker enforcement**:
```typescript
// Always inject user_id from auth token, never from request
const user = await authenticateRequest(req, env);
const plan = await getPlan(env.DB, user.id, planId);  // user.id from JWT
```

## Future Enhancements (Phase 3+)

### Real-Time Sync (Durable Objects)

```typescript
export class SyncSession {
  async fetch(request: Request) {
    const ws = new WebSocketPair();
    this.handleWebSocket(ws[1]);
    return new Response(null, { status: 101, webSocket: ws[0] });
  }

  async handleWebSocket(ws: WebSocket) {
    ws.addEventListener('message', async (event) => {
      const delta = JSON.parse(event.data);
      await this.applyDelta(delta);
      this.broadcast(delta);  // To other devices
    });
  }
}
```

### Collaborative Plans

Share plans with other users (read-only or edit):
```sql
CREATE TABLE plan_shares (
    plan_id TEXT,
    shared_with_user_id TEXT,
    permission TEXT,  -- 'read' or 'edit'
    created_at INTEGER,
    PRIMARY KEY (plan_id, shared_with_user_id)
);
```
