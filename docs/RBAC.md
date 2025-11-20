# Role-Based Access Control (RBAC) Guide

GoConnect implements a comprehensive RBAC system with two privilege levels beyond regular users.

## Roles

### 1. Regular User
**Default role for all users**

**Permissions:**
- ✅ Send chat messages
- ✅ Edit own messages (within 15 minutes)
- ✅ Delete own messages
- ✅ View all messages in accessible scopes
- ✅ Create and manage own networks
- ✅ Join public networks

**Restrictions:**
- ❌ Cannot edit messages after 15-minute window
- ❌ Cannot delete other users' messages
- ❌ Cannot redact messages
- ❌ Cannot edit/delete messages in restricted networks

---

### 2. Moderator (`is_moderator: true`)
**Content moderation privileges**

**All Regular User permissions, plus:**
- ✅ Delete any user's messages (soft or hard delete)
- ✅ Redact inappropriate content
- ✅ Access edit history for all messages
- ✅ Moderate chat in all scopes

**Restrictions:**
- ❌ Cannot edit other users' messages
- ❌ Limited to content moderation (no admin privileges)

**Use Cases:**
- Community moderators
- Content reviewers
- Customer support staff

---

### 3. Administrator (`is_admin: true`)
**Full system privileges**

**All Moderator permissions, plus:**
- ✅ Edit any message (no time limit)
- ✅ Full network management
- ✅ User management capabilities
- ✅ System configuration access

**Use Cases:**
- System administrators
- Platform owners
- DevOps team members

---

## JWT Token Structure

```json
{
  "user_id": "usr_123abc",
  "tenant_id": "tnt_456def",
  "email": "user@example.com",
  "is_admin": false,
  "is_moderator": true,
  "type": "access",
  "exp": 1698765432,
  "iat": 1698764532
}
```

---

## API Endpoints by Role

### Public Endpoints (No Auth)
```http
POST /v1/auth/register
POST /v1/auth/login
POST /v1/auth/refresh
```

### Authenticated Endpoints
```http
# All authenticated users
GET  /v1/chat
POST /v1/chat
GET  /v1/chat/:id
PATCH /v1/chat/:id          # Own messages only (15min limit)
DELETE /v1/chat/:id         # Own messages or admin/moderator
```

### Moderator-Only Endpoints
```http
POST /v1/chat/:id/redact    # RequireModerator middleware
```

### Admin-Only Endpoints
```http
# Future: User management, system settings
```

---

## Middleware Usage

### In Route Registration

```go
// Public routes
r.POST("/v1/auth/login", authHandler.Login)

// Authenticated routes
chat := r.Group("/v1/chat")
chat.Use(AuthMiddleware(authService))
{
    chat.GET("", handler.ListMessages)
    chat.POST("", handler.SendMessage)
    
    // Moderator-only routes
    chat.POST("/:id/redact", RequireModerator(), handler.RedactMessage)
}

// Admin-only routes
admin := r.Group("/v1/admin")
admin.Use(AuthMiddleware(authService))
admin.Use(RequireAdmin())
{
    admin.GET("/users", handler.ListUsers)
}
```

### In Handler Code

```go
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
    userID := c.GetString("user_id")
    isAdmin := c.GetBool("is_admin")
    isModerator := c.GetBool("is_moderator")
    
    // Service enforces permissions
    err := h.chatService.DeleteMessage(ctx, messageID, userID, mode, isAdmin, isModerator)
}
```

---

## WebSocket Permissions

WebSocket clients automatically inherit permissions from their JWT token:

```go
type Client struct {
    userID      string
    tenantID    string
    isAdmin     bool        // From JWT
    isModerator bool        // From JWT
    rooms       map[string]bool
}
```

**WebSocket message handlers use these flags:**
```go
// chat.edit
chatService.EditMessage(ctx, messageID, client.userID, newBody, client.isAdmin)

// chat.delete
chatService.DeleteMessage(ctx, messageID, client.userID, mode, client.isAdmin, client.isModerator)

// chat.redact
chatService.RedactMessage(ctx, messageID, client.userID, client.isAdmin, client.isModerator, mask)
```

---

## Granting Moderator Role

### Method 1: Database Direct Update (PostgreSQL)
```sql
UPDATE users 
SET is_moderator = true, updated_at = NOW() 
WHERE email = 'moderator@example.com';
```

### Method 2: Admin API (Future)
```http
POST /v1/admin/users/:user_id/roles
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "is_moderator": true
}
```

### Method 3: Registration Override (Development)
```go
// In auth.go Register() method - temporary for testing
user := &domain.User{
    Email:       req.Email,
    IsModerator: req.Email == "mod@example.com", // Quick test override
}
```

---

## Security Best Practices

### 1. Token Validation
- ✅ JWT tokens signed with HS256
- ✅ Access tokens expire in 15 minutes
- ✅ Refresh tokens expire in 7 days
- ✅ Permissions extracted from validated tokens only

### 2. Permission Checks
- ✅ Middleware enforces role requirements
- ✅ Service layer validates ownership + permissions
- ✅ Domain layer provides permission helper methods

### 3. Audit Trail
- ✅ All moderator actions logged
- ✅ Redaction includes moderator ID + reason
- ✅ Edit history preserved

---

## Testing Roles

### Create Test Users

```bash
# Regular user
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"pass123","locale":"en"}'

# Moderator (set via database)
psql -d goconnect -c "UPDATE users SET is_moderator = true WHERE email = 'user@test.com';"

# Admin (set via database)
psql -d goconnect -c "UPDATE users SET is_admin = true WHERE email = 'user@test.com';"
```

### Test Permission Enforcement

```http
# Login as moderator
POST http://localhost:8080/v1/auth/login
Content-Type: application/json

{"email":"mod@test.com","password":"pass123"}

# Use token to redact message
POST http://localhost:8080/v1/chat/msg_123/redact
Authorization: Bearer {moderator_token}
Content-Type: application/json

{"reason":"Spam content"}
```

---

## Common Issues

### 1. "Moderator privileges required" error
**Cause:** User's `is_moderator` flag is false in JWT token

**Solution:**
1. Update database: `UPDATE users SET is_moderator = true WHERE id = 'user_id';`
2. Get new token: `POST /v1/auth/login` (old tokens won't have new permission)
3. Use new access token

### 2. Permission works in REST but not WebSocket
**Cause:** WebSocket client created before permission grant

**Solution:** Reconnect WebSocket connection to get new JWT with updated permissions

### 3. Cannot edit own message after granting moderator
**Cause:** Moderators still have 15-minute edit limit on own messages (use admin for unlimited)

**Solution:** Grant `is_admin` for unlimited edit time, or wait for message to age and use redact instead

---

## Roadmap

### Phase 1 (Current)
- ✅ Basic RBAC (Admin + Moderator)
- ✅ JWT permission extraction
- ✅ Middleware enforcement
- ✅ WebSocket permission support

### Phase 2 (Next)
- ⏳ Admin API for user management
- ⏳ Role assignment UI
- ⏳ Permission audit dashboard
- ⏳ Moderator action history

### Phase 3 (Future)
- ⏳ Custom role definitions
- ⏳ Fine-grained permissions
- ⏳ Network-scoped moderators
- ⏳ Temporary role grants

---

## See Also

- [TECH_SPEC.md](./TECH_SPEC.md) - Technical specification
- [API_EXAMPLES.http](./API_EXAMPLES.http) - API usage examples
- [WS_MESSAGES.md](./WS_MESSAGES.md) - WebSocket message types
- [SECURITY.md](./SECURITY.md) - Security guidelines
