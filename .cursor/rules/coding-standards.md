# Coding Standards

## Go Code Style

### Naming
- **Exported**: `PascalCase` (e.g., `UserService`, `GetUser`)
- **Internal**: `camelCase` (e.g., `userRepo`, `validateInput`)
- **Constants**: `PascalCase` or `SCREAMING_SNAKE_CASE` for true constants
- **Interfaces**: Noun or `-er` suffix (e.g., `Reader`, `UserRepository`)

### Error Handling
```go
// ✅ Good - wrap with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ Bad - naked return
if err != nil {
    return err
}
```

### Struct Tags
```go
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

### Testing
- Use table-driven tests
- Test file: `foo_test.go` in same package
- Use `testify/assert` for assertions
- Mock interfaces, not implementations

## TypeScript/React Style

### Components
```typescript
// ✅ Good - functional component with types
interface Props {
  title: string;
  onClick: () => void;
}

export function Button({ title, onClick }: Props) {
  return <button onClick={onClick}>{title}</button>;
}
```

### State Management
- Use React hooks (`useState`, `useEffect`, `useCallback`)
- TanStack Query for server state
- Keep state close to where it's used

## Rust Style

### Error Handling
```rust
// ✅ Good - use Result and ? operator
fn process_data(data: &str) -> Result<Output, Error> {
    let parsed = parse(data)?;
    let validated = validate(parsed)?;
    Ok(transform(validated))
}
```

### Naming
- Functions: `snake_case`
- Types/Traits: `PascalCase`
- Constants: `SCREAMING_SNAKE_CASE`

