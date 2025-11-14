# Release Notes Generator - API Documentation

**Base URL:** `http://localhost:8080`  
**API Version:** `v1`  
**API Prefix:** `/api/v1`

---

## Table of Contents

1. [Authentication Flow](#authentication-flow)
2. [User Endpoints](#user-endpoints)
3. [Bug Management Endpoints](#bug-management-endpoints)
4. [Bugsby Sync Endpoints](#bugsby-sync-endpoints)
5. [Release Notes Workflow Endpoints](#release-notes-workflow-endpoints)
6. [AI-Powered Release Note Generation](#ai-powered-release-note-generation)
7. [Complete Workflow Example](#complete-workflow-example)
8. [Error Responses](#error-responses)

---

## Authentication Flow

All API endpoints (except login, refresh, and health check) require JWT authentication.

### Standard Response Format

**Success Response:**
```json
{
  "success": true,
  "data": { /* response data */ },
  "message": "Optional success message"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "error_code",
  "message": "Human-readable error message"
}
```

---

## User Endpoints

### 1. Login

**Endpoint:** `POST /api/v1/user/login`  
**Authentication:** None (Public)  
**Description:** Login with email and role (no password required for hackathon)

**Request Body:**
```json
{
  "email": "developer@arista.com",
  "role": "developer"
}
```

**Validation:**
- `email`: Required, must be valid email format
- `role`: Required, must be either `"developer"` or `"manager"`

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "developer@arista.com",
      "role": "developer",
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  }
}
```

**Token Details:**
- `token`: Access token (expires in 24 hours)
- `refresh_token`: Refresh token (expires in 7 days)

**Usage:**
Include the access token in all subsequent requests:
```
Authorization: Bearer <token>
```

---

### 2. Refresh Token

**Endpoint:** `POST /api/v1/user/refresh`  
**Authentication:** None (Public)  
**Description:** Get new access and refresh tokens

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "token": "new_access_token...",
    "refresh_token": "new_refresh_token..."
  }
}
```

---

### 3. Get Current User

**Endpoint:** `GET /api/v1/user/me`  
**Authentication:** Required  
**Description:** Get current logged-in user details

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "developer@arista.com",
    "role": "developer",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

---

### 4. Logout

**Endpoint:** `POST /api/v1/user/logout`  
**Authentication:** None (Public)  
**Description:** Logout user (client-side token removal)

**Success Response (200):**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

### 5. Delete Current User

**Endpoint:** `DELETE /api/v1/user/me`  
**Authentication:** Required  
**Description:** Delete current user account

**Success Response (200):**
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

---

## Bug Management Endpoints

### 6. List Bugs

**Endpoint:** `GET /api/v1/bugs`
**Authentication:** Required
**Description:** Get paginated list of bugs with optional filters

**Query Parameters:**
- `release` (string): Filter by release name (e.g., "wifi.nainital")
- `status` (array): Filter by status (e.g., "open", "resolved")
- `assigned_to` (UUID): Filter by assigned user ID
- `manager_id` (UUID): Filter by manager ID
- `severity` (array): Filter by severity (e.g., "critical", "major")
- `bug_type` (array): Filter by bug type
- `component` (string): Filter by component
- `has_release_note` (boolean): Filter bugs with/without release notes
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `sort_by` (string): Sort field (default: "created_at")
- `sort_order` (string): "asc" or "desc" (default: "desc")

**Example Request:**
```
GET /api/v1/bugs?release=wifi.nainital&has_release_note=false&page=1&limit=20
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "bugs": [
      {
        "id": "fd108a72-a6d4-4b85-9266-54286309421f",
        "created_at": "2025-01-15T10:30:00Z",
        "updated_at": "2025-01-15T10:30:00Z",
        "bugsby_id": "1184600",
        "bugsby_url": "https://bugs.arista.io/bugzilla/show_bug.cgi?id=1184600",
        "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments",
        "description": "Bug description text...",
        "severity": "major",
        "priority": "P2",
        "bug_type": "defect",
        "cve_number": null,
        "assigned_to": "550e8400-e29b-41d4-a716-446655440000",
        "manager_id": null,
        "release": "wifi.nainital",
        "component": "wifi-mwm",
        "status": "resolved",
        "last_synced_at": "2025-01-15T10:30:00Z",
        "sync_status": "synced",
        "release_note": null
      }
    ],
    "total": 18,
    "page": 1,
    "limit": 20,
    "total_pages": 1
  }
}
```

---

### 7. Get Bug by ID

**Endpoint:** `GET /api/v1/bugs/:id`
**Authentication:** Required
**Description:** Get detailed information about a specific bug

**Path Parameters:**
- `id` (UUID): Bug ID

**Example Request:**
```
GET /api/v1/bugs/fd108a72-a6d4-4b85-9266-54286309421f
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "bugsby_id": "1184600",
    "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments",
    "description": "Bug description...",
    "severity": "major",
    "priority": "P2",
    "release": "wifi.nainital",
    "status": "resolved",
    "release_note": {
      "id": "abc123...",
      "content": "Fixed PCAP fragment cleanup...",
      "status": "ai_generated",
      "version": 1
    }
  }
}
```

---

### 8. Update Bug (Manager Only)

**Endpoint:** `PATCH /api/v1/bugs/:id`
**Authentication:** Required (Manager role)
**Description:** Update bug details

**Path Parameters:**
- `id` (UUID): Bug ID

**Request Body:**
```json
{
  "status": "resolved",
  "assigned_to": "550e8400-e29b-41d4-a716-446655440000",
  "manager_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "status": "resolved",
    "assigned_to": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

### 9. Delete Bug (Manager Only)

**Endpoint:** `DELETE /api/v1/bugs/:id`
**Authentication:** Required (Manager role)
**Description:** Delete a bug

**Success Response (200):**
```json
{
  "success": true,
  "message": "Bug deleted successfully"
}
```

---

## Bugsby Sync Endpoints

All sync endpoints require **Manager role**.

⚡ **NEW: Auto-Generation Feature** - All sync endpoints now automatically generate AI release notes in the background after syncing bugs. This means:
- ✅ Sync returns immediately (non-blocking)
- ✅ AI generation runs asynchronously in background
- ✅ Release notes are ready within seconds after sync
- ✅ Frontend can immediately fetch release notes without additional API calls
- ✅ Skips bugs that already have release notes (no duplicates)
- ✅ Graceful error handling (logs failures, continues with other bugs)

### 10. Sync Bug by Bugsby ID

**Endpoint:** `POST /api/v1/bugsby/sync/:bugsby_id`
**Authentication:** Required (Manager only)
**Description:** Sync a single bug from Bugsby by its Bugsby ID. **Automatically generates AI release note in background.**

**Path Parameters:**
- `bugsby_id` (int): Bugsby bug ID (e.g., 1184600)

**Example Request:**
```
POST /api/v1/bugsby/sync/1184600
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "total_fetched": 1,
    "new_bugs": 1,
    "updated_bugs": 0,
    "failed_bugs": 0,
    "synced_at": "2025-01-15T10:30:00Z",
    "synced_bug_ids": ["fd108a72-a6d4-4b85-9266-54286309421f"],
    "errors": []
  }
}
```

**What Happens After Sync:**
1. ✅ Sync response returns immediately
2. 🤖 AI generation starts in background (async)
3. ⏱️ Release note ready within 2-5 seconds
4. 📋 Frontend can fetch release notes using `GET /api/v1/release-notes`

**Server Logs (Background Process):**
```
🤖 Starting background AI release note generation bug_count=1 source=SyncBugByID
✅ Successfully auto-generated AI release note bug_id=fd108a72-... source=SyncBugByID
🎉 Background AI release note generation completed total=1 success=1 skipped=0 failed=0
```

---

### 11. Sync Bugs by Custom Query

**Endpoint:** `POST /api/v1/bugsby/sync-by-query`
**Authentication:** Required (Manager only)
**Description:** Sync bugs using a custom Bugsby query. **Automatically generates AI release notes for all synced bugs in background.**

**Request Body:**
```json
{
  "query": "blocks==wifi.nainital",
  "limit": 25
}
```

**Fields:**
- `query` (string, required): Bugsby query string
- `limit` (int, optional): Maximum number of bugs to sync (default: 25, max: 100)

**Query Examples:**
- `"blocks==wifi.nainital"` - Get bugs blocking wifi.nainital release
- `"assignee==user@arista.com"` - Get bugs assigned to user
- `"release==wifi.nainital AND severity==critical"` - Complex query

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "total_fetched": 25,
    "new_bugs": 20,
    "updated_bugs": 5,
    "failed_bugs": 0,
    "synced_at": "2025-01-15T10:30:00Z",
    "synced_bug_ids": [
      "fd108a72-a6d4-4b85-9266-54286309421f",
      "abc123-def456-...",
      "..."
    ],
    "errors": []
  }
}
```

**What Happens After Sync:**
1. ✅ Sync response returns immediately with synced bug IDs
2. 🤖 AI generation starts for all 25 bugs in background (async)
3. ⏱️ Release notes ready within 30-60 seconds (depending on bug count)
4. 📋 Frontend can poll `GET /api/v1/release-notes` to see notes as they appear
5. ⏭️ Skips bugs that already have release notes

**Server Logs (Background Process):**
```
🤖 Starting background AI release note generation bug_count=25 source=SyncByQuery
✅ Successfully auto-generated AI release note bug_id=fd108a72-... source=SyncByQuery
✅ Successfully auto-generated AI release note bug_id=abc123-... source=SyncByQuery
...
🎉 Background AI release note generation completed total=25 success=23 skipped=2 failed=0
```

**Note:** Default limit changed from 100 to 25 for faster demo/testing. Increase limit for production use.

---

### 12. Get Sync Status

**Endpoint:** `GET /api/v1/bugsby/status?release=wifi.nainital`
**Authentication:** Required (Manager only)
**Description:** Get sync status for a release

**Query Parameters:**
- `release` (string): Release name

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "release": "wifi.nainital",
    "total_bugs": 18,
    "synced_bugs": 18,
    "pending_bugs": 0,
    "failed_bugs": 0,
    "last_synced_at": "2025-01-15T10:30:00Z"
  }
}
```

---

## Release Notes Workflow Endpoints

### 13. Get Release Notes (Kanban View)

**Endpoint:** `GET /api/v1/release-notes`
**Authentication:** Required
**Description:** Get bugs WITH release notes, filtered by status for Kanban board view. This endpoint supports filtering by developer assignment, manager assignment, status, release, and component.

**Query Parameters:**
- `assigned_to_me` (boolean): Filter by bugs assigned to current user
- `manager_id` (boolean): Filter by bugs managed by current user (use `true` for current user)
- `status` (array): Filter by release note status (`ai_generated`, `dev_approved`, `manager_approved`, `rejected`, `draft`)
- `release` (string): Filter by release
- `component` (string): Filter by component
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20)
- `sort_by` (string): Sort field
- `sort_order` (string): "asc" or "desc"

**Example Requests:**

**Developer Kanban View - Column 1 (AI Generated):**
```
GET /api/v1/release-notes?assigned_to_me=true&status=ai_generated&page=1&limit=20
Authorization: Bearer <token>
```

**Developer Kanban View - Column 2 (Dev Approved):**
```
GET /api/v1/release-notes?assigned_to_me=true&status=dev_approved&page=1&limit=20
Authorization: Bearer <token>
```

**Developer Kanban View - Column 3 (Manager Approved):**
```
GET /api/v1/release-notes?assigned_to_me=true&status=manager_approved&page=1&limit=20
Authorization: Bearer <token>
```

**Manager Kanban View - Needs Approval:**
```
GET /api/v1/release-notes?manager_id=true&status=dev_approved&page=1&limit=20
Authorization: Bearer <token>
```

**Filter by Release:**
```
GET /api/v1/release-notes?release=wifi.nainital&status=ai_generated
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "release_notes": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
        "content": "Fixed an issue where WM was not deleting stale PCAP fragments...",
        "category": "bug_fix",
        "status": "ai_generated",
        "generated_by": "openai",
        "created_by_id": "550e8400-e29b-41d4-a716-446655440001",
        "approved_by_dev_id": null,
        "approved_by_mgr_id": null,
        "created_at": "2025-01-15T10:30:00Z",
        "updated_at": "2025-01-15T10:30:00Z",
        "bug": {
          "id": "fd108a72-a6d4-4b85-9266-54286309421f",
          "bugsby_id": "1184600",
          "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments",
          "severity": "major",
          "priority": "P2",
          "release": "wifi.nainital",
          "component": "wifi-manager",
          "assigned_to": "550e8400-e29b-41d4-a716-446655440001",
          "manager_id": "550e8400-e29b-41d4-a716-446655440002"
        }
      }
    ],
    "total": 15,
    "page": 1,
    "limit": 20,
    "total_pages": 1
  }
}
```

**Use Cases:**

1. **Developer Kanban Board:**
   - Column 1: `GET /api/v1/release-notes?assigned_to_me=true&status=ai_generated`
   - Column 2: `GET /api/v1/release-notes?assigned_to_me=true&status=dev_approved`
   - Column 3: `GET /api/v1/release-notes?assigned_to_me=true&status=manager_approved`

2. **Manager Kanban Board:**
   - Column 1: `GET /api/v1/release-notes?manager_id=true&status=dev_approved` (Needs my approval)
   - Column 2: `GET /api/v1/release-notes?manager_id=true&status=manager_approved` (I approved)
   - Column 3: `GET /api/v1/release-notes?manager_id=true&status=rejected` (I rejected)

3. **Optimistic UI Updates:**
   - Fetch each column separately
   - When user approves a bug, update the status via `PUT /api/v1/release-notes/:id`
   - Optimistically move the card in UI
   - Optionally refetch only the affected columns

**Error Responses:**
- `401 Unauthorized`: Missing or invalid token
- `400 Bad Request`: Invalid query parameters

---

### 14. Get Pending Bugs (Bugs Without Release Notes)

**Endpoint:** `GET /api/v1/release-notes/pending`
**Authentication:** Required
**Description:** Get list of bugs that don't have release notes yet

**Query Parameters:**
- `assigned_to_me` (boolean): Filter by current user's assignments
- `release` (string): Filter by release
- `status` (array): Filter by status
- `severity` (array): Filter by severity
- `component` (string): Filter by component
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20)
- `sort_by` (string): Sort field
- `sort_order` (string): "asc" or "desc"

**Example Request:**
```
GET /api/v1/release-notes/pending?assigned_to_me=true&release=wifi.nainital&page=1&limit=20
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "bugs": [
      {
        "id": "fd108a72-a6d4-4b85-9266-54286309421f",
        "bugsby_id": "1184600",
        "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments",
        "severity": "major",
        "priority": "P2",
        "release": "wifi.nainital",
        "status": "resolved",
        "release_note": null
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 20,
    "total_pages": 1
  }
}
```

---

### 15. Get Bug Context with Commit Information

**Endpoint:** `GET /api/v1/release-notes/bug/:bug_id/context`
**Authentication:** Required
**Description:** Get bug details with parsed commit information from Gerrit comments

**Path Parameters:**
- `bug_id` (UUID): Bug ID

**Example Request:**
```
GET /api/v1/release-notes/bug/fd108a72-a6d4-4b85-9266-54286309421f/context
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "bug": {
      "id": "fd108a72-a6d4-4b85-9266-54286309421f",
      "bugsby_id": "1184600",
      "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments",
      "description": "Bug description...",
      "severity": "major",
      "priority": "P2",
      "release": "wifi.nainital",
      "status": "resolved"
    },
    "comments": [
      {
        "commit_hash": "475484",
        "gerrit_url": "https://gerrit.corp.arista.io/c/wifi-mwm/+/475484",
        "repository": "wifi-mwm.git",
        "branch": "master",
        "title": "server/packet_capture: Remove stale PCAP fragments.",
        "message": "This commit addresses the issue of unmerged packet capture fragments...",
        "change_id": "I18c855353065953a894de05b22d805948791d982",
        "merged_by": "om.nikam",
        "comment_id": 26919159,
        "commented_at": "2025-07-29T11:45:54Z"
      }
    ],
    "commit_count": 1,
    "ready_for_generation": true
  }
}
```

**Response Fields:**
- `bug`: Full bug details
- `comments`: Array of parsed commit information from gerrit@arista.com comments
- `commit_count`: Number of commits found
- `ready_for_generation`: `true` if bug has commits and is ready for AI generation

**Use Case:**
This endpoint is crucial for the release notes generation workflow. It:
1. Fetches bug details from the database
2. Queries Bugsby API for comments by `gerrit@arista.com`
3. Parses commit information (Gerrit URL, commit message, change ID, etc.)
4. Returns structured data ready for AI processing

---

### 16. Generate Release Note

**Endpoint:** `POST /api/v1/release-notes/generate`
**Authentication:** Required
**Description:** Generate a release note for a bug using AI (Google Gemini 2.5 Pro) or manual content. The AI follows **AID1711 mandatory release note guidelines** to produce customer-facing, professional release notes.

**AI Generation Process:**
1. Fetches bug details (title, severity, component, description)
2. Retrieves related Gerrit commits from bug comments
3. Sends structured prompt to Google Gemini 2.5 Pro with AID1711 guidelines
4. AI returns JSON with release note, confidence score, reasoning, and alternatives
5. Stores release note with AI metadata (model, confidence)

**Request Body:**
```json
{
  "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
  "manual_content": "Optional manual content if not using AI"
}
```

**Fields:**
- `bug_id` (UUID, required): Bug ID to generate release note for
- `manual_content` (string, optional): Manual release note content. If not provided, AI will generate it automatically.

**Success Response - AI Generated (201):**
```json
{
  "success": true,
  "data": {
    "id": "abc123-def456-...",
    "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "content": "Resolved issue where the Wireless Manager failed to delete stale packet capture fragments on C-360 access points, causing storage to fill up over time",
    "version": 1,
    "generated_by": "ai",
    "ai_model": "gemini-2.5-pro",
    "ai_confidence": 0.87,
    "ai_reasoning": "Bug description clearly states the issue. Commit information provides technical context about PCAP fragment cleanup. High confidence based on clear symptom and fix details.",
    "ai_alternative_versions": "[\"Fixed packet capture fragment cleanup issue on C-360 access points\", \"Resolved stale PCAP fragment accumulation problem on C-360 APs\"]",
    "status": "ai_generated",
    "created_by_id": "550e8400-e29b-41d4-a716-446655440000",
    "approved_by_dev_id": null,
    "approved_by_mgr_id": null,
    "dev_approved_at": null,
    "mgr_approved_at": null,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

**Success Response - Manual (201):**
```json
{
  "success": true,
  "data": {
    "id": "abc123-def456-...",
    "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "content": "Custom manual release note content",
    "version": 1,
    "generated_by": "manual",
    "ai_model": null,
    "ai_confidence": null,
    "ai_reasoning": null,
    "ai_alternative_versions": null,
    "status": "draft",
    "created_by_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

**Generated By Values:**
- `ai`: Generated by Google Gemini AI
- `manual`: Manually written by developer
- `placeholder`: Fallback placeholder (when AI fails)

**Status Values:**
- `draft`: Manually created draft
- `ai_generated`: Generated by AI (needs developer review)
- `dev_approved`: Approved by developer (ready for manager)
- `mgr_approved`: Approved by manager (final, ready for release)
- `rejected`: Rejected by manager

**AI Confidence Score:**
- Range: `0.0` to `1.0` (capped at `0.95`)
- `0.3-0.5`: Low confidence (limited information)
- `0.5-0.7`: Medium confidence (basic information available)
- `0.7-0.85`: High confidence (detailed bug + commits)
- `0.85-0.95`: Very high confidence (comprehensive information)

**Error Response - Bug Not Found (404):**
```json
{
  "success": false,
  "error": "not_found",
  "message": "Bug not found"
}
```

**Error Response - AI Generation Failed (500):**
```json
{
  "success": false,
  "error": "ai_generation_failed",
  "message": "AI generation failed: timeout"
}
```

**Note:** If AI generation fails, the system automatically falls back to placeholder generation with `generated_by: "placeholder"` and `status: "draft"`.

---

### 17. Get Release Note by Bug ID

**Endpoint:** `GET /api/v1/release-notes/bug/:bug_id`
**Authentication:** Required
**Description:** Get the release note for a specific bug

**Path Parameters:**
- `bug_id` (UUID): Bug ID

**Example Request:**
```
GET /api/v1/release-notes/bug/fd108a72-a6d4-4b85-9266-54286309421f
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "abc123-def456-...",
    "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "content": "Fixed issue with stale PCAP fragments...",
    "version": 2,
    "generated_by": "ai",
    "status": "dev_approved",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T11:00:00Z",
    "bug": {
      "id": "fd108a72-a6d4-4b85-9266-54286309421f",
      "bugsby_id": "1184600",
      "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments"
    }
  }
}
```

**Error Response (404):**
```json
{
  "success": false,
  "error": "not_found",
  "message": "Release note not found for this bug"
}
```

---

### 18. Update Release Note

**Endpoint:** `PUT /api/v1/release-notes/:id`
**Authentication:** Required
**Description:** Update release note content or status

**Path Parameters:**
- `id` (UUID): Release note ID

**Request Body:**
```json
{
  "content": "Updated release note content...",
  "status": "dev_approved"
}
```

**Fields:**
- `content` (string, required): Updated release note content
- `status` (string, optional): New status (draft, ai_generated, dev_approved, mgr_approved, rejected)

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "abc123-def456-...",
    "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "content": "Updated release note content...",
    "version": 3,
    "status": "dev_approved",
    "updated_at": "2025-01-15T12:00:00Z"
  }
}
```

