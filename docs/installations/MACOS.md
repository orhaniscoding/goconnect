# 🍎 macOS Kurulum Rehberi

macOS için GoConnect kurulumu, yapılandırması ve sorun giderme.

---

## 📑 İçindekiler

- [Desteklenen Sürümler](#desteklenen-sürümler)
- [Sistem Gereksinimleri](#sistem-gereksinimleri)
- [Kurulum Yöntemleri](#kurulum-yöntemleri)
- [Desktop Application](#desktop-application)
- [CLI Application](#cli-application)
- [Self-Hosted Server](#self-hosted-server)
- [Sorun Giderme](#sorun-giderme)

---

## 🍎 Desteklenen Sürümler

| Sürüm | Codename | Destek | Notlar |
|-------|----------|--------|--------|
| **macOS 15** | Sequoia | ✅ Full support | Intel + Apple Silicon |
| **macOS 14** | Sonoma | ✅ Full support | Intel + Apple Silicon |
| **macOS 13** | Ventura | ✅ Full support | Intel + Apple Silicon |
| **macOS 12** | Monterey | ⚠️ Supported | Intel + Apple Silicon |
| **macOS 11** | Big Sur | ⚠️ Minimum | Intel + Apple Silicon |

**Not:** macOS 10.15 (Catalina) ve öncesi desteklenmiyor.

---

## 💻 Sistem Gereksinimleri

### Minimum Gereksinimler

| Bileşen | Intel Mac | Apple Silicon |
|---------|-----------|---------------|
| **RAM** | 4 GB | 4 GB |
| **Disk** | 100 MB | 100 MB |
| **macOS** | 11+ | 11+ |

### Gerekli Araçlar

**Xcode Command Line Tools:**
```bash
xcode-select --install
```

**Homebrew (önerilen):**
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

---

## 📦 Kurulum Yöntemleri

| Yöntem | Zorluk | Avantaj |
|--------|--------|---------|
| **DMG Installer** | ⭐ Basit | Sürükle-bırak |
| **Homebrew Cask** | ⭐ Basit | Package manager |
| **Binary** | ⭐⭐ Orta | Manuel kontrol |

---

## 🖥️ Desktop Application

### Yöntem 1: DMG Installer (Önerilen)

#### Apple Silicon (M1/M2/M3/M4)

```bash
# İndirin
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/GoConnect-aarch64.dmg

# Açın
open GoConnect-aarch64.dmg

# Drag & Drop
# GoConnect simgesini "Applications" klasörüne sürükleyin
```

#### Intel Mac

```bash
# İndirin
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/GoConnect-x64.dmg

# Açın
open GoConnect-x64.dmg

# Drag & Drop
# GoConnect simgesini "Applications" klasörüne sürükleyin
```

---

#### İlk Çalıştırma

```bash
# Uygulamayı açın
open /Applications/GoConnect.app

# "GoConnect" has been damaged uyarısı çıkarsa:
sudo xattr -cr /Applications/GoConnect.app
```

---

### Yöntem 2: Homebrew Cask

```bash
# Kurun
brew install --cask goconnect

# Başlatın
open /Applications/GoConnect.app

# Güncelleyin
brew upgrade --cask goconnect

# Kaldırın
brew uninstall --cask goconnect
```

---

## 💻 CLI Application

### Yöntem 1: Binary Download

#### Apple Silicon

```bash
# İndirin
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_darwin_arm64.tar.gz

# Çıkarın
tar -xzf goconnect_darwin_arm64.tar.gz
cd goconnect-darwin-arm64

# PATH'e ekleyin
sudo mv goconnect /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect

# Kullanın
goconnect
```

#### Intel

```bash
# İndirin
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_darwin_amd64.tar.gz

# Çıkarın
tar -xzf goconnect_darwin_amd64.tar.gz
cd goconnect-darwin-amd64

# PATH'e ekleyin
sudo mv goconnect /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect

# Kullanın
goconnect
```

---

### Yöntem 2: Homebrew

```bash
# Homebrew kurun (yoksa)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# GoConnect kurun
brew install goconnect

# Güncelleyin
brew upgrade goconnect

# Kaldırın
brew uninstall goconnect
```

---

### Shell Completion

**Zsh:**
```bash
# Completion script'i ekleyin
goconnect completion zsh > ~/.zfunc/_goconnect

# ~/.zshrc'ye ekleyin
echo "fpath=(~/.zfunc \$fpath)" >> ~/.zshrc
echo "autoload -U compinit && compinit" >> ~/.zshrc
```

**Bash:**
```bash
# Completion script'i ekleyin
goconnect completion bash > /usr/local/etc/bash_completion.d/goconnect
source ~/.bash_profile
```

---

## 🏢 Self-Hosted Server

### Docker ile Kurulum (Önerilen)

#### Docker Desktop Kurulumu

```bash
# Homebrew ile
brew install --cask docker

# Veya indirin
# https://www.docker.com/products/docker-desktop
```

#### Apple Silicon (M1/M2/M3)

```bash
# docker-compose.yml indirin
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# .env oluşturun
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 16)
PUBLIC_IP=$(curl -s ifconfig.me)

cat > .env << EOF
JWT_SECRET=$JWT_SECRET
DATABASE_URL=postgres://goconnect:$DB_PASSWORD@db:5432/goconnect?sslmode=disable
WG_SERVER_ENDPOINT=$PUBLIC_IP:51820
EOF

# Başlatın
docker compose up -d

# Logları görün
docker compose logs -f
```

---

### Manual LaunchAgent Service

#### Binary Installation

```bash
# İndirin
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server-darwin-arm64.tar.gz
tar -xzf goconnect-server-darwin-arm64.tar.gz
cd goconnect-server-darwin-arm64

# Kopyalayın
sudo cp goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# Konfigürasyon dizini oluşturun
sudo mkdir -p /etc/goconnect
sudo cp config.example.env /etc/goconnect/.env
sudo nano /etc/goconnect/.env  # Edit config
```

---

#### LaunchAgent

```bash
# LaunchAgent dosyası oluşturun
cat > ~/Library/LaunchAgents/com.goconnect.server.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.goconnect.server</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/goconnect-server</string>
        <string>-config</string>
        <string>/etc/goconnect/.env</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <key>StandardOutPath</key>
    <string>/tmp/goconnect.stdout.log</string>

    <key>StandardErrorPath</key>
    <string>/tmp/goconnect.stderr.log</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin</string>
    </dict>

    <key>ProcessType</key>
    <string>Interactive</string>
</dict>
</plist>
EOF

# Yükleyin
launchctl load ~/Library/LaunchAgents/com.goconnect.server.plist

# Başlatın
launchctl start com.goconnect.server

# Durumu kontrol edin
launchctl list | grep goconnect
```

---

## ⚙️ Yapılandırma

### macOS Firewall

```bash
# Firewall'ı açın (System Preferences → Security & Privacy → Firewall)
# Veya komut satırından:

# GoConnect'e izin verin
/usr/libexec/ApplicationFirewall/socketfilterfw --add /usr/local/bin/goconnect
/usr/libexec/ApplicationFirewall/socketfilterfw --unblockapp /usr/local/bin/goconnect

# Durumu kontrol edin
/usr/libexec/ApplicationFirewall/socketfilterfw --listapps
```

---

### Security & Privacy

**Full Disk Access (gerekirse):**
```
System Preferences → Security & Privacy
→ Privacy → Full Disk Access
→ "+" → GoConnect'i ekleyin
```

**Accessibility (gerekirse):**
```
System Preferences → Security & Privacy
→ Privacy → Accessibility
→ "+" → GoConnect'i ekleyin
```

---

## 🔧 Sorun Giderme

### "GoConnect is damaged"

**Açıklama:** macOS quarantine ve notarization kontrolü.

**Çözüm:**
```bash
# Quarantine'i kaldırın
sudo xattr -cr /Applications/GoConnect.app

# Uygulamayı açın
open /Applications/GoConnect.app
```

---

### "Cannot verify developer"

**Çözüm:**
```
System Preferences → Security & Privacy
→ General
→ "Open Anyway" butonuna tıklayın
```

---

### "Command not found: goconnect"

**Çözüm:**
```bash
# PATH'i kontrol edin
echo $PATH

# /usr/local/bin PATH'te mi?
ls -la /usr/local/bin/goconnect

# Manuel PATH'e ekleyin
export PATH="/usr/local/bin:$PATH"

# Kalıcı hale getirmek için ~/.zshrc'ye ekleyin
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

---

### "WireGuard kernel extension not loaded"

**Çözüm:**
```bash
# WireGuard tools kurun (Homebrew)
brew install wireguard-tools

# Kernel modülünü yükleyin
sudo kextload /Library/Extensions/wireguard.kext

# Veya GoConnect'in built-in WireGuard'ı kullanın
# (Ek kurulum gerekmez)
```

---

### "Service crashes on startup"

**Çözüm:**
```bash
# Logları kontrol edin
cat /tmp/goconnect.stderr.log
cat /tmp/goconnect.stdout.log

# Console.app'te logları görün
open /Applications/Utilities/Console.app
# → "GoConnect" filtreleyin

# Manuel test
/usr/local/bin/goconnect-server -config /etc/goconnect/.env
```

---

## 🗑️ Kaldırma

### Desktop App

```bash
# Uygulamayı silin
rm -rf /Applications/GoConnect.app

# User data sil (opsiyonel)
rm -rf ~/Library/Application Support/com.goconnect.app
rm -rf ~/Library/Caches/com.goconnect.app
rm -rf ~/Library/Preferences/com.goconnect.app.plist
```

---

### CLI

```bash
# Homebrew ile kurulduysa
brew uninstall goconnect

# Manuel ise
rm /usr/local/bin/goconnect
rm -rf ~/.config/goconnect
```

---

### Server (LaunchAgent)

```bash
# Service'i durdurun ve unload edin
launchctl unload ~/Library/LaunchAgents/com.goconnect.server.plist

# Dosyaları silin
rm ~/Library/LaunchAgents/com.goconnect.server.plist
rm /usr/local/bin/goconnect-server
sudo rm -rf /etc/goconnect

# Logları silin
rm /tmp/goconnect.*.log
```

---

## 📚 Ek Kaynaklar

- [Genel Kurulum Rehberi](../INSTALLATION.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Geliştirme Rehberi](../DEVELOPMENT.md)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
