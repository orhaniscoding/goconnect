# Code Standards

This document defines coding standards for all languages used in GoConnect.

## Go Code Style

### Naming Conventions
| Type | Convention | Example |
|------|------------|---------|
| Exported | PascalCase | `CreateNetwork`, `UserService` |
| Internal | camelCase | `validateInput`, `parseConfig` |
| Constants | PascalCase or ALL_CAPS | `MaxRetries`, `DEFAULT_PORT` |
| Packages | lowercase, singular | `service`, `handler`, `repository` |

### Formatting
- Always run `gofmt` or `goimports` before committing
- Use `golangci-lint` for static analysis (see `.golangci.yml`)
- Maximum line length: 120 characters (soft limit)

### Import Organization
Group imports in this order, separated by blank lines:
```go
import (
    // Standard library
    "context"
    "fmt"
    
    // External packages
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    
    // Internal packages
    "github.com/orhaniscoding/goconnect/server/internal/service"
)
```

### Function Guidelines
```go
// ✅ Good: Clear name, context first, error last
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // validation first
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    // main logic
    // ...
}

// ❌ Bad: Poor naming, no context
func (s *UserService) DoThing(data interface{}) interface{} {
    // ...
}
```

### Struct Tags
```go
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

---

## TypeScript/React Code Style

### Component Structure
```typescript
// ✅ Functional components with TypeScript interfaces
interface NetworkCardProps {
  network: Network;
  onJoin: (id: string) => void;
  isLoading?: boolean;
}

export function NetworkCard({ network, onJoin, isLoading = false }: NetworkCardProps) {
  // hooks first
  const [isExpanded, setIsExpanded] = useState(false);
  
  // handlers
  const handleJoin = () => onJoin(network.id);
  
  // render
  return (
    <div className="network-card">
      {/* ... */}
    </div>
  );
}
```

### Naming Conventions
| Type | Convention | Example |
|------|------------|---------|
| Components | PascalCase | `NetworkCard.tsx` |
| Hooks | camelCase with use prefix | `useNetworkStatus.ts` |
| Utilities | camelCase | `formatDate.ts` |
| Constants | UPPER_SNAKE_CASE | `API_BASE_URL` |
| Types/Interfaces | PascalCase | `NetworkStatus`, `UserProfile` |

### File Organization
```
src/
├── components/          # Reusable UI components
│   ├── NetworkCard.tsx
│   └── NetworkCard.test.tsx
├── hooks/              # Custom React hooks
├── pages/              # Page components
├── services/           # API calls and external services
├── types/              # TypeScript type definitions
└── utils/              # Helper functions
```

---

## Rust Code Style (Tauri)

### Naming Conventions
| Type | Convention | Example |
|------|------------|---------|
| Functions | snake_case | `get_network_status` |
| Types/Structs | PascalCase | `NetworkConfig` |
| Constants | UPPER_SNAKE_CASE | `MAX_CONNECTIONS` |
| Modules | snake_case | `network_manager` |

### Error Handling
```rust
// ✅ Always use Result with ?
fn connect_to_network(config: &NetworkConfig) -> Result<Connection, NetworkError> {
    let socket = create_socket(&config.address)?;
    let connection = socket.connect()?;
    Ok(connection)
}

// ✅ Use custom error types
#[derive(Debug, thiserror::Error)]
pub enum NetworkError {
    #[error("Failed to connect: {0}")]
    ConnectionFailed(String),
    #[error("Invalid configuration")]
    InvalidConfig,
}
```

### Tauri Commands
```rust
#[tauri::command]
async fn get_networks(state: State<'_, AppState>) -> Result<Vec<Network>, String> {
    state.daemon_client
        .list_networks()
        .await
        .map_err(|e| e.to_string())
}
```

---

## General Rules

### Comments
```go
// ✅ Good: Explains WHY, not WHAT
// Use bcrypt for password hashing because it's resistant to GPU attacks
// and automatically handles salt generation.
hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

// ❌ Bad: States the obvious
// Hash the password
hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
```

### Magic Numbers
```go
// ❌ Bad
if retries > 3 {
    return errors.New("too many retries")
}

// ✅ Good
const MaxRetries = 3

if retries > MaxRetries {
    return errors.New("too many retries")
}
```

### Boolean Naming
```go
// ✅ Good: Clear intent
isActive, hasPermission, canDelete, shouldRetry

// ❌ Bad: Ambiguous
active, permission, delete, retry
```

---

## Linting & Formatting

### Go
```bash
# Format code
gofmt -w .
goimports -w .

# Run linter
golangci-lint run ./...
```

### TypeScript
```bash
# Format with Prettier
npm run format

# Lint with ESLint  
npm run lint
```

### Rust
```bash
# Format code
cargo fmt

# Run linter
cargo clippy
```