**Note:** Each update increments the version number for audit trail.

---

### 19. Bulk Generate Release Notes

**Endpoint:** `POST /api/v1/release-notes/bulk-generate`
**Authentication:** Required
**Description:** Generate release notes for multiple bugs at once

**Request Body:**
```json
{
  "bug_ids": [
    "fd108a72-a6d4-4b85-9266-54286309421f",
    "abc123-def456-...",
    "xyz789-uvw012-..."
  ],
  "release": "wifi.nainital"
}
```

**Fields:**
- `bug_ids` (array of UUIDs): List of bug IDs to generate release notes for
- `release` (string, optional): If provided, generates for all bugs in the release

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "total": 3,
    "generated": 2,
    "failed": 1,
    "results": [
      {
        "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
        "release_note_id": "abc123-...",
        "status": "success",
        "error": null
      },
      {
        "bug_id": "abc123-def456-...",
        "release_note_id": "def456-...",
        "status": "success",
        "error": null
      },
      {
        "bug_id": "xyz789-uvw012-...",
        "release_note_id": null,
        "status": "failed",
        "error": "No commit information found"
      }
    ]
  }
}
```

---

### 20. Approve/Reject Release Note (Manager Only)

**Endpoint:** `POST /api/v1/release-notes/:id/approve`
**Authentication:** Required (Manager role)
**Description:** Approve or reject a release note

**Path Parameters:**
- `id` (UUID): Release note ID

**Request Body:**
```json
{
  "action": "approve",
  "feedback": "Looks good, approved!"
}
```

**Fields:**
- `action` (string, required): Either `"approve"` or `"reject"`
- `feedback` (string, optional): Manager's feedback

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "abc123-def456-...",
    "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "status": "mgr_approved",
    "approved_by_mgr_id": "660e8400-e29b-41d4-a716-446655440001",
    "mgr_approved_at": "2025-01-15T13:00:00Z"
  }
}
```

