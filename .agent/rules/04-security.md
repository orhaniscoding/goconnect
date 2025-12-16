# Security Guidelines

GoConnect handles sensitive networking and encryption. Security is paramount.

## Core Security Principles

1. **Defense in Depth** - Multiple layers of security
2. **Least Privilege** - Minimal permissions required
3. **Fail Secure** - Default to denying access
4. **Never Trust Input** - Validate everything from external sources

---

## WireGuard Security

### Private Key Protection
```go
// ❌ NEVER do this
log.Printf("Private key: %s", privateKey)
fmt.Sprintf("Key: %s", wireguardKey)

// ❌ NEVER store in plain config
config := Config{
    PrivateKey: "abc123...",  // BAD!
}

// ✅ Always use secure storage
key, err := keyring.Get("goconnect", "wireguard_private_key")
if err != nil {
    return fmt.Errorf("failed to retrieve key from secure storage: %w", err)
}
```

### Pre-Shared Keys (PSK)
- Generate unique PSK for each peer connection
- Rotate PSKs every 6-12 months
- Never commit PSKs to version control
- Provides post-quantum resistance layer

### Key Rotation
```go
// Implement key rotation for long-lived connections
const KeyRotationInterval = 6 * 30 * 24 * time.Hour // ~6 months

func (m *KeyManager) ShouldRotate(keyCreatedAt time.Time) bool {
    return time.Since(keyCreatedAt) > KeyRotationInterval
}
```

---

## Secrets Management

### What MUST be Protected
| Secret Type | Storage | Never Log |
|-------------|---------|-----------|
| WireGuard private keys | OS keyring | ✅ |
| API tokens | Environment variables | ✅ |
| Database passwords | Environment variables | ✅ |
| Session tokens | Secure memory | ✅ |
| User passwords | Hash only (bcrypt) | ✅ |

### Secure Patterns
```go
// ✅ Good: Use environment variables
dbPassword := os.Getenv("DB_PASSWORD")
if dbPassword == "" {
    return errors.New("DB_PASSWORD not set")
}

// ✅ Good: Implement Stringer to hide secrets
type SecretString string

func (s SecretString) String() string {
    return "[REDACTED]"
}

func (s SecretString) Value() string {
    return string(s)
}

// ✅ Good: Clear secrets from memory when done
func (c *Client) Close() {
    // Zero out sensitive data
    for i := range c.apiKey {
        c.apiKey[i] = 0
    }
}
```

---

## Input Validation

### SQL Injection Prevention
```go
// ❌ NEVER do this - SQL injection vulnerability
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)
db.Query(query)

// ✅ Always use parameterized queries
query := "SELECT * FROM users WHERE id = $1"
db.Query(query, userID)

// ✅ Use query builders
user, err := db.User.
    Query().
    Where(user.ID(userID)).
    Only(ctx)
```

### API Input Validation
```go
type CreateNetworkRequest struct {
    Name        string `json:"name" validate:"required,min=3,max=50"`
    Description string `json:"description" validate:"max=500"`
    IsPrivate   bool   `json:"is_private"`
}

func (h *Handler) CreateNetwork(c *gin.Context) {
    var req CreateNetworkRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request body"})
        return
    }
    
    // Validate using validator
    if err := h.validator.Struct(req); err != nil {
        c.JSON(400, gin.H{"error": "validation failed", "details": err.Error()})
        return
    }
    
    // Sanitize input
    req.Name = strings.TrimSpace(req.Name)
    req.Description = sanitize.HTML(req.Description)
    
    // Process...
}
```

### Path Traversal Prevention
```go
// ❌ NEVER trust user-provided paths
filePath := filepath.Join(baseDir, userProvidedPath)

// ✅ Validate and clean paths
func SafeJoin(baseDir, userPath string) (string, error) {
    // Clean the path
    cleanPath := filepath.Clean(userPath)
    
    // Ensure no directory traversal
    if strings.Contains(cleanPath, "..") {
        return "", errors.New("invalid path: directory traversal detected")
    }
    
    fullPath := filepath.Join(baseDir, cleanPath)
    
    // Verify result is still under baseDir
    if !strings.HasPrefix(fullPath, baseDir) {
        return "", errors.New("invalid path: outside base directory")
    }
    
    return fullPath, nil
}
```

---

## Network Security

### TLS Requirements
- Minimum TLS 1.2 for all connections
- TLS 1.3 preferred where supported
- Valid certificates required (no skip verify in production)

```go
// ✅ Proper TLS configuration
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
    },
}

// ❌ NEVER in production
tlsConfig := &tls.Config{
    InsecureSkipVerify: true,  // DANGEROUS!
}
```

### Rate Limiting
```go
// Apply rate limiting to prevent abuse
limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 req/sec

func RateLimitMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
            return
        }
        c.Next()
    }
}
```

---

## Authentication & Authorization

### Password Handling
```go
import "golang.org/x/crypto/bcrypt"

// ✅ Hash passwords with bcrypt
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// ✅ Verify passwords securely
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Token Security
```go
// ✅ Use secure random for tokens
func GenerateSecureToken(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

// ✅ Set proper token expiration
const (
    AccessTokenExpiry  = 15 * time.Minute
    RefreshTokenExpiry = 7 * 24 * time.Hour
)
```

---

## Security Checklist

Before merging any PR, verify:

- [ ] No secrets or credentials in code
- [ ] All user input is validated
- [ ] SQL queries use parameterized statements
- [ ] Sensitive data is never logged
- [ ] TLS is properly configured
- [ ] Authentication is required where needed
- [ ] Rate limiting is applied to public endpoints
- [ ] Error messages don't leak internal details
