# 🔗 GoConnect

> **"Discord, but for networks."**

GoConnect, internetteki insanların sanki aynı yerel ağdaymış gibi görünmesini sağlayan kullanıcı dostu bir sanal LAN platformudur.

[![Release](https://img.shields.io/github/v/release/orhaniscoding/goconnect?style=flat-square)](https://github.com/orhaniscoding/goconnect/releases)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev)

---

## 📖 İçindekiler

- [GoConnect Nedir?](#-goconnect-nedir)
- [Kimler İçin?](#-kimler-için)
- [Nasıl Çalışır?](#-nasıl-çalışır)
- [Kurulum](#-kurulum)
- [Kullanım](#-kullanım)
- [Özellikler](#-özellikler)
- [Mimari](#-mimari)
- [Geliştirme](#-geliştirme)
- [SSS](#-sss)
- [Katkıda Bulunma](#-katkıda-bulunma)
- [Lisans](#-lisans)

---

## 🤔 GoConnect Nedir?

GoConnect, **tek bir uygulama** ile:

- 🌐 **Kendi ağını oluştur** - Arkadaşlarınla özel LAN partisi
- 🔗 **Başka ağlara katıl** - Davet linki ile tek tıkla bağlan
- 💬 **Sohbet et** - Discord benzeri metin kanalları
- 🎮 **Oyun oyna** - LAN oyunları internet üzerinden

### Discord ile Karşılaştırma

| Discord | GoConnect |
|---------|-----------|
| Ses/Video sunucuları | **Ağ sunucuları** |
| Ses kanalları | **Sanal LAN'lar** |
| Sunucu oluştur | **Ağ oluştur** |
| Sunucuya katıl | **Ağa bağlan** |
| Metin kanalları | **Metin kanalları** ✓ |

---

## 👥 Kimler İçin?

### 🎮 Oyuncular
- Minecraft LAN dünyalarını arkadaşlarla paylaş
- Eski LAN oyunlarını internet üzerinden oyna
- Düşük gecikmeli oyun deneyimi

### 💼 Uzaktan Çalışanlar
- Ofis kaynaklarına güvenli erişim
- Ekip içi dosya paylaşımı
- Basit VPN alternatifi

### 🏠 Ev Kullanıcıları
- Evdeki cihazlara dışarıdan erişim
- Aile ile güvenli dosya paylaşımı
- NAS'a uzaktan bağlantı

### 👨‍💻 Geliştiriciler
- Test ortamları oluşturma
- Mikroservis iletişimi
- Konteyner ağları

---

## ⚙️ Nasıl Çalışır?

```
┌─────────────────────────────────────────────────────────────────┐
│                        GoConnect App                             │
│                                                                  │
│  ┌──────────────────┐          ┌──────────────────┐             │
│  │  Ağ Oluştur 🌐   │          │   Ağa Katıl 🔗   │             │
│  │                  │          │                  │             │
│  │ Kendi sunucunu   │          │ Davet linki ile  │             │
│  │ başlat ve        │          │ başka birisinin  │             │
│  │ arkadaşlarını    │          │ ağına bağlan     │             │
│  │ davet et         │          │                  │             │
│  └────────┬─────────┘          └────────┬─────────┘             │
│           │                             │                        │
│           ▼                             ▼                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              WireGuard Güvenli Tünel                     │    │
│  │         (Otomatik yapılandırma - siz bir şey             │    │
│  │          yapmanıza gerek yok!)                           │    │
│  └─────────────────────────────────────────────────────────┘    │
│           │                             │                        │
│           ▼                             ▼                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Sanal Yerel Ağ                         │   │
│  │                                                           │   │
│  │   👤 Sen          👤 Arkadaş 1      👤 Arkadaş 2         │   │
│  │   10.0.1.1        10.0.1.2          10.0.1.3             │   │
│  │                                                           │   │
│  │   Artık hepiniz aynı LAN'dasınız!                        │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Basit Adımlar

1. **İndir** → GoConnect uygulamasını indir
2. **Aç** → Uygulamayı çalıştır
3. **Seç** → "Ağ Oluştur" veya "Ağa Katıl"
4. **Bağlan** → Tek tıkla bağlan!

---

## 📥 Kurulum

### Seçenek 1: Masaüstü Uygulaması (Önerilen)

En kolay yol! Tek bir uygulama ile her şeyi yapabilirsin.

| Platform | İndir |
|----------|-------|
| **Windows** | [GoConnect-Windows.exe](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **macOS (Intel)** | [GoConnect-macOS-Intel.dmg](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **macOS (Apple Silicon)** | [GoConnect-macOS-ARM.dmg](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **Linux (Debian/Ubuntu)** | [GoConnect-Linux.deb](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **Linux (AppImage)** | [GoConnect-Linux.AppImage](https://github.com/orhaniscoding/goconnect/releases/latest) |

### Seçenek 2: Terminal Uygulaması

Terminal kullanmayı sevenler için interaktif CLI.

```bash
# Linux/macOS
curl -fsSL https://get.goconnect.io | sh

# veya manuel indirme
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-linux-amd64
chmod +x goconnect-cli-linux-amd64
./goconnect-cli-linux-amd64
```

```powershell
# Windows (PowerShell)
irm https://get.goconnect.io/windows | iex

# veya manuel indirme
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-windows-amd64.exe" -OutFile "goconnect.exe"
.\goconnect.exe
```

### Seçenek 3: Docker

Sunucu olarak çalıştırmak için.

```bash
docker run -d \
  --name goconnect \
  --cap-add NET_ADMIN \
  -p 8080:8080 \
  -p 51820:51820/udp \
  ghcr.io/orhaniscoding/goconnect:latest
```

---

## 🎯 Kullanım

### Ağ Oluşturma (Host)

**Masaüstü Uygulaması:**
1. GoConnect'i aç
2. "Ağ Oluştur" butonuna tıkla
3. Ağ adı gir (örn: "Minecraft Sunucum")
4. "Oluştur" butonuna tıkla
5. Davet linkini arkadaşlarınla paylaş!

**Terminal:**
```bash
$ goconnect

  🔗 GoConnect - Discord, but for networks

  ? Ne yapmak istiyorsun?
  ❯ Ağ Oluştur
    Ağa Katıl
    Ayarlar
    Çıkış

# "Ağ Oluştur" seç ve yönergeleri takip et
```

### Ağa Katılma (Client)

**Masaüstü Uygulaması:**
1. GoConnect'i aç
2. "Ağa Katıl" butonuna tıkla
3. Davet linkini yapıştır
4. "Bağlan" butonuna tıkla
5. Artık ağdasın!

**Terminal:**
```bash
$ goconnect join gc://invite.goconnect.io/abc123

✓ Bağlantı başarılı!
  Ağ: Minecraft Sunucum
  IP Adresin: 10.0.1.5
  Çevrimiçi: 3 kişi
```

### Hızlı Komutlar (Terminal)

| Komut | Açıklama |
|-------|----------|
| `goconnect` | İnteraktif mod |
| `goconnect create "Ağ Adı"` | Hızlı ağ oluştur |
| `goconnect join <link>` | Hızlı katıl |
| `goconnect list` | Ağlarını listele |
| `goconnect status` | Bağlantı durumu |
| `goconnect disconnect` | Bağlantıyı kes |
| `goconnect help` | Yardım |

---

## ✨ Özellikler

### Temel Özellikler (Ücretsiz)

| Özellik | Açıklama |
|---------|----------|
| 🌐 **Ağ Oluşturma** | Kendi sanal LAN'ını oluştur |
| 🔗 **Ağa Katılma** | Davet linki ile tek tıkla katıl |
| 💬 **Metin Sohbeti** | Discord benzeri sohbet kanalları |
| 👥 **Üye Yönetimi** | Davet, çıkarma, yasaklama |
| 🔒 **Güvenli Bağlantı** | WireGuard şifreleme |
| 🖥️ **Çoklu Platform** | Windows, macOS, Linux |
| 📱 **Çoklu Cihaz** | Aynı hesapla birden fazla cihaz |

### Gelecek Özellikler

| Özellik | Durum |
|---------|-------|
| 📱 Mobil Uygulama | 🔜 Yakında |
| 🎤 Sesli Sohbet | 📋 Planlandı |
| 📹 Görüntülü Görüşme | 📋 Planlandı |
| 🎮 Oyun Entegrasyonu | 📋 Planlandı |

---

## 🏗️ Mimari

GoConnect üç ana bileşenden oluşur:

```
┌─────────────────────────────────────────────────────────────┐
│                     GoConnect Mimarisi                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              GoConnect App (Tauri)                   │    │
│  │                                                      │    │
│  │  • Masaüstü uygulaması (Windows/macOS/Linux)        │    │
│  │  • Hem host hem client olabilir                     │    │
│  │  • Modern Discord benzeri arayüz                    │    │
│  │  • Sistem tepsisinde çalışır                        │    │
│  └─────────────────────────────────────────────────────┘    │
│                            │                                 │
│                            │                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              GoConnect CLI                           │    │
│  │                                                      │    │
│  │  • Terminal uygulaması                              │    │
│  │  • İnteraktif TUI arayüz                            │    │
│  │  • Aynı özellikler, terminal'den                    │    │
│  │  • Sunucu/headless ortamlar için ideal             │    │
│  └─────────────────────────────────────────────────────┘    │
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              GoConnect Core (Go)                     │    │
│  │                                                      │    │
│  │  • WireGuard yönetimi                               │    │
│  │  • Ağ oluşturma ve yönetim                          │    │
│  │  • Kullanıcı kimlik doğrulama                       │    │
│  │  • P2P bağlantı koordinasyonu                       │    │
│  │  • Sohbet ve mesajlaşma                             │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Teknoloji Yığını

| Katman | Teknoloji | Neden? |
|--------|-----------|--------|
| **Desktop App** | Tauri + React | Küçük boyut, native performans |
| **CLI** | Go + Bubbletea | Çapraz platform, tek binary |
| **Core** | Go | Hızlı, güvenli, çapraz platform |
| **Networking** | WireGuard | Modern, hızlı VPN protokolü |
| **Database** | SQLite/PostgreSQL | Gömülü veya ölçeklenebilir |

---

## 🛠️ Geliştirme

### Gereksinimler

- Go 1.24+
- Node.js 20+ (Desktop App için)
- Rust (Desktop App için)

### Kaynak Koddan Derleme

```bash
# Repo'yu klonla
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# CLI derle
cd goconnect-cli
go build -o goconnect ./cmd/goconnect

# Desktop App derle
cd ../desktop-client
npm install
npm run tauri build
```

### Proje Yapısı

```
goconnect/
├── desktop-client/        # Tauri masaüstü uygulaması
│   ├── src/               # React frontend
│   ├── src-tauri/         # Rust backend
│   └── package.json
├── goconnect-cli/         # Terminal uygulaması (Go)
│   ├── cmd/goconnect/     # Ana komut
│   ├── internal/          # İç paketler
│   └── go.mod
├── goconnect-core/        # Ortak kütüphane (Go)
│   ├── network/           # Ağ yönetimi
│   ├── wireguard/         # WireGuard entegrasyonu
│   ├── auth/              # Kimlik doğrulama
│   └── go.mod
├── docs/                  # Dokümantasyon
├── README.md              # Bu dosya
└── LICENSE                # MIT Lisansı
```

---

## ❓ SSS

### Genel Sorular

<details>
<summary><b>GoConnect ücretsiz mi?</b></summary>

Evet! Temel özellikler tamamen ücretsiz. Gelecekte premium özellikler eklenebilir ama çekirdek işlevsellik her zaman ücretsiz kalacak.
</details>

<details>
<summary><b>Hangi platformlarda çalışır?</b></summary>

- ✅ Windows 10/11
- ✅ macOS 11+ (Intel ve Apple Silicon)
- ✅ Linux (Ubuntu 20.04+, Debian 11+, Fedora 35+)
- 🔜 Android (yakında)
- 🔜 iOS (yakında)
</details>

<details>
<summary><b>VPN ile arasındaki fark nedir?</b></summary>

GoConnect bir VPN değil, sanal LAN platformudur:
- **VPN**: Tüm trafiği bir sunucu üzerinden yönlendirir
- **GoConnect**: Sadece ağdaki cihazlar arasında doğrudan bağlantı kurar

Bu sayede daha düşük gecikme ve daha yüksek hız elde edilir.
</details>

<details>
<summary><b>Güvenli mi?</b></summary>

Evet! GoConnect, endüstri standardı WireGuard şifreleme kullanır:
- ChaCha20 simetrik şifreleme
- Curve25519 anahtar değişimi
- Blake2s hash fonksiyonu
- Poly1305 mesaj kimlik doğrulama
</details>

### Teknik Sorular

<details>
<summary><b>Port yönlendirme gerekli mi?</b></summary>

Çoğu durumda hayır! GoConnect, NAT traversal teknikleri kullanır:
- UDP hole punching
- STUN/TURN sunucuları
- Relay sunucuları (son çare)

Eğer doğrudan bağlantı kurulamazsa otomatik olarak relay kullanılır.
</details>

<details>
<summary><b>Bant genişliği limiti var mı?</b></summary>

GoConnect sunucuları üzerinden geçen trafik için limit yoktur çünkü trafik doğrudan cihazlar arasında akar. Relay kullanılması durumunda bazı limitler olabilir.
</details>

<details>
<summary><b>Kaç cihaz bağlanabilir?</b></summary>

Tek bir ağa teorik olarak 65.534 cihaz bağlanabilir (/16 subnet). Pratik limit donanım ve bant genişliğinize bağlıdır.
</details>

---

## 🤝 Katkıda Bulunma

Katkılarınızı bekliyoruz! 

### Nasıl Katkıda Bulunabilirim?

1. **Bug Raporla**: [Issue aç](https://github.com/orhaniscoding/goconnect/issues/new)
2. **Özellik Öner**: [Discussion başlat](https://github.com/orhaniscoding/goconnect/discussions)
3. **Kod Katkısı**: Fork → Branch → PR

### Geliştirme Kuralları

- Conventional Commits kullan (`feat:`, `fix:`, `docs:` vb.)
- Testleri çalıştır: `make test`
- Lint kontrolü: `make lint`

Detaylar için [CONTRIBUTING.md](CONTRIBUTING.md) dosyasına bak.

---

## 📄 Lisans

Bu proje [MIT Lisansı](LICENSE) altında lisanslanmıştır.

```
MIT License - Özgürce kullanın, değiştirin, dağıtın!
```

---

## 🙏 Teşekkürler

- [WireGuard](https://www.wireguard.com/) - Modern VPN protokolü
- [Tauri](https://tauri.app/) - Masaüstü uygulama framework'ü
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework'ü
- Tüm açık kaynak katkıda bulunanlar

---

## 📞 İletişim

- **GitHub**: [@orhaniscoding](https://github.com/orhaniscoding)
- **Issues**: [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues)
- **Discussions**: [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)

---

<div align="center">

**[⬆ Başa Dön](#-goconnect)**

❤️ ile yapıldı

</div>