**Reject Example:**
```json
{
  "action": "reject",
  "feedback": "Please add more technical details about the fix"
}
```

**Reject Response:**
```json
{
  "success": true,
  "data": {
    "id": "abc123-def456-...",
    "status": "rejected",
    "approved_by_mgr_id": "660e8400-e29b-41d4-a716-446655440001",
    "mgr_approved_at": "2025-01-15T13:00:00Z"
  }
}
```

---

## AI-Powered Release Note Generation

### Overview

The system uses **Google Gemini 2.5 Pro** (via Vertex AI) to automatically generate professional release notes from bug information and Gerrit commits. The AI follows **AID1711 mandatory release note guidelines** to ensure customer-facing, compliant release notes.

---

### AID1711 Release Note Guidelines

The AI is trained to follow these mandatory organizational guidelines:

#### **Audience & Focus**
- ✅ Write for **customers and field teams**, NOT internal engineering
- ✅ Focus on **customer-visible symptoms**, not internal fix details
- ✅ Answer: "What will customers notice?" and "What conditions trigger this?"

#### **Format & Content**
- ✅ Brief (1-2 sentences)
- ✅ MUST include: when problem occurs (required configuration) + impact
- ✅ Use past tense (Resolved, Fixed, Corrected)
- ✅ Include workarounds if they exist (do NOT say "no known workarounds")

