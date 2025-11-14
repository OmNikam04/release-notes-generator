# Release Notes Generator API Documentation

## Overview

This API provides endpoints for managing bugs, release notes, and Bugsby integration for the Release Notes Generator system.

**Base URL**: `http://localhost:8080/api/v1`

---

## 🔐 Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

### Get JWT Token

**Endpoint**: `POST /user/login`

**Request Body**:
```json
{
  "email": "om.nikam@arista.com",
  "role": "developer"  // or "manager"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGci...",
    "refresh_token": "njRg...",
    "user": {
      "id": "uuid",
      "email": "om.nikam@arista.com",
      "role": "developer"
    }
  }
}
```

---

## 📊 Understanding the Sync System

### **Why Do We Need Sync?**

The system has **two sources of bug data**:

1. **Bugsby API** (External Source)
   - Live bug tracking system at Arista
   - Contains ALL bugs with full details
   - Always up-to-date
   - Requires API calls (slower, rate-limited)

2. **Local Database** (Our Cache)
   - Stores bugs we're working with
   - Fast queries for our application
   - Allows us to add custom fields (status, release notes, etc.)
   - Needs to be synced with Bugsby

### **What Does Sync Do?**

**Sync** = Copy bugs from Bugsby → Our Database

```
Bugsby API (Source of Truth)
         ↓
    [Sync Process]
         ↓
Our Database (Local Cache + Custom Data)
```

### **Three Sync Operations**:

1. **SyncRelease** - Sync ALL bugs for a release (e.g., "wifi-ooty")
2. **SyncBugByID** - Sync ONE specific bug by its Bugsby ID
3. **GetSyncStatus** - Check how many bugs are synced for a release

---

## 🔄 Bugsby Sync Endpoints (Manager Only)

### 1. Sync Release

Fetches ALL bugs for a specific release from Bugsby and stores them in our database.

**Endpoint**: `POST /bugsby/sync`

**Auth**: Required (Manager role only)

**Request Body**:
```json
{
  "release": "wifi-ooty",
  "status": "ASSIGNED",      // optional
  "severity": "sev3",        // optional
  "bug_type": "BUG",         // optional
  "component": "gnutls"      // optional
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "total_fetched": 150,
    "new_bugs": 120,
    "updated_bugs": 30,
    "failed_bugs": 0,
    "synced_at": "2025-11-13T16:52:00Z",
    "errors": []
  },
  "message": "Successfully synced 150 bugs for release wifi-ooty"
}
```

**What Happens**:
1. Calls Bugsby API with filters
2. For each bug:
   - Creates user accounts if they don't exist (assignee, manager)
   - Checks if bug exists in our DB (by `bugsby_id`)
   - If exists → UPDATE the bug
   - If new → CREATE the bug
3. Returns summary of sync operation

---

### 2. Sync Single Bug

Fetches ONE specific bug from Bugsby by its ID.

**Endpoint**: `POST /bugsby/sync/:bugsby_id`

**Auth**: Required (Manager role only)

**Example**: `POST /bugsby/sync/1092263`

**Response**:
```json
{
  "success": true,
  "message": "Bug synced successfully",
  "data": {
    "id": "uuid",
    "bugsby_id": "1092263",
    "title": "Remove Redundent Locationid Query Param",
    "status": "pending",
    "severity": "sev3",
    "release": "main",
    "assigned_to": "uuid",
    "last_synced_at": "2025-11-13T16:52:00Z",
    "sync_status": "synced"
  }
}
```

---

### 3. Get Sync Status

Check sync status for a release (how many bugs are synced).

**Endpoint**: `GET /bugsby/status?release=wifi-ooty`

**Auth**: Required (Manager role only)

**Response**:
```json
{
  "success": true,
  "data": {
    "release": "wifi-ooty",
    "total_bugs": 150,
    "synced_bugs": 145,
    "pending_bugs": 3,
    "failed_bugs": 2,
    "last_synced_at": "2025-11-13T16:52:00Z"
  }
}
```

