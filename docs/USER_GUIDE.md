# 📖 GoConnect User Guide

GoConnect ile sanal LAN ağları oluşturmak ve katılmak için kapsamlı kullanım kılavuzu.

---

## 📑 İçindekiler

1. [Başlarken](#-başlarken)
2. [Desktop App Kullanımı](#-desktop-app-kullanımı)
3. [CLI/Terminal Kullanımı](#-cliterminal-kullanımı)
4. [Ağ Oluşturma](#-ağ-oluşturma)
5. [Ağa Katılma](#-ağa-katılma)
6. [Sohbet](#-sohbet)
7. [Dosya Transferi](#-dosya-transferi)
8. [Ayarlar](#%EF%B8%8F-ayarlar)
9. [Sorun Giderme](#-sorun-giderme)

---

## 🚀 Başlarken

### Hangi Versiyonu Kullanmalıyım?

| Kullanıcı Tipi | Önerilen |
|----------------|----------|
| Günlük kullanıcı | **Desktop App** - Görsel arayüz, kolay kullanım |
| Sunucu/headless | **CLI** - Terminal tabanlı, script desteği |
| Geliştirici | **CLI** - Otomatizasyon ve entegrasyon |

### İndirme

**Desktop App:**
- Windows: [GoConnect_x64-setup.exe](https://github.com/orhaniscoding/goconnect/releases/latest)
- macOS: [GoConnect_aarch64.dmg](https://github.com/orhaniscoding/goconnect/releases/latest) (Apple Silicon)
- Linux: [GoConnect_amd64.deb](https://github.com/orhaniscoding/goconnect/releases/latest)

**CLI:**
- Windows: `goconnect_*_windows_amd64.zip`
- macOS: `goconnect_*_darwin_arm64.tar.gz`
- Linux: `goconnect_*_linux_amd64.tar.gz`

---

## 🖥️ Desktop App Kullanımı

### İlk Açılış

1. GoConnect'i indirin ve kurun
2. Uygulamayı başlatın
3. "Hesap Oluştur" veya "Giriş Yap" seçeneklerinden birini seçin

### Ana Ekran

```
┌────────────────────────────────────────────────────────────┐
│  GoConnect                                        ─ □ ✕   │
├────┬───────────────────────────────────────────────────────┤
│    │                                                       │
│ 🏠 │  Hoş Geldiniz!                                        │
│    │                                                       │
│ 🎮 │  ┌─────────────────────────────────────────────────┐  │
│    │  │  🌐 Yeni Ağ Oluştur                             │  │
│ 💼 │  │  Kendi sanal LAN ağınızı başlatın               │  │
│    │  └─────────────────────────────────────────────────┘  │
│    │                                                       │
│    │  ┌─────────────────────────────────────────────────┐  │
│ +  │  │  🔗 Ağa Katıl                                   │  │
│    │  │  Davet linki ile mevcut ağa katılın             │  │
│    │  └─────────────────────────────────────────────────┘  │
│    │                                                       │
│ ⚙️ │                                                       │
└────┴───────────────────────────────────────────────────────┘
```

### System Tray

GoConnect arka planda çalışır. System tray'den:
- Bağlantı durumunu görüntüleyin
- Hızlıca ağ değiştirin
- Uygulamayı tamamen kapatın

---

## 💻 CLI/Terminal Kullanımı

### Kurulum

```bash
# Linux/macOS
tar -xzf goconnect_*.tar.gz
sudo mv goconnect /usr/local/bin/

# PATH'e eklendikten sonra
goconnect --version
```

### Temel Komutlar

```bash
# İnteraktif mod (TUI arayüz)
goconnect

# Hızlı komutlar
goconnect create "Ağ Adı"    # Ağ oluştur
goconnect join <link>        # Ağa katıl
goconnect networks           # Ağları listele
goconnect peers              # Peerleri listele
goconnect status             # Bağlantı durumu
goconnect doctor             # Sorun giderme ve teşhis
goconnect help               # Yardım
```

### TUI Navigasyonu

| Tuş | İşlev |
|-----|-------|
| `↑` / `k` | Yukarı |
| `↓` / `j` | Aşağı |
| `Enter` | Seç |
| `Tab` | Panel değiştir |
| `q` | Çıkış |
| `/` | Arama |
| `?` | Yardım |

---

## 🌐 Ağ Oluşturma

### Desktop App ile

1. Sol menüden **"+"** butonuna tıklayın
2. **"Yeni Ağ Oluştur"** seçin
3. Ağ bilgilerini girin:
   - **Ağ Adı**: Örn. "Gaming Night"
   - **Açıklama**: (Opsiyonel)
   - **Gizlilik**: Public veya Private
4. **"Oluştur"** butonuna tıklayın

### CLI ile

```bash
# İnteraktif
goconnect create "Gaming Night"

# Detaylı
goconnect create \
  --name "Gaming Night" \
  --description "Friday gaming sessions" \
  --private
```

### Davet Linki Oluşturma

Ağ oluşturduktan sonra:

```bash
# Link oluştur
goconnect invite
```

Çıktı: `goconnect://join/abc123xyz`

Bu linki arkadaşlarınızla paylaşın!

---

## 🔗 Ağa Katılma

### Desktop App ile

1. Davet linkine tıklayın, veya
2. **"Ağa Katıl"** butonuna tıklayın
3. Davet kodunu yapıştırın
4. **"Katıl"** butonuna tıklayın

### CLI ile

```bash
# Link ile
goconnect join --invite goconnect://join/abc123xyz

# Kod ile
goconnect join --invite abc123xyz
```

### Otomatik Bağlanma

Bağlanma başarılı olduğunda:
- ✅ VPN tüneli kurulur
- ✅ Sanal IP adresi atanır
- ✅ Diğer üyelerle iletişim başlar

---

## 💬 Sohbet

### Text Channels

Her ağda varsayılan **#general** kanalı bulunur. Adminler ek kanallar oluşturabilir:

- `#general` - Genel sohbet
- `#gaming` - Oyun koordinasyonu
- `#announcements` - Duyurular

### Mesaj Gönderme

**Desktop:**
- Kanal seçin → Mesaj yazın → Enter

**CLI:**
- `goconnect` (interactive mode) ile arayüzü başlatın.
- Tab tuşu ile sohbet paneline geçin.
- Mesajınızı yazıp Enter'a basın.

### Özellikler

- ✅ Gerçek zamanlı mesajlaşma
- ✅ Emoji desteği 🎮
- ✅ @mention bildirimleri
- ✅ Mesaj geçmişi (yerel)

---

## 📁 Dosya Transferi

GoConnect, P2P dosya transferi destekler.

### Desktop App ile

1. Sağ panelde üye listesinden kişi seçin
2. **"Dosya Gönder"** butonuna tıklayın
3. Dosya seçin
4. Transfer başlar

### CLI ile

Dosya transferi şu anda sadece **Desktop App** üzerinden veya **Interactive CLI** (planlanıyor) üzerinden yapılabilir.

---

## ⚙️ Ayarlar

### Desktop App Ayarları

**Genel:**
- 🌙 Karanlık/Aydınlık tema
- 🔔 Bildirim tercihleri
- 🚀 Başlangıçta otomatik başlat

**Ağ:**
- 🔄 Otomatik yeniden bağlanma
- 📊 Bant genişliği limiti
- 🌐 Proxy ayarları

**Hesap:**
- 👤 Profil düzenleme
- 🔐 Şifre değiştirme
- 🛡️ 2FA etkinleştirme

### CLI Ayarları

Ayar dosyası: `~/.config/goconnect/config.yaml`

```yaml
# config.yaml
server:
  url: "https://api.goconnect.io"

ui:
  theme: "dark"
  
notifications:
  enabled: true
  sound: true

auto_connect: true
```

---

## 🔧 Sorun Giderme

### Tanı Aracı (Doctor)

Kurulum veya bağlantı sorunları yaşıyorsanız, dahili tanı aracını kullanın:

```bash
goconnect doctor
```

Bu komut:
- Sistem gereksinimlerini kontrol eder
- WireGuard kurulumunu doğrular
- Sunucu bağlantısını test eder
- Config dosyasını analiz eder

### Bağlantı Sorunları

| Sorun | Çözüm |
|-------|-------|
| "Bağlanılamıyor" | İnternet bağlantınızı kontrol edin |
| "Timeout" | Firewall ayarlarını kontrol edin (UDP 51820) |
| "Authentication failed" | Tekrar giriş yapın (`goconnect login`) |
| "Peer unreachable" | Karşı tarafın bağlı olduğundan emin olun |

### Firewall Ayarları

WireGuard için UDP 51820 portunu açın:

```bash
# Linux (UFW)
sudo ufw allow 51820/udp

# Windows (PowerShell - Admin)
New-NetFirewallRule -DisplayName "WireGuard" -Direction Inbound -Protocol UDP -LocalPort 51820 -Action Allow
```

### Log Dosyaları

```bash
# CLI logs
~/.config/goconnect/logs/

# Desktop logs
# Windows: %APPDATA%/goconnect/logs/
# macOS: ~/Library/Application Support/goconnect/logs/
# Linux: ~/.local/share/goconnect/logs/
```

### Sıfırlama

Tüm ayarları sıfırlamak için:

```bash
# CLI
rm -rf ~/.config/goconnect

# Sonra tekrar başlatın
goconnect
```

---

## ❓ Sık Sorulan Sorular

**S: GoConnect ücretsiz mi?**
A: Evet, GoConnect açık kaynak ve ücretsizdir.

**S: Kaç kişi aynı ağa bağlanabilir?**
A: Varsayılan olarak 256 üye. Self-hosted sunucularda sınırsız.

**S: Verilerim güvende mi?**
A: Evet, tüm trafik WireGuard ile uçtan uca şifrelenir.

**S: Hangi işletim sistemlerini destekliyorsunuz?**
A: Windows 10+, macOS 12+, Ubuntu 20.04+ ve diğer Linux dağıtımları.

**S: Mobil uygulama var mı?**
A: Henüz yok, roadmap'te planlanıyor.

---

## 📞 Destek

- 📚 [Dokümantasyon](https://github.com/orhaniscoding/goconnect/docs)
- 🐛 [Bug Bildirimi](https://github.com/orhaniscoding/goconnect/issues)
- 💬 [Discord Topluluğu](https://discord.gg/goconnect)
- 📧 [Email](mailto:support@goconnect.io)

---

## 📄 Lisans

MIT License - Detaylar için [LICENSE](../LICENSE) dosyasına bakın.