#### **Avoid Internal Jargon**
- ❌ NO internal architectural names (e.g., "HW LAG", "SW LAG")
- ❌ NO codenames (e.g., Jericho, Sand, Broadcom chip numbers)
- ❌ NO bug IDs in the note text
- ❌ NO specific EOS version numbers in the note text
- ❌ AVOID: crash, segfault, assert, race condition

#### **Agent/System Language**
- ✅ If agent dies: "the [Agent Name] agent can restart unexpectedly"
- ✅ If system goes down: "the system can restart unexpectedly" or "reset unexpectedly"

#### **Spelling & Capitalization**
- ✅ Use American English spelling
- ✅ Protocol names/acronyms in ALL CAPS (BGP, OSPF, MLAG, VXLAN)
- ✅ Specific spellings: "running config", "route map", "next hop", "port channel" (not hyphenated)
- ✅ Use "workaround" as a noun

#### **Likelihood**
- ❌ Do NOT comment on likelihood (avoid "rare", "infrequently", etc.)

---

### AI Response Format

The AI returns a structured JSON response with multiple components:

```json
{
  "release_note": "Resolved issue preventing packet capture on C-360 APs in Dual 5G mode",
  "confidence": 0.85,
  "reasoning": "Bug affects C-360 AP packet capture in specific mode. Combined info from 3 commits.",
  "alternative_versions": [
    "Fixed packet capture for C-360 APs in Dual 5G mode",
    "Resolved packet capture failure on C-360 access points"
  ]
}
```

