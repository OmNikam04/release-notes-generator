# Understanding the Bugsby Sync System

## 🤔 The Problem

We have **two sources** of bug data:

1. **Bugsby API** (External)
   - Arista's bug tracking system
   - Contains ALL bugs with complete information
   - Always up-to-date
   - Slow to query (network calls, rate limits)
   - We don't control it

2. **Our Database** (Local)
   - PostgreSQL database in our application
   - Fast to query (local database)
   - We can add custom fields (release note status, AI-generated notes, etc.)
   - But needs to be kept in sync with Bugsby

## 💡 The Solution: Sync System

**Sync** = Copy bugs from Bugsby → Our Database

Think of it like downloading emails:
- **Bugsby** = Gmail server (source of truth)
- **Our Database** = Your local email client (fast access, offline capability)
- **Sync** = Downloading emails to your computer

## 📊 Database Schema

Our `bugs` table has **two types of fields**:

### 1. Bugsby Fields (copied from Bugsby)
```sql
bugsby_id       -- Bug ID from Bugsby (e.g., "1092263")
title           -- Bug title
description     -- Bug description
severity        -- "sev1", "sev2", "sev3"
priority        -- "P0", "P1", "P2"
bug_type        -- "BUG", "TASK", "FEATURE"
release         -- "wifi-ooty", "main"
component       -- "gnutls", "wifi-network-config"
assigned_to     -- User UUID (from Bugsby assignee email)
```

### 2. Our Custom Fields (for release notes workflow)
```sql
status          -- "pending", "ai_generated", "dev_approved", "mgr_approved"
sync_status     -- "synced", "pending", "failed"
last_synced_at  -- When we last synced from Bugsby
```

## 🔄 Three Sync Operations

### 1. **SyncRelease** - Bulk Sync
**Endpoint**: `POST /bugsby/sync`

**What it does**:
- Fetches ALL bugs for a release from Bugsby
- Stores them in our database
- Creates user accounts if needed
- Updates existing bugs if they changed

**When to use**:
- First time setting up a release
- Periodic refresh to get latest bug updates
- After major changes in Bugsby

**Example**:
```bash
# Sync all bugs for wifi-ooty release
POST /bugsby/sync
{
  "release": "wifi-ooty",
  "status": "ASSIGNED"  # optional filter
}

# Response:
{
  "total_fetched": 150,
  "new_bugs": 120,      # Newly added to our DB
  "updated_bugs": 30,   # Already existed, updated
  "failed_bugs": 0
}
```

---

### 2. **SyncBugByID** - Single Bug Sync
**Endpoint**: `POST /bugsby/sync/:bugsby_id`

**What it does**:
- Fetches ONE specific bug from Bugsby
- Stores/updates it in our database

**When to use**:
- Someone reports a bug is missing
- Need to refresh a specific bug's data
- Testing sync functionality

**Example**:
```bash
# Sync bug #1092263
POST /bugsby/sync/1092263

# Response:
{
  "id": "uuid",
  "bugsby_id": "1092263",
  "title": "Remove Redundent Locationid Query Param",
  "status": "pending",
  "sync_status": "synced"
}
```

---

### 3. **GetSyncStatus** - Check Progress
**Endpoint**: `GET /bugsby/status?release=wifi-ooty`

**What it does**:
- Counts bugs in our database for a release
- Shows how many are synced vs pending vs failed

**When to use**:
- After running a sync, check if it completed
- Monitor sync health
- Debugging sync issues

**Example**:
```bash
GET /bugsby/status?release=wifi-ooty

# Response:
{
  "release": "wifi-ooty",
  "total_bugs": 150,
  "synced_bugs": 145,
  "pending_bugs": 3,
  "failed_bugs": 2,
  "last_synced_at": "2025-11-13T16:52:00Z"
}
```

## 🔍 Sync vs Direct Query

### Sync Endpoints (`/bugsby/sync`)
- **Purpose**: Store bugs in our database
- **Auth**: Manager only
- **Result**: Bugs saved to database
- **Use case**: Setting up release, periodic updates

