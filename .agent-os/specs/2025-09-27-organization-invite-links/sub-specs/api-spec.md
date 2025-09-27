# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-09-27-organization-invite-links/spec.md

## Endpoints

### POST /api/organization/invites

**Purpose:** Generate a new organization invite link
**Parameters:** None (uses authenticated user's organization)
**Response:** 
```json
{
  "invite_id": "uuid",
  "token": "secure-token",
  "invite_url": "https://app.example.com/invite/secure-token",
  "expires_at": "2025-09-28T10:30:00Z",
  "created_at": "2025-09-27T10:30:00Z"
}
```
**Errors:** 
- 403 Forbidden (user not admin)
- 500 Internal Server Error

### GET /api/organization/invites

**Purpose:** List active invite links for the organization
**Parameters:** None (uses authenticated user's organization)
**Response:**
```json
{
  "invites": [
    {
      "invite_id": "uuid",
      "token": "secure-token",
      "invite_url": "https://app.example.com/invite/secure-token",
      "created_by": "admin-username",
      "created_at": "2025-09-27T10:30:00Z",
      "expires_at": "2025-09-28T10:30:00Z",
      "used_at": null,
      "used_by": null,
      "active": true
    }
  ]
}
```
**Errors:**
- 403 Forbidden (user not admin)
- 500 Internal Server Error

### DELETE /api/organization/invites/{invite_id}

**Purpose:** Revoke an active invite link
**Parameters:** 
- invite_id (path): UUID of the invite to revoke
**Response:** 204 No Content
**Errors:**
- 403 Forbidden (user not admin)
- 404 Not Found (invite not found)
- 400 Bad Request (invite already used or expired)
- 500 Internal Server Error

### GET /invite/{token}

**Purpose:** Display registration form for invited users
**Parameters:**
- token (path): Invite token from the invite link
**Response:** HTML page with registration form
**Errors:**
- 404 Not Found (invalid or expired token)
- 400 Bad Request (token already used)

### POST /invite/{token}/register

**Purpose:** Complete user registration using invite link
**Parameters:**
- token (path): Invite token
- Body: Registration form data
```json
{
  "name": "John Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword"
}
```
**Response:** 
- 302 Redirect to dashboard on success
- HTML form with validation errors on failure
**Errors:**
- 404 Not Found (invalid or expired token)
- 400 Bad Request (validation errors, token already used)
- 409 Conflict (username or email already exists)

## Controllers

### Organization Invite Controller

**Actions:**
- `GenerateInvite`: Creates new invite link for organization
- `ListInvites`: Retrieves active invites for organization
- `RevokeInvite`: Deactivates an invite link
- `ValidateInvite`: Validates invite token and returns organization info
- `UseInvite`: Marks invite as used and creates user account

**Business Logic:**
- Only ROLE_ADMIN users can generate/revoke invites
- Invite tokens are cryptographically secure (UUID + random)
- Expiration is exactly 24 hours from creation
- Used invites cannot be reused
- Username and email must be unique across all organizations

**Error Handling:**
- Comprehensive validation of invite tokens
- Proper HTTP status codes for different error conditions
- User-friendly error messages for registration failures
- Security: No information leakage about organization structure