**Fields:**
- `release_note`: Primary release note following AID1711 guidelines
- `confidence`: AI's self-assessed confidence score (0.0-1.0)
- `reasoning`: Explanation of why the AI gave this confidence score
- `alternative_versions`: 2 alternative phrasings for developer choice

---

### AI Generation Process

**Step-by-step process:**

1. **Fetch Bug Context**
   - Retrieves bug details (title, severity, component, description)
   - Fetches Gerrit commits from bug comments (posted by gerrit@arista.com)
   - Parses commit information (subject, message, change ID)

2. **Build Prompt**
   - Includes AID1711 guidelines in prompt
   - Adds bug information (title, severity, component, description)
   - Adds commit details (if available)
   - Specifies JSON output format with example

3. **Call Gemini API**
   - Sends prompt to Google Gemini 2.5 Pro via Vertex AI
   - Uses retry logic (3 attempts with exponential backoff)
   - Timeout: 60 seconds per attempt
   - Generation parameters:
     - Temperature: 0.7 (balanced creativity)
     - Max tokens: 1000
     - TopP: 0.95
     - TopK: 40

4. **Parse Response**
   - Extracts JSON from AI response
   - Validates confidence score (0.0-1.0 range)
   - Applies additional confidence adjustments:
     - +0.05 if commits available
     - +0.05 if detailed bug description
     - +0.05 if well-formed release note
   - Caps confidence at 0.95 (never 100% certain)

