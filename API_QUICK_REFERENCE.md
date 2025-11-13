# API Quick Reference Card

**Base URL:** `http://localhost:8080/api/v1`

---

## 🔐 Authentication

```bash
# Login
POST /user/login
Body: { "email": "dev@arista.com", "role": "developer" }
Response: { "token": "...", "refresh_token": "...", "user": {...} }

# Use token in all requests
Header: Authorization: Bearer <token>
```

---

## 📋 Main Workflow Endpoints

### 1. Get Release Notes (Kanban View)
```bash
# Developer Kanban - Column 1: AI Generated
GET /release-notes?assigned_to_me=true&status=ai_generated

# Developer Kanban - Column 2: Dev Approved
GET /release-notes?assigned_to_me=true&status=dev_approved

# Developer Kanban - Column 3: Manager Approved
GET /release-notes?assigned_to_me=true&status=manager_approved

# Manager Kanban - Needs Approval
GET /release-notes?manager_id=true&status=dev_approved
```

### 2. Get Pending Bugs
```bash
GET /release-notes/pending?assigned_to_me=true&limit=20
```

### 3. Get Bug Context (with commits)
```bash
GET /release-notes/bug/{bug_id}/context
```
**Returns:** Bug details + parsed Gerrit commits

### 4. Generate Release Note
```bash
POST /release-notes/generate
Body: { "bug_id": "uuid..." }
```

### 5. Get Release Note
```bash
GET /release-notes/bug/{bug_id}
```

### 6. Update Release Note
```bash
PUT /release-notes/{id}
Body: { "content": "...", "status": "dev_approved" }
```

### 7. Approve/Reject (Manager)
```bash
POST /release-notes/{id}/approve
Body: { "action": "approve", "feedback": "..." }
```

---

## 🐛 Bug Endpoints

```bash
# List bugs with filters
GET /bugs?release=wifi.nainital&has_release_note=false&page=1&limit=20

# Get bug by ID
GET /bugs/{id}

# Update bug (Manager only)
PATCH /bugs/{id}
Body: { "status": "resolved", "assigned_to": "uuid..." }
```

---

## 🔄 Bugsby Sync (Manager Only)

```bash
# Sync single bug
POST /bugsby/sync/{bugsby_id}

# Sync by query
POST /bugsby/sync-by-query
Body: { "query": "blocks==1229583", "limit": 100 }

# Get sync status
GET /bugsby/status?release=wifi.nainital
```

---

## 📊 Response Format

**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": "error_code",
  "message": "Human-readable message"
}
```

---

## 🎯 Status Values

**Release Note Status:**
- `draft` - Manual draft
- `ai_generated` - AI generated
- `dev_approved` - Developer approved
- `mgr_approved` - Manager approved (final)
- `rejected` - Rejected by manager

---

## 🔑 Common Query Parameters

**Pagination:**
- `page` (int, default: 1)
- `limit` (int, default: 20, max: 100)

**Sorting:**
- `sort_by` (string, default: "created_at")
- `sort_order` ("asc" | "desc", default: "desc")

**Filters:**
- `release` (string)
- `status` (array)
- `severity` (array)
- `component` (string)
- `has_release_note` (boolean)
- `assigned_to_me` (boolean)

---

## 🚀 Quick Test Commands

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@arista.com","role":"developer"}' \
  | jq -r '.data.token')

# 2. Get pending bugs
curl -X GET "http://localhost:8080/api/v1/release-notes/pending?limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq

# 3. Get bug context
BUG_ID="your-bug-uuid"
curl -X GET "http://localhost:8080/api/v1/release-notes/bug/$BUG_ID/context" \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. Generate release note
curl -X POST http://localhost:8080/api/v1/release-notes/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"bug_id\":\"$BUG_ID\"}" | jq
```

---

## 📝 Frontend Integration Checklist

- [ ] Set up axios with base URL and interceptors
- [ ] Implement login and token storage
- [ ] Add auto token refresh on 401
- [ ] Create pending bugs list page
- [ ] Create bug context viewer
- [ ] Create release note generator
- [ ] Create release note editor
- [ ] Add manager approval page (if manager role)
- [ ] Handle loading and error states
- [ ] Add pagination controls

---

## 🔗 Full Documentation

- **Complete API Reference:** `API_DOCUMENTATION.md`
- **Frontend Integration Guide:** `FRONTEND_INTEGRATION_GUIDE.md`

---

**Last Updated:** 2025-01-15

