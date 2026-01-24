# ❓ Sık Sorulan Sorular (FAQ)

GoConnect hakkında en sık sorulan sorular ve yanıtları.

---

## 📑 İçindekiler

- [Genel Sorular](#genel-sorular)
- [Kurulum ve Kurulum](#kurulum-ve-kurulum)
- [Kullanım](#kullanım)
- [Teknik Sorular](#teknik-sorular)
- [Güvenlik ve Gizlilik](#güvenlik-ve-gizlilik)
- [Platform-Specific](#platform-specific)

---

## 🌐 Genel Sorular

### GoConnect nedir?

GoConnect, internet üzerindeki cihazların aynı yerel ağda (LAN) gibi görünmesini sağlayan bir **virtual LAN platformu**dur.

**Farkı nedir?**
- **VPN** degildir - Tüm trafiği yönlendirmez
- **P2P** - Direkt cihazlar arası bağlantı
- **Kullanımı kolay** - Tek tıkla bağlanın
- **Özellik zengin** - Sohbet, dosya transferi, sesli görüşme

---

### Ücretsiz mi?

**Evet!** Temel özellikler tamamen ücretsizdir:

- ✅ Sınırsız ağ oluşturma
- ✅ Sınırsız üye ekleme
- ✅ Tüm chat özellikleri
- ✅ Dosya transferi
- ✅ Sesli görüşme

Gelecekte bazı **premium** özellikler eklenebilir ancak temel işlevler her zaman ücretsiz kalacaktır.

---

### Hangi platformları destekliyor?

| Platform | Durum | Notlar |
|----------|--------|--------|
| **Windows 10/11** | ✅ Tam destek | Native app |
| **macOS 11+** | ✅ Tam destek | Intel + Apple Silicon |
| **Linux** | ✅ Tam destek | Ubuntu, Debian, Fedora, Arch |
| **Android** | 🔜 Yakında | Beta testi devam ediyor |
| **iOS** | 🔜 Planlandı | 2025 Q2 |

---

### Kaç cihaz bağlanabilir?

Teorik olarak **65,534 cihaz** (/16 subnet) bir ağa bağlanabilir.

**Pratik limitler:**
- Sunucu kapasitesi
- Network bant genişliği
- Donanım performansı

**Önerilen:**
- Küçük ağlar: 2-10 cihaz
- Orta ölçekli: 10-100 cihaz
- Büyük ağlar: 100-1000 cihaz

---

### Offline çalışabilir mi?

**Hayır.** GoConnect internet bağlantısı gerektirir.

**Neden?**
- Peer discovery (diğer cihazları bulmak)
- NAT traversal (bağlantı kurmak)
- Signaling ( handshake için)
- Relay (fallback)

Ancak bir kez bağlantı kurulduktan sonra:
- P2P bağlantısı **internet olmadan** çalışabilir
- Local file sharing yapabilir
- Chat geçmişini görebilirsiniz

---

## 🚀 Kurulum ve Kurulum

### Hangi sürümü indirmeliyim?

**Desktop App (Önerilen):**
- GUI arayüzü
- En kolay kullanım
- Tüm özellikler

**CLI (Terminal):**
- Sunucular için
- Geliştiriciler için
- Scripting

**Self-Hosted:**
- Kendi sunucunuzu kurun
- Tam kontrol
- Gizlilik

---

### Kurulum admin权限 gerektiriyor mu?

**Windows:** Evet, ilk kurulum için
- Driver yükleme (WireGuard)
- Firewall kuralı ekleme

**macOS:** Hayır (genellikle)
- Sadece first run'da password sorar

**Linux:** Evet (bazı durumlarda)
- Network interface oluşturma
- systemd service kurulumu

---

### Portable sürüm var mı?

**Evet!** Windows için portable version mevcut:

```powershell
# İndirin
goconnect-portable-windows-amd64.zip

# Çıkarın ve çalıştırın (kurulum gerektirmez)
.\goconnect.exe
```

---

### Kurulumu nasıl kaldırırım?

**Windows:**
```
Settings → Apps → GoConnect → Uninstall
```

**macOS:**
```bash
# App'i sil
rm -rf /Applications/GoConnect.app

# User data sil (isteğe bağlı)
rm -rf ~/Library/Application Support/com.goconnect.app
```

**Linux:**
```bash
# Debian/Ubuntu
sudo apt remove goconnect

# AppImage
rm GoConnect-amd64.AppImage
```

---

## 🎯 Kullanım

### İlk ağımı nasıl oluştururum?

**Desktop App:**
1. GoConnect'i açın
2. "Create Network" butonuna tıklayın
3. Ağ adı girin (örn: "Ailem")
4. "Create"e tıklayın
5. Davet linkini paylaşın

**CLI:**
```bash
goconnect
# "Create Network" seçeneğini seçin
# veya
goconnect create "Ağ Adı"
```

---

### Bir ağa nasıl katılırım?

**Davet linki ile:**
```
gc://invite.goconnect.io/abc123
```

Bu linke tıkladığınızda otomatik olarak Desktop App açılır.

**Manuel:**
1. "Join Network" butonuna tıklayın
2. Davet kodunu yapıştırın
3. "Connect"e tıklayın

---

### Birden fazla ağa katılabilir miyim?

**Evet!** Sınırsız ağa katılabilirsiniz.

**Not:** Her ağ farklı bir IP adresi alır.
- Ağ 1: 10.0.1.5
- Ağ 2: 10.0.2.5

---

### Ağ şifresi nasıl ayarlarım?

**Şu anda parola koruması yok.** Ağlar **davet linki** ile korunur.

**Gelecek:**
- Password protection (v3.1)
- 2FA (v3.2)
- SSO integration (roadmap)

---

### Üyeleri nasıl yönetirim?

**Mevcut özellikler:**
- ✅ Üye listesi görme
- ✅ Online/offline durum
- ✅ IP adresi görme

**Yakında:**
- 🔜 Kick member (v3.1)
- 🔜 Ban member (v3.2)
- 🔜 Admin roles (v3.2)

---

### Chat geçmişi nerede saklanır?

**Yerel olarak:**
- Desktop App: `~/AppData/GoConnect/chat.db`
- CLI: `~/.config/goconnect/chat.db`

**Sunucu:**
- Message history sunucuda saklanır (self-hosted için)
- 90 gün retention (ayarlanabilir)

---

## 🔧 Teknik Sorular

### WireGuard nedir?

**WireGuard** modern, güvenli, hızlı bir VPN protokolüdür.

**Özellikler:**
- ⚡ Çok hızlı (kernel-space)
- 🔒 Güvenli (modern kriptografi)
- 📦 Küçük kod tabanı (~4,000 satır)
- 🔐 Open source

**GoConnect neden WireGuard kullanıyor?**
- Native kernel desteği (Linux)
- Düşük latency
- Yüksek throughput
- Mobile-friendly

---

### Port forwarding gerekli mi?

**Genellikle hayır.** GoConnect **NAT traversal** kullanır.

**Techniques:**
- UDP hole punching
- STUN servers
- UPnP
- PCP

**Eğer başarısız olursa:** Relay kullanılır

---

### Relay nedir?

Relay, **son çare** olarak kullanılan bir sunucudur.

**Ne zaman devreye girer?**
- NAT traversal başarısız olduğunda
- Symmetric NAT (kısıtlayıcı firewall)
- Corporate network

**Dezavantajları:**
- Daha yavaş (tüm trafik sunucudan geçer)
- Daha fazla latency

---

### Hangi portları kullanıyor?

| Port | Protokol | Kullanım |
|------|----------|----------|
| **8080** | TCP | HTTP API |
| **51820** | UDP | WireGuard VPN |

**Firewall kuralları:**
```bash
# TCP 8080 (opsiyonel - sadece self-hosted)
# UDP 51820 (zorunlu)
```

---

### IPv6 desteği var mı?

**Evet!** GoConnect IPv6 destekler.

**Not:** Şu anda default olarak IPv4 kullanıyor.

---

## 🔒 Güvenlik ve Gizlilik

### Güvenli mi?

**Evet, çok güvenli.**

**Özellikler:**
- 🔒 End-to-end şifreleme (WireGuard)
- 🔐 Secure key exchange (Curve25519)
- 🛡️ Perfect Forward Secrecy
- ✅ No hardcoded secrets
- 🚨 Regular security audits

---

### Verilerimi görebilir miyim?

**Evet, tam kontrol.**

**Desktop/CLI:**
- Tüm veri yerel diskte
- İstediğiniz zaman export edin

**Self-hosted:**
- Veritabanı sizde
- Loglara erişin

---

### Kayıt tutuyor musunuz?

**Minimal logging:**
- ✅ Connection timestamps
- ✅ Error logs (debugging için)
- ✅ Security events (failed attempts)
- ❌ Chat içeriği loglanmaz
- ❌ File transfer içeriği loglanmaz

**Self-hosted:** Kendi loglama politikanızı belirleyin.

---

### Açık kaynak mı?

**Evet!** Tüm kodlar GitHub'da açık:

- [Core (Server)](https://github.com/orhaniscoding/goconnect/tree/main/core)
- [CLI](https://github.com/orhaniscoding/goconnect/tree/main/cli)
- [Desktop]((https://github.com/orhaniscoding/goconnect/tree/main/desktop)

**Lisans:** MIT License

---

### Third-party analytics kullanıyor musunuz?

**Hayır.**

- ❌ Google Analytics
- ❌ Telemetry
- ❌ Crash reporting (opsiyonel)

**Exception:** Anonim crash reporting (kullanıcı onayıyla).

---

## 🖥️ Platform-Specific

### Windows: Defender engelliyor

Bu **yanlış pozitif**. GoConnect güvenlidir.

**Geçici çözüm:**
```
Windows Security → Virus & threat protection
→ "Allow on device"
```

**Kalıcı çözüm:** Exclusion ekleyin (bkz: [Troubleshooting](TROUBLESHOOTING.md))

---

### macOS: "Damaged" hatası

**Çözüm:**
```bash
sudo xattr -cr /Applications/GoConnect.app
```

---

### Linux: Turkish karakter sorunu

**Çözüm:** UTF-8 locale kullanın

```bash
export LANG=tr_TR.UTF-8
export LC_ALL=tr_TR.UTF-8
```

---

## 📚 Ek Kaynaklar

Daha fazla bilgi için:

- 📖 [Kurulum Rehberi](INSTALLATION.md)
- 🔧 [Troubleshooting](TROUBLESHOOTING.md)
- 🛠️ [Geliştirme Rehberi](DEVELOPMENT.md)
- 🏠 [Self-Hosted Setup](SELF_HOSTED_SETUP.md)
- 📞 [Destek](https://github.com/orhaniscoding/goconnect/discussions)

---

### Sorunuz yok mu?

- 🐙 [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues/new) - Bug bildirin
- 💬 [Discussions](https://github.com/orhaniscoding/goconnect/discussions) - Sorun sorun
- 📧 [E-posta](mailto:support@goconnect.io) - Özel destek

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