5. **Store Release Note**
   - Saves release note content to database
   - Stores AI metadata:
     - `ai_model`: "gemini-2.5-pro"
     - `ai_confidence`: Adjusted confidence score (0.0-1.0)
     - `ai_reasoning`: AI's explanation for the confidence score
     - `ai_alternative_versions`: JSON array of 2-3 alternative phrasings
     - `generated_by`: "ai"
     - `status`: "ai_generated"

6. **Fallback on Failure**
   - If AI fails, automatically generates placeholder
   - Sets `generated_by: "placeholder"` and `status: "draft"`
   - Logs error for debugging

---

### AI Metadata Fields

When a release note is generated by AI (`generated_by: "ai"`), the following metadata fields are populated:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| **`ai_model`** | string | AI model used for generation | `"gemini-2.5-pro"` |
| **`ai_confidence`** | float | Confidence score (0.0-1.0) | `0.87` |
| **`ai_reasoning`** | string | AI's explanation for confidence score | `"Bug description clearly states the issue..."` |
| **`ai_alternative_versions`** | string (JSON array) | 2-3 alternative phrasings | `"[\"Alternative 1\", \"Alternative 2\"]"` |

**Example AI Metadata:**
```json
{
  "ai_model": "gemini-2.5-pro",
  "ai_confidence": 0.87,
  "ai_reasoning": "Bug description clearly states the issue. Commit information provides technical context about PCAP fragment cleanup. High confidence based on clear symptom and fix details.",
  "ai_alternative_versions": "[\"Fixed packet capture fragment cleanup issue on C-360 access points\", \"Resolved stale PCAP fragment accumulation problem on C-360 APs\"]"
}
```

**Parsing Alternative Versions:**
The `ai_alternative_versions` field is a JSON-encoded string array. To use it in your frontend:

```javascript
// Parse the JSON string
const alternatives = JSON.parse(releaseNote.ai_alternative_versions);
// alternatives = ["Alternative 1", "Alternative 2"]

// Display alternatives to user
alternatives.forEach((alt, index) => {
  console.log(`Option ${index + 1}: ${alt}`);
});
```

---

### Confidence Score Interpretation

| Score Range | Interpretation | Typical Scenario |
|-------------|----------------|------------------|
| 0.3 - 0.5 | Low confidence | Limited bug information, no commits |
| 0.5 - 0.7 | Medium confidence | Basic bug info, few or no commits |
| 0.7 - 0.85 | High confidence | Detailed bug description + commits |
| 0.85 - 0.95 | Very high confidence | Comprehensive info, multiple commits, well-formed note |

**Factors that increase confidence:**
- ✅ Multiple Gerrit commits with detailed messages
- ✅ Detailed bug description (>100 characters)
- ✅ Well-formed release note (proper length, structure)
- ✅ Clear customer impact described in bug

**Factors that decrease confidence:**
- ❌ No commit information available
- ❌ Vague or missing bug description
- ❌ Unclear customer impact
- ❌ Internal jargon in bug title/description

---

### Configuration

**Environment Variables:**
```env
GCP_PROJECT_ID=anetorg-kduda-amod
GCP_LOCATION=us-central1
GEMINI_MODEL=gemini-2.5-pro
```

**AI Service Initialization:**
- If GCP credentials are missing, AI service is disabled
- System falls back to placeholder generation
- No errors thrown - graceful degradation

---

### Example AI-Generated Release Notes

**Example 1: High Confidence (0.87)**
```
Input:
  Bug: [Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments
  Severity: major
  Commits: 3 commits with fix details

Output:
  "Resolved issue where the Wireless Manager failed to delete stale
   packet capture fragments on C-360 access points, causing storage
   to fill up over time"
```