### Testing Endpoints (`/bugsby-api`)
- **Purpose**: Test Bugsby API directly
- **Auth**: None (for testing)
- **Result**: Data returned but NOT stored
- **Use case**: Testing, debugging, quick lookups

### Bug Management Endpoints (`/bugs`)
- **Purpose**: Query our local database
- **Auth**: All authenticated users
- **Result**: Fast queries from our DB
- **Use case**: Daily operations, viewing bugs

## 📝 Complete Workflow Example

### Scenario: Manager sets up wifi-ooty release

**Step 1: Manager syncs all bugs**
```bash
POST /bugsby/sync
{
  "release": "wifi-ooty",
  "status": "ASSIGNED"
}

# Result: 150 bugs copied to our database
```

**Step 2: Check sync status**
```bash
GET /bugsby/status?release=wifi-ooty

# Result: 
# - 145 synced successfully
# - 3 pending
# - 2 failed
```

**Step 3: Developer views their assigned bugs**
```bash
GET /bugs?release=wifi-ooty&assigned_to=<user_id>

# Result: Fast query from our database
# Shows bugs with custom status field
```

**Step 4: AI generates release notes**
```bash
# (Future feature)
# AI reads bugs from our DB
# Generates release notes
# Updates bug.status = "ai_generated"
```

**Step 5: Developer approves release note**
```bash
PUT /bugs/:id
{
  "status": "dev_approved"
}

# This custom field only exists in our DB!
# Bugsby doesn't know about it
```

**Step 6: Periodic re-sync (weekly)**
```bash
POST /bugsby/sync
{
  "release": "wifi-ooty"
}

# Updates bugs if they changed in Bugsby
# Adds any new bugs
# Preserves our custom fields (status, etc.)
```

## 🎯 Key Insights

### Why Not Query Bugsby Directly Every Time?

❌ **Problems with direct queries**:
- Slow (network latency)
- Rate limited (Bugsby API has limits)
- Can't add custom fields
- Can't track our workflow state
- Requires Bugsby to be available

✅ **Benefits of sync + local DB**:
- Fast queries (local database)
- No rate limits
- Custom fields for our workflow
- Works even if Bugsby is down
- Can do complex queries and joins

### When Does Data Get Out of Sync?

Our database can become stale if:
- Bug is updated in Bugsby after we synced
- New bugs are added to release in Bugsby
- Bug is reassigned in Bugsby

**Solution**: Periodic re-sync (daily or weekly)

### What Happens on Re-Sync?

When you sync again:
1. Fetches latest data from Bugsby
2. **Updates** existing bugs (by `bugsby_id`)
3. **Preserves** our custom fields (`status`, etc.)
4. **Adds** any new bugs
5. Does NOT delete bugs (even if removed from Bugsby)

## 🔐 Security: Why Manager-Only?

Sync endpoints require **manager role** because:
- Prevents unauthorized bulk data imports
- Sync operations are expensive (many API calls)
- Managers control which releases are tracked
- Prevents accidental database pollution

## 🧪 Testing Endpoints

The `/bugsby-api` endpoints are for **testing only**:
- No authentication (for easy testing)
- Directly query Bugsby
- Do NOT store data
- Should be disabled in production

**Remember to re-enable auth** by uncommenting:
```go
bugsbyAPI.Use(middleware.AuthMiddleware(cfg.JWTSecret))
```

## 📊 Data Flow Summary

```
Bugsby API (Source of Truth)
    ↓
[Manager runs sync]
    ↓
Our Database (Local Cache + Custom Fields)
    ↓
[Developers query our DB]
    ↓
Fast responses with custom workflow data
```

## 🚀 Next Steps

1. **Test the sync**: Use Postman collection to sync a small release
2. **Check status**: Verify bugs are in database
3. **Query bugs**: Use `/bugs` endpoints to query local data
4. **Build workflow**: Add AI generation, approval flows, etc.
5. **Schedule re-sync**: Set up periodic sync (cron job, etc.)

