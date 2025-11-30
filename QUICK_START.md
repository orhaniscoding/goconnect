# 🚀 GoConnect Hızlı Başlangıç

Bu kılavuz, GoConnect'i 5 dakikada kullanmaya başlamanızı sağlar.

---

## 📋 İçindekiler

1. [İndir](#1-i̇ndir)
2. [Kur](#2-kur)
3. [Başlat](#3-başlat)
4. [Ağ Oluştur veya Katıl](#4-ağ-oluştur-veya-katıl)
5. [Kullan](#5-kullan)

---

## 1. İndir

### Masaüstü Uygulaması (Önerilen)

[GitHub Releases](https://github.com/orhaniscoding/goconnect/releases/latest) sayfasından işletim sisteminize uygun dosyayı indirin:

| İşletim Sistemi | Dosya |
|-----------------|-------|
| Windows | `GoConnect-Setup.exe` |
| macOS Intel | `GoConnect-Intel.dmg` |
| macOS Apple Silicon | `GoConnect-ARM.dmg` |
| Linux Debian/Ubuntu | `GoConnect.deb` |
| Linux Diğer | `GoConnect.AppImage` |

### Terminal Uygulaması

```bash
# Linux/macOS
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-$(uname -s)-$(uname -m)
chmod +x goconnect-cli-*
sudo mv goconnect-cli-* /usr/local/bin/goconnect
```

```powershell
# Windows PowerShell
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-windows-amd64.exe" -OutFile "$env:LOCALAPPDATA\goconnect.exe"
```

---

## 2. Kur

### Windows
1. `GoConnect-Setup.exe` dosyasını çift tıklayın
2. Kurulum sihirbazını takip edin
3. "Finish" butonuna tıklayın

### macOS
1. `.dmg` dosyasını açın
2. GoConnect ikonunu Applications klasörüne sürükleyin
3. İlk açılışta "Open Anyway" seçeneğini onaylayın

### Linux (Debian/Ubuntu)
```bash
sudo dpkg -i GoConnect.deb
```

### Linux (AppImage)
```bash
chmod +x GoConnect.AppImage
./GoConnect.AppImage
```

---

## 3. Başlat

### Masaüstü Uygulaması

1. GoConnect uygulamasını başlatın
2. Karşılama ekranı görünecek:

```
┌──────────────────────────────────────┐
│         🔗 GoConnect'e Hoşgeldiniz   │
│                                      │
│    "Discord, but for networks."      │
│                                      │
│   ┌────────────────────────────┐     │
│   │     🌐 Ağ Oluştur          │     │
│   │     Kendi ağını başlat     │     │
│   └────────────────────────────┘     │
│                                      │
│   ┌────────────────────────────┐     │
│   │     🔗 Ağa Katıl           │     │
│   │     Davet linki ile katıl  │     │
│   └────────────────────────────┘     │
│                                      │
└──────────────────────────────────────┘
```

### Terminal Uygulaması

```bash
$ goconnect

  🔗 GoConnect v2.28.2

  ? Ne yapmak istiyorsun?
  ❯ 🌐 Ağ Oluştur
    🔗 Ağa Katıl
    📋 Ağlarım
    ⚙️  Ayarlar
    ❌ Çıkış
```

---

## 4. Ağ Oluştur veya Katıl

### Seçenek A: Yeni Ağ Oluştur

**Ne zaman kullanmalı?**
- Arkadaşlarınla oyun oynamak istiyorsun
- Kendi özel LAN'ını kurmak istiyorsun
- Dosya paylaşımı için ağ lazım

**Adımlar:**

1. "Ağ Oluştur" seçeneğini seç
2. Ağ bilgilerini gir:
   - **Ağ Adı**: `Minecraft Sunucum`
   - **Açıklama**: `Arkadaşlarla survival dünyası`
3. "Oluştur" butonuna tıkla
4. Davet linkini kopyala ve arkadaşlarına gönder!

```
✅ Ağ oluşturuldu!

📋 Davet Linki:
   gc://join.goconnect.io/abc123xyz

🔗 Bu linki arkadaşlarınla paylaş!
```

### Seçenek B: Mevcut Ağa Katıl

**Ne zaman kullanmalı?**
- Birileri sana davet linki gönderdi
- Başka birinin ağına katılmak istiyorsun

**Adımlar:**

1. "Ağa Katıl" seçeneğini seç
2. Davet linkini yapıştır: `gc://join.goconnect.io/abc123xyz`
3. "Bağlan" butonuna tıkla
4. Bağlantı kurulacak!

```
✅ Bağlantı başarılı!

🌐 Ağ: Minecraft Sunucum
🖥️ IP Adresin: 10.0.1.5
👥 Çevrimiçi: 3 kişi

Artık aynı LAN'dasınız!
```

---

## 5. Kullan

### Bağlantı Durumunu Kontrol Et

**Masaüstü:**
- Sistem tepsisindeki GoConnect ikonuna bak
- 🟢 Yeşil = Bağlı
- 🔴 Kırmızı = Bağlı değil

**Terminal:**
```bash
$ goconnect status

🌐 Bağlı Ağlar:
   • Minecraft Sunucum (10.0.1.0/24)
     IP: 10.0.1.5
     Çevrimiçi: 3 kişi
```

### Diğer Cihazlara Eriş

Artık ağdaki diğer cihazlara IP adresleriyle erişebilirsin:

```bash
# Ping at
ping 10.0.1.2

# SSH bağlantısı
ssh user@10.0.1.3

# Dosya paylaşımı
\\10.0.1.4\shared  # Windows
smb://10.0.1.4/shared  # macOS
```

### Minecraft LAN Örneği

1. Minecraft'ı aç
2. Dünyayı aç → "Open to LAN"
3. Port numarasını not al (örn: 25565)
4. Arkadaşların "Direct Connect" ile bağlanır: `10.0.1.1:25565`

---

## 🎉 Tebrikler!

GoConnect'i başarıyla kurdun ve kullanmaya başladın!

### Sonraki Adımlar

- 📖 [Tam Kullanım Kılavuzu](docs/USER_GUIDE.md)
- ⚙️ [Gelişmiş Ayarlar](docs/ADVANCED.md)
- ❓ [SSS](README.md#-sss)
- 🐛 [Sorun Bildir](https://github.com/orhaniscoding/goconnect/issues)

### Yardım Gerekiyor mu?

- 💬 [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- 📧 Destek: issues sayfasından ulaşın

---

<div align="center">

**[← Ana Sayfa](README.md)**

</div>