---

## 🧪 Bugsby Testing Endpoints (No Auth - For Testing Only)

### 1. Get Bugs by Assignee

Fetch bugs directly from Bugsby API for a specific assignee.

**Endpoint**: `GET /bugsby-api/bugs/assignee/:email`

**Auth**: None (disabled for testing)

**Query Parameters**:
- `limit` (optional, default: 100) - Number of bugs to fetch
- `sortBy` (optional, default: "id") - Sort field
- `order` (optional, default: "asc") - Sort order

**Example**: `GET /bugsby-api/bugs/assignee/om.nikam@arista.com?limit=3`

**Response**:
```json
{
  "success": true,
  "data": {
    "bugs": [
      {
        "id": 825754,
        "title": "[Intern Project] Enhance packet Capture feature...",
        "assignee": "om.nikam@arista.com",
        "status": "RESOLVED",
        "priority": "MU (Must understand)",
        "severity": "non-escape",
        "version": "main",
        "description": "..."
      }
    ],
    "count": 3,
    "cursor": 1092263,
    "has_next": true,
    "next_link": "/v3/bugs?cursor=1092263&..."
  },
  "message": "Found 3 bugs for om.nikam@arista.com"
}
```

---

### 2. Custom Bugsby Query

Execute a custom query against Bugsby API with full control over parameters.

**Endpoint**: `POST /bugsby-api/bugs/query`

**Auth**: None (disabled for testing)

**Request Body**:
```json
{
  "query": "assignee==om.nikam@arista.com AND status==ASSIGNED",
  "limit": "10",
  "sortBy": "lastUpdateTime",
  "order": "desc",
  "source": "mysql",
  "textQueryMode": "default"
}
```

**Query Syntax Examples**:
- `assignee==om.nikam@arista.com`
- `version=="wifi-ooty" AND status=="ASSIGNED"`
- `severity=="sev1" OR severity=="sev2"`
- `status=="ASSIGNED" AND priority=="P0"`

**Response**: Same as "Get Bugs by Assignee"

---

## 📝 Bug Management Endpoints

### 1. List Bugs

Get bugs from our local database with filtering and pagination.

**Endpoint**: `GET /bugs`

**Auth**: Required

**Query Parameters**:
- `release` - Filter by release name
- `status` - Filter by status (pending, ai_generated, dev_approved, mgr_approved, rejected)
- `severity` - Filter by severity
- `bug_type` - Filter by bug type
- `assigned_to` - Filter by assignee UUID
- `manager_id` - Filter by manager UUID
- `page` (default: 1)
- `limit` (default: 20)

**Example**: `GET /bugs?release=wifi-ooty&status=pending&limit=10`

**Response**:
```json
{
  "success": true,
  "data": {
    "bugs": [
      {
        "id": "uuid",
        "bugsby_id": "1092263",
        "bugsby_url": "https://bugs.arista.io/1092263",
        "title": "Remove Redundent Locationid Query Param",
        "description": "...",
        "severity": "sev3",
        "priority": "MU (Must understand)",
        "bug_type": "BUG",
        "assigned_to": "uuid",
        "manager_id": "uuid",
        "release": "main",
        "component": "wifi-network-config",
        "status": "pending",
        "sync_status": "synced",
        "last_synced_at": "2025-11-13T16:52:00Z",
        "created_at": "2025-11-13T16:00:00Z",
        "updated_at": "2025-11-13T16:52:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 150,
      "total_pages": 15
    }
  }
}
```

---

### 2. Get Bug by ID

Get a single bug from our database by its UUID.

**Endpoint**: `GET /bugs/:id`

**Auth**: Required

