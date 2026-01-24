# Developer Setup Tutorial

**Complete guide to setting up GoConnect development environment for contributing to the project.**

---

## 📋 Overview

This tutorial shows you how to:
- Clone and build GoConnect from source
- Set up your development environment
- Run and test the application locally
- Make your first contribution

**Time:** 30-45 minutes
**Difficulty:** Intermediate
**Prerequisites:**
- Git installed
- Go 1.21+ installed
- Node.js 18+ and npm installed
- Basic command line knowledge

---

## 📦 Step 1: Fork and Clone

### 1.1 Fork the Repository

1. Visit [github.com/oeo/goconnect](https://github.com/oeo/goconnect)
2. Click **Fork** button in the top-right
3. Select your account as the destination

### 1.2 Clone Your Fork

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/goconnect.git
cd goconnect

# Add upstream remote
git remote add upstream https://github.com/oeo/goconnect.git

# Verify remotes
git remote -v
```

**Expected output:**
```
origin    https://github.com/YOUR_USERNAME/goconnect.git (fetch)
origin    https://github.com/YOUR_USERNAME/goconnect.git (push)
upstream  https://github.com/oeo/goconnect.git (fetch)
upstream  https://github.com/oeo/goconnect.git (push)
```

---

## 🛠️ Step 2: Install Dependencies

### 2.1 Check Go Version

```bash
go version
```

**Required:** Go 1.21 or higher

If you need to install/update Go:
- **macOS:** `brew install go`
- **Windows:** Download from [go.dev/dl](https://go.dev/dl/)
- **Linux:** `sudo snap install go --classic`

### 2.2 Check Node.js Version

```bash
node --version
npm --version
```

**Required:** Node.js 18+ and npm 9+

If you need to install/update Node.js:
- **macOS:** `brew install node`
- **Windows:** Download from [nodejs.org](https://nodejs.org/)
- **Linux:** `sudo apt install nodejs npm` or use [nvm](https://github.com/nvm-sh/nvm)

### 2.3 Install Go Dependencies

```bash
# From project root
go mod download
go mod verify
```

### 2.4 Install Node.js Dependencies

```bash
# Navigate to desktop app
cd desktop
npm install
```

---

## 🏗️ Step 3: Build the Project

### 3.1 Build Core Service

```bash
# From project root
cd core
go build -o goconnect-core ./cmd/goconnect
```

**Expected output:**
```
# Successful build creates: core/goconnect-core (or goconnect-core.exe on Windows)
```

### 3.2 Build Desktop App

```bash
# From desktop directory
cd ../desktop
npm run build
```

**Expected output:**
```
✓ built in XXms
```

### 3.3 Verify Builds

```bash
# Test core service
cd ../core
./goconnect-core --version

# Test desktop app
cd ../desktop
npm run dev
```

---

## 🧪 Step 4: Run Tests

### 4.1 Run Go Tests

```bash
# From core directory
cd core

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/p2p/...
```

**Expected output:**
```
ok      github.com/oeo/goconnect/core/internal/p2p      0.234s  coverage: 87.5% of statements
ok      github.com/oeo/goconnect/core/internal/network  0.156s  coverage: 92.3% of statements
...
```

### 4.2 Run Frontend Tests

```bash
# From desktop directory
cd ../desktop

# Run tests
npm test

# Run with coverage
npm run test:coverage
```

---

## 🔧 Step 5: Development Workflow

### 5.1 Run in Development Mode

**Terminal 1 - Core Service:**
```bash
cd core
go run ./cmd/goconnect
```

**Terminal 2 - Desktop App:**
```bash
cd desktop
npm run dev
```

**Terminal 3 - Watch Tests (Optional):**
```bash
cd core
# Install watch tool
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### 5.2 Code Style and Linting

**Go Code:**
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run ./...

# Auto-format code
go fmt ./...
```

**JavaScript/TypeScript:**
```bash
cd desktop

# Run ESLint
npm run lint

# Auto-fix issues
npm run lint:fix

# Format with Prettier
npm run format
```

### 5.3 Protocol Buffers (If Modified)

```bash
# Install protoc compiler
# macOS: brew install protobuf
# Windows: Download from github.com/protocolbuffers/protobuf/releases
# Linux: sudo apt install protobuf-compiler

# Install Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
cd core
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

---

## 🌿 Step 6: Create a Feature Branch

### 6.1 Sync with Upstream

```bash
# Fetch latest changes
git fetch upstream

# Update your main branch
git checkout main
git merge upstream/main
git push origin main
```

### 6.2 Create Feature Branch

```bash
# Create and checkout new branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

**Branch naming conventions:**
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements

---

## 📝 Step 7: Make Changes

### 7.1 Follow Code Standards

**Zero-Dependency Policy:**
- ❌ No ORM libraries (gorm, sqlx)
- ❌ No heavy abstraction libraries
- ✅ Standard library preferred
- ✅ Custom implementations for simple tasks

See [DEPENDENCIES.md](../../DEPENDENCIES.md) for details.

**Code Quality:**
- Single Responsibility Principle
- No magic numbers
- No silent try/catch
- Readable code over clever code
- Comments only where logic isn't self-evident

See [DEVELOPMENT.md](../../DEVELOPMENT.md) for full guidelines.

### 7.2 Write Tests

**Every change must include tests:**

```go
// Example test structure
func TestNetworkCreation(t *testing.T) {
    // Setup
    service := setupTestService(t)
    defer service.Cleanup()

    // Execute
    network, err := service.CreateNetwork("test-net")

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, network)
    assert.Equal(t, "test-net", network.Name)
}
```

**Minimum coverage:** 80% global, 100% for critical business logic

---

## ✅ Step 8: Submit Your Changes

### 8.1 Run Pre-Commit Checks

```bash
# Run all tests
cd core && go test ./...
cd ../desktop && npm test

