# 🔗 GoConnect

> **"Virtual LAN made simple."** / **"Sanal LAN basitliği."**

---

[English](#english) | [Türkçe](#türkçe)

---

## English

## 🤔 What is GoConnect?

GoConnect is a **user-friendly virtual LAN platform** that makes devices on the internet appear as if they're on the same local network.

### What Does This Mean in Simple Terms?

**Imagine:** You and your friend live in different cities. You want to play a game together, but the game only works on the same local network (LAN).

**Without GoConnect:** You cannot play together because you're not in the same house/building.

**With GoConnect:** The game *thinks* you're on the same network! You can play, share files, chat, and do anything you could do if you were sitting next to each other.

---

## 🎯 Why Would You Use GoConnect?

### For Gamers 🎮
- **Play LAN games over the internet** - Old games that only support LAN
- **Minecraft LAN worlds** - Share your Minecraft world with friends anywhere
- **Low latency** - Direct connection means less lag than online servers

### For Remote Workers 💼
- **Access office resources** - Connect to your office network from home
- **Team file sharing** - Share files with your team securely
- **VPN alternative** - Simpler than traditional VPNs

### For Home Users 🏠
- **Access home devices** - Connect to your home NAS, server, or computer from anywhere
- **Family file sharing** - Share photos and videos with family securely
- **Remote desktop** - Help family members with computer problems

### For Developers 👨‍💻
- **Test environments** - Simulate network topologies
- **Microservices** - Test distributed systems locally
- **Container networking** - Connect containers across different machines

---

## ✨ Key Features

| Feature | What It Means | Why It Matters |
|---------|---------------|----------------|
| 🌐 **Create Network** | Host your own virtual LAN | You control who joins |
| 🔗 **Join Networks** | Connect with one click | No technical knowledge needed |
| 💬 **Text Chat** | Built-in messaging | Don't need separate chat apps |
| 🗣️ **Voice Chat** | Real-time voice communication | Talk while gaming or working |
| 📁 **File Transfer** | P2P file sharing | Send files directly to peers |
| 👥 **Member Management** | Invite, kick, ban users | Control your network |
| 🔒 **Secure** | WireGuard encryption | Your data is safe |
| 🖥️ **Cross-Platform** | Windows, macOS, Linux | Works on all major OS |
| 📱 **Multi-Device** | Multiple devices per account | Connect from anywhere |
| 🔄 **Auto-Update** | Seamless background updates | Always have the latest version |

---

## 🚀 Quick Start (5 Minutes)

### What You'll Need:
- **Computer** with Windows 10+, macOS 11+, or Linux (Ubuntu 20.04+, Debian 11+, Fedora 35+)
- **Internet connection**
- **10 minutes** of your time

### Step-by-Step:

#### Step 1: Download GoConnect

**What is downloading?** Downloading means getting the GoConnect application file from the internet to your computer.

**How to download:**

1. Open your web browser (Chrome, Firefox, Edge, Safari, etc.)
2. Go to: https://github.com/orhaniscoding/goconnect/releases/latest
3. Find your operating system and click the download link

| Your Operating System | What to Download |
|----------------------|------------------|
| Windows | `GoConnect-Setup.exe` |
| Mac (Apple Silicon M1/M2/M3) | `GoConnect-aarch64.dmg` |
| Mac (Intel) | `GoConnect-x64.dmg` |
| Linux (Ubuntu/Debian) | `GoConnect.deb` |
| Linux (Any) | `GoConnect.AppImage` |

**Don't know which one?** Here's how to check:

**Windows:** Press `Win + R`, type `winver`, press Enter. You'll see "Windows 10" or "Windows 11"

**Mac:** Click Apple menu → About This Mac. Look at "Processor" or "Chip":
- If it says "Intel" → Download `GoConnect-x64.dmg`
- If it says "M1", "M2", "M3" → Download `GoConnect-aarch64.dmg`

**Linux:** Open terminal and run: `uname -m`
- If output is `x86_64` → Your system is 64-bit (most common)
- If output is `aarch64` or `arm64` → Your system is ARM-based

#### Step 2: Install GoConnect

**What is installing?** Installing means setting up the application so your computer can run it.

**How to install:**

**Windows:**
1. Find the downloaded `GoConnect-Setup.exe` file (usually in Downloads folder)
2. Double-click it
3. If Windows asks "Do you want to allow this app?" → Click **Yes**
4. Click **Next** through the installation wizard
5. Click **Finish** when done

**macOS:**
1. Find the downloaded `.dmg` file in Downloads
2. Double-click it (a window opens with GoConnect icon)
3. Drag GoConnect icon to Applications folder
4. Close the window
5. Eject the disk (drag it to Trash)

**Linux (Debian/Ubuntu with .deb):**
1. Open terminal (press `Ctrl + Alt + T`)
2. Type: `sudo dpkg -i Downloads/GoConnect.deb`
3. Type your password (you won't see it while typing, that's normal)
4. Press Enter

**Linux (AppImage - Any distro):**
1. Right-click the downloaded `GoConnect.AppImage`
2. Properties → Permissions → Check "Allow executing file as program"
3. Close and double-click the AppImage

#### Step 3: Open GoConnect

**What happens when you open it?** GoConnect starts running on your computer.

**How to open:**

**Windows:** Press Start, type "GoConnect", press Enter

**macOS:** Open Applications folder, double-click GoConnect

**Linux:** Press `Alt + F2`, type `goconnect`, press Enter
Or from terminal: `goconnect`

**What you'll see:**
```
┌─────────────────────────────────────┐
│         GoConnect                   │
├─────────────────────────────────────┤
│                                     │
│  What would you like to do?         │
│                                     │
│  ⦿ Create Network                  │
│    Join Network                     │
│    Settings                         │
│    Exit                             │
│                                     │
└─────────────────────────────────────┘
```

#### Step 4: Create Your First Network

**What is creating a network?** You're making a virtual private LAN that you and your friends can join.

**How to create:**

1. Click **"Create Network"** (or press Enter)
2. Type a name for your network (e.g., "My Gaming Network")
3. Press Enter
4. Wait a few seconds while GoConnect sets up the network

**What happens now:**
- GoConnect creates a virtual network interface on your computer
- You get a virtual IP address (like `10.0.1.1`)
- GoConnect generates an invite link

**What you'll see:**
```
✓ Network created successfully!

Network Name: My Gaming Network
Your IP: 10.0.1.1

Invite Link:
gc://invite.goconnect.io/abc123def456

Share this link with friends to let them join!
```

#### Step 5: Invite Friends

**What is inviting?** Sharing a special link that lets others join your network.

**How to invite:**

1. Copy the invite link (click it and press `Ctrl + C` or `Cmd + C`)
2. Send it to your friends via:
   - Email
   - Discord
   - WhatsApp
   - Any messaging app

**What happens when they click it:**
- Their GoConnect opens automatically
- They click "Join"
- They're now on your virtual LAN!

#### Step 6: You're Connected!

**What can you do now?**

**Play LAN games:**
1. Start your game (e.g., Minecraft)
2. Choose "LAN Game" or "Multiplayer → LAN"
3. Your friends will see your game in their server list
4. They join and play!

**Share files:**
1. Open GoConnect chat
2. Click the attachment icon
3. Select a file
4. Send to anyone on your network

**Voice chat:**
1. Click the microphone icon in GoConnect
2. Talk to your friends

---

## 📖 Detailed Installation Guides

Need help? We have detailed guides for every platform:

| Platform | Guide |
|----------|-------|
| **Windows** | [Windows Installation Guide](docs/en/installations/windows.md) |
| **macOS** | [macOS Installation Guide](docs/en/installations/macos.md) |
| **Linux** | [Linux Installation Guide](docs/en/installations/linux.md) |
| **Docker** | [Docker Installation Guide](docs/en/installations/docker.md) |

Each guide includes:
- ✅ System requirements
- ✅ Step-by-step instructions with screenshots
- ✅ Troubleshooting common problems
- ✅ Advanced configuration

---

## 🎓 Usage Guides

Learn how to use specific features:

| Guide | Description |
|-------|-------------|
| [Creating a Network](docs/en/guides/creating-network.md) | Host your own virtual LAN |
| [Joining a Network](docs/en/guides/joining-network.md) | Connect to existing networks |
| [Text Chat](docs/en/guides/chat.md) | Use built-in messaging |
| [Voice Chat](docs/en/guides/voice-chat.md) | Real-time voice communication |
| [File Transfer](docs/en/guides/file-transfer.md) | Share files peer-to-peer |
| [Member Management](docs/en/guides/member-management.md) | Manage network members |

---

## 🏠 Self-Hosting

Want to run your own GoConnect server?

**Why self-host?**
- 🔒 **Privacy** - Your data stays on your server
- 🎛️ **Control** - You control everything
- 🚀 **Performance** - No third-party dependencies
- 💰 **Cost** - Can be cheaper long-term

**Quick Start with Docker:**

```bash
# Download docker-compose file
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# Create environment file
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 16)
WG_SERVER_ENDPOINT=your-domain.com:51820
EOF

# Start server
docker compose up -d
```

**📖 Full Guide:** [Self-Hosting Guide](docs/en/self-hosting/overview.md)

The guide covers:
- Docker installation (recommended)
- Manual binary installation
- Configuration options
- Reverse proxy setup (Nginx/Caddy)
- Security checklist
- Monitoring and troubleshooting

---

## 🛠️ Development

Want to contribute or build from source?

### Requirements

| Tool | Version | Why? |
|------|---------|------|
| **Go** | 1.24+ | Core language (cli and core modules) |
| **Node.js** | 20+ | Desktop app frontend |
| **Rust** | Latest | Desktop app backend (Tauri) |
| **protoc** | Latest | Protocol Buffers compiler |

### Quick Start

```bash
# Clone repository
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# Build CLI
cd cli
go build -o goconnect ./cmd/goconnect
./goconnect

# Build Server
cd ../core
go build -o goconnect-server ./cmd/server
./goconnect-server

# Build Desktop App
cd ../desktop
npm install
npm run tauri build
```

**📖 Full Guide:** [Development Guide](docs/en/development/introduction.md)

---

## ❓ Frequently Asked Questions

<details>
<summary><b>Is GoConnect free?</b></summary>

Yes! GoConnect is completely free and open-source (MIT License). Core features will always remain free.
</details>

<details>
<summary><b>What platforms are supported?</b></summary>

✅ Windows 10/11
✅ macOS 11+ (Intel and Apple Silicon)
✅ Linux (Ubuntu 20.04+, Debian 11+, Fedora 35+)
🔜 Mobile apps coming soon
</details>

<details>
<summary><b>What's the difference from a VPN?</b></summary>

**VPN (Virtual Private Network):**
- Routes ALL your traffic through a server
- Server sees everything you do
- Slower because of server bottleneck
- Good for privacy, not great for speed

**GoConnect:**
- Creates direct connections between devices
- Only GoConnect traffic goes through the network
- Faster because it's peer-to-peer
- Perfect for gaming, file sharing, and LAN applications

**Think of it this way:**
- VPN = All your internet traffic goes through a tunnel
- GoConnect = Only specific apps/games go through the tunnel, everything else uses normal internet
</details>

<details>
<summary><b>Is it secure?</b></summary>

Yes! GoConnect uses industry-standard security:

**Encryption:**
- WireGuard protocol (used by militaries and corporations)
- ChaCha20 encryption (same algorithm used in HTTPS)
- Perfect Forward Secrecy (even if someone records your traffic, they can't decrypt it later)

**Authentication:**
- Each device has unique cryptographic keys
- No passwords that can be guessed
- Invite links are cryptographically signed

**Privacy:**
- No central server sees your data (peer-to-peer)
- You can self-host for complete control
- Open-source code anyone can audit

**But remember:** Like any tool, security depends on proper use. Always:
- Only join networks from people you trust
- Keep GoConnect updated
- Use strong passwords on your self-hosted server
</details>

<details>
<summary><b>Do I need port forwarding?</b></summary>

Usually **no!** GoConnect uses advanced techniques to connect without port forwarding:

**NAT Traversal:**
- UDP hole punching
- STUN servers (help discover your public IP)
- TURN relay (fallback when direct connection fails)

**When might you need port forwarding?**
- If you're behind a very restrictive firewall
- If both peers have symmetric NAT (rare)
- For self-hosted servers

**How to check if you need it:**
Just try connecting first! If it doesn't work, GoConnect will tell you.

**How to port forward (if needed):**
See our [Network Configuration Guide](docs/en/guides/network-config.md)
</details>

<details>
<summary><b>How many devices can connect?</b></summary>

**Theoretical limit:** 65,534 devices per network (/16 subnet)

**Practical limit:** Depends on your hardware and internet connection

**Realistic estimates:**
- Gaming: 10-50 players (depends on game)
- File sharing: 100+ users
- Chat: 1000+ users

**For larger deployments:** Consider running multiple networks or our enterprise edition (coming soon).
</details>

<details>
<summary><b>Does it work with NAT/CGNAT?</b></summary>

**Yes, usually!** GoConnect is designed to work through:

✅ NAT (Network Address Translation)
✅ CGNAT (Carrier-Grade NAT)
✅ Firewall
✅ Symmetric NAT (harder, but we try)

**How it works:**
1. We try direct connection first
2. If that fails, we use STUN to discover public IP
3. If that fails, we use TURN relay

**Success rate:** ~95% of connections succeed without any configuration
</details>

**More questions?** See [Full FAQ](docs/en/faq.md)

---

## 🤝 Contributing

We welcome contributions!

### How to Contribute

1. **Report Bugs:** [Open an issue](https://github.com/orhaniscoding/goconnect/issues/new?template=bug_report.md)
2. **Suggest Features:** [Start a discussion](https://github.com/orhaniscoding/goconnect/discussions)
3. **Submit Code:** Fork → Branch → PR

### Development Guidelines

- Follow [Conventional Commits](https://www.conventionalcommits.org/)
- Write tests for new features
- Update documentation
- Keep it simple (see our [Zero-Dependency Policy](docs/development/zero-dependency.md))

**📖 Full Guide:** [Contributing Guide](CONTRIBUTING.md)

---

## 📄 License

MIT License - See [LICENSE](LICENSE) for details.

**What this means:**
- ✅ Use for free
- ✅ Modify as you want
- ✅ Distribute (even commercially)
- ✅ Keep this license notice

**In short:** Do whatever you want, just don't sue us if it breaks.

---

## 🙏 Acknowledgments

Built with amazing open-source tools:

- [WireGuard](https://www.wireguard.com/) - Modern VPN protocol
- [Tauri](https://tauri.app/) - Desktop application framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI
- [Go](https://go.dev/) - Programming language
- [React](https://react.dev/) - Frontend library

---

## 📞 Contact & Support

**Get Help:**
- 📖 [Documentation](docs/)
- 💬 [Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- 🐛 [Bug Reports](https://github.com/orhaniscoding/goconnect/issues)
- ✉️ Email: support@goconnect.io (coming soon)

**Follow Us:**
- GitHub: [@orhaniscoding](https://github.com/orhaniscoding)
- Website: https://goconnect.io (coming soon)

---

## 🗺️ Roadmap

### v1.2.0 (In Progress)
- [ ] Mobile apps (iOS/Android)
- [ ] Video chat
- [ ] Performance improvements

### v1.3.0 (Planned)
- [ ] Game integration plugins
- [ ] Advanced network management
- [ ] Web client

### v2.0.0 (Future)
- [ ] Enterprise features
- [ ] Custom network topologies
- [ ] API for third-party apps

**Suggest a feature:** [Open a feature request](https://github.com/orhaniscoding/goconnect/issues/new?template=feature_request.md)

---

<div align="center">

**⭐ Star us on GitHub!**

Made with ❤️ by [orhaniscoding](https://github.com/orhaniscoding)

[⬆ Back to Top](#-goconnect)

</div>

---

## Türkçe

## 🤔 GoConnect Nedir?

GoConnect, **kullanıcı dostu bir sanal LAN platformudur**. İnternet üzerindeki cihazların, aynı yerel ağdaymış gibi görünmesini sağlar.

### Bu Basit Terimlerle Ne Anlama Gelir?

**Hayal Edin:** Siz ve arkadaşınız farklı şehirlerde yaşıyorsunuz. Birlikte oyun oynamak istiyorsunuz ama oyun sadece aynı yerel ağda (LAN) çalışıyor.

**GoConnect Olmadan:** Aynı ev/binada olmadığınız için birlikte oynayamazsınız.

**GoConnect ile:** Oyun, sizi **aynı ağda sanıyor**! Yan yana oturuyormuşsunuz gibi oyun oynayabilir, dosya paylaşabilir, sohbet edebilir ve yan yanda yapabileceğiniz her şeyi yapabilirsiniz.

---

## 🎯 Neden GoConnect Kullanmalısınız?

### Oyuncular İçin 🎮
- **İnternet üzerinden LAN oyunları oyna** - Sadece LAN destekli eski oyunlar
- **Minecraft LAN dünyaları** - Minecraft dünyanı her yerden arkadaşlarla paylaş
- **Düşük gecikme** - Doğrudan bağlantı = online sunuculardan daha az lag

### Uzaktan Çalışanlar İçin 💼
- **Ofis kaynaklarına eriş** - Evden ofis ağına bağlan
- **Ekip dosya paylaşımı** - Ekiple güvenli dosya paylaşımı
- **VPN alternatifi** - Geleneksel VPN'lerden daha basit

### Ev Kullanıcıları İçin 🏠
- **Ev cihazlarına eriş** - Her yerden evdeki NAS'a, sunucuya veya bilgisayara bağlan
- **Aile dosya paylaşımı** - Aileyle güvenli fotoğraf ve video paylaşımı
- **Uzak masaüstü** - Aile üyelerinin bilgisayar sorunlarına yardım et

### Geliştiriciler İçin 👨‍💻
- **Test ortamları** - Ağ topolojilerini simüle et
- **Mikroservisler** - Dağıtık sistemleri yerel olarak test et
- **Konteyner ağları** - Farklı makinelerdeki konteynerleri bağla

---

## ✨ Temel Özellikler

| Özellik | Ne Anlama Gelir? | Neden Önemli? |
|---------|-----------------|---------------|
| 🌐 **Ağ Oluştur** | Kendi sanal LAN'ını host et | Kimin katıldığını sen kontrol edersin |
| 🔗 **Ağlara Katıl** | Tek tıkla bağlan | Teknik bilgi gerektirmez |
| 💬 **Metin Sohbeti** | Yerleşik mesajlaşma | Ayrı sohbet uygulamalarına gerek yok |
| 🗣️ **Sesli Sohbet** | Gerçek zamanlı sesli iletişim | Oyun veya çalışırken konuşma |
| 📁 **Dosya Transferi** | P2P dosya paylaşımı | Dosyaları doğrudan akranlara gönder |
| 👥 **Üye Yönetimi** | Kullanıcı davet et, at, engelle | Ağını kontrol et |
| 🔒 **Güvenli** | WireGuard şifreleme | Verilerin güvende |
| 🖥️ **Çapraz Platform** | Windows, macOS, Linux | Tüm major işletim sistemlerinde çalışır |
| 📱 **Çoklu Cihaz** | Hesap başına birden fazla cihaz | Her yerden bağlan |
| 🔄 **Otomatik Güncelleme** - Sorunsuz arka plan güncellemeleri | Her zaman en son sürüm |

---

## 🚀 Hızlı Başlangıç (5 Dakika)

### İhtiyacınız Olanlar:
- **Windows 10+**, **macOS 11+**, veya **Linux** (Ubuntu 20.04+, Debian 11+, Fedora 35+) bilgisayar
- **İnternet bağlantısı**
- **10 dakika** zaman

### Adım Adım:

#### Adım 1: GoConnect'i İndir

**İndirmek nedir?** İndirmek, GoConnect uygulama dosyasını internetten bilgisayarınıza alma işlemidir.

**Nasıl indirilir:**

1. Web tarayıcınızı açın (Chrome, Firefox, Edge, Safari vb.)
2. Şu adrese gidin: https://github.com/orhaniscoding/goconnect/releases/latest
3. İşletim sisteminizi bulun ve indirme linkine tıklayın

| İşletim Sisteminiz | İndirilecek Olan |
|-------------------|------------------|
| Windows | `GoConnect-Setup.exe` |
| Mac (Apple Silicon M1/M2/M3) | `GoConnect-aarch64.dmg` |
| Mac (Intel) | `GoConnect-x64.dmg` |
| Linux (Ubuntu/Debian) | `GoConnect.deb` |
| Linux (Herhangi) | `GoConnect.AppImage` |

**Hangisini indireceğinizi bilmiyor musunuz?** İşte nasıl kontrol edeceksiniz:

**Windows:** `Win + R` tuşuna basın, `winver` yazın, Enter'a basın. "Windows 10" veya "Windows 11" göreceksiniz

**Mac:** Apple menüsüne tıklayın → Bu Mac Hakkında. "İşlemci" veya "Çip" kısmına bakın:
- Eğer "Intel" diyorsa → `GoConnect-x64.dmg` indirin
- Eğer "M1", "M2", "M3" diyorsa → `GoConnect-aarch64.dmg` indirin

**Linux:** Terminal açın ve şu komutu çalıştırın: `uname -m`
- Çıktı `x86_64` ise → Sisteminiz 64-bit (en yaygın)
- Çıktı `aarch64` veya `arm64` ise → Sisteminiz ARM tabanlı

#### Adım 2: GoConnect'i Kur

**Kurulum nedir?** Kurulum, bilgisayarınızın uygulamayı çalıştırabilmesi için hazırlanmasıdır.

**Nasıl kurulur:**

**Windows:**
1. İndirilen `GoConnect-Setup.exe` dosyasını bulun (genellikle İndirilenler klasöründe)
2. Çift tıklayın
3. Windows "Bu uygulamaya izin vermek istiyor musunuz?" diye sorarsa → **Evet**'e tıklayın
4. Kurulum sihirbazında **İleri**'ye tıklayın
5. Bittiğinde **Bitir**'e tıklayın

**macOS:**
1. İndirilen `.dmg` dosyasını İndirilenler'de bulun
2. Çift tıklayın (GoConnect ikonu olan bir pencere açılır)
3. GoConnect ikonunu Uygulamalar klasörüne sürükleyin
4. Pencereyi kapatın
5. Diski çıkarın (Çöpe sürükleyin)

**Linux (Debian/Ubuntu .deb ile):**
1. Terminal açın (`Ctrl + Alt + T`)
2. Şunu yazın: `sudo dpkg -i İndirilenler/GoConnect.deb`
3. Şifrenizi girin (yazarken göremezsiniz, bu normal)
4. Enter'a basın

**Linux (AppImage - Her distro):**
1. İndirilen `GoConnect.AppImage` dosyasına sağ tıklayın
2. Özellikler → İzinler → "Dosyayı program olarak çalıştırmaya izin ver" işaretleyin
3. Kapatın ve AppImage'e çift tıklayın

#### Adım 3: GoConnect'i Aç

**Açınca ne olur?** GoConnect bilgisayarınızda çalışmaya başlar.

**Nasıl açılır:**

**Windows:** Start tuşuna basın, "GoConnect" yazın, Enter'a basın

**macOS:** Uygulamalar klasörünü açın, GoConnect'e çift tıklayın

**Linux:** `Alt + F2` tuşlarına basın, `goconnect` yazın, Enter'a basın
Veya terminalden: `goconnect`

**Ne göreceksiniz:**
```
┌─────────────────────────────────────┐
│         GoConnect                   │
├─────────────────────────────────────┤
│                                     │
│  Ne yapmak istersiniz?              │
│                                     │
│  ⦿ Ağ Oluştur                      │
│    Ağa Katıl                       │
│    Ayarlar                         │
│    Çıkış                           │
│                                     │
└─────────────────────────────────────┘
```

#### Adım 4: İlk Ağınızı Oluşturun

**Ağ oluşturmak nedir?** Sen ve arkadaşlarının katılabileceği sanal bir özel LAN oluşturuyorsun.

**Nasıl oluşturulur:**

1. **"Ağ Oluştur"**'a tıklayın (veya Enter'a basın)
2. Ağınıza bir isim verin (örn: "Oyun Ağım")
3. Enter'a basın
4. GoConnect ağ kurarken birkaç saniye bekleyin

**Şimdi ne oldu:**
- GoConnect bilgisayarınızda sanal bir ağ arayüzü oluşturdu
- Sanal bir IP adresi aldınız (örn: `10.0.1.1`)
- GoConnect bir davet linki oluşturdu

**Ne göreceksiniz:**
```
✓ Ağ başarıyla oluşturuldu!

Ağ Adı: Oyun Ağım
IP Adresiniz: 10.0.1.1

Davet Linki:
gc://invite.goconnect.io/abc123def456

Arkadaşlarınızın katılması için bu linki paylaşın!
```

#### Adım 5: Arkadaşlarınızı Davet Edin

**Davet etmek nedir?** Başkalarının ağana katılmasını sağlayan özel link paylaşmak.

**Nasıl davet edersiniz:**

1. Davet linkini kopyalayın (tıklayın ve `Ctrl + C` veya `Cmd + C` tuşlarına basın)
2. Arkadaşlarınıza şunlarla gönderin:
   - E-posta
   - Discord
   - WhatsApp
   - Herhangi bir mesajlaşma uygulaması

**Tıkladıklarında ne olur:**
- GoConnect'leri otomatik açılır
- "Katıl"a tıklarlar
- Artık sanal LAN'inizedeler!

#### Adım 6: Bağlandınız!

**Şimdi ne yapabilirsiniz?**

**LAN oyunları oyna:**
1. Oyununu başlat (örn: Minecraft)
2. "LAN Oyunu" veya "Çok Oyunculu → LAN" seç
3. Arkadaşların sunucu listende oyununu görür
4. Katılırlar ve oynarsınız!

**Dosya paylaş:**
1. GoConnect sohbetini aç
2. Eklenti ikonuna tıkla
3. Bir dosya seç
4. Ağındaki herhangi birine gönder

**Sesli sohbet:**
1. GoConnect'te mikrofon ikonuna tıkla
2. Arkadaşlarınla konuş

---

## 📖 Detaylı Kurulum Rehberleri

Yardıma mı ihtiyacın var? Her platform için detaylı rehberlerimiz var:

| Platform | Rehber |
|----------|-------|
| **Windows** | [Windows Kurulum Rehberi](docs/tr/installations/windows.md) |
| **macOS** | [macOS Kurulum Rehberi](docs/tr/installations/macos.md) |
| **Linux** | [Linux Kurulum Rehberi](docs/tr/installations/linux.md) |
| **Docker** | [Docker Kurulum Rehberi](docs/tr/installations/docker.md) |

Her rehber şunları içerir:
- ✅ Sistem gereksinimleri
- ✅ Adım adım talimatlar (ekran görüntüleriyle)
- ✅ Yaygın sorunların giderilmesi
- ✅ Gelişmiş yapılandırma

---

## 🎓 Kullanım Rehberleri

Özel özellikleri nasıl kullanacağınızı öğrenin:

| Rehber | Açıklama |
|-------|---------|
| [Ağ Oluşturma](docs/tr/guides/creating-network.md) | Kendi sanal LAN'ını host et |
| [Ağa Katılma](docs/tr/guides/joining-network.md) | Mevcut ağlara bağlan |
| [Metin Sohbeti](docs/tr/guides/chat.md) | Yerleşik mesajlaşmayı kullan |
| [Sesli Sohbet](docs/tr/guides/voice-chat.md) | Gerçek zamanlı sesli iletişim |
| [Dosya Transferi](docs/tr/guides/file-transfer.md) | Peer-to-peer dosya paylaşımı |
| [Üye Yönetimi](docs/tr/guides/member-management.md) | Ağ üyelerini yönet |

---

## 🏠 Self-Hosting

Kendi GoConnect sunucunu çalıştırmak ister misin?

**Neden self-host?**
- 🔒 **Gizlilik** - Verileriniz sunucunuzda kalır
- 🎛️ **Kontrol** - Her şeye siz karar verirsiniz
- 🚀 **Performans** - Üçüncü taraf bağımlılığı yok
- 💰 **Maliyet** - Uzun vadede daha ucuz olabilir

**Docker ile Hızlı Başlangıç:**

```bash
# docker-compose dosyasını indir
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# ortam dosyası oluştur
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 16)
WG_SERVER_ENDPOINT=domainadiniz.com:51820
EOF

# sunucuyu başlat
docker compose up -d
```

**📖 Tam Rehber:** [Self-Hosting Rehberi](docs/tr/self-hosting/overview.md)

Rehber şunları kapsar:
- Docker kurulumu (önerilen)
- Manuel binary kurulum
- Yapılandırma seçenekleri
- Reverse proxy kurulumu (Nginx/Caddy)
- Güvenlik checklist'i
- İzleme ve sorun giderme

---

## 🛠️ Geliştirme

Katkıda bulunmak veya kaynak koddan derlemek ister misin?

### Gereksinimler

| Araç | Sürüm | Neden? |
|------|-------|-------|
| **Go** | 1.24+ | Ana dil (cli ve core modülleri) |
| **Node.js** | 20+ | Masaüstü uygulaması frontend'i |
| **Rust** | Son | Masaüstü uygulaması backend'i (Tauri) |
| **protoc** | Son | Protocol Buffers derleyicisi |

### Hızlı Başlangıç

```bash
# Repoyu klonla
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# CLI derle
cd cli
go build -o goconnect ./cmd/goconnect
./goconnect

# Sunucu derle
cd ../core
go build -o goconnect-server ./cmd/server
./goconnect-server

# Masaüstü Uygulaması derle
cd ../desktop
npm install
npm run tauri build
```

**📖 Tam Rehber:** [Geliştirme Rehberi](docs/tr/development/introduction.md)

---

## ❓ Sıkça Sorulan Sorular

<details>
<summary><b>GoConnect ücretsiz mi?</b></summary>

Evet! GoConnect tamamen ücretsiz ve açık kaynaktır (MIT Lisansı). Temel özellikler her zaman ücretsiz kalacaktır.
</details>

<details>
<summary><b>Hangi platformlar destekleniyor?</b></summary>

✅ Windows 10/11
✅ macOS 11+ (Intel ve Apple Silicon)
✅ Linux (Ubuntu 20.04+, Debian 11+, Fedora 35+)
🔜 Mobil uygulamalar yakında
</details>

<details>
<summary><b>VPN'den farkı nedir?</b></summary>

**VPN (Sanal Özel Ağ):**
- Tüm trafiğin bir sunucudan geçmesini sağlar
- Sunucu yaptığın her şeyi görür
- Sunucu darboğazı yüzünden daha yavaş
- Gizlilik için iyi, hız için değil çok iyi değil

**GoConnect:**
- Cihazlar arasında doğrudan bağlantı oluşturur
- Sadece GoConnect trafiği ağdan geçer
- Peer-to-peer olduğu için daha hızlı
- Oyun, dosya paylaşımı ve LAN uygulamaları için mükemmel

**Şöyle düşünün:**
- VPN = Tüm internet trafiğin bir tünelden geçer
- GoConnect = Sadece belirli uygulamalar/oyunlar tünelden geçer, her şey başka normal interneti kullanır
</details>

<details>
<summary><b>Güvenli mi?</b></summary>

Evet! GoConnect endüstri standardı güvenlik kullanır:

**Şifreleme:**
- WireGuard protokolü (ordular ve şirketler tarafından kullanılır)
- ChaCha20 şifreleme (HTTPS'de kullanılanla aynı algoritma)
- Mükemmel İleri Gizlilik (Biri trafiğini kaydetse bile, daha sonra çöncemez)

**Kimlik Doğrulama:**
- Her cihazın benzersiz kriptografik anahtarları vardır
- Tahmin edilebilecek şifre yoktur
- Davet linkleri kriptografik olarak imzalanmıştır

**Gizlilik:**
- Merkezi sunucu verilerinizi görmez (peer-to-peer)
- Tam kontrol için self-host edebilirsiniz
- Herkesin denetleyebileceği açık kaynak kod

**Ama unutmayın:** Her araç gibi, güvenlik düzgün kullanıma bağlıdır. Her zaman:
- Sadece güvendiğiniz kişilerden ağlara katılın
- GoConnect'i güncel tutun
- Self-host sunucunuzda güçlü şifreler kullanın
</details>

<details>
<summary><b>Port yönlendirme gerekir mi?</b></summary>

Genellikle **hayır!** GoConnect port yönlendirme olmadan bağlanmak için gelişmiş teknikler kullanır:

**NAT Geçişi:**
- UDP delme
- STUN sunucuları (genel IP'nizi bulmaya yardımcı olur)
- TURN röle (doğrudan bağlantı başarısız olduğunda yedek)

**Ne zaman port yönlendirme gerekebilir?**
- Çok kısıtlayıcı bir güvenlik duvarının arkadaysanız
- Her iki peer da simetrik NAT'e sahipse (nadir)
- Self-host sunucular için

**Gerekip gerekmediğini nasıl kontrol edersiniz:**
Sadece bağlanmayı deneyin! Çalışmazsa, GoConnect size söyleyecektir.

**Port yönlendirme nasıl yapılır (gerekirse):**
[Ağ Yapılandırma Rehberi](docs/tr/guides/network-config.md)'ne bakın
</details>

<details>
<summary><b>Kaç cihaz bağlanabilir?</b></summary>

**Teorik limit:** Ağ başına 65.534 cihaz (/16 subnet)

**Pratik limit:** Donanımınıza ve internet bağlantınıza bağlı

**Gerçekçi tahminler:**
- Oyun: 10-50 oyuncu (oyuna bağlı)
- Dosya paylaşımı: 100+ kullanıcı
- Sohbet: 1000+ kullanıcı

**Daha büyük dağıtımlar için:** Birden fazla ağ çalıştırmayı veya kurumsal sürümümüzü (yakında) düşünün.
</details>

<details>
<summary><b>NAT/CGNAT ile çalışır mı?</b></summary>

**Evet, genellikle!** GoConnect şunlarla çalışacak şekilde tasarlanmıştır:

✅ NAT (Ağ Adresi Çevirisi)
✅ CGNAT (Taşıyıcı Sınıfı NAT)
✅ Güvenlik Duvarı
✅ Simetrik NAT (daha zor, ama deniyoruz)

**Nasıl çalışır:**
1. Önce doğrudan bağlantıyı deneriz
2. Başarısız olursa, genel IP'yi bulmak için STUN kullanırız
3. Başarısız olursa, TURN rölesini kullanırız

**Başarı oranı:** ~%95 bağlantı herhangi bir yapılandırma olmadan başarılı olur
</details>

**Daha fazla soru?** [Tam SSS'ye bakın](docs/tr/faq.md)

---

## 🤝 Katkıda Bulunma

Katkılara hoş geldiniz!

### Nasıl Katkıda Bulunursunuz

1. **Hata Bildir:** [Issue açın](https://github.com/orhaniscoding/goconnect/issues/new?template=bug_report.md)
2. **Özellik Öner:** [Tartışma başlatın](https://github.com/orhaniscoding/goconnect/discussions)
3. **Kod Gönder:** Fork → Branch → PR

### Geliştirme Yönergeleri

- [Conventional Commits](https://www.conventionalcommits.org/) takip edin
- Yeni özellikler için test yazın
- Dokümantasyonu güncelleyin
- Basit tutun (bkz. [Sıfır Bağımlılık Politikası](docs/development/zero-dependency.md))

**📖 Tam Rehber:** [Katkıda Bulunma Rehberi](CONTRIBUTING.md)

---

## 📄 Lisans

MIT Lisansı - Ayrıntılar için [LICENSE](LICENSE) dosyasına bakın.

**Bu ne anlama gelir:**
- ✅ Ücretsiz kullan
- ✅ İstediğin gibi değiştir
- ✅ Dağıt (ticari olarak bile)
- ✅ Bu lisans notunu koru

**Kısaca:** İstediğini yap, bozulursa bizi dava etme.

---

## 🙏 Teşekkürler

Harika açık kaynak araçlarla inşa edildi:

- [WireGuard](https://www.wireguard.com/) - Modern VPN protokolü
- [Tauri](https://tauri.app/) - Masaüstü uygulama çerçevesi
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI
- [Go](https://go.dev/) - Programlama dili
- [React](https://react.dev/) - Frontend kütüphanesi

---

## 📞 İletişim ve Destek

**Yardım Al:**
- 📖 [Dokümantasyon](docs/)
- 💬 [Tartışmalar](https://github.com/orhaniscoding/goconnect/discussions)
- 🐛 [Hata Bildirleri](https://github.com/orhaniscoding/goconnect/issues)
- ✉️ E-posta: support@goconnect.io (yakında)

**Bizi Takip Edin:**
- GitHub: [@orhaniscoding](https://github.com/orhaniscoding)
- Web sitesi: https://goconnect.io (yakında)

---

## 🗺️ Yol Haritası

### v1.2.0 (Devam Ediyor)
- [ ] Mobil uygulamalar (iOS/Android)
- [ ] Görüntülü sohbet
- [ ] Performans iyileştirmeleri

### v1.3.0 (Planlandı)
- [ ] Oyun entegrasyon eklentileri
- [ ] Gelişmiş ağ yönetimi
- [ ] Web istemcisi

### v2.0.0 (Gelecek)
- [ ] Kurumsal özellikler
- [ ] Özel ağ topolojileri
- [ ] Üçüncü taraf uygulamalar için API

**Özellik önerin:** [Özellik isteği açın](https://github.com/orhaniscoding/goconnect/issues/new?template=feature_request.md)

---

<div align="center">

**⭐ GitHub'da bize yıldız verin!**

[orhaniscoding](https://github.com/orhaniscoding) tarafından ❤️ ile yapıldı

[⬆ Başa Dön](#-goconnect)

</div>
