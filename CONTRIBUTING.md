# 🤝 Contributing to GoConnect / GoConnect'e Katkıda Bulunma

---

[English](#english) | [Türkçe](#türkçe)

---

## English

## 📋 Overview

**What is contributing?**

Contributing means helping improve GoConnect. This could be:
- Reporting bugs
- Suggesting new features
- Writing code
- Improving documentation
- Helping other users
- Translating documentation

**Who can contribute?**

Everyone! Whether you're a beginner or an expert, your help is valuable.

**Why contribute?**

- 🌟 Help thousands of users
- 📚 Learn new skills (Go, Rust, React, WireGuard)
- 🏆 Build your portfolio
- 👥 Join a great community
- 🎁 Get swag (stickers, t-shirt)

---

## 🚀 Quick Start (5 Minutes)

### First Time Contributing?

Welcome! Here's how to get started:

#### Step 1: Understand the Project

**What is GoConnect?**

GoConnect is a virtual LAN platform with 3 main parts:
- **CLI** - Command-line application (Go)
- **Core** - Server backend (Go)
- **Desktop** - Desktop app (Tauri + React)

**What technologies do we use?**

| Component | Technology | Why? |
|-----------|------------|------|
| **CLI** | Go + Bubbletea | Cross-platform, single binary |
| **Core** | Go + WireGuard | Fast, secure networking |
| **Desktop** | Tauri + React | Small, fast desktop apps |
| **Protocol** | Protocol Buffers | Type-safe communication |

#### Step 2: Set Up Your Development Environment

**What you'll need:**

1. **Git** - Version control
2. **Go 1.24+** - For CLI and Core
3. **Node.js 20+** - For Desktop app
4. **Rust** - For Tauri (Desktop app backend)
5. **protoc** - Protocol Buffers compiler
6. **Editor** - VS Code, GoLand, or similar

**How to install:**

**Git:**
- Windows: https://git-scm.com/download/win
- macOS: `brew install git`
- Linux: `sudo apt install git`

**Go:**
- Download: https://go.dev/dl/
- Verify: `go version`

**Node.js:**
- Download: https://nodejs.org/
- Verify: `node --version`

**Rust:**
- Download: https://rustup.rs/
- Verify: `rustc --version`

**protoc:**
- macOS: `brew install protobuf`
- Linux: `sudo apt install protobuf-compiler`
- Windows: https://github.com/protocolbuffers/protobuf/releases

#### Step 3: Fork and Clone

**What is forking?**

Forking creates your own copy of GoConnect on GitHub.

**How to fork:**

1. Go to https://github.com/orhaniscoding/goconnect
2. Click "Fork" button (top right)
3. Wait for fork to complete

**How to clone:**

```bash
# Clone YOUR fork
git clone https://github.com/YOUR_USERNAME/goconnect.git
cd goconnect

# Add original repository as upstream
git remote add upstream https://github.com/orhaniscoding/goconnect.git

# Verify
git remote -v
```

**Expected output:**
```
origin    https://github.com/YOUR_USERNAME/goconnect.git (fetch)
origin    https://github.com/YOUR_USERNAME/goconnect.git (push)
upstream  https://github.com/orhaniscoding/goconnect.git (fetch)
upstream  https://github.com/orhaniscoding/goconnect.git (push)
```

#### Step 4: Create a Branch

**What is a branch?**

A branch is a separate version of the code where you make your changes.

**How to create a branch:**

```bash
# Update from upstream
git fetch upstream

# Create branch from main
git checkout -b feature/your-feature-name

# Or for bug fix
git checkout -b fix/your-bug-fix
```

**Branch naming:**

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Adding tests

#### Step 5: Make Your Changes

**What can you change?**

- Add new features
- Fix bugs
- Improve documentation
- Add tests
- Refactor code
- Update dependencies

**How to make changes:**

1. Edit files in your editor
2. Test your changes (see Testing section)
3. Commit your changes (see Commits section)

#### Step 6: Test Your Changes

**Why test?**

To ensure your changes work and don't break anything.

**How to test:**

```bash
# Run all tests
make test

# Run specific module tests
cd cli && go test ./...
cd core && go test ./...
cd desktop && npm test

# Run with coverage
go test -cover ./...
```

#### Step 7: Commit Your Changes

**What is committing?**

Saving your changes to Git history.

**Commit message format:**

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

**Examples:**

Good:
```
feat: add dark mode to desktop app

Implements dark mode toggle in settings.
Uses system preference by default.

Closes #123
```

Bad:
```
fixed bug
update
changes
```

**How to commit:**

```bash
# Stage changes
git add .

# Commit with message
git commit -m "feat: add user profile page"
```

#### Step 8: Push and Create Pull Request

**What is a Pull Request (PR)?**

A request to merge your changes into the main project.

**How to push:**

```bash
# Push your branch
git push origin feature/your-feature-name
```

**How to create PR:**

1. Go to https://github.com/orhaniscoding/goconnect
2. You'll see "Compare & pull request" button
3. Click it
4. Fill in PR template
5. Click "Create pull request"

**PR Template:**

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How did you test these changes?

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added to complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added/updated
- [ ] All tests pass
```

#### Step 9: Review and Merge

**What happens next?**

1. **Automated checks** - CI runs tests
2. **Code review** - Maintainers review your code
3. **Feedback** - We might request changes
4. **Approval** - Once approved, we merge

**How long does it take?**

Usually 1-7 days, depending on complexity.

---

## 📝 Development Guidelines

### Code Standards

#### Go Code (CLI and Core)

**Formatting:**

```bash
# Format code
go fmt ./...

# Or use golangci-lint
golangci-lint run
```

**Naming conventions:**

```go
// Packages: lowercase, single word
package network

// Constants: PascalCase or UPPER_SNAKE_CASE
const MaxRetries = 3
const API_BASE_URL = "https://api.goconnect.io"

// Variables: camelCase
var userCount int

// Functions: PascalCase (exported), camelCase (private)
func ConnectToServer() {}
func parseResponse() {}

// Interfaces: PascalCase, usually -er suffix
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Structs: PascalCase
type User struct {
    ID       string
    Username string
}
```

**File organization:**

```
cli/
├── cmd/
│   └── goconnect/
│       └── main.go          # Entry point
├── internal/
│   ├── tui/                 # TUI code
│   ├── daemon/              # Daemon code
│   ├── chat/                # Chat logic
│   └── config/              # Config handling
└── pkg/                     # Public packages
    └── api/                 # API client
```

**Comments:**

```go
// Package comment (explains what this package does)
package network

// Comment explains WHY, not WHAT
// Bad: Increment count by 1
// Good: Increment count to track active connections
func incrementCount() {
    count++
}

// Exported functions MUST have comments
// ConnectToServer establishes a connection to the GoConnect server.
// It returns an error if the connection fails.
func ConnectToServer(addr string) error {
    // ...
}
```

#### React/TypeScript Code (Desktop)

**Formatting:**

```bash
cd desktop
npm run format
```

**Naming conventions:**

```typescript
// Components: PascalCase
function UserProfile() {
  // ...
}

// Hooks: camelCase with 'use' prefix
function useUserData() {
  // ...
}

// Variables/Functions: camelCase
const userCount = 0;

function fetchUserData() {
  // ...
}

// Constants: UPPER_SNAKE_CASE
const MAX_RETRIES = 3;

// Interfaces/Types: PascalCase
interface User {
  id: string;
  username: string;
}
```

**File organization:**

```
desktop/
├── src/
│   ├── components/          # Reusable components
│   ├── pages/               # Page components
│   ├── hooks/               # Custom hooks
│   ├── services/            # API calls
│   ├── types/               # TypeScript types
│   └── utils/               # Utility functions
```

#### Rust Code (Desktop - Tauri Backend)

**Formatting:**

```bash
cd desktop/src-tauri
cargo fmt
```

**Naming conventions:**

```rust
// Functions: snake_case
fn connect_to_server() {
    // ...
}

// Types: PascalCase
struct User {
    id: String,
    username: String,
}

// Constants: UPPER_SNAKE_CASE
const MAX_RETRIES: u32 = 3;

// Modules: snake_case
mod network_config;
```

### Testing Guidelines

#### Go Tests

**What to test:**

- Business logic
- Edge cases
- Error handling
- Public API

**How to write tests:**

```go
// File: network_test.go
package network

import (
    "testing"
)

func TestConnectToServer(t *testing.T) {
    // Arrange
    addr := "localhost:8080"

    // Act
    err := ConnectToServer(addr)

    // Assert
    if err != nil {
        t.Errorf("ConnectToServer() error = %v; want nil", err)
    }
}

func TestConnectToServerInvalidAddr(t *testing.T) {
    tests := []struct {
        name    string
        addr    string
        wantErr bool
    }{
        {
            name:    "empty address",
            addr:    "",
            wantErr: true,
        },
        {
            name:    "invalid address",
            addr:    "invalid://address",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ConnectToServer(tt.addr)
            if (err != nil) != tt.wantErr {
                t.Errorf("ConnectToServer() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Test naming:**

- `Test<FunctionName>` - Happy path
- `Test<FunctionName><Scenario>` - Specific scenario
- `Test<FunctionName><ErrorCondition>` - Error case

#### React Tests

**What to test:**

- Component rendering
- User interactions
- State changes
- API calls (mocked)

**How to write tests:**

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders button text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);

    fireEvent.click(screen.getByText('Click me'));

    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('is disabled when loading', () => {
    render(<Button loading>Loading</Button>);

    expect(screen.getByRole('button')).toBeDisabled();
  });
});
```

### Documentation Guidelines

#### Code Comments

**When to add comments:**

- **WHY, not WHAT** - Explain reasoning, not obvious code
- **Public API** - Document exported functions/types
- **Complex logic** - Explain algorithms
- **TODO/FIXME** - Mark temporary work

**Examples:**

Good:
```go
// Argon2id is used instead of bcrypt because it's memory-hard,
// making GPU-based attacks infeasible.
hash, err := argon2id.CreateHash(password)
```

Bad:
```go
// Hash password
hash, err := argon2id.CreateHash(password)
```

#### README/Docs

**When to update docs:**

- New feature added → Update README.md
- API changed → Update API docs
- Breaking change → Update migration guide
- Config option added → Update config reference

**Documentation style:**

- ✅ Clear and concise
- ✅ Include examples
- ✅ Explain "why" and "how"
- ✅ Use consistent formatting
- ❌ Assume technical knowledge
- ❌ Leave out edge cases

---

## 🔍 Finding Things to Work On

### Good First Issues

**What are they?**

Issues labeled `good first issue` are perfect for beginners.

**How to find:**

1. Go to https://github.com/orhaniscoding/goconnect/issues
2. Click "Labels"
3. Select "good first issue"
4. Pick one that interests you

**What you'll learn:**

- How the codebase works
- Our development process
- Git and GitHub workflow

### Help Wanted

**What are they?**

Issues we want help with, but might need more experience.

**How to find:**

1. Go to https://github.com/orhaniscoding/goconnect/issues
2. Click "Labels"
3. Select "help wanted"

**Examples:**

- New features
- Performance improvements
- Documentation
- Testing

### Roadmap

**What's planned?**

See [ROADMAP.md](ROADMAP.md) for upcoming features.

**Can I work on something not in issues?**

Yes! But please:
1. Open an issue first to discuss
2. Wait for approval
3. Then start working

This avoids duplicate work and ensures your PR will be accepted.

---

## 📌 Types of Contributions

### Reporting Bugs

**How to report:**

1. Search existing issues first
2. Use bug report template
3. Provide clear steps to reproduce
4. Include environment details

**Bug Report Template:**

```markdown
**Description**
Clear description of the bug

**To Reproduce**
Steps to reproduce:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected Behavior**
What you expected to happen

**Screenshots**
If applicable, add screenshots

**Environment**
- OS: [e.g. Windows 11]
- GoConnect Version: [e.g. v1.2.0]
- Browser (if desktop app): [e.g. Chrome 120]

**Additional Context**
Add any other context about the problem here
```

### Suggesting Features

**How to suggest:**

1. Check if feature already exists
2. Search existing feature requests
3. Use feature request template
4. Explain the use case

**Feature Request Template:**

```markdown
**Is your feature request related to a problem?**
A clear and concise description of what the problem is.

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
```

### Writing Code

**Before you start coding:**

1. Check if issue is assigned to someone
2. Comment on the issue that you want to work on it
3. Wait for maintainer approval
4. Create a branch from `main`

**While coding:**

1. Follow code standards (see above)
2. Write tests for your changes
3. Update documentation
4. Keep commits atomic (one logical change per commit)
5. Write clear commit messages

**Before submitting PR:**

1. Rebase from upstream `main`
2. Ensure all tests pass
3. Run linters
4. Self-review your changes
5. Update PR description

### Improving Documentation

**Types of documentation:**

- **README** - Main project README
- **API Docs** - API reference (if applicable)
- **Guides** - How-to guides
- **Tutorials** - Step-by-step tutorials
- **Comments** - Code comments

**How to improve docs:**

1. Find confusing or incomplete docs
2. Open issue describing improvement
3. Fork and edit docs
4. Submit PR

**Documentation style:**

See [Documentation Style Guide](docs/en/style-guide.md) (coming soon)

---

## ✅ Pull Request Checklist

Before submitting your PR, ensure:

### Code Quality
- [ ] Code follows project style guidelines
- [ ] No unnecessary comments
- [ ] No commented-out code
- [ ] No console.log or debug statements
- [ ] Proper error handling

### Testing
- [ ] Tests added for new features
- [ ] Tests updated for bug fixes
- [ ] All tests pass locally
- [ ] No test failures in CI

### Documentation
- [ ] README updated (if needed)
- [ ] API docs updated (if needed)
- [ ] Comments added to complex code
- [ ] CHANGELOG.md updated (if breaking change)

### Commits
- [ ] Commit messages follow Conventional Commits
- [ ] Commits are atomic (one change per commit)
- [ ] No merge commits in PR
- [ ] Commit history is clean

### Branch
- [ ] Branch is up-to-date with main
- [ ] Branch name follows convention
- [ ] Branch is not ahead of upstream

---

## 🔄 Pull Request Process

### What Happens After You Submit PR?

#### 1. Automated Checks (CI)

**What runs:**

- Go tests
- React tests
- Linters (golangci-lint, ESLint)
- Code coverage checks
- Build checks

**If checks fail:**

- View failure logs
- Fix issues locally
- Push fixes to branch
- CI runs again automatically

#### 2. Code Review

**Who reviews:**

- Maintainers
- Project experts
- Community members (for now)

**What we look for:**

- Code quality
- Test coverage
- Documentation
- Breaking changes
- Security implications
- Performance impact

**Review outcomes:**

- ✅ **Approved** - Ready to merge
- 🔄 **Changes requested** - Make updates and resubmit
- ❌ **Rejected** - Closing PR (will explain why)

#### 3. Addressing Feedback

**How to address:**

1. Read review comments carefully
2. Ask questions if anything is unclear
3. Make requested changes
4. Push to branch
5. Comment "Ready for review"

**What if you disagree?**

- Explain your reasoning
- Provide evidence/alternatives
- We'll discuss and decide together

#### 4. Merge

**When do we merge?**

- All checks pass
- At least one maintainer approves
- No outstanding objections

**How do we merge?**

- Squash and merge (commits are combined)
- Delete branch after merge
- Update CHANGELOG.md

---

## 🎖️ Recognition

### How Contributors Are Recognized

**Credits:**

- **Contributors list** - In README.md
- **Release notes** - Mentioned in version updates
- **Git history** - Your name in commit log
- **Hall of Fame** - Coming soon to website

**Swag:**

After significant contributions:
- 🎁 GoConnect stickers
- 👕 GoConnect t-shirt
- 🏆 Special badges

**References:**

Can we list you as a reference? Yes! After multiple quality PRs, we'll be happy to serve as a reference for future job opportunities.

---

## ❓ Getting Help

### Where to Ask

**For contribution questions:**

- GitHub Issues: Use "question" label
- GitHub Discussions: https://github.com/orhaniscoding/goconnect/discussions
- Discord: (Coming soon)

**Before asking:**

1. Search existing issues/discussions
2. Read relevant documentation
3. Check if your question is already answered

**How to ask effectively:**

- **Be specific** - Include code, error messages, screenshots
- **Explain what you tried** - Show research effort
- **Provide context** - What are you trying to accomplish?
- **Use code blocks** - Format code properly

**Example:**

Bad:
```
My code doesn't work. Help!
```

Good:
```
I'm trying to add a new button to the settings page following
the CONTRIBUTING.md guide, but I'm getting this error:

TypeError: Cannot read property 'onClick' of undefined

Here's my code:
[paste code]

I've tried:
- Reinstalling dependencies
- Checking for similar buttons in the codebase

Any suggestions?
```

---

## 📜 Code of Conduct

### Our Pledge

In the interest of fostering an open and welcoming environment, we pledge to make participation in our project and our community a harassment-free experience for everyone.

### Our Standards

**Positive behavior:**

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behavior:**

- The use of sexualized language or imagery
- Trolling or insulting/derogatory comments
- Personal or political attacks
- Public or private harassment
- Publishing others' private information
- Other unethical or unprofessional conduct

### Responsibilities

**Project maintainers:**

- Clarify standards of acceptable behavior
- Respond to all reports of harassment
- Take appropriate corrective action

**Participants:**

- Follow the standards
- Report violations to maintainers

### Enforcement

**How to report:**

Email [conduct@goconnect.io](mailto:conduct@goconnect.io)

**What happens:**

1. We investigate the report
2. We determine if violation occurred
3. We take appropriate action (warning, ban, etc.)
4. We report back to reporter

**Confidentiality:**

All reports will be kept confidential.

---

## 🌟 Becoming a Maintainer

### What is a Maintainer?

A maintainer is a trusted contributor with:
- Write access to the repository
- Responsibility to review PRs
- Authority to make project decisions
- Duty to keep the project healthy

### How to Become a Maintainer

**Requirements:**

- Consistent quality contributions (6+ months)
- Deep understanding of codebase
- Active participation in reviews
- Positive community interaction
- Endorsed by current maintainers

**Process:**

1. Contribute consistently over time
2. Show interest in taking more responsibility
3. Current maintainers will discuss internally
4. If consensus, we'll invite you to join
5. You'll start with limited permissions
6. Over time, you'll get full access

**Expectations:**

- Review PRs in your area of expertise
- Triage issues
- Participate in project decisions
- Mentor new contributors
- Follow Code of Conduct

---

## 🔧 Development Tools

### Useful Commands

**Go (CLI and Core):**

```bash
# Build
go build ./cmd/goconnect

# Run tests
go test ./...
go test -v ./...
go test -cover ./...

# Run specific test
go test -run TestConnectToServer

# Benchmark
go test -bench=. -benchmem

# Race detector
go test -race ./...

# Format
go fmt ./...

# Lint
golangci-lint run

# Dependency update
go get -u ./...
go mod tidy

# View dependencies
go mod graph
go mod why <package>
```

**Node.js (Desktop):**

```bash
# Install dependencies
npm install

# Run development server
npm run tauri dev

# Build
npm run tauri build

# Test
npm test

# Lint
npm run lint

# Format
npm run format
```

**Git:**

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# View branches
git branch -a

# View changes
git log
git diff
git status
```

### Recommended VS Code Extensions

**Go development:**
- Go (Google)
- Go Tests Explorer
- golangci-lint

**React/TypeScript:**
- ESLint
- Prettier
- TypeScript Importer
- Auto Rename Tag

**General:**
- GitLens
- GitHub Pull Requests
- Better Comments
- Error Lens

---

## 📚 Learning Resources

### Go Resources

- [A Tour of Go](https://go.dev/tour/welcome/1)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Proverbs](https://go-proverbs.github.io)

### React Resources

- [React Documentation](https://react.dev/)
- [React Tutorial](https://react.dev/learn)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html)

### WireGuard Resources

- [WireGuard Quick Start](https://www.wireguard.com/quickstart/)
- [WireGuard Protocol](https://www.wireguard.com/protocol/)
- [WireGuard Whitepaper](https://www.wireguard.com/papers/wireguard.pdf)

### General Programming

- [Clean Code](https://www.amazon.com/Clean-Code-Handbook-Software-Craftsmanship/dp/0132350882)
- [The Pragmatic Programmer](https://www.amazon.com/Pragmatic-Programmer-Journey-Mastery/dp/020161622X)

---

## 🙏 Thank You

**Every contribution matters!**

Whether you're:
- Fixing a typo
- Reporting a bug
- Writing a feature
- Helping a user
- Translating docs

**You're making GoConnect better for everyone.**

**We appreciate you!** 🎉

---

**Last Updated:** 2025-01-24
**Language:** English
**Version:** 1.0.0

---

## Türkçe

## 📋 Genel Bakış

**Katkıda bulunmak nedir?**

Katkıda bulunmak, GoConnect'i iyileştirmeye yardım etmek demektir. Şunları içerebilir:
- Hata bildirme
- Yeni özellik önerme
- Kod yazma
- Dokümantasyon iyileştirme
- Diğer kullanıcılara yardım etme
- Dokümantasyon çevirisi

**Kim katkıda bulunabilir?**

Herkes! İster acemi ister uzman olun, yardımınız değerli.

**Neden katkıda bulunmalısınız?**

- 🌟 Binlerce kullanıcıya yardım et
- 📚 Yeni beceriler öğren (Go, Rust, React, WireGuard)
- 🏆 Portföyönüzü oluşturun
- 👥 Harika bir topluluğa katılın
- 🎁 Swag alın (stickler, tişört)

---

## 🚀 Hızlı Başlangıç (5 Dakika)

### İlk Kez Katkıda Bulunanlar mı?

Hoş geldiniz! Başlamak için işte rehber:

#### Adım 1: Projeyi Anlayın

**GoConnect nedir?**

GoConnect, 3 ana parçası olan bir sanal LAN platformudur:
- **CLI** - Komut satırı uygulaması (Go)
- **Core** - Sunucu backend'i (Go)
- **Desktop** - Masaüstü uygulaması (Tauri + React)

**Hangi teknolojileri kullanıyoruz?**

| Bileşen | Teknoloji | Neden? |
|-----------|------------|------|
| **CLI** | Go + Bubbletea | Çapraz platform, tek binary |
| **Core** | Go + WireGuard | Hızlı, güvenli ağ |
| **Desktop** | Tauri + React | Küçük, hızlı masaüstü uygulamaları |
| **Protocol** | Protocol Buffers | Tip-güvenli iletişim |

#### Adım 2: Geliştirme Ortamınızı Kurun

**İhtiyacınız olanlar:**

1. **Git** - Sürüm kontrolü
2. **Go 1.24+** - CLI ve Core için
3. **Node.js 20+** - Desktop uygulaması için
4. **Rust** - Tauri için (Desktop uygulaması backend'i)
5. **protoc** - Protocol Buffers derleyicisi
6. **Editör** - VS Code, GoLand veya benzeri

**Nasıl kurulur?**

**Git:**
- Windows: https://git-scm.com/download/win
- macOS: `brew install git`
- Linux: `sudo apt install git`

**Go:**
- İndir: https://go.dev/dl/
- Doğrula: `go version`

**Node.js:**
- İndir: https://nodejs.org/
- Doğrula: `node --version`

**Rust:**
- İndir: https://rustup.rs/
- Doğrula: `rustc --version`

**protoc:**
- macOS: `brew install protobuf`
- Linux: `sudo apt install protobuf-compiler`
- Windows: https://github.com/protocolbuffers/protobuf/releases

#### Adım 3: Fork Edin ve Klonlayın

**Fork nedir?**

Forkleme, GoConnect'in GitHub üzerinde kendi kopyanızı oluşturur.

**Nasıl fork edilir?**

1. https://github.com/orhaniscoding/goconnect adresine gidin
2. "Fork" butonuna tıklayın (sağ üst)
3. Fork tamamlanmasını bekleyin

**Nasıl klonlanır?**

```bash
# Fork'ınızı klonlayın
git clone https://github.com/SIZIN_KULLANICI_ADINIZ/goconnect.git
cd goconnect

# Orijinal repository'yu upstream olarak ekleyin
git remote add upstream https://github.com/orhaniscoding/goconnect.git

# Doğrulayın
git remote -v
```

**Beklenen çıktı:**
```
origin    https://github.com/SIZIN_KULLANICI_ADINIZ/goconnect.git (fetch)
origin    https://github.com/SIZIN_KULLANICI_ADINIZ/goconnect.git (push)
upstream  https://github.com/orhaniscoding/goconnect.git (fetch)
upstream  https://github.com/orhaniscoding/goconnect.git (push)
```

#### Adım 4: Bir Branş Oluşturun

**Branş nedir?**

Branş, değişikliklerinizi yaptığınız kodun ayrı bir versiyonudur.

**Nasıl oluşturulur?**

```bash
# Upstream'dan güncelleyin
git fetch upstream

# Main'den branş oluşturun
git checkout -b feature/feature-adiniz

# Veya bug fix için
git checkout -b fix/bug-fix-adiniz
```

**Branş adlandırma:**

- `feature/` - Yeni özellikler
- `fix/` - Bug düzeltmeleri
- `docs/` - Dokümantasyon değişiklikleri
- `refactor/` - Kod refactor'ı
- `test/` - Test ekleme

#### Adım 5: Değişikliklerinizi Yapın

**Ne değiştirebilirsiniz?**

- Yeni özellikler ekleyin
- Bugları düzeltin
- Dokümantasyonu iyileştirin
- Testler ekleyin
- Kodu refactor edin
- Bağımlılıkları güncelleyin

**Nasıl değişiklik yapılır?**

1. Editörünüzde dosyaları düzenleyin
2. Değişikliklerinizi test edin (Testing bölümüne bakın)
3. Değişikliklerinizi commit edin (Commits bölümüne bakın)

#### Adım 6: Değişikliklerinizi Test Edin

**Neden test etmeliyiz?**

Değişikliklerinizin çalıştığından ve hiçbir şeyi bozmadığından emin olmak için.

**Nasıl test edilir?**

```bash
# Tüm testleri çalıştır
make test

# Belirli modül testlerini çalıştır
cd cli && go test ./...
cd core && go test ./...
cd desktop && npm test

# Coverage ile çalıştır
go test -cover ./...
```

#### Adım 7: Değişikliklerinizi Commit Edin

**Commit etmek nedir?**

Değişikliklerinizi Git geçmişine kaydetmek.

**Commit mesajı formatı:**

[Conventional Commits](https://www.conventionalcommits.org/) kullanıyoruz:

```
<tip>: <açıklama>

[isteğe bağlı gövde]

[isteğe bağlı alt bilgi]
```

**Tipler:**
- `feat:` - Yeni özellik
- `fix:` - Bug düzeltmesi
- `docs:` - Dokümantasyon değişiklikleri
- `style:` - Kod stili değişiklikleri (formatlama)
- `refactor:` - Kod refactor'ı
- `test:` - Test ekleme veya güncelleme
- `chore:` - Bakım görevleri

**Örnekler:**

İyi:
```
feat: desktop uygulamasına karanlık mod ekle

Ayarlar bölümünde karanlık mod toggle'ı uygular.
Varsayılan olarak sistem tercihini kullanır.

Closes #123
```

Kötü:
```
bug düzeltildi
güncelleme
değişiklikler
```

**Nasıl commit edilir?**

```bash
# Değişiklikleri hazırla
git add .

# Mesajla commit et
git commit -m "feat: kullanıcı profil sayfası ekle"
```

#### Adım 8: Push Edin ve Pull Request Oluşturun

**Pull Request (PR) nedir?**

Değişikliklerinizi ana projeye birleştirme isteği.

**Nasıl push edilir?**

```bash
# Branşinizi push edin
git push origin feature/feature-adiniz
```

**Nasıl PR oluşturulur?**

1. https://github.com/orhaniscoding/goconnect adresine gidin
2. "Compare & pull request" butonunu göreceksiniz
3. Tıklayın
4. PR şablonunu doldurun
5. "Create pull request"e tıklayın

**PR Şablonu:**

```markdown
## Açıklama
Değişikliklerin kısa açıklaması

## Değişiklik Tipi
- [ ] Bug fix
- [ ] Yeni özellik
- [ ] Breaking change
- [ ] Dokümantasyon güncellemesi

## Testler
Bu değişiklikleri nasıl test ettiniz?

## Kontrol Listesi
- [ ] Kod stil yönergelerine uyuyor
- [ ] Self-review tamamlandı
- [ ] Karmaşık koda yorumlar eklendi
- [ ] Dokümantasyon güncellendi
- [ ] Yeni uyarı oluşturulmadı
- [ ] Testler eklendi/güncellendi
- [ ] Tüm testler geçiyor
```

#### Adım 9: Review ve Birleştirme

**Sırada ne olacak?**

1. **Otomatik kontroller** - CI testleri çalıştırır
2. **Kod incelemesi** - Maintainer'lar kodunuzu inceler
3. **Geribildirim** - Değişiklik isteyebiliriz
4. **Onay** - Onaylandıktan sonra birleştiririz

**Ne kadar sürer?**

Genellikle 1-7 gün, karmaşıklığa bağlı.

---

## 📝 Geliştirme Yönergeleri

### Kod Standartları

#### Go Kodu (CLI ve Core)

**Formatlama:**

```bash
# Kodu formatla
go fmt ./...

# Veya golangci-lint kullan
golangci-lint run
```

**Adlandırma kuralları:**

```go
// Paketler: küçük harf, tek kelime
package network

// Sabitler: PascalCase veya UPPER_SNAKE_CASE
const MaxRetries = 3
const API_BASE_URL = "https://api.goconnect.io"

// Değişkenler: camelCase
var userCount int

// Fonksiyonlar: PascalCase (dışa aktarılan), camelCase (özel)
func ConnectToServer() {}
func parseResponse() {}

// Arayüzler: PascalCase, genellikle -er soneki
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Struct'lar: PascalCase
type User struct {
    ID       string
    Username string
}
```

**Dosya organizasyonu:**

```
cli/
├── cmd/
│   └── goconnect/
│       └── main.go          # Giriş noktası
├── internal/
│   ├── tui/                 # TUI kodu
│   ├── daemon/              # Daemon kodu
│   ├── chat/                # Sohbet mantığı
│   └── config/              # Config yönetimi
└── pkg/                     # Herkese açık paketler
    └── api/                 # API istemcisi
```

**Yorumlar:**

```go
// Paket yorumu (bu paketin ne yaptığını açıklar)
package network

// Yorum NEDENİ açıklar, NEYİ değil
// Kötü: Sayacı 1 artır
// İyi: Aktif bağlantıları takip etmek için sayacı artır
func incrementCount() {
    count++
}

// Dışa aktarılan fonksiyonlar YORUM ZORUNLU
// ConnectToServer GoConnect sunucusuna bağlantı kurar.
// Bağlantı başarısız olursa hata döndürür.
func ConnectToServer(addr string) error {
    // ...
}
```

#### React/TypeScript Kodu (Desktop)

**Formatlama:**

```bash
cd desktop
npm run format
```

**Adlandırma kuralları:**

```typescript
// Komponentler: PascalCase
function UserProfile() {
  // ...
}

// Hook'lar: 'use' öneki ile camelCase
function useUserData() {
  // ...
}

// Değişkenler/Fonksiyonlar: camelCase
const userCount = 0;

function fetchUserData() {
  // ...
}

// Sabitler: UPPER_SNAKE_CASE
const MAX_RETRIES = 3;

// Arayüzler/Tipler: PascalCase
interface User {
  id: string;
  username: string;
}
```

**Dosya organizasyonu:**

```
desktop/
├── src/
│   ├── components/          # Yeniden kullanılabilir komponentler
│   ├── pages/               # Sayfa komponentleri
│   ├── hooks/               # Özel hook'lar
│   ├── services/            # API çağrıları
│   ├── types/               # TypeScript tipleri
│   └── utils/               # Yardımcı fonksiyonlar
```

#### Rust Kodu (Desktop - Tauri Backend)

**Formatlama:**

```bash
cd desktop/src-tauri
cargo fmt
```

**Adlandırma kuralları:**

```rust
// Fonksiyonlar: snake_case
fn connect_to_server() {
    // ...
}

// Tipler: PascalCase
struct User {
    id: String,
    username: String,
}

// Sabitler: UPPER_SNAKE_CASE
const MAX_RETRIES: u32 = 3;

// Modüller: snake_case
mod network_config;
```

### Test Yönergeleri

#### Go Testleri

**Ne test edilmeli?**

- İş mantığı
- Kenar durumlar
- Hata yönetimi
- Herkese açık API

**Nasıl yazılır?**

```go
// Dosya: network_test.go
package network

import (
    "testing"
)

func TestConnectToServer(t *testing.T) {
    // Hazırlık
    addr := "localhost:8080"

    // Uygulama
    err := ConnectToServer(addr)

    // Doğrulama
    if err != nil {
        t.Errorf("ConnectToServer() error = %v; want nil", err)
    }
}

func TestConnectToServerInvalidAddr(t *testing.T) {
    tests := []struct {
        name    string
        addr    string
        wantErr bool
    }{
        {
            name:    "boş adres",
            addr:    "",
            wantErr: true,
        },
        {
            name:    "geçersiz adres",
            addr:    "invalid://address",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ConnectToServer(tt.addr)
            if (err != nil) != tt.wantErr {
                t.Errorf("ConnectToServer() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Test adlandırma:**

- `Test<FonksiyonAdı>` - İyi durum
- `Test<FonksiyonAdı><Senaryo>` - Özel senaryo
- `Test<FonksiyonAdı><HataDurumu>` - Hata durumu

#### React Testleri

**Ne test edilmeli?**

- Komponent render'ı
- Kullanıcı etkileşimleri
- Durum değişiklikleri
- API çağrıları (mock edilmiş)

**Nasıl yazılır?**

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('buton metnini render eder', () => {
    render(<Button>Bana tıkla</Button>);
    expect(screen.getByText('Bana tıkla')).toBeInTheDocument();
  });

  it('tıklandığında onClick çağırır', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Bana tıkla</Button>);

    fireEvent.click(screen.getByText('Bana tıkla'));

    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('yükleme durumunda devre dışı bırakılır', () => {
    render(<Button loading>Yükleniyor</Button>);

    expect(screen.getByRole('button')).toBeDisabled();
  });
});
```

### Dokümantasyon Yönergeleri

#### Kod Yorumları

**Ne zaman yorum eklenmeli?**

- **NEDEN, ne DEĞİL** - Mantığı açıklayın, açık kodu değil
- **Herkese açık API** - Dışa aktarılan fonksiyon/tipleri belgeleyin
- **Karmaşık mantık** - Algoritmaları açıklayın
- **TODO/FIXME** - Geçici çalışmaları işaretleyin

**Örnekler:**

İyi:
```go
// Argon2id bcrypt yerine kullanılır çünkü bellek-ağırdır,
// GPU tabanlı saldırıları imkansız hale getirir.
hash, err := argon2id.CreateHash(password)
```

Kötü:
```go
// Şifreyi hashle
hash, err := argon2id.CreateHash(password)
```

#### README/Dokümanlar

**Ne zaman güncellenmeli?**

- Yeni özellik eklendi → README.md güncelle
- API değişti → API dokümanlarını güncelle
- Breaking change → Migration guide güncelle
- Config seçeneği eklendi → Config referansı güncelle

**Dokümantasyon stili:**

- ✅ Açık ve öz
- ✅ Örnekler içer
- ✅ "Neden" ve "Nasıl" açıklar
- ✅ Tutarlı formatlama kullan
- ❌ Teknik bilgi varsayar
- ❌ Kenar durumları dışarıda bırakır

---

## 🔍 Üzerinde Çalışılacak Şeyleri Bulma

### İyi Başlangıç İçin Sorunlar

**Ne oldukları?**

"good first issue" etiketli sorunlar acemiler için mükemmeldir.

**Nasıl bulunur?**

1. https://github.com/orhaniscoding/goconnect/issues adresine gidin
2. "Labels"e tıklayın
3. "good first issue" seçin
4 İlginizi birini seçin

**Ne öğreneceksiniz?**

- Kod tabanının nasıl çalıştığını
- Geliştirme sürecimizi
- Git ve GitHub iş akışı

### Yardım İstenen

**Ne oldukları?**

Yardım istediğimiz ama daha fazla deneyim gerektiren sorunlar.

**Nasıl bulunur?**

1. https://github.com/orhaniscoding/goconnect/issues adresine gidin
2. "Labels"e tıklayın
3. "help wanted" seçin

**Örnekler:**

- Yeni özellikler
- Performans iyileştirmeleri
- Dokümantasyon
- Test

### Yol Haritası

**Ne planlanıyor?**

Yaklaşan özellikler için [ROADMAP.md](ROADMAP.md) dosyasına bakın.

**Sorun olmayan bir şey üzerinde çalışabilir miyim?**

Evet! Ama lütfen:
1. Önce tartışmak için issue açın
2. Onay bekleyin
3. Sonra çalışmaya başlayın

Bu, duplicate çalışmayı önler ve PR'inizin kabul edilmesini sağlar.

---

## 📌 Katkı Türleri

### Bug Bildirme

**Nasıl bildirilir?**

1. Önce mevcut issue'lara bakın
2. Bug raporu şablonunu kullanın
3. Net üretim adımları sağlayın
4. Ortam detaylarını ekleyin

**Bug Raporu Şablonu:**

```markdown
**Açıklama**
Bugin net açıklaması

**Yeniden Üretme**
Adımlar:
1. '...'a gidin
2. '....'e tıklayın
3. '....'e kadar aşağı kaydırın
4. Hata görün

**Beklenen Davranış**
Ne olmasını beklediniz

**Ekran Görüntüleri**
Mümkünse, ekran görüntüleri ekleyin

**Ortam**
- OS: [örn. Windows 11]
- GoConnect Sürümü: [örn. v1.2.0]
- Tarayıcı (desktop uygulamasıysa): [örn. Chrome 120]

**Ek Bağlam**
Sorun hakkında başka bağlam veya ekran görüntüleri buraya ekleyin
```

### Özellik Önerme

**Nasıl önerilir?**

1. Özelliğin zaten var olup olmadığını kontrol edin
2. Mevcut özellik taleplerini arayın
3. Özellik talebi şablonunu kullanın
4. Kullanım durumunu açıklayın

**Özellik Talebi Şablonu:**

```markdown
**Bu özellik talebi bir sorunla ilgili mi?**
Sorunun net ve öz açıklaması

**İstediğiniz çözümü açıklayın**
Ne olmak istediğinizin net ve öz açıklaması

**Düşündüğünüz alternatifleri açıklayın**
Düşündüğünüz diğer çözümler veya özelliklerin net açıklaması

**Ek Bağlam**
Özellik talebi hakkında başka bağlam veya ekran görüntüleri buraya ekleyin
```

### Kod Yazma

**Kod yazmaya başlamadan:**

1. Issue başkasına atanmış mı kontrol edin
2. Üzerinde çalışmak istediğiniz issue'ye yorum yapın
3. Maintainer onayı bekleyin
4. `main`'den bir branş oluşturun

**Kod yazarken:**

1. Kod standartlarını takip edin (yukarıya bakın)
2. Değişiklikleriniz için test yazın
3. Dokümantasyonu güncelleyin
4. Commit'leri atomik tutun (her commit'te bir mantıksal değişiklik)
5. Açık commit mesajları yazın

**PR göndermeden önce:**

1. Upstream `main`'den rebase edin
2. Tüm testlerin geçtiğinden emin olun
3. Linter'ları çalıştırın
4. Değişikliklerinizi self-review edin
5. PR açıklamasını güncelleyin

### Dokümantasyon İyileştirme

**Dokümantasyon türleri:**

- **README** - Ana proje README
- **API Docs** - API referansı (uyguluyorsa)
- **Rehberler** - Nasıl yapılır rehberleri
- **Tutorial'lar** - Adım adım tutorial'lar
- **Yorumlar** - Kod yorumları

**Nasıl iyileştirilir?**

1. Kafa karıştırık veya eksik dokümanları bulun
2. İyileştirmeyi açıklayan issue açın
3. Fork edin ve dokümanları düzenleyin
4. PR gönderin

**Dokümantasyon stili:**

[Dokümantasyon Stil Rehberi](docs/tr/style-guide.md) (yakında gelecek)

---

## ✅ Pull Request Kontrol Listesi

PR'nizi göndermeden önce şunları sağlayın:

### Kod Kalitesi
- [ ] Kod proje stil yönergelerine uyuyor
- [ ] Gereksiz yorumlar yok
- [ ] Yorumlanmış kod yok
- [ ] console.log veya debug ifadeleri yok
- [ ] Uygun hata yönetimi

### Test
- [ ] Yeni özellikler için test eklendi
- [ ] Bug düzeltmeleri için test güncellendi
- [ ] Tüm testler yerelde geçiyor
- [ ] CI'de test hatası yok

### Dokümantasyon
- [ ] README güncellendi (gerekliyse)
- [ ] API dokümanları güncellendi (gerekliyse)
- [ ] Karmaşık koda yorumlar eklendi
- [ ] CHANGELOG.md güncellendi (breaking change ise)

### Commit'ler
- [ ] Commit mesajları Conventional Commits takip ediyor
- [ ] Commit'ler atomik (her commit'te bir değişiklik)
- [ ] PR'de merge commit yok
- [ ] Commit geçmişi temiz

### Branş
- [ ] Branş main ile güncel
- [ ] Branş adı kurala uygun
- [ ] Branş upstream'dan ileride değil

---

## 🔄 Pull Request Süreci

### PR Gönderdikten Sonra Ne Olur?

#### 1. Otomatik Kontroller (CI)

**Ne çalışır?**

- Go testleri
- React testleri
- Linter'lar (golangci-lint, ESLint)
- Code coverage kontrolleri
- Build kontrolleri

**Kontroller başarısız olursa:**

- Hata loglarını görüntüleyin
- Sorunları yerel olarak düzeltin
- Düzeltmeleri branşa push edin
- CI otomatik olarak tekrar çalışır

#### 2. Kod İncelemesi

**Kim inceler?**

- Maintainer'lar
- Proje uzmanları
- Topluluk üyeleri (şimdilik)

**Ne arıyoruz?**

- Kod kalitesi
- Test coverage
- Dokümantasyon
- Breaking değişiklikler
- Güvenlik etkileri
- Performans etkisi

**İnceleme sonuçları:**

- ✅ **Onaylandı** - Birleştirilmeye hazır
- 🔄 **Değişiklik istendi** - Güncelleyin ve yeniden gönderin
- ❌ **Reddedildi** - PR kapatılıyor (nedenini açıklayacağız)

#### 3. Geribildirimle Uğraşma

**Nasıl uğraşırsınız?**

1. İnceleme yorumlarını dikkatlice okuyun
2. Bir şey belirsizse sorun sorun
3. İstenen değişiklikleri yapın
4. Branşa push edin
5. "Review için hazır" yorumu yapın

**Katılmıyorsanız ne yapmalısınız?**

- Nedeninizi açıklayın
- Kanıt/alternatifler sağlayın
- Tartışacağız ve birlikte karar vereceğiz

#### 4. Birleştirme

**Ne zaman birleştiriyoruz?**

- Tüm kontroller geçerse
- En az bir maintainer onaylarsa
- Sürmeyen karşı itiraz yoksa

**Nasıl birleştiriyoruz?**

- Squash and merge (commit'ler birleştirilir)
- Birleştirmeden sonra branş silinir
- CHANGELOG.md güncellenir

---

## 🎖️ Takdir

### Katkıda Bulunanlar Nasıl Takdir Edilir?

**Krediler:**

- **Katkıda bulunanlar listesi** - README.md'de
- **Sürüm notları** - Sürüm güncellemelerinde bahsedilir
- **Git geçmişi** - Commit logunda isminiz
- **Onur Listesi** - Yakında web sitesinde

**Swag:**

Önemli katkılardan sonra:
- 🎁 GoConnect stickelleri
- 👕 GoConnect tişörtü
- 🏆 Özel rozetler

**Referanslar:**

Sizi referans olarak listeleyebilir miyiz? Evet! Kaliteli birkaç PR'den sonra, gelecekteki iş fırsatları için referans olarak hizmet etmemekten mutluluk duyarız.

---

## ❓ Yardım Alma

### Nereye Sormalı

**Katkı soruları için:**

- GitHub Issues: "question" etiketini kullanın
- GitHub Discussions: https://github.com/orhaniscoding/goconnect/discussions
- Discord: (Çok yakında)

**Sormadan önce:**

1. Mevcut issue'ları/discussion'ları arayın
2. İlgili dokümantasyonu okuyun
3. Sorunuzun zaten cevaplanıp olmadığını kontrol edin

**Etkili nasıl sorulur?**

- **Spesifik olun** - Kod, hata mesajları, ekran görüntüleri dahil edin
- **Ne denediğinizi açıklayın** - Araştırma çabanızı gösterin
- **Bağlam sağlayın** - Neyi başarmaya çalışıyorsunuz?
- **Kod blokları kullanın** - Kodu düzgün formatlayın

**Örnek:**

Kötü:
```
Kodum çalışmıyor. Yardım!
```

İyi:
```
CONTRIBUTING.md rehberini takip ederek ayarlar sayfasına yeni bir
button eklemeye çalışıyorum ama şu hatayı alıyorum:

TypeError: Cannot read property 'onClick' of undefined

Kodum:
[kod yapıştır]

Şunları denedim:
- Bağımlılıkları yeniden yükledim
- Kod tabanında benzer button'ları aradım

Öneriniz var mı?
```

---

## 📜 Davranış Kuralları

### Sözümüz

Açık ve kapsayıcı bir ortam teşvik etmek için, projemize ve topluluğumuzda katılımda herkes için taciz deneyimi yaşama sözü veriyoruz.

### Standartlarımız

**Pozitif davranış:**

- Kapsayıcı ve hoşgörülü dil kullanmak
- Farklı görüş ve deneyimlere saygılı olmak
- Yapıcı eleştiriyi zarif karşılamak
- Topluluk için en iyiyi odaklanmak
- Diğer topluluk üyelerine empati göstermek

**Kabul edilemez davranış:**

- Cinsel dil veya görüntü kullanımı
- Trollleme veya aşağılayıcı/eleştirel yorumlar
- Kişisel veya siyasi saldırılar
- Herkese açık veya özel taciz
- Başkalarının özel bilgilerini yayınlamak
- Diğer gayri profesyonel davranışlar

### Sorumluluklar

**Proje sahipleri:**

- Kabul edilebilir davranış standartlarını netleştirir
- Tüm taciz raporlarını yanıtlar
- Uygun düzeltici eylem alır

**Katılımcılar:**

- Standartları takip eder
- İhlalleri maintainer'lara bildirir

### Uygulama

**Nasıl raporlanır?**

[conduct@goconnect.io](mailto:conduct@goconnect.io) adresine e-posta gönderin

**Ne olur?**

1. Raporu inceleriz
2. İhlal olup olmadığına karar veririz
3. Uygun düzeltici eylem yaparız (uyarı, yasaklama vb.)
4. Raporlayana geri bildirimde bulunuruz

**Gizlilik:**

Tüm raporlar gizli tutulacaktır.

---

## 🌟 Maintainer Olmak

### Maintainer Nedir?

Maintainer, şunlara sahip güvenilir bir katkıda bulunandır:
- Repository için yazma erişimi
- PR'leri inceleme sorumluluğu
- Proje kararları verme yetkisi
- Projeyi sağlı tutma görevi

### Maintainer Nasıl Olunur?

**Gereksinimler:**

- Tutarlı kaliteli katkılar (6+ ay)
- Kod tabanının derinlemesi
- İncelemelerde aktif katılım
- Pozitif topluluk etkileşimi
- Mevcut maintainer'lar tarafından desteklenme

**Süreç:**

1. Zaman içinde tutarlı katkıda bulunun
2. Daha fazla sorumluluk alma isteği gösterin
3. Mevcut maintainer'lar dahili olarak tartışır
4. Konsensus varsa, sizi katılmaya davet ederiz
5. Sınırlı izinlerle başlarsınız
6. Zamanla tam erişim elde edersiniz

**Beklentiler:**

- Uzmanlık alanınızdaki PR'leri inceleyin
- Issue'leri triaj edin
- Proje kararlarına katılın
- Yeni katkıda bulunanlara mentorluk yapın
- Davranış kurallarına uyun

---

## 🔧 Geliştirme Araçları

### Yararlı Komutlar

**Go (CLI ve Core):**

```bash
# Build
go build ./cmd/goconnect

# Testleri çalıştır
go test ./...
go test -v ./...
go test -cover ./...

# Belirli testi çalıştır
go test -run TestConnectToServer

# Benchmark
go test -bench=. -benchmem

# Race detector
go test -race ./...

# Format
go fmt ./...

# Lint
golangci-lint run

# Bağımlılık güncelleme
go get -u ./...
go mod tidy

# Bağımlılıkları görüntüle
go mod graph
go mod why <paket>
```

**Node.js (Desktop):**

```bash
# Bağımlılıkları yükle
npm install

# Geliştirme sunucusunu çalıştır
npm run tauri dev

# Build
npm run tauri build

# Test
npm test

# Lint
npm run lint

# Format
npm run format
```

**Git:**

```bash
# Upstream ile senkronize
git fetch upstream
git checkout main
git merge upstream/main

# Branşları görüntüle
git branch -a

# Değişiklikleri görüntüle
git log
git diff
git status
```

### Önerilen VS Code Eklentileri

**Go geliştirme:**
- Go (Google)
- Go Tests Explorer
- golangci-lint

**React/TypeScript:**
- ESLint
- Prettier
- TypeScript Importer
- Auto Rename Tag

**Genel:**
- GitLens
- GitHub Pull Requests
- Better Comments
- Error Lens

---

## 📚 Öğrenme Kaynakları

### Go Kaynakları

- [A Tour of Go](https://go.dev/tour/welcome/1)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Proverbs](https://go-proverbs.github.io)

### React Kaynakları

- [React Documentation](https://react.dev/)
- [React Tutorial](https://react.dev/learn)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html)

### WireGuard Kaynakları

- [WireGuard Quick Start](https://www.wireguard.com/quickstart/)
- [WireGuard Protocol](https://www.wireguard.com/protocol/)
- [WireGuard Whitepaper](https://www.wireguard.com/papers/wireguard.pdf)

### Genel Programlama

- [Clean Code](https://www.amazon.com/Clean-Code-Handbook-Software-Craftsmanship/dp/0132350882)
- [The Pragmatic Programmer](https://www.amazon.com/Pragmatic-Programmer-Journey-Mastery/dp/020161622X)

---

## 🙏� Teşekkürler

**Her katkı önemlidir!**

İster şunları yapıyor olun:
- Bir yazım hatası düzeltmek
- Bir bug bildirmek
- Bir özellik yazmak
- Bir kullanıcıya yardım etmek
- Dokümantasyon çevirmek

**Siz GoConnect'i herkes için daha iyi hale getiriyorsunuz.**

**Size minnettarız!** 🎉

---

**Son Güncelleme:** 2025-01-24
**Dil:** Türkçe
**Sürüm:** 1.0.0
