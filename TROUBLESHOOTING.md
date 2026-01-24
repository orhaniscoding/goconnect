# 🔧 Sorun Giderme (Troubleshooting)

GoConnect ile karşılaşabileceğiniz sorunlar ve çözümleri.

---

## 📑 İçindekiler

- [Hızlı Çözümler](#hızlı-çözümler)
- [Kurulum Sorunları](#kurulum-sorunları)
- [Bağlantı Sorunları](#bağlantı-sorunları)
- [Performans Sorunları](#performans-sorunları)
- [Platform-Specific Sorunlar](#platform-specific-sorunlar)
- [Debugging Araçları](#debugging-araçları)

---

## ⚡ Hızlı Çözümler

### Genel Kontrol Listesi

Sorun yaşadığınızda bu adımları sırayla deneyin:

```bash
# 1. GoConnect sürümünü kontrol edin
goconnect version

# 2. İnternet bağlantısını test edin
ping -c 4 api.goconnect.io
ping -c 4 8.8.8.8

# 3. Firewall durumunu kontrol edin
# Windows
netsh advfirewall show allprofiles

# macOS
sudo pfctl -s info

# Linux
sudo ufw status

# 4. Log dosyalarını kontrol edin
# Windows
type %APPDATA%\goconnect\logs\goconnect.log

# macOS/Linux
tail -f ~/.config/goconnect/goconnect.log

# 5. Portu kontrol edin
# Linux/macOS
lsof -i :51820
netstat -tuln | grep 51820

# Windows
netstat -an | findstr 51820
```

---

## 🚀 Kurulum Sorunları

### Windows

#### "Windows Defender virüs buldu"

**Açıklama:** Yanlış pozitif. GoConnect zararsızdır.

**Çözüm:**
1. İndirilen dosyaya sağ tıklayın
2. "Daha fazla bilgi" → "Yine de çalıştır" seçin

**Kalıcı çözüm (AV'de hariç tut):**
```
Windows Security → Virus & threat protection
→ Manage settings → Exclusions
→ Add or remove exclusions
→ Add an exclusion → Folder
→ C:\Users\YourName\AppData\Local\GoConnect
```

---

#### "MSI dosyası açılamıyor"

**Açıklama:** Windows Installer eksik.

**Çözüm:**
```powershell
# Windows Installer'i yeniden başlat
net stop msiserver
net start msiserver
```

---

#### "WebView2 eksik"

**Açıklama:** Desktop app WebView2 gerektirir.

**Çözüm:**
```powershell
# İndirin ve kurun
winget install Microsoft.WebView2.Runtime
```

---

### macOS

#### "GoConnect has been damaged"

**Açıklama:** macOS quarantine ve notarization kontrolü.

**Çözüm:**
```bash
# Quarantine'i kaldırın
sudo xattr -cr /Applications/GoConnect.app

# Uygulamayı açın
open /Applications/GoConnect.app
```

---

#### "Developer cannot be verified"

**Açıklama:** Güvenlik ayarları.

**Çözüm:**
```
System Preferences → Security & Privacy
→ General → "Open Anyway"
```

---

### Linux

#### "Permission denied"

**Açıklama:** Binary çalıştırılabilir değil.

**Çözüm:**
```bash
chmod +x goconnect
```

---

#### "libwebkit2gtk-4.1 not found"

**Açıklama:** Desktop app bağımlılıkları eksik.

**Çözüm:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y libwebkit2gtk-4.1-dev libappindicator3-dev librsvg2-dev

# Fedora
sudo dnf install webkit2gtk4.1-devel libappindicator-gtk3-devel librsvg2-devel

# Arch
sudo pacman -S webkit2gtk-4.1 libappindicator-gtk3 librsvg
```

---

## 🔌 Bağlantı Sorunları

### "Bağlanamıyor"

#### Olası Nedenler ve Çözümler

**1. Internet bağlantısı yok**
```bash
# Test et
ping -c 4 8.8.8.8

# DNS kontrolü
nslookup api.goconnect.io
```

**2. Firewall engelliyor**

**Windows:**
```powershell
# Kural ekle
netsh advfirewall firewall add rule name="GoConnect" dir=in action=allow program="C:\Users\YourName\AppData\Local\GoConnect\goconnect.exe" enable=yes

# Veya Windows Security'den manuel ekle
```

**Linux:**
```bash
# UFW (Ubuntu)
sudo ufw allow 51820/udp
sudo ufw allow from any to any port 51820 proto udp

# firewalld (Fedora/CentOS)
sudo firewall-cmd --permanent --add-port=51820/udp
sudo firewall-cmd --reload
```

**macOS:**
```bash
# Firewall kontrolü
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate

# App'i izin verilenlere ekle (System Preferences)
```

**3. VPN/proxy kullanıyorsunuz**

**Çözüm:**
- VPN'i kapatmayı deneyin
- Proxy ayarlarını kontrol edin
-Corporate network'teyseniz, IT desteğine başvurun

---

### "NAT traversal başarısız"

**Açıklama:** GoConnect peer-to-peer bağlantı kuramıyor.

**Çözümler:**

**1. UPnP'yi etkinleştirin**
```
Router admin panel → NAT/UPnP
→ UPnP'yi Enable yapın
→ Kaydedin ve router'ı yeniden başlatın
```

**2. Port forwarding yapın**
```
Router admin panel → Port Forwarding
→ External Port: 51820
→ Internal Port: 51820
→ Protocol: UDP
→ Internal IP: [Bilgisayarınızın IP'si]
→ Enable
```

**3. DMZ kullanın (son çare)**
```
Router admin panel → DMZ
→ Bilgisayarınızın IP'sini DMZ'ye ekleyin
```

---

### "Handshake timeout"

**Açıklama:** WireGuard handshake timeout.

**Çözüm:**
```bash
# Logları kontrol edin
goconnect logs

# Debug modda çalıştırın
LOG_LEVEL=debug goconnect

# Network delay'ı kontrol edin
ping -c 10 api.goconnect.io
```

---

## ⚡ Performans Sorunları

### "Yavaş dosya transferi"

**Olası nedenler:**

**1. Relay kullanılıyor**
```
Peer-to-peer bağlantı kurulamadı, relay kullanılıyor.
Relay yavaştır çünkü tüm trafik sunucudan geçer.
```

**Çözüm:** Port forwarding yapın (yukarıya bakın)

**2. Network throttling**
```bash
# Bandwidth test
speedtest-cli

# QoS kontrolü (router ayarları)
```

---

### "Yüksek CPU kullanımı"

**Açıklama:** GoConnect %50+ CPU kullanıyor.

**Çözüm:**
```bash
# Çok sayıda bağlantı var mı?
goconnect status

# WireGuard interface'leri kontrol edin
# Linux
sudo wg show

# Windows/macOS
goconnect network status

# Gereksiz bağlantıları kapatın
goconnect disconnect <network-id>
```

---

### "Memory leak"

**Belirtiler:**
- Uygulama zamanla yavaşlar
- RAM kullanımı sürekli artar

**Çözüm:**
```bash
# Restart yapın
goconnect quit
goconnect

# Hala sorun varsa logları toplayın ve issue açın
goconnect logs > bug-report.log
```

---

## 🖥️ Platform-Specific Sorunlar

### Windows

#### "Windows Update sonrası çalışmıyor"

**Çözüm:**
```powershell
# Windows Firewall sıfırlanmış, kuralı yeniden ekleyin
netsh advfirewall firewall add rule name="GoConnect" dir=in action=allow program="C:\Users\YourName\AppData\Local\GoConnect\goconnect.exe"
```

---

#### "System tray'de kayboldu"

**Çözüm:**
```powershell
# Process'i restart edin
taskkill /IM goconnect.exe /F
goconnect
```

---

### macOS

#### "Gatekeeper engelliyor"

**Çözüm:**
```bash
# xattr'ı temizle
sudo xattr -cr /Applications/GoConnect.app

# Sistemi yeniden başlat
sudo reboot
```

---

#### "Network permissions"

**Çözüm:**
```
System Preferences → Security & Privacy
→ Privacy → Full Disk Access
→ GoConnect'i ekleyin
```

---

### Linux

#### "WireGuard module yüklenemiyor"

**Çözüm:**
```bash
# WireGuard kernel module'ü kontrol edin
lsmod | grep wireguard

# Yüklü değilse:
# Ubuntu/Debian
sudo apt install wireguard-dkms

# Fedora/CentOS
sudo dnf install wireguard-tools kernel-devel
sudo dkms autoinstall
```

---

#### "Systemd service başlamıyor"

**Çözüm:**
```bash
# Logları kontrol edin
sudo journalctl -u goconnect -n 50

# Konfigürasyon dosyasını kontrol edin
sudo cat /etc/goconnect/.env

# Manual test
sudo -u goconnect /usr/local/bin/goconnect-server -config /etc/goconnect/.env
```

---

## 🐛 Debugging Araçları

### Loglar

**CLI:**
```bash
# Son 50 satır
goconnect logs --tail 50

# Canlı takip
goconnect logs --follow

# Debug level
LOG_LEVEL=debug goconnect
```

**Desktop:**
```
Help → Show Logs in Folder
```

**Server:**
```bash
# Docker
docker logs -f goconnect

# Systemd
sudo journalctl -u goconnect -f
```

---

### Network Diagnostics

```bash
# Port kontrolü
nc -zuv 51820  # Linux/macOS
Test-NetConnection -Port 51820  # Windows

# Trace route
traceroute api.goconnect.io

# DNS kontrolü
nslookup api.goconnect.io
dig api.goconnect.io

# Network delay
ping -c 100 api.goconnect.io | tail -1
```

---

### WireGuard Debugging

```bash
# Interface durumu
sudo wg show

# Handshake durumu
sudo wg show wg0

# Latest handshake zamanı
sudo wg show wg0 | grep peer

# Transfer istatistikleri
sudo wg show wg0 | grep transfer
```

---

## 📞 Yardım Alın

Sorun hala çözülmedi mi?

### Bilgi Toplayın

**Sistem bilgileri:**
```bash
# GoConnect sürümü
goconnect version

# İşletim sistemi
# Windows
systeminfo | findstr /B /C:"OS Name" /C:"OS Version"

# macOS
sw_vers

# Linux
cat /etc/os-release
```

**Log dosyaları:**
- GoConnect logs
- Network diagnostics
- Screenshot (mümkünse)

### Destek Kanalları

- 📖 [Dokümantasyon](README.md)
- ❓ [FAQ](FAQ.md)
- 🐙 [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues/new?template=bug_report.md)
- 💬 [Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- 📧 E-posta: support@goconnect.io

---

## 📚 Sorun Giderme Rehberleri

Platforma özgü detaylı rehberler:

- 🪟 [Windows Sorun Giderme](docs/installations/WINDOWS.md#sorun-giderme)
- 🍎 [macOS Sorun Giderme](docs/installations/MACOS.md#sorun-giderme)
- 🐧 [Linux Sorun Giderme](docs/installations/LINUX.md#sorun-giderme)
- 🐳 [Docker Sorun Giderme](docs/installations/DOCKER.md#sorun-giderme)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