# Run linters
cd core && golangci-lint run ./...
cd ../desktop && npm run lint

# Check formatting
cd core && go fmt ./...
cd ../desktop && npm run format
```

### 8.2 Commit Changes

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: Add network discovery feature

- Implement peer discovery protocol
- Add discovery tests with 95% coverage
- Update network service documentation

Closes #123"
```

**Commit message format:**
```
<type>: <subject>

<body>

<footer>
```

**Types:** feat, fix, docs, refactor, test, chore

### 8.3 Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name
```

**Then:**
1. Visit your fork on GitHub
2. Click **Compare & Pull Request**
3. Fill out the PR template (auto-loaded)
4. Submit for review

---

## 🐛 Step 9: Debug Common Issues

### 9.1 Build Failures

**"Package not found"**
```bash
# Clean and reinstall
cd core
go clean -modcache
go mod download
```

**"Node modules missing"**
```bash
# Reinstall dependencies
cd desktop
rm -rf node_modules package-lock.json
npm install
```

### 9.2 Test Failures

**"Port already in use"**
```bash
# Find and kill process
# macOS/Linux:
lsof -ti:8080 | xargs kill -9

# Windows:
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

**"Database locked"**
```bash
# Clean test database
rm -rf core/test.db core/*.db-*
```

### 9.3 Runtime Issues

**"Permission denied"**
```bash
# Make binary executable
chmod +x core/goconnect-core
```

**"CGO errors"**
```bash
# Disable CGO (if not needed)
CGO_ENABLED=0 go build ./cmd/goconnect
```

---

## 📚 Additional Resources

### Documentation
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Full contribution guidelines
- [DEVELOPMENT.md](../../DEVELOPMENT.md) - Development practices
- [DEPENDENCIES.md](../../DEPENDENCIES.md) - Dependency policy
- [TROUBLESHOOTING.md](../../TROUBLESHOOTING.md) - Common issues

### Code Structure
```
goconnect/
├── core/                   # Go backend service
│   ├── cmd/                # Command-line entry points
│   ├── internal/           # Internal packages
│   │   ├── p2p/            # P2P networking
│   │   ├── network/        # Network management
│   │   ├── database/       # Database layer
│   │   └── api/            # API handlers
│   └── proto/              # Protocol buffers
│
└── desktop/                # Electron desktop app
    ├── src/                # TypeScript source
    ├── renderer/           # UI components
    └── main/               # Electron main process
```

### Communication
- **Issues:** [github.com/oeo/goconnect/issues](https://github.com/oeo/goconnect/issues)
- **Discussions:** [github.com/oeo/goconnect/discussions](https://github.com/oeo/goconnect/discussions)
- **Security:** See [SECURITY.md](../../SECURITY.md)

---

## 🎯 Next Steps

Now that your development environment is set up:

1. **Pick an issue** - Look for `good first issue` label
2. **Ask questions** - Don't hesitate to ask in discussions
3. **Start small** - Fix typos, improve docs, add tests
4. **Learn codebase** - Read existing code and tests
5. **Contribute** - Submit your first PR!

**Happy coding! 🚀**

---

*Last updated: 2026-01-25*
