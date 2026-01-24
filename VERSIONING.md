# 🔢 Versioning ve Release Politikası

GoConnect'in sürüm numaralandırması ve release süreci.

---

## 📋 Semantic Versioning

GoConnect **Semantic Versioning 2.0.0** kullanır.

### Format

```
MAJOR.MINOR.PATCH

Örnek: 3.0.0
```

| Bileşen | Anlamı | Örnek |
|---------|--------|-------|
| **MAJOR** | Breaking changes | 3.0.0 → 4.0.0 |
| **MINOR** | Yeni özellikler (backward compatible) | 3.0.0 → 3.1.0 |
| **PATCH** | Bug fixes | 3.0.0 → 3.0.1 |

---

## 📈 Release Cycle

### Sürüm Türleri

| Tür | Sıklık | Örnek |
|-----|--------|-------|
| **Major** | Yılda 1-2 kez | 3.0 → 4.0 |
| **Minor** | 3 ayda bir | 3.0 → 3.1 |
| **Patch** | Gerekirse | 3.0.0 → 3.0.1 |

### Release Kanalları

**Stable:**
- Production için
- Tam test edilmiş
- Kararlı API

**Beta:**
- Yeni özellikler
- Community testi
- Stabil olabilir

**Alpha:**
- Erken erişim
- Deneysel
- Breaking changes olabilir

---

## 🔄 Release Süreci

### 1. Development

```bash
# Feature branch oluştur
git checkout -b feature/add-voice-chat

# Geliştir
# Test et
# PR aç
```

### 2. Release Branch

```bash
# Minor release için
git checkout -b release/3.1.0

# Version bump
# Changelog update
# Testing
```

### 3. Release

```bash
# Tag oluştur
git tag -a v3.1.0 -m "Release v3.1.0: Add voice chat"

# Push
git push origin v3.1.0

# GitHub Actions automatic:
# - Build binaries
# - Create GitHub Release
# - Upload assets
```

### 4. Post-Release

```bash
# Merge to main
git checkout main
git merge release/3.1.0

# Delete branch
git branch -d release/3.1.0
```

---

## 📝 Changelog Format

```markdown
# [3.1.0] - 2025-01-24

## Added
- Voice chat feature
- Screen sharing
- File transfer progress bar

## Changed
- UI redesign
- Improved performance

## Fixed
- Crash on network join
- Memory leak in chat

## Security
- Updated WireGuard to v1.0.0

## Breaking Changes
- API endpoint /v1/networks → /v2/networks
```

---

## 🚨 Breaking Change Politikası

### Ne Zaman Breaking Change?

**MAJOR version'lar:**
- API değişiklikleri
- Database schema değişiklikleri
- Konfigürasyon formatı değişiklikleri
- Removed features

**Örnek:**
```go
// v2.0
CreateNetwork(name string) (*Network, error)

// v3.0 (breaking change)
CreateNetwork(ctx context.Context, req CreateNetworkRequest) (*Network, error)
```

### Migration Guide

Her breaking change için:
1. **Migration guide** yazın
2. **Deprecation warning** ekleyin (en az 2 minor version önce)
3. **Upgrade tool** sağlayın

**Örnek:**
```markdown
# Migration Guide: v2.0 → v3.0

## Breaking Changes

### API Endpoints

Old: POST /v1/auth/register
New: POST /v2/auth/register

### Migration Steps

1. Update API base URL
2. Migrate database: `goconnect migrate --to=v3`
3. Update config: add JWT_SECRET
```

---

## 🔄 Deprecation Policy

### Ömür Döngüsü

| Durum | Süre | Örnek |
|------|------|-------|
| **Announced** | Release | "Deprecated in v3.0" |
| **Soft Deprecated** | 2 minor version | v3.0, v3.1 (uyarı) |
| **Hard Deprecated** | 1 minor version | v3.2 (hata) |
| **Removed** | Next major | v4.0 (kaldırıldı) |

### Örnek Timeline

```
v3.0.0 - Feature announced
v3.1.0 - Feature added (alongside old)
v3.2.0 - Old feature deprecated (warning)
v3.3.0 - Old feature causes error
v4.0.0 - Old feature removed
```

---

## 🔙 Backward Compatibility

### API Compatibility

**Guarantee:** PATCH ve MINOR sürümler backward compatible.

**Exception:** Security patches gerektiriyorsa.

### Database Schema

**Migration:** Otomatik migration on first run.

**Rollback:** Manuel mümkündür (backup gerektirir).

### Configuration

**Old config:** Yeni sürümde de çalışır.

**New config:** Varsayılan değerlerle oluşturulur.

---

## 📊 Support Policy

### Destek Süreleri

| Sürüm | Destek | Son Güncelleme |
|-------|--------|----------------|
| **3.x** | ✅ Aktif | Her 3 ayda bir |
| **2.x** | ⚠️ Maintenance | Sadece security fixes |
| **1.x** | ❌ End-of-Life | Desteklenmiyor |

### Security Patches

**Kritik security issues:** 48 saat içinde patch

**Non-critical:** Bir sonraki PATCH veya MINOR release

---

## 🧪 Testing Before Release

### Pre-Release Checklist

- [ ] Unit tests pass (%100)
- [ ] Integration tests pass
- [ ] E2E tests pass
- [ ] Manual testing (Windows, macOS, Linux)
- [ ] Security audit (gosec, nancy)
- [ ] Performance benchmarks
- [ ] Documentation updated
- [ ] Changelog written
- [ ] Release notes drafted

### Beta Testing Period

**Süre:** 2-4 hafta

**Kapsam:**
- Community testing
- Bug reports
- Performance feedback
- UX improvements

---

## 🎯 Release Planning

### Roadmap

**3.1.0 (Yakında)**
- Voice chat improvements
- Screen sharing
- Performance optimizations

**3.2.0 (Q2 2025)**
- Mobile apps (Android beta)
- End-to-end encryption for chat
- Custom themes

**4.0.0 (Q4 2025)**
- Breaking API changes
- New architecture
- Enhanced security

---

## 📦 Release Assets

Her release şunları içerir:

### Binaries
- Windows (x64)
- macOS (Intel + Apple Silicon)
- Linux (x64, ARM64)

### Checksums
- SHA256
- GPG signature (opsiyonel)

### Documentation
- Release notes
- Migration guide (eğer gerekli)
- Upgrade instructions

---

## 🚀 Automatic Updates

### Desktop App

**Check frequency:** Her 24 saat

**Update process:**
1. Background check
2. Yeni sürüm varsa bildirim
3. User onayı
4. Download
5. Automatic install
6. Restart (opsiyonel)

### CLI

**Manual check:**
```bash
goconnect check-update
```

**Update:**
```bash
# Package manager ile
# macOS (Homebrew)
brew upgrade goconnect

# Linux (apt)
sudo apt update && sudo apt install goconnect

# Manual download
# GitHub releases
```

---

## 📞 Feedback

### Beta Testing

Katılmak için:
- 🐙 [Discussions](https://github.com/orhaniscoding/goconnect/discussions/categories/beta-testing)
- 📧 [E-posta](mailto:beta@goconnect.io)

### Bug Reports

[GitHub Issues](https://github.com/orhaniscoding/goconnect/issues/new?template=bug_report.md)

### Feature Requests

[GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)

---

## 📚 Referanslar

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Release Notes Archive](CHANGELOG.md)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