**Example 2: Medium Confidence (0.65)**
```
Input:
  Bug: BGP session flaps on route update
  Severity: minor
  Commits: None

Output:
  "Fixed issue where BGP sessions could restart unexpectedly when
   receiving route updates with specific attributes"
```

**Example 3: Low Confidence (0.45)**
```
Input:
  Bug: System crash
  Severity: critical
  Commits: None
  Description: (empty)

Output:
  "Resolved issue where the system could restart unexpectedly under
   certain conditions"
```

---

## Complete Workflow Example

Here's a complete workflow from login to generating release notes:

---

## Workflow Option 1: Auto-Generation (Recommended - New Feature!)

This is the **recommended workflow** that leverages the new auto-generation feature during sync.

### Step 1: Login as Manager

```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "manager@arista.com",
    "role": "manager"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": { "id": "...", "email": "manager@arista.com", "role": "manager" }
  }
}
```

Save the token as `MANAGER_TOKEN`.

---

### Step 2: Sync Bugs (AI Release Notes Auto-Generated!)

```bash
curl -X POST http://localhost:8080/api/v1/bugsby/sync-by-query \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "blocks==wifi.nainital",
    "limit": 25
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_fetched": 25,
    "new_bugs": 25,
    "updated_bugs": 0,
    "failed_bugs": 0,
    "synced_at": "2025-01-15T10:30:00Z",
    "synced_bug_ids": ["fd108a72-...", "abc123-...", "..."]
  }
}
```

**What Happens:**
- ✅ Bugs synced to database
- 🤖 AI generation starts in background (async)
- ⏱️ Release notes ready within 30-60 seconds

---

### Step 3: Wait a Few Seconds (Optional)

Wait 30-60 seconds for AI generation to complete. You can monitor server logs:

```
🤖 Starting background AI release note generation bug_count=25 source=SyncByQuery
✅ Successfully auto-generated AI release note bug_id=fd108a72-...
✅ Successfully auto-generated AI release note bug_id=abc123-...
...
🎉 Background AI release note generation completed total=25 success=25 skipped=0 failed=0
```

---

### Step 4: Login as Developer

```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@arista.com",
    "role": "developer"
  }'
```

Save the token as `DEV_TOKEN`.

---

### Step 5: View AI-Generated Release Notes (Kanban View)

```bash
curl -X GET "http://localhost:8080/api/v1/release-notes?assigned_to_me=true&status=ai_generated" \
  -H "Authorization: Bearer $DEV_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "release_notes": [
      {
        "id": "abc123-...",
        "bug_id": "fd108a72-...",
        "content": "Resolved issue where the Wireless Manager failed to delete stale packet capture fragments...",
        "status": "ai_generated",
        "generated_by": "ai",
        "ai_model": "gemini-2.5-pro",
        "ai_confidence": 0.87,
        "ai_reasoning": "Bug description clearly states the issue. Commit information provides technical context. High confidence based on clear symptom and fix.",
        "ai_alternative_versions": "[\"Fixed packet capture fragment cleanup issue\", \"Resolved stale PCAP fragment accumulation problem\"]",
        "bug": {
          "bugsby_id": "1184600",
          "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments",
          "severity": "major"
        }
      }
    ],
    "total": 25
  }
}
```

**Benefits:**
- ✅ Release notes already generated and ready!
- ✅ No need to call `/api/v1/release-notes/generate` for each bug
- ✅ Faster workflow for developers
- ✅ Better UX - notes ready on home screen

---

### Step 6: Developer Reviews and Approves

```bash
curl -X PUT http://localhost:8080/api/v1/release-notes/abc123-... \
  -H "Authorization: Bearer $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Resolved issue where the Wireless Manager failed to delete stale packet capture fragments...",
    "status": "dev_approved"
  }'
```

---

### Step 7: Manager Approves

```bash
curl -X POST http://localhost:8080/api/v1/release-notes/abc123-.../approve \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "approve",
    "feedback": "Looks good!"
  }'
```

---

## Workflow Option 2: Manual Generation (Legacy)

This is the **legacy workflow** for manual generation (still supported).

### Step 1: Login

```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@arista.com",
    "role": "developer"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": { "id": "...", "email": "developer@arista.com", "role": "developer" }
  }
}
```

Save the token for subsequent requests.

---

### Step 2: Get Pending Bugs (Bugs Without Release Notes)

```bash
curl -X GET "http://localhost:8080/api/v1/release-notes/pending?assigned_to_me=true&limit=20" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "bugs": [
      {
        "id": "fd108a72-a6d4-4b85-9266-54286309421f",
        "bugsby_id": "1184600",
        "title": "[Systest][SWAT-Wifi] WM is not deleting stale PCAP fragments",
        "release_note": null
      }
    ],
    "total": 5
  }
}
```

---

### Step 3: Get Bug Context with Commit Information

```bash
curl -X GET "http://localhost:8080/api/v1/release-notes/bug/fd108a72-a6d4-4b85-9266-54286309421f/context" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "bug": { "id": "...", "title": "..." },
    "comments": [
      {
        "commit_hash": "475484",
        "gerrit_url": "https://gerrit.corp.arista.io/c/wifi-mwm/+/475484",
        "title": "server/packet_capture: Remove stale PCAP fragments.",
        "message": "This commit addresses..."
      }
    ],
    "commit_count": 1,
    "ready_for_generation": true
  }
}
```

