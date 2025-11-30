# 📚 GoConnect Kullanım Kılavuzu

Bu kılavuz, GoConnect'in tüm özelliklerini detaylı şekilde açıklar.

---

## 📋 İçindekiler

1. [Giriş](#1-giriş)
2. [Kurulum](#2-kurulum)
3. [İlk Kullanım](#3-i̇lk-kullanım)
4. [Ağ Yönetimi](#4-ağ-yönetimi)
5. [Üye Yönetimi](#5-üye-yönetimi)
6. [Sohbet](#6-sohbet)
7. [Ayarlar](#7-ayarlar)
8. [Sorun Giderme](#8-sorun-giderme)

---

## 1. Giriş

### GoConnect Nedir?

GoConnect, internetteki cihazları sanki aynı yerel ağdaymış gibi birbirine bağlayan bir platformdur. 

**Temel Kavramlar:**

| Kavram | Açıklama | Örnek |
|--------|----------|-------|
| **Ağ (Network)** | Sanal LAN ortamı | "Minecraft Sunucum" |
| **Host** | Ağı oluşturan kişi | Sunucu sahibi |
| **Üye (Member)** | Ağa katılan kişi | Oyuncular |
| **Davet Linki** | Ağa katılım bağlantısı | `gc://join.goconnect.io/abc123` |
| **IP Adresi** | Ağ içindeki adres | `10.0.1.5` |

### Desteklenen Platformlar

| Platform | Masaüstü App | Terminal App | Durum |
|----------|--------------|--------------|-------|
| Windows 10/11 | ✅ | ✅ | Hazır |
| macOS 11+ | ✅ | ✅ | Hazır |
| Linux | ✅ | ✅ | Hazır |
| Android | 📱 | - | Yakında |
| iOS | 📱 | - | Yakında |

---

## 2. Kurulum

### 2.1 Sistem Gereksinimleri

**Minimum:**
- İşlemci: 1 GHz
- RAM: 512 MB
- Disk: 100 MB
- Ağ: İnternet bağlantısı

**Önerilen:**
- İşlemci: 2+ GHz
- RAM: 2 GB
- Disk: 500 MB
- Ağ: 10+ Mbps

### 2.2 İndirme

[GitHub Releases](https://github.com/orhaniscoding/goconnect/releases/latest) sayfasından indirin.

### 2.3 Platform Bazlı Kurulum

#### Windows

1. `GoConnect-Setup.exe` dosyasını çalıştırın
2. "Next" butonlarıyla ilerleyin
3. Kurulum konumunu seçin (varsayılan önerilir)
4. "Install" butonuna tıklayın
5. "Finish" ile tamamlayın

**Not:** Windows Defender uyarısı çıkarsa "More info" → "Run anyway" seçin.

#### macOS

1. `.dmg` dosyasını açın
2. GoConnect ikonunu Applications'a sürükleyin
3. İlk açılışta Gatekeeper uyarısı çıkacak
4. System Preferences → Security → "Open Anyway" tıklayın

**Not:** Apple Silicon (M1/M2/M3) için ARM sürümünü indirin.

#### Linux

**Debian/Ubuntu:**
```bash
sudo dpkg -i goconnect_*.deb
sudo apt-get install -f  # Bağımlılıkları çöz
```

**Fedora/RHEL:**
```bash
sudo rpm -i goconnect_*.rpm
```

**AppImage (Tüm dağıtımlar):**
```bash
chmod +x GoConnect-*.AppImage
./GoConnect-*.AppImage
```

**Snap:**
```bash
sudo snap install goconnect
```

---

## 3. İlk Kullanım

### 3.1 Uygulamayı Başlatma

**Masaüstü:**
- Windows: Başlat menüsünden "GoConnect"
- macOS: Applications → GoConnect
- Linux: Uygulama menüsünden veya `goconnect` komutuyla

**Terminal:**
```bash
goconnect
```

### 3.2 Ana Ekran

```
┌────────────────────────────────────────────────────────────┐
│  🔗 GoConnect                                    ─ □ ✕    │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                                                     │  │
│  │            🌐 Hoş Geldiniz!                        │  │
│  │                                                     │  │
│  │   GoConnect ile arkadaşlarınla aynı ağda ol.       │  │
│  │                                                     │  │
│  │   ┌───────────────┐    ┌───────────────┐          │  │
│  │   │ Ağ Oluştur    │    │  Ağa Katıl    │          │  │
│  │   │     🌐        │    │     🔗        │          │  │
│  │   └───────────────┘    └───────────────┘          │  │
│  │                                                     │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                            │
│  ────────────────────────────────────────────────────────  │
│  📡 Ağlarım (0)                                           │
│  ────────────────────────────────────────────────────────  │
│                                                            │
│  Henüz hiçbir ağa bağlı değilsiniz.                       │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## 4. Ağ Yönetimi

### 4.1 Ağ Oluşturma

**Adımlar:**

1. "Ağ Oluştur" butonuna tıklayın
2. Ağ bilgilerini doldurun:

| Alan | Zorunlu | Açıklama | Örnek |
|------|---------|----------|-------|
| Ağ Adı | ✅ | Ağınızın ismi | "Minecraft Sunucum" |
| Açıklama | ❌ | Kısa açıklama | "Survival dünyası" |
| Alt Ağ | ❌ | IP aralığı | `10.0.1.0/24` (otomatik) |
| Şifre | ❌ | Katılım şifresi | Boş = şifresiz |

3. "Oluştur" butonuna tıklayın
4. Davet linkini kopyalayın

**Terminal:**
```bash
$ goconnect create "Minecraft Sunucum"

✅ Ağ oluşturuldu!

📋 Bilgiler:
   Ağ Adı: Minecraft Sunucum
   Alt Ağ: 10.0.1.0/24
   IP Adresin: 10.0.1.1

🔗 Davet Linki:
   gc://join.goconnect.io/abc123xyz

   Bu linki arkadaşlarınla paylaş!
```

### 4.2 Ağa Katılma

**Adımlar:**

1. "Ağa Katıl" butonuna tıklayın
2. Davet linkini yapıştırın
3. Şifre varsa girin
4. "Bağlan" butonuna tıklayın

**Terminal:**
```bash
$ goconnect join gc://join.goconnect.io/abc123xyz

🔗 Bağlanılıyor: Minecraft Sunucum...

✅ Bağlantı başarılı!

📋 Bilgiler:
   Ağ Adı: Minecraft Sunucum
   Alt Ağ: 10.0.1.0/24
   IP Adresin: 10.0.1.5
   Çevrimiçi: 3 kişi
```

### 4.3 Bağlantıyı Yönetme

**Bağlantıyı Kesme:**
- Ağ kartındaki "Bağlantıyı Kes" butonuna tıklayın
- veya `goconnect disconnect`

**Yeniden Bağlanma:**
- Ağ kartındaki "Bağlan" butonuna tıklayın
- veya `goconnect connect "Ağ Adı"`

### 4.4 Ağ Ayarları (Host)

Host olarak ağ ayarlarını değiştirebilirsiniz:

| Ayar | Açıklama |
|------|----------|
| Ağ Adı | İsmi değiştir |
| Açıklama | Açıklamayı güncelle |
| Şifre | Katılım şifresi ekle/kaldır |
| Davet Linki | Yeni link oluştur |
| Ağı Sil | Kalıcı olarak sil |

---

## 5. Üye Yönetimi

### 5.1 Üyeleri Görüntüleme

Ağ detay ekranında "Üyeler" sekmesinden tüm üyeleri görebilirsiniz:

```
┌─────────────────────────────────────────┐
│ 👥 Üyeler (5)                           │
├─────────────────────────────────────────┤
│ 🟢 Ahmet (Host)        10.0.1.1         │
│ 🟢 Mehmet              10.0.1.2         │
│ 🟢 Ayşe                10.0.1.3         │
│ 🟡 Fatma (Boşta)       10.0.1.4         │
│ ⚫ Ali (Çevrimdışı)    10.0.1.5         │
└─────────────────────────────────────────┘
```

**Durum Göstergeleri:**
- 🟢 Çevrimiçi
- 🟡 Boşta (5+ dakika aktivite yok)
- ⚫ Çevrimdışı

### 5.2 Üye Yönetimi (Host)

Host olarak üyeler üzerinde işlem yapabilirsiniz:

| İşlem | Açıklama |
|-------|----------|
| **Çıkar** | Üyeyi ağdan çıkarır (tekrar katılabilir) |
| **Yasakla** | Üyeyi kalıcı olarak yasaklar |
| **Yasağı Kaldır** | Yasaklı üyenin yasağını kaldırır |

---

## 6. Sohbet

### 6.1 Metin Kanalları

Her ağda varsayılan sohbet kanalları bulunur:

- **#genel** - Genel sohbet
- **#duyurular** - Sadece host yazabilir (opsiyonel)

### 6.2 Mesaj Gönderme

1. Kanal listesinden bir kanal seçin
2. Alt kısımdaki metin kutusuna yazın
3. Enter'a basın veya "Gönder" butonuna tıklayın

**Desteklenen Özellikler:**
- 📎 Dosya paylaşımı (5 MB'a kadar)
- 😀 Emoji
- @mention (kullanıcı etiketleme)
- Mesaj düzenleme/silme (kendi mesajlarınız)

---

## 7. Ayarlar

### 7.1 Genel Ayarlar

| Ayar | Açıklama | Varsayılan |
|------|----------|------------|
| Başlangıçta çalıştır | Bilgisayar açıldığında başlat | ✅ |
| Sistem tepsisine küçült | Kapatınca tepsiye git | ✅ |
| Bildirimler | Masaüstü bildirimleri | ✅ |
| Dil | Arayüz dili | Türkçe |
| Tema | Karanlık/Aydınlık | Karanlık |

### 7.2 Ağ Ayarları

| Ayar | Açıklama | Varsayılan |
|------|----------|------------|
| Otomatik bağlan | Uygulama açıldığında bağlan | ✅ |
| Yeniden bağlanma | Bağlantı koparsa tekrar dene | ✅ |
| DNS ayarları | Özel DNS sunucusu | Sistem |

### 7.3 Gelişmiş Ayarlar

| Ayar | Açıklama |
|------|----------|
| WireGuard arayüzü | Ağ arayüzü adı |
| Loglama seviyesi | Debug/Info/Warning/Error |
| Veri klasörü | Yapılandırma dosyaları konumu |

---

## 8. Sorun Giderme

### 8.1 Sık Karşılaşılan Sorunlar

<details>
<summary><b>❌ Bağlantı kurulamıyor</b></summary>

**Olası Nedenler:**
1. İnternet bağlantısı yok
2. Güvenlik duvarı engelliyor
3. Host çevrimdışı

**Çözümler:**
1. İnternet bağlantınızı kontrol edin
2. Güvenlik duvarında GoConnect'e izin verin
3. Host'un çevrimiçi olduğundan emin olun

```bash
# Windows Güvenlik Duvarı
netsh advfirewall firewall add rule name="GoConnect" dir=in action=allow program="C:\Program Files\GoConnect\goconnect.exe"

# Linux UFW
sudo ufw allow 51820/udp
```
</details>

<details>
<summary><b>❌ Diğer cihazlara ping atamıyorum</b></summary>

**Olası Nedenler:**
1. Hedef cihaz çevrimdışı
2. Güvenlik duvarı ping'i engelliyor
3. IP adresi yanlış

**Çözümler:**
1. Hedef cihazın çevrimiçi olduğunu kontrol edin
2. Her iki tarafta da ICMP'ye izin verin
3. IP adresini "Üyeler" listesinden doğrulayın
</details>

<details>
<summary><b>❌ Uygulama açılmıyor</b></summary>

**Çözümler:**
1. Bilgisayarı yeniden başlatın
2. Uygulamayı yeniden yükleyin
3. Günlük dosyalarını kontrol edin:
   - Windows: `%APPDATA%\GoConnect\logs`
   - macOS: `~/Library/Logs/GoConnect`
   - Linux: `~/.local/share/goconnect/logs`
</details>

### 8.2 Günlükleri Görüntüleme

**Masaüstü:**
Ayarlar → Gelişmiş → "Günlükleri Aç"

**Terminal:**
```bash
goconnect logs
goconnect logs --level debug
```

### 8.3 Destek Alma

1. [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues) - Bug raporları
2. [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions) - Sorular
3. [FAQ](../README.md#-sss) - Sık sorulan sorular

---

<div align="center">

**[← Ana Sayfa](../README.md)** | **[Hızlı Başlangıç →](../QUICK_START.md)**

</div>
