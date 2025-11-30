# 💻 GoConnect CLI

GoConnect'in terminal uygulaması. İnteraktif TUI arayüzü ile ağ oluşturun veya mevcut ağlara katılın.

> **Not:** Bu dizin `goconnect-cli` olarak yeniden adlandırılacak.

## ✨ Özellikler

- 🖥️ **İnteraktif TUI** - Bubbletea ile modern terminal arayüzü
- 🌐 **Ağ Oluştur** - Terminal'den ağ oluştur ve yönet
- 🔗 **Ağa Katıl** - Davet linki ile bağlan
- 📊 **Durum Görüntüle** - Bağlantı durumu, üyeler, IP adresleri
- 🔧 **Headless Mod** - Sunucularda arka planda çalıştır

## 🚀 Hızlı Başlangıç

### İndirme

```bash
# Linux
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-linux-amd64
chmod +x goconnect-cli-linux-amd64
sudo mv goconnect-cli-linux-amd64 /usr/local/bin/goconnect

# macOS
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-darwin-arm64
chmod +x goconnect-cli-darwin-arm64
sudo mv goconnect-cli-darwin-arm64 /usr/local/bin/goconnect
```

### Kullanım

```bash
# İnteraktif mod
goconnect

# Hızlı komutlar
goconnect create "Ağ Adı"    # Ağ oluştur
goconnect join <link>        # Ağa katıl
goconnect list               # Ağları listele
goconnect status             # Bağlantı durumu
goconnect disconnect         # Bağlantıyı kes
```

## 🎨 TUI Arayüzü

```
┌──────────────────────────────────────────────────────────────┐
│                    🔗 GoConnect CLI                          │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│   ? Ne yapmak istiyorsun?                                    │
│                                                              │
│   ❯ 🌐 Ağ Oluştur                                           │
│     🔗 Ağa Katıl                                            │
│     📋 Ağlarım                                              │
│     ⚙️  Ayarlar                                              │
│     ❌ Çıkış                                                 │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│   ↑/↓: seç  •  Enter: onayla  •  q: çık                     │
└──────────────────────────────────────────────────────────────┘
```

## 🛠️ Geliştirme

### Gereksinimler

- Go 1.24+
- WireGuard araçları (`wg`, `wg-quick`)

### Derleme

```bash
# Tek platform
go build -o goconnect ./cmd/daemon

# Tüm platformlar
make build-all
```

### Proje Yapısı

```
client-daemon/  (→ goconnect-cli)
├── cmd/
│   └── daemon/
│       └── main.go         # Giriş noktası
├── internal/
│   ├── tui/                # Terminal UI
│   │   ├── model.go        # TUI modeli
│   │   ├── views.go        # Ekranlar
│   │   └── styles.go       # Stiller
│   ├── network/            # Ağ yönetimi
│   ├── wireguard/          # WireGuard entegrasyonu
│   └── config/             # Yapılandırma
└── go.mod
```

## ⚙️ Yapılandırma

Yapılandırma dosyası konumları:
- **Linux:** `~/.config/goconnect/config.yaml`
- **macOS:** `~/Library/Application Support/GoConnect/config.yaml`
- **Windows:** `%APPDATA%\GoConnect\config.yaml`

### Örnek Yapılandırma

```yaml
# GoConnect CLI Yapılandırma
server:
  url: ""  # Boş = yeni ağ oluştur

wireguard:
  interface_name: goconnect0

daemon:
  local_port: 12345
  health_check_interval: 30s
```

## 🔧 Sistem Servisi

### Linux (systemd)

```bash
sudo ./goconnect install
sudo systemctl enable goconnect
sudo systemctl start goconnect
```

### macOS (launchd)

```bash
sudo ./goconnect install
```

### Windows (Windows Service)

```powershell
# Admin olarak çalıştır
.\goconnect.exe install
```

## 📄 Lisans

MIT License - Detaylar için [LICENSE](../LICENSE) dosyasına bakın.