**Example**: `GET /bugs/61912d22-ffa9-4c09-92d6-df3a4d3541a4`

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "61912d22-ffa9-4c09-92d6-df3a4d3541a4",
    "bugsby_id": "1092263",
    "title": "Remove Redundent Locationid Query Param",
    "status": "pending",
    "release": "main"
  }
}
```

---

### 3. Get Bug by Bugsby ID

Get a bug from our database by its Bugsby ID.

**Endpoint**: `GET /bugs/bugsby/:bugsby_id`

**Auth**: Required

**Example**: `GET /bugs/bugsby/1092263`

**Response**: Same as "Get Bug by ID"

---

### 4. Update Bug

Update a bug in our database (e.g., change status, add notes).

**Endpoint**: `PUT /bugs/:id`

**Auth**: Required

**Request Body**:
```json
{
  "status": "dev_approved",
  "description": "Updated description"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Bug updated successfully",
  "data": {
    "id": "uuid",
    "status": "dev_approved",
    "updated_at": "2025-11-13T17:00:00Z"
  }
}
```

---

### 5. Delete Bug

Delete a bug from our database.

**Endpoint**: `DELETE /bugs/:id`

**Auth**: Required

**Response**:
```json
{
  "success": true,
  "message": "Bug deleted successfully"
}
```

---

## 👥 User Endpoints

### 1. Login

**Endpoint**: `POST /user/login`

**Request Body**:
```json
{
  "email": "om.nikam@arista.com",
  "role": "developer"
}
```

**Response**: See Authentication section above

---

### 2. Refresh Token

**Endpoint**: `POST /user/refresh`

**Request Body**:
```json
{
  "refresh_token": "your_refresh_token"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "token": "new_access_token",
    "refresh_token": "new_refresh_token"
  }
}
```

---

## 📋 Postman Collection

### Import Instructions

1. Open Postman
2. Click "Import" button
3. Select "Raw text"
4. Paste the JSON below
5. Click "Import"

### Postman Collection JSON

```json
{
  "info": {
    "name": "Release Notes Generator API",
    "description": "API endpoints for Release Notes Generator with Bugsby integration",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080/api/v1",
      "type": "string"
    },
    {
      "key": "token",
      "value": "",
      "type": "string"
    }
  ],
  "item": [
    {
      "name": "Authentication",
      "item": [
        {
          "name": "Login",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "if (pm.response.code === 200) {",
                  "    var jsonData = pm.response.json();",
                  "    pm.collectionVariables.set('token', jsonData.data.token);",
                  "}"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"om.nikam@arista.com\",\n  \"role\": \"developer\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/user/login",
              "host": ["{{base_url}}"],
              "path": ["user", "login"]
            }
          }
        },
        {
          "name": "Refresh Token",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"refresh_token\": \"your_refresh_token\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/user/refresh",
              "host": ["{{base_url}}"],
              "path": ["user", "refresh"]
            }
          }
        }
      ]
    },
    {
      "name": "Bugsby Sync (Manager Only)",
      "item": [
        {
          "name": "Sync Release",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              },
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"release\": \"wifi-ooty\",\n  \"status\": \"ASSIGNED\",\n  \"severity\": \"sev3\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/bugsby/sync",
              "host": ["{{base_url}}"],
              "path": ["bugsby", "sync"]
            }
          }
        },
        {
          "name": "Sync Single Bug",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/bugsby/sync/1092263",
              "host": ["{{base_url}}"],
              "path": ["bugsby", "sync", "1092263"]
            }
          }
        },
        {
          "name": "Get Sync Status",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/bugsby/status?release=wifi-ooty",
              "host": ["{{base_url}}"],
              "path": ["bugsby", "status"],
              "query": [
                {
                  "key": "release",
                  "value": "wifi-ooty"
                }
              ]
            }
          }
        }
      ]
    },
    {
      "name": "Bugsby Testing (No Auth)",
      "item": [
        {
          "name": "Get Bugs by Assignee",
          "request": {
            "method": "GET",
            "header": [],
            "url": {
              "raw": "{{base_url}}/bugsby-api/bugs/assignee/om.nikam@arista.com?limit=3",
              "host": ["{{base_url}}"],
              "path": ["bugsby-api", "bugs", "assignee", "om.nikam@arista.com"],
              "query": [
                {
                  "key": "limit",
                  "value": "3"
                }
              ]
            }
          }
        },
        {
          "name": "Custom Bugsby Query",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"query\": \"assignee==om.nikam@arista.com AND status==ASSIGNED\",\n  \"limit\": \"10\",\n  \"sortBy\": \"lastUpdateTime\",\n  \"order\": \"desc\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/bugsby-api/bugs/query",
              "host": ["{{base_url}}"],
              "path": ["bugsby-api", "bugs", "query"]
            }
          }
        }
      ]
    },
    {
      "name": "Bug Management",
      "item": [
        {
          "name": "List Bugs",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/bugs?release=wifi-ooty&status=pending&limit=10",
              "host": ["{{base_url}}"],
              "path": ["bugs"],
              "query": [
                {
                  "key": "release",
                  "value": "wifi-ooty"
                },
                {
                  "key": "status",
                  "value": "pending"
                },
                {
                  "key": "limit",
                  "value": "10"
                }
              ]
            }
          }
        },
        {
          "name": "Get Bug by ID",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/bugs/61912d22-ffa9-4c09-92d6-df3a4d3541a4",
              "host": ["{{base_url}}"],
              "path": ["bugs", "61912d22-ffa9-4c09-92d6-df3a4d3541a4"]
            }
          }
        },
        {
          "name": "Get Bug by Bugsby ID",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/bugs/bugsby/1092263",
              "host": ["{{base_url}}"],
              "path": ["bugs", "bugsby", "1092263"]
            }
          }
        },
        {
          "name": "Update Bug",
          "request": {
            "method": "PUT",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              },
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"status\": \"dev_approved\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/bugs/61912d22-ffa9-4c09-92d6-df3a4d3541a4",
              "host": ["{{base_url}}"],
              "path": ["bugs", "61912d22-ffa9-4c09-92d6-df3a4d3541a4"]
            }
          }
        },
        {
          "name": "Delete Bug",
          "request": {
            "method": "DELETE",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/bugs/61912d22-ffa9-4c09-92d6-df3a4d3541a4",
              "host": ["{{base_url}}"],
              "path": ["bugs", "61912d22-ffa9-4c09-92d6-df3a4d3541a4"]
            }
          }
        }
      ]
    }
  ]
}
```

---

## 🔍 Common Use Cases

### Use Case 1: Manager Syncs Bugs for a Release

1. Login as manager
2. Call `POST /bugsby/sync` with release name
3. Check sync status with `GET /bugsby/status?release=wifi-ooty`
4. View synced bugs with `GET /bugs?release=wifi-ooty`

### Use Case 2: Developer Views Assigned Bugs

1. Login as developer
2. Call `GET /bugs?assigned_to=<your_user_id>`
3. View bug details with `GET /bugs/:id`

### Use Case 3: Testing Bugsby Integration

1. Call `GET /bugsby-api/bugs/assignee/your.email@arista.com` (no auth needed)
2. Verify bugs are returned from Bugsby
3. Use custom query for complex filters

---

## ⚠️ Important Notes

1. **Testing Endpoints**: `/bugsby-api/*` endpoints have authentication disabled for testing. Re-enable before production!

2. **Sync vs Direct Fetch**:
   - Use **Sync** (`/bugsby/sync`) to store bugs in our database
   - Use **Testing** (`/bugsby-api`) to directly query Bugsby without storing

3. **Manager Role**: Sync endpoints require manager role to prevent unauthorized data imports

4. **Rate Limiting**: Bugsby API has rate limits. Use sync sparingly for large releases.

5. **Pagination**: When syncing large releases, the API handles pagination automatically.

---

## 🐛 Error Responses

All errors follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

Common error codes:
- `invalid_request` - Bad request body or parameters
- `unauthorized` - Missing or invalid JWT token
- `forbidden` - Insufficient permissions (e.g., not a manager)
- `not_found` - Resource not found
- `sync_failed` - Bugsby sync operation failed
- `decode_failed` - Failed to parse Bugsby response


