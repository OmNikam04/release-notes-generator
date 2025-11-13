# API Quick Reference Card

## 🔐 Authentication
```bash
# Login
POST /api/v1/user/login
{"email": "om.nikam@arista.com", "role": "developer"}

# Use token in all requests
Authorization: Bearer <token>
```

---

## 🔄 Bugsby Sync (Manager Only)

### Sync All Bugs for Release
```bash
POST /api/v1/bugsby/sync
{
  "release": "wifi-ooty",
  "status": "ASSIGNED",    # optional
  "severity": "sev3"       # optional
}
```

### Sync Single Bug
```bash
POST /api/v1/bugsby/sync/1092263
```

### Check Sync Status
```bash
GET /api/v1/bugsby/status?release=wifi-ooty
```

---

## 🧪 Bugsby Testing (No Auth)

### Get Bugs by Assignee
```bash
GET /api/v1/bugsby-api/bugs/assignee/om.nikam@arista.com?limit=3
```

### Custom Query
```bash
POST /api/v1/bugsby-api/bugs/query
{
  "query": "assignee==om.nikam@arista.com AND status==ASSIGNED",
  "limit": "10"
}
```

---

## 📝 Bug Management

### List Bugs
```bash
GET /api/v1/bugs?release=wifi-ooty&status=pending&limit=10
```

### Get Bug by ID
```bash
GET /api/v1/bugs/:id
GET /api/v1/bugs/bugsby/:bugsby_id
```

### Update Bug
```bash
PUT /api/v1/bugs/:id
{"status": "dev_approved"}
```

### Delete Bug
```bash
DELETE /api/v1/bugs/:id
```

---

## 📊 Common Filters

**Bug Status** (our workflow):
- `pending` - Not processed yet
- `ai_generated` - AI created release note
- `dev_approved` - Developer approved
- `mgr_approved` - Manager approved
- `rejected` - Rejected

**Bugsby Status** (from Bugsby):
- `ASSIGNED`, `RESOLVED`, `CLOSED`, `VERIFIED`

**Severity**:
- `sev1`, `sev2`, `sev3`, `sev4`

**Bug Type**:
- `BUG`, `TASK`, `FEATURE`, `ENHANCEMENT`

---

## 🎯 Common Workflows

### 1. Manager: Sync New Release
```bash
# 1. Sync bugs
POST /bugsby/sync {"release": "wifi-ooty"}

# 2. Check status
GET /bugsby/status?release=wifi-ooty

# 3. View bugs
GET /bugs?release=wifi-ooty
```

### 2. Developer: View My Bugs
```bash
# 1. Login
POST /user/login {"email": "me@arista.com", "role": "developer"}

# 2. Get my bugs
GET /bugs?assigned_to=<my_user_id>&release=wifi-ooty

# 3. Update status
PUT /bugs/:id {"status": "dev_approved"}
```

### 3. Testing: Query Bugsby Directly
```bash
# No auth needed
GET /bugsby-api/bugs/assignee/om.nikam@arista.com?limit=5
```

---

## 🔑 Key Differences

| Endpoint | Auth | Stores Data | Use Case |
|----------|------|-------------|----------|
| `/bugsby/sync` | Manager | ✅ Yes | Setup release, periodic sync |
| `/bugsby-api` | None | ❌ No | Testing, debugging |
| `/bugs` | All users | N/A | Daily operations |

---

## 📦 Postman Import

Copy the JSON from `API_DOCUMENTATION.md` and import into Postman.

---

## 🐛 Troubleshooting

**403 Forbidden**: Check JWT token and role
**404 Not Found**: Bug not synced yet, run sync first
**500 Sync Failed**: Check Bugsby API connectivity
**Decode Failed**: Bugsby response format changed

---

## 📚 Full Documentation

- **Complete API Docs**: `API_DOCUMENTATION.md`
- **Sync System Explained**: `SYNC_SYSTEM_EXPLAINED.md`
- **Migration Guide**: `MIGRATIONS_GUIDE.md`

