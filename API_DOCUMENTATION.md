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
6. [Complete Workflow Example](#complete-workflow-example)
7. [Error Responses](#error-responses)

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

### 10. Sync Bug by Bugsby ID

**Endpoint:** `POST /api/v1/bugsby/sync/:bugsby_id`
**Authentication:** Required (Manager only)
**Description:** Sync a single bug from Bugsby by its Bugsby ID

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
    "errors": []
  }
}
```

---

### 11. Sync Bugs by Custom Query

**Endpoint:** `POST /api/v1/bugsby/sync-by-query`
**Authentication:** Required (Manager only)
**Description:** Sync bugs using a custom Bugsby query

**Request Body:**
```json
{
  "query": "blocks==1229583",
  "limit": 100
}
```

**Query Examples:**
- `"blocks==1229583"` - Get bugs blocking bug 1229583
- `"assignee==user@arista.com"` - Get bugs assigned to user
- `"release==wifi.nainital AND severity==critical"` - Complex query

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "total_fetched": 10,
    "new_bugs": 8,
    "updated_bugs": 2,
    "failed_bugs": 0,
    "synced_at": "2025-01-15T10:30:00Z",
    "errors": []
  }
}
```

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

### 13. Get Pending Bugs (Bugs Without Release Notes)

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

### 14. Get Bug Context with Commit Information

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

### 15. Generate Release Note

**Endpoint:** `POST /api/v1/release-notes/generate`
**Authentication:** Required
**Description:** Generate a release note for a bug (AI or manual)

**Request Body:**
```json
{
  "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
  "manual_content": "Optional manual content if not using AI"
}
```

**Fields:**
- `bug_id` (UUID, required): Bug ID to generate release note for
- `manual_content` (string, optional): Manual release note content. If not provided, AI will generate it.

**Success Response (201):**
```json
{
  "success": true,
  "data": {
    "id": "abc123-def456-...",
    "bug_id": "fd108a72-a6d4-4b85-9266-54286309421f",
    "content": "Fixed issue with stale PCAP fragments not being cleaned up after WM upgrade...",
    "version": 1,
    "generated_by": "ai",
    "ai_model": "gpt-4",
    "ai_confidence": 0.95,
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

**Status Values:**
- `draft`: Manually created draft
- `ai_generated`: Generated by AI
- `dev_approved`: Approved by developer
- `mgr_approved`: Approved by manager (final)
- `rejected`: Rejected by manager

---

### 16. Get Release Note by Bug ID

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

### 17. Update Release Note

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

### 18. Bulk Generate Release Notes

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

### 19. Approve/Reject Release Note (Manager Only)

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

## Complete Workflow Example

Here's a complete workflow from login to generating release notes:

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

### Step 4: Generate Release Note

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

**Last Updated:** 2025-01-15
**API Version:** 1.0.0