---

### Step 4: Generate Release Note Manually

```bash
curl -X POST http://localhost:8080/api/v1/release-notes/generate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "abc123-...",
    "content": "Fixed issue with stale PCAP fragments...",
    "status": "ai_generated",
    "version": 1
  }
}
```

---

### Step 5: Update Release Note (if needed)

```bash
curl -X PUT http://localhost:8080/api/v1/release-notes/abc123-... \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Improved release note content...",
    "status": "dev_approved"
  }'
```

---

### Step 6: Manager Approves (Manager Role Required)

```bash
curl -X POST http://localhost:8080/api/v1/release-notes/abc123-.../approve \
  -H "Authorization: Bearer <manager_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "approve",
    "feedback": "Looks good!"
  }'
```

---

## Error Responses

### Common Error Codes

**401 Unauthorized:**
```json
{
  "success": false,
  "error": "unauthorized",
  "message": "Missing or invalid authentication token"
}
```

**403 Forbidden:**
```json
{
  "success": false,
  "error": "forbidden",
  "message": "Insufficient permissions. Manager role required."
}
```

**404 Not Found:**
```json
{
  "success": false,
  "error": "not_found",
  "message": "Bug not found"
}
```

**400 Bad Request:**
```json
{
  "success": false,
  "error": "validation_error",
  "message": "Invalid request body: email is required"
}
```

**500 Internal Server Error:**
```json
{
  "success": false,
  "error": "internal_error",
  "message": "An unexpected error occurred"
}
```

---

## Additional Notes

### Authentication Headers

All authenticated endpoints require:
```
Authorization: Bearer <access_token>
```

### Content Type

For POST/PUT/PATCH requests:
```
Content-Type: application/json
```

### Pagination

Default pagination values:
- `page`: 1
- `limit`: 20 (max: 100)

### Date Format

All timestamps use ISO 8601 format:
```
2025-01-15T10:30:00Z
```

### UUID Format

All IDs use UUID v4 format:
```
fd108a72-a6d4-4b85-9266-54286309421f
```

---

## Health Check

**Endpoint:** `GET /health`
**Authentication:** None
**Description:** Check if the API is running

**Response:**
```json
{
  "status": "ok",
  "service": "release-notes-generator",
  "version": "1.0.0"
}
```

---

---

## Summary of Key Changes

### ⚡ New Features (v1.2.0)

#### **1. AI Reasoning and Alternative Versions (v1.2.0)**

**What Changed:**
- Added `ai_reasoning` field - AI explains why it gave a specific confidence score
- Added `ai_alternative_versions` field - AI provides 2-3 alternative phrasings
- Increased token limit from 1000 to 4096 to prevent response truncation
- All AI-generated release notes now include complete metadata

**New Fields:**
```json
{
  "ai_reasoning": "Bug description clearly states the issue. Commit information provides technical context...",
  "ai_alternative_versions": "[\"Alternative 1\", \"Alternative 2\", \"Alternative 3\"]"
}
```

**Benefits:**
- ✅ **Transparency** - Understand why AI gave a specific confidence score
- ✅ **Flexibility** - Choose from multiple phrasings
- ✅ **Better UX** - Developers can pick the best version or mix-and-match
- ✅ **Learning** - See AI's reasoning to improve future manual edits

---

#### **2. Auto-Generation During Sync (v1.1.0)**

**What Changed:**
- All sync endpoints (`/api/v1/bugsby/sync/:bugsby_id`, `/api/v1/bugsby/sync-by-query`) now automatically generate AI release notes in the background
- Sync operations return immediately (non-blocking)
- AI generation runs asynchronously using goroutines
- Release notes are ready within seconds after sync completes

**Benefits:**
- ✅ **Fewer API calls** - Frontend only needs to call sync endpoint
- ✅ **Better UX** - Notes ready immediately on home screen
- ✅ **Faster workflow** - No waiting for manual generation
- ✅ **Demo-friendly** - Default limit changed from 100 to 25 bugs

**Migration Guide:**

**Old Workflow (v1.0.0):**
```bash
# Step 1: Sync bugs
POST /api/v1/bugsby/sync-by-query

# Step 2: Get pending bugs
GET /api/v1/release-notes/pending

# Step 3: Generate release note for each bug (slow!)
POST /api/v1/release-notes/generate (for each bug)

# Step 4: Fetch release notes
GET /api/v1/release-notes
```

**New Workflow (v1.1.0):**
```bash
# Step 1: Sync bugs (AI generation happens automatically in background!)
POST /api/v1/bugsby/sync-by-query

# Step 2: Wait a few seconds (optional)

# Step 3: Fetch release notes (already generated!)
GET /api/v1/release-notes
```

**Backward Compatibility:**
- ✅ All existing endpoints still work
- ✅ Manual generation endpoint (`POST /api/v1/release-notes/generate`) still available
- ✅ No breaking changes to API contracts
- ✅ New fields (`ai_reasoning`, `ai_alternative_versions`) are nullable and optional
- ✅ Old release notes without these fields will have `null` values

---

**Last Updated:** 2025-11-14
**API Version:** 1.2.0 (AI Reasoning & Alternative Versions)

