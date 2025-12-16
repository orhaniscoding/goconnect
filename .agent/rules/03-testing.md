# Testing Strategy

This document defines testing practices to ensure code quality and reliability in GoConnect.

## Test Coverage Goals

| Component | Target Coverage | Priority Areas |
|-----------|----------------|----------------|
| core/internal/service/ | 80%+ | Business logic |
| core/internal/repository/ | 70%+ | Database operations |
| core/internal/handler/ | 70%+ | API endpoints |
| cli/internal/daemon/ | 70%+ | Daemon lifecycle |
| cli/internal/engine/ | 75%+ | Network engine |
| desktop/src/ | 60%+ | Critical components |

## Go Testing

### Unit Tests

**File Convention:** `*_test.go` in the same package

```go
// ✅ Table-driven tests (idiomatic Go)
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserRequest
        wantErr bool
        errType error
    }{
        {
            name:    "valid user",
            input:   CreateUserRequest{Email: "test@example.com", Password: "secure123"},
            wantErr: false,
        },
        {
            name:    "empty email",
            input:   CreateUserRequest{Email: "", Password: "secure123"},
            wantErr: true,
            errType: ErrInvalidEmail,
        },
        {
            name:    "weak password",
            input:   CreateUserRequest{Email: "test@example.com", Password: "123"},
            wantErr: true,
            errType: ErrWeakPassword,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := NewUserService(mockRepo)
            _, err := svc.CreateUser(context.Background(), tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                if tt.errType != nil {
                    require.ErrorIs(t, err, tt.errType)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Integration Tests

Use build tags to separate integration tests:

```go
//go:build integration

package repository_test

func TestUserRepository_Integration(t *testing.T) {
    // Setup real database
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    // Test with real database operations
    user, err := repo.Create(ctx, testUser)
    require.NoError(t, err)
    require.NotEmpty(t, user.ID)
}
```

### Mocking

Use interfaces for dependency injection:

```go
// Define interface
type UserRepository interface {
    Create(ctx context.Context, user *User) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
}

// Mock implementation for tests
type MockUserRepository struct {
    CreateFunc  func(ctx context.Context, user *User) (*User, error)
    GetByIDFunc func(ctx context.Context, id string) (*User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) (*User, error) {
    return m.CreateFunc(ctx, user)
}
```

### Test Helpers

```go
// testutil/helpers.go
func NewTestContext(t *testing.T) context.Context {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    t.Cleanup(cancel)
    return ctx
}

func RequireEventually(t *testing.T, condition func() bool, timeout time.Duration) {
    t.Helper()
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if condition() {
            return
        }
        time.Sleep(10 * time.Millisecond)
    }
    t.Fatal("condition not met within timeout")
}
```

---

## Running Tests

### Quick Tests
```bash
# Run all unit tests (fast)
make test

# Or directly
cd core && go test ./... -short
cd cli && go test ./... -short
```

### With Coverage
```bash
# Generate coverage report
cd core && go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Check coverage percentage
go tool cover -func=coverage.out | grep total
```

### Integration Tests
```bash
# Run integration tests (requires database)
cd core && go test ./... -tags=integration

# With race detection
go test ./... -race -short
```

### Specific Package
```bash
# Test specific package
go test ./internal/service/... -v

# Test specific function
go test ./internal/service/... -run TestUserService_CreateUser -v
```

---

## TypeScript Testing

### Vitest Configuration
Tests use Vitest (configured in `vitest.config.ts`):

```typescript
// Component test example
import { render, screen, fireEvent } from '@testing-library/react';
import { NetworkCard } from './NetworkCard';

describe('NetworkCard', () => {
  it('displays network name', () => {
    const network = { id: '1', name: 'Test Network' };
    render(<NetworkCard network={network} onJoin={() => {}} />);
    
    expect(screen.getByText('Test Network')).toBeInTheDocument();
  });
  
  it('calls onJoin when button clicked', async () => {
    const onJoin = vi.fn();
    const network = { id: '1', name: 'Test Network' };
    
    render(<NetworkCard network={network} onJoin={onJoin} />);
    await fireEvent.click(screen.getByRole('button', { name: /join/i }));
    
    expect(onJoin).toHaveBeenCalledWith('1');
  });
});
```

### Running Frontend Tests
```bash
cd desktop
npm test           # Run tests
npm run test:watch # Watch mode
npm run test:coverage # With coverage
```

---

## Test Quality Rules

### DO
- ✅ Test edge cases and error conditions
- ✅ Use descriptive test names that explain the scenario
- ✅ Keep tests independent (no shared state)
- ✅ Clean up resources in test teardown
- ✅ Use `t.Parallel()` for independent tests
- ✅ Test public interfaces, not internal implementations

### DON'T
- ❌ Don't test trivial code (getters/setters)
- ❌ Don't use `time.Sleep` in tests (use channels/conditions)
- ❌ Don't ignore test failures
- ❌ Don't commit broken tests
- ❌ Don't mock everything (integration tests matter)

---

## CI Integration

Tests run automatically on every PR via GitHub Actions:

```yaml
# .github/workflows/test.yml
- name: Run Tests
  run: |
    cd core && go test ./... -race -coverprofile=coverage.out
    cd ../cli && go test ./... -race -coverprofile=coverage.out
```

Coverage reports are generated and should not decrease on PRs.
