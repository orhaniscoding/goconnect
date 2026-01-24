# 🛠️ Geliştirme Ortamı Kurulumu

GoConnect projeye katkıda bulunmak veya yerel olarak geliştirmek için gerekli tüm adımlar.

---

## 📑 İçindekiler

- [Genel Bakış](#genel-bakış)
- [Sistem Gereksinimleri](#sistem-gereksinimleri)
- [Geliştirme Ortamı Kurulumu](#geliştirme-ortamı-kurulumu)
- [Projeyi Clone Etme](#projeyi-clone-etme)
- [Bileşenleri Kurma](#bileşenleri-kurma)
- [Çalıştırma](#çalıştırma)
- [Test Etme](#test-etme)
- [Debugging](#debugging)

---

## 👁️ Genel Bakış

GoConnect üç ana bileşenden oluşur:

```
goconnect/
├── cli/           # Terminal uygulaması (Go)
├── core/          # Backend sunucu (Go)
└── desktop/       # Desktop uygulaması (Tauri + React)
```

Her bileşen ayrı ayrı geliştirilebilir ve çalıştırılabilir.

---

## 💻 Sistem Gereksinimleri

### Minimum Gereksinimler

| Araç | Minimum Sürüm | Önerilen |
|------|---------------|----------|
| **Go** | 1.24+ | En son stabil |
| **Node.js** | 20+ | LTS sürümü |
| **Rust** | 1.70+ | En son stabil |
| **Git** | 2.30+ | En son stabil |

### Platform-Specific Gereksinimler

#### Windows
- Windows 10/11 (64-bit)
- [WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) (genellikle yüklü)
- [Visual Studio Build Tools](https://visualstudio.microsoft.com/downloads/) (C++ araçları)
- [WiX Toolset](https://wixtoolset.org/) (MSI installer için)

#### macOS
- macOS 11+ (Big Sur)
- Xcode Command Line Tools:
  ```bash
  xcode-select --install
  ```

#### Linux (Ubuntu/Debian)
```bash
# Temel bağımlılıklar
sudo apt update
sudo apt install -y \
  build-essential \
  git \
  curl \
  wget \
  pkg-config \
  libssl-dev \
  libwebkit2gtk-4.1-dev \
  libappindicator3-dev \
  librsvg2-dev \
  libcairo2-dev \
  libpango1.0-dev \
  libgdk-pixbuf2.0-dev
```

---

## 🚀 Geliştirme Ortamı Kurulumu

### 1. Go Kurulumu

**Linux/macOS:**
```bash
# İndirin ve kurun
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# PATH'e ekleyin (~/.bashrc veya ~/.zshrc)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# Doğrulayın
go version
```

**Windows:**
- [Go indir](https://go.dev/dl/)
- MSI installer ile kurun
- PATH otomatik eklenir

### 2. Node.js Kurulumu

**Node Version Manager (NVM) kullanarak (Önerilen):**

```bash
# NVM kurun
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash

# Yeni terminal açın ve Node.js kurun
nvm install 20
nvm use 20

# Doğrulayın
node --version
npm --version
```

**Windows:**
- [nvm-windows](https://github.com/coreybutler/nvm-windows) indirin
```powershell
nvm install 20
nvm use 20
```

### 3. Rust Kurulumu

```bash
# rustup ile kurun
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# PATH'i yenileyin
source $HOME/.cargo/env

# Doğrulayın
rustc --version
cargo --version
```

### 4. Diğer Araçlar

**Protocol Buffers Compiler (protoc):**

```bash
# Ubuntu/Debian
sudo apt install -y protobuf-compiler

# macOS
brew install protobuf

# Windows
# İndirin: https://github.com/protocolbuffers/protobuf/releases
```

**WireGuard Tools (Linux/macOS):**

```bash
# Ubuntu/Debian
sudo apt install -y wireguard-tools

# macOS
brew install wireguard-tools
```

---

## 📥 Projeyi Clone Etme

```bash
# Fork'layın ve clone edın
git clone https://github.com/YOUR_USERNAME/goconnect.git
cd goconnect

# Upstream remote ekleyin
git remote add upstream https://github.com/orhaniscoding/goconnect.git

# Branch'iniz kontrol edin
git checkout -b feature/your-feature-name
```

---

## 🔧 Bileşenleri Kurma

### CLI (Go) Kurulumu

```bash
cd cli

# Dependencies
go mod download

# Development tools
go install github.com/cosmtrek/air@latest  # Hot reload
go install golang.org/x/tools/cmd/goimports@latest

# Kurulumu doğrulayın
go version
go list -m all
```

### Core (Server) Kurulumu

```bash
cd core

# Dependencies
go mod download

# Database (SQLite development için yeterli)
# Production için PostgreSQL kurun:
# sudo apt install postgresql postgresql-contrib

# Migrations aracı
go install github.com/pressly/goose/v3/cmd/goose@latest

# Kurulumu doğrulayın
go version
```

### Desktop (Tauri + React) Kurulumu

```bash
cd desktop

# Node dependencies
npm install

# Development tools
npm install -g vite  # Fast dev server

# Kurulumu doğrulayın
npm list --depth=0
```

---

## ▶️ Çalıştırma

### CLI Development Mode

```bash
cd cli

# Hot reload ile çalıştır (air)
air

# Veya manuel
go run ./cmd/goconnect
```

**CLI Komutları:**
```bash
# TUI başlat
goconnect

# Hızlı komutlar
goconnect create "Test Network"
goconnect join gc://invite.goconnect.io/abc123
goconnect list
goconnect status
```

### Core (Server) Development Mode

```bash
cd core

# Environment variables
export JWT_SECRET="dev-secret-change-in-production"
export DATABASE_URL="sqlite:///tmp/goconnect.db"
export HTTP_PORT="8080"
export LOG_LEVEL="debug"

# Migrations çalıştır
goose -dir migrations/sqlite sqlite "tmp/goconnect.db" up

# Server'ı başlat
go run ./cmd/server

# Veya hot reload ile
air
```

**API Test:**
```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!"}'
```

### Desktop Development Mode

```bash
cd desktop

# Development server başlat
npm run tauri dev

# Bu komut:
# 1. Vite dev server başlatır (localhost:1420)
# 2. Tauri backend başlatır
# 3. Uygulamayı otomatik yeniler (hot reload)
```

**Desktop Hızlı Başlat:**
```bash
# Vite only (frontend only)
npm run dev

# Tauri build (production)
npm run tauri build
```

---

## 🧪 Test Etme

### Unit Tests

**CLI:**
```bash
cd cli

# Tüm testler
go test ./...

# Coverage ile
go test -cover ./...

# Verbose
go test -v ./internal/daemon

# Specific package
go test -run TestCreateNetwork ./internal/network
```

**Core:**
```bash
cd core

# Tüm testler
go test ./...

# Integration tests
go test -tags=integration ./...

# Benchmark
go test -bench=. -benchmem
```

**Desktop:**
```bash
cd desktop

# Unit tests
npm test

# Type checking
npm run typecheck

# Lint
npm run lint

# Format check
npm run format:check
```

### E2E Tests

```bash
# Desktop E2E (Playwright)
cd desktop
npm run test:e2e

# Manuel E2E test
# 1. CLI'yi başlat (terminal 1)
cd cli && go run ./cmd/goconnect

# 2. Desktop'ı başlat (terminal 2)
cd desktop && npm run tauri dev

# 3. Manually test all features
```

---

## 🐛 Debugging

### VS Code Kurulumu

**.vscode/launch.json:**

```json
{
  "version": "0.3.0",
  "configurations": [
    {
      "name": "Launch CLI",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cli/cmd/goconnect",
      "env": {
        "LOG_LEVEL": "debug"
      },
      "args": ["dev"]
    },
    {
      "name": "Launch Core Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/core/cmd/server",
      "env": {
        "JWT_SECRET": "dev-secret",
        "DATABASE_URL": "sqlite:///tmp/goconnect.db",
        "LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Launch Desktop",
      "type": "node",
      "request": "launch",
      "cwd": "${workspaceFolder}/desktop",
      "runtimeExecutable": "npm",
      "runtimeArgs": ["run", "tauri", "dev"],
      "console": "integratedTerminal"
    }
  ]
}
```

### Debugging Tips

**Go (CLI/Core):**
- Breakpoint koyun ve F5'e basın
- `dbg` paketini kullanın:
  ```go
  import "github.com/fatih/color"
  color.Red("DEBUG: %+v", variable)
  ```

**Desktop (React):**
- Browser DevTools'u kullanın (Ctrl+Shift+I)
- React DevTools eklentisini kurun
- Console.log yerine:
  ```javascript
  console.debug('Debug info:', data);
  ```

**Tauri Backend (Rust):**
- Rust Analyzer eklentisini kullanın
- `println!` veya `log::debug!` kullanın
  ```rust
  log::debug!("Debug info: {:?}", variable);
  ```

---

## 📝 Geliştirme İpuçları

### Kod Kalitesi

**Go:**
```bash
# Format
go fmt ./...
goimports -w .

# Lint
golangci-lint run

# Tidy dependencies
go mod tidy
```

**JavaScript/TypeScript:**
```bash
# Format
npm run format

# Lint
npm run lint

# Type check
npm run typecheck
```

### Verimli Çalışma Akışı

**1. Feature Branch Oluştur:**
```bash
git checkout -b feature/add-network-rename
```

**2. Development Loop:**
```bash
# Terminal 1: CLI (hot reload)
cd cli && air

# Terminal 2: Desktop (hot reload)
cd desktop && npm run tauri dev

# Terminal 3: Tests (watch mode)
cd cli && go test -watch ./...
```

**3. Commit Conventions:**
```bash
# Feature
git commit -m "feat(network): add network rename functionality"

# Bug fix
git commit -m "fix(cli): resolve crash on empty network name"

# Docs
git commit -m "docs(readme): update installation instructions"

# Breaking change
git commit -m "feat(api)!: change authentication flow

BREAKING CHANGE: Client must now include API key in headers"
```

**4. Push ve PR:**
```bash
git push origin feature/add-network-rename
# GitHub'da Pull Request açın
```

---

## 🔗 Faydalı Kaynaklar

### Dokümantasyon
- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)
- [Tauri Documentation](https://tauri.app/v1/guides/)
- [WireGuard Documentation](https://www.wireguard.com/docs/)

### Araçlar
- [Go Tools](https://golang.org/tools/)
- [Air (Go hot reload)](https://github.com/cosmtrek/air)
- [Vite (Frontend dev server)](https://vitejs.dev/)
- [gosec (Go security)](https://github.com/securego/gosec)

### Standartlar
- [Effective Go](https://golang.org/doc/effective_go)
- [React Best Practices](https://react.dev/learn/thinking-in-react)
- [CLAUDE.md](../CLAUDE.md) (Proje standartları)

---

## ❓ Yardım

Geliştirme sırasında sorun yaşarsanız:

- 📖 [CONTRIBUTING.md](CONTRIBUTING.md) okuyun
- 💬 [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions) sorun
- 🐙 [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues) bildirin

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
