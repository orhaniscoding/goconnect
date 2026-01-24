# 📦 Bağımlılık Listesi

GoConnect projesinin kullandığı tüm bağımlılıkların tam listesi ve açıklamaları.

---

## 📋 İçindekiler

- [Zero-Dependency Politikası](#zero-dependency-politikası)
- [Production Bağımlılıkları](#production-bağımlılıkları)
- [Development Bağımlılıkları](#development-bağımlılıkları)
- [Platform-Specific Build Tools](#platform-specific-build-tools)
- [Dependency Yönetimi](#dependency-yönetimi)
- [Security Audit](#security-audit)

---

## 🎯 Zero-Dependency Politikası

GoConnect **production binary'sinde sıfır dış bağımlılık** ilkesini benimser.

### Neden?

| Sebep | Açıklama |
|-------|----------|
| **Güvenlik** | Supply chain saldırı riski azalır |
| **Boyut** | Binary boyutu küçülür (~15MB) |
| **Performans** | Stdlib optimize edilmiş, hızlı |
| **Sürdürülebilirlik** | Uzun vadeli bakım kolay |
| **Bağımsızlık** | External kütüphane güncellemesi gerekmez |

### Kural

**Production code → Stdlib ONLY**
```go
// ✅ İZİN VERİLEN
import "crypto/rand"
import "database/sql"
import "encoding/json"

// ❌ YASAK
import "github.com/gorilla/mux"     // HTTP router
import "github.com/jmoiron/sqlx"    // Database wrapper
import "github.com/gin-gonic/gin"   // Web framework
```

**Development code → External libraries OK**
```go
// ✅ Testlerde kullanılabilir
import "github.com/stretchr/testify/assert" // Test assertions
```

---

## 🚀 Production Bağımlılıkları

### Core (Go Stdlib Only)

| Paket | Sürüm | Kullanım |
|-------|-------|----------|
| `std` | Go 1.24+ | **Tüm production kod** |

**Hangi stdlib paketleri kullanılıyor?**

```go
// Networking
import ("net"; "net/http"; "net/url")

// Cryptography
import ("crypto"; "crypto/rand"; "crypto/rsa"; "crypto/tls")

// Database
import ("database/sql"; "database/sql/driver")

// Encoding
import ("encoding/json"; "encoding/base64"; "encoding/xml")

// I/O
import ("io"; "bufio"; "bytes")

// Concurrency
import ("sync"; "context"; "time")

// Logging
import ("log"; "log/syslog")
```

### Desktop (Tauri + React)

**Core Framework:**
| Araç | Sürüm | Boyut | Kullanım |
|------|-------|-------|----------|
| **Tauri** | 2.0+ | ~3MB | Desktop framework |
| **React** | 19.0+ | ~45KB | UI framework |
| **Rust Stdlib** | Stable | - | Backend |

**Minimum Varsayılan:**
```json
{
  "dependencies": {
    "react": "^19.0.0",
    "@tauri-apps/api": "^2.0.0"
  }
}
```

**Neden bu kadar az?**
- Tauri web view kullanır (system WebView)
- React çok küçüktür (~45KB minified + gzip)
- Rust backend zaten compiled binary'dir

---

## 🛠️ Development Bağımlılıkları

### Go Development Tools

| Araç | Sürüm | Kullanım | Dev-only |
|------|-------|----------|----------|
| **testify** | v1.9.0 | Test assertions | ✅ |
| **golangci-lint** | v1.55+ | Linting | ✅ |
| **gosec** | v2.18+ | Security scanning | ✅ |
| **air** | v1.50+ | Hot reload | ✅ |
| **goose** | v3.17+ | Database migrations | ✅ |
| **protoc** | v24+ | Protocol buffers | ✅ |
| **mockgen** | v1.6+ | Mock generation | ✅ |

**Kurulum:**
```bash
# Test framework
go install github.com/stretchr/testify/assert@latest

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security scanner
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Hot reload
go install github.com/cosmtrek/air@latest

# Migration tool
go install github.com/pressly/goose/v3/cmd/goose@latest

# Mock generator
go install github.com/golang/mock/mockgen@latest
```

### Frontend Development Tools

| Araç | Sürüm | Kullanım |
|------|-------|----------|
| **Vite** | 5.0+ | Dev server, bundling |
| **TypeScript** | 5.0+ | Type checking |
| **Tailwind CSS** | 4.0+ | Styling |
| **Vitest** | 1.0+ | Unit testing |
| **Playwright** | 1.40+ | E2E testing |
| **ESLint** | 8.50+ | Linting |
| **Prettier** | 3.0+ | Formatting |

**Kurulum:**
```bash
cd desktop
npm install --save-dev vite typescript tailwindcss vitest
npm install --save-dev @playwright/test eslint prettier
```

### Code Generation Tools

| Araç | Sürüm | Kullanım |
|------|-------|----------|
| **protoc-gen-go** | v1.32+ | Go protobuf generation |
| **sqlc** | v1.25+ | Type-safe SQL generation |
| **wire** | v0.5+ | Dependency injection |
| **oapi-codegen** | v1.0+ | OpenAPI generation |

---

## 🔧 Platform-Specific Build Tools

### Windows Build Tools

| Araç | Amaç | Gerekli mi? |
|------|-------|------------|
| **WiX Toolset** | MSI installer oluşturma | Opsiyonel |
| **signtool.exe** | Code signing | Production için evet |
| **Visual Studio Build Tools** | C++ derleme | Zorunlu |

**Kurulum:**
```powershell
# WiX Toolset
winget install WiX.Toolset

# VS Build Tools
winget install Microsoft.VisualStudio.2022.BuildTools
```

### macOS Build Tools

| Araç | Amaç | Gerekli mi? |
|------|-------|------------|
| **Xcode Command Line Tools** | C derleme, code signing | Zorunlu |
| **Xcode** (full) | iOS development | Opsiyonel |

**Kurulum:**
```bash
xcode-select --install
```

### Linux Build Tools

| Araç | Amaç | Gerekli mi? |
|------|-------|------------|
| **build-essential** | GCC, Make | Zorunlu |
| **pkg-config** | Library discovery | Zorunlu |
| **webkit2gtk-4.1** | Tauri webview | Zorunlu |
| **libappindicator3** | System tray | Zorunlu |

**Kurulum (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install -y \
  build-essential \
  pkg-config \
  libwebkit2gtk-4.1-dev \
  libappindicator3-dev \
  librsvg2-dev
```

---

## 📊 Dependency Yönetimi

### Go Modules

**CLI go.mod:**
```go
module github.com/orhaniscoding/goconnect/cli

go 1.24

require (
  // NO external dependencies for production
)

// Dev-only dependencies
dev (
  github.com/stretchr/testify v1.9.0
)
```

**Core go.mod:**
```go
module github.com/orhaniscoding/goconnect/core

go 1.24

require (
  // NO external dependencies for production
)

dev (
  github.com/stretchr/testify v1.9.0
  github.com/pressly/goose/v3 v3.17.0
)
```

### Node.js package.json

**desktop/package.json:**
```json
{
  "name": "goconnect-desktop",
  "dependencies": {
    "react": "^19.0.0",
    "@tauri-apps/api": "^2.0.0"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "typescript": "^5.0.0",
    "tailwindcss": "^4.0.0",
    "vitest": "^1.0.0",
    "@playwright/test": "^1.40.0",
    "eslint": "^8.50.0",
    "prettier": "^3.0.0"
  }
}
```

### Rust Cargo.toml

**desktop/src-tauri/Cargo.toml:**
```toml
[package]
name = "goconnect"
version = "3.0.0"

[dependencies]
tauri = { version = "2.0", features = ["shell-open"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

# NO external networking deps - use std only
```

---

## 🔒 Security Audit

### Dependency Scanning

**Go:**
```bash
# Vulnerability scan
go list -json -m all | nancy sleuth

# SBOM oluştur
syft goconnect-cli -o spdx-json > sbom.json
```

**Node.js:**
```bash
# Audit
npm audit

# Fix vulnerabilities
npm audit fix

# SBOM
syft goconnect-desktop -o spdx-json > desktop-sbom.json
```

**Rust:**
```bash
# Audit
cargo audit

# Check outdated
cargo outdated
```

### Dependabot

Dependabot otomatik olarak güncellemeleri kontrol eder:
- `.github/dependabot.yml` yapılandırılmış
- Haftalıkalık check
- Otomatik PR'ler

### Manual Review Checklist

Her dependency güncellemesinden önce:

- [ ] Changelog'i oku
- [ ] Breaking change'leri kontrol et
- [ ] Security advisories kontrol et
- [ ] Testleri çalıştır
- [ ] Manual test yap
- [ ] Binary boyutunu ölç

---

## 🚨 Yasaklı Bağımlılıklar

Bu bağımlılıkları **ASLA eklemeyin**:

### Go

| Kütüphane | Neden Yasak? | Alternatif |
|-----------|-------------|------------|
| `gorm` | ORM, çok büyük | Custom 150-line scanner |
| `sqlx` | Gereksiz wrapper | stdlib `database/sql` |
| `gin` | Web framework | `net/http` stdlib |
| `gorilla/mux` | HTTP router | `http.ServeMux` stdlib |
| `grpc-go` | Ağır RPC | Custom protobuf + stdlib |
| `viper` | Config management | Custom env parser |

### JavaScript

| Kütüphane | Neden Yasak? | Alternatif |
|-----------|-------------|------------|
| `axios` | Gereksiz wrapper | `fetch` native |
| `lodash` | Çok büyük | Native methods |
| `moment.js` | Deprecated | `Intl.DateTimeFormat` |
| `redux` | Gereksiz karmaşıklık | Zustand (daha küçük) |

### Rust

| Kütüphane | Neden Yasak? | Alternatif |
|-----------|-------------|------------|
| `reqwest` | Gereksiz HTTP client | `hyper` veya `curl` |
| `diesel` | ORM, çok büyük | Custom SQL |
| `actix-web` | Web framework | Custom minimal HTTP |

---

## 📈 Binary Size Impact

### Current Sizes

| Bileşen | Boyut | Notlar |
|---------|-------|--------|
| **CLI (Linux)** | ~8MB | Stripped, stdlib only |
| **CLI (Windows)** | ~10MB | + PE header |
| **CLI (macOS)** | ~9MB | + Mach-O |
| **Desktop (Windows)** | ~15MB | Tauri + React |
| **Desktop (macOS)** | ~12MB | Tauri + React |
| **Desktop (Linux)** | ~14MB | Tauri + React |

### Size Reduction Techniques

```bash
# Strip debug info
go build -ldflags="-s -w" ./cmd/goconnect

# UPX compression (opsiyonel)
upx --best --lzma goconnect

# Result: ~30% smaller
```

---

## 🔄 Dependency Update Policy

### Update Frequency

- **Security patches**: İlk 24 saat içinde
- **Minor updates**: Haftalık
- **Major updates**: Manuel inceleme sonra

### Update Process

1. **Check for updates:**
   ```bash
   go get -u ./...
   npm update
   cargo update
   ```

2. **Test everything:**
   ```bash
   make test
   make lint
   ```

3. **Manual testing:**
   - CLI'yi çalıştır
   - Desktop'ı aç
   - E2E testleri çalıştır

4. **Update documentation:**
   - DEPENDENCIES.md
   - CHANGELOG.md

5. **Commit:**
   ```bash
   git commit -m "chore(deps): update Go to 1.24.2"
   ```

---

## 📚 Ek Kaynaklar

- [Go Modules Reference](https://golang.org/ref/mod)
- [Tauri Performance](https://tauri.app/v1/guides/performance/)
- [Node.js Security Best Practices](https://github.com/lirantal/awesome-sec-devtools#nodejs-security)
- [Rust Security](https://doc.rust-lang.org/book/ch12-00-an-io-project.html)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
