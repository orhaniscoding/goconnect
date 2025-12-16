# Error Handling

Consistent error handling ensures debuggability and user experience.

## Go Error Handling

### Error Wrapping
Always add context when propagating errors:

```go
// ✅ Good: Adds context with %w for error chain
func (s *NetworkService) JoinNetwork(ctx context.Context, networkID string) error {
    network, err := s.repo.GetByID(ctx, networkID)
    if err != nil {
        return fmt.Errorf("failed to get network %s: %w", networkID, err)
    }
    
    if err := s.engine.Connect(network); err != nil {
        return fmt.Errorf("failed to connect to network: %w", err)
    }
    
    return nil
}

// ❌ Bad: Loses original error context
func (s *NetworkService) JoinNetwork(ctx context.Context, networkID string) error {
    network, err := s.repo.GetByID(ctx, networkID)
    if err != nil {
        return errors.New("failed to join network")  // Lost the why!
    }
    // ...
}
```

### Sentinel Errors
Use sentinel errors for expected error conditions:

```go
// Define package-level sentinel errors
var (
    ErrNotFound      = errors.New("resource not found")
    ErrUnauthorized  = errors.New("unauthorized access")
    ErrAlreadyExists = errors.New("resource already exists")
    ErrInvalidInput  = errors.New("invalid input")
)

// Check with errors.Is
func (h *Handler) GetNetwork(c *gin.Context) {
    network, err := h.service.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            c.JSON(404, gin.H{"error": "network not found"})
            return
        }
        c.JSON(500, gin.H{"error": "internal error"})
        return
    }
}
```

### Custom Error Types
For complex error data, use custom types:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

// Check with errors.As
var validErr *ValidationError
if errors.As(err, &validErr) {
    c.JSON(400, gin.H{
        "error": "validation_error",
        "field": validErr.Field,
        "message": validErr.Message,
    })
}
```

---

## Panic Handling

### When to Panic
```go
// ✅ OK: Unrecoverable initialization errors
func MustLoadConfig(path string) *Config {
    cfg, err := LoadConfig(path)
    if err != nil {
        panic(fmt.Sprintf("failed to load required config: %v", err))
    }
    return cfg
}

// ✅ OK: Programming errors (should never happen)
func (s *Service) Process(item Item) {
    if item.ID == "" {
        panic("Process called with empty item ID - this is a bug")
    }
}
```

### Panic Recovery
```go
// Recover panics in HTTP handlers
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                log.Error().
                    Interface("panic", r).
                    Str("stack", string(debug.Stack())).
                    Msg("panic recovered")
                
                c.JSON(500, gin.H{"error": "internal server error"})
            }
        }()
        c.Next()
    }
}
```

---

## Logging Standards

### Structured Logging
```go
import "github.com/rs/zerolog/log"

// ✅ Good: Structured with context
log.Info().
    Str("network_id", networkID).
    Str("user_id", userID).
    Int("peer_count", len(peers)).
    Msg("user joined network")

// ✅ Good: Error with context
log.Error().
    Err(err).
    Str("network_id", networkID).
    Str("operation", "join_network").
    Msg("failed to join network")

// ❌ Bad: Unstructured, hard to parse
log.Printf("User %s joined network %s with %d peers", userID, networkID, len(peers))
```

### Log Levels
| Level | Usage |
|-------|-------|
| `Error` | Errors requiring attention |
| `Warn` | Unusual but handled situations |
| `Info` | Normal operations (start/stop, connections) |
| `Debug` | Detailed debugging info |
| `Trace` | Very detailed, performance-heavy logging |

### What NOT to Log
```go
// ❌ NEVER log sensitive data
log.Info().Str("password", password).Msg("login attempt")     // BAD!
log.Debug().Str("api_key", apiKey).Msg("making request")      // BAD!
log.Info().Str("private_key", wgKey).Msg("wireguard config")  // BAD!

// ✅ Log identifiers, not secrets
log.Info().Str("user_email", email).Msg("login attempt")
log.Debug().Str("key_id", keyID).Msg("using API key")
```

---

## User-Facing vs Internal Errors

### HTTP API
```go
func (h *Handler) CreateNetwork(c *gin.Context) {
    network, err := h.service.Create(ctx, req)
    if err != nil {
        // Log full error internally
        log.Error().Err(err).Str("request_id", requestID).Msg("create network failed")
        
        // Return safe message to user
        switch {
        case errors.Is(err, ErrInvalidInput):
            c.JSON(400, gin.H{"error": "Invalid network configuration"})
        case errors.Is(err, ErrAlreadyExists):
            c.JSON(409, gin.H{"error": "Network name already taken"})
        default:
            // Never expose internal errors to users!
            c.JSON(500, gin.H{"error": "An unexpected error occurred"})
        }
        return
    }
    
    c.JSON(201, network)
}
```

### CLI/TUI
```go
func (cmd *JoinCommand) Execute() error {
    if err := daemon.JoinNetwork(networkID); err != nil {
        // Show user-friendly message
        if errors.Is(err, ErrNetworkNotFound) {
            return fmt.Errorf("network '%s' not found - check the ID and try again", networkID)
        }
        if errors.Is(err, ErrConnectionFailed) {
            return fmt.Errorf("could not connect - check your internet connection")
        }
        // Generic fallback
        return fmt.Errorf("failed to join network: please try again later")
    }
    return nil
}
```

---

## Error Handling Checklist

- [ ] All errors are either handled or wrapped with context
- [ ] Sentinel errors used for expected conditions
- [ ] Custom error types for complex error data
- [ ] Panics only for unrecoverable/programming errors
- [ ] Structured logging with appropriate levels
- [ ] No sensitive data in logs or error messages
- [ ] User-facing errors are helpful but don't leak internals
