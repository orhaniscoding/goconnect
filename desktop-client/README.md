# 🖥️ GoConnect Desktop

GoConnect'in masaüstü uygulaması. Tek bir uygulama ile hem ağ oluşturun (host) hem de başka ağlara katılın (client).

## ✨ Özellikler

- 🌐 **Ağ Oluştur** - Kendi sanal LAN'ını başlat
- 🔗 **Ağa Katıl** - Davet linki ile tek tıkla bağlan
- 💬 **Sohbet** - Discord benzeri metin kanalları
- 👥 **Üye Yönetimi** - Davet, çıkarma, yasaklama
- 🎨 **Modern UI** - Karanlık tema, kullanıcı dostu

## 🛠️ Teknolojiler

| Katman | Teknoloji |
|--------|-----------|
| Framework | Tauri 2.0 |
| Frontend | React 19 + TypeScript |
| Styling | Tailwind CSS |
| Backend | Rust |

## 📦 Geliştirme

### Gereksinimler

- Node.js 20+
- Rust (rustup ile)
- Platform bağımlılıkları:
  - **Windows:** WebView2 (genellikle yüklü)
  - **macOS:** Xcode Command Line Tools
  - **Linux:** `webkit2gtk`, `libappindicator`

### Kurulum

```bash
# Bağımlılıkları yükle
npm install

# Geliştirme modunda çalıştır
npm run tauri dev

# Production build
npm run tauri build
```

### Proje Yapısı

```
desktop-client/
├── src/                # React frontend
│   ├── App.tsx         # Ana uygulama
│   ├── main.tsx        # Giriş noktası
│   └── index.css       # Global stiller
├── src-tauri/          # Rust backend
│   ├── src/
│   │   └── main.rs     # Tauri uygulaması
│   ├── Cargo.toml      # Rust bağımlılıkları
│   └── tauri.conf.json # Tauri yapılandırma
├── package.json
├── tailwind.config.js
└── vite.config.ts
```

## 🎨 UI Yapısı

```
┌────────────────────────────────────────────────────────────┐
│  GoConnect                                        ─ □ ✕   │
├────┬──────────────┬────────────────────────────────────────┤
│ 🏠 │  Ağ Adı      │  Ana içerik alanı                     │
│────│              │                                        │
│ 🎮 │  AĞLAR       │  Bağlantı durumu, üyeler,             │
│ 💼 │  • Minecraft │  sohbet vb.                           │
│ 👥 │  • Work VPN  │                                        │
│    │              │                                        │
│ +  │  KANALLAR    │                                        │
│    │  # genel     │                                        │
│ 👤 │  # duyurular │                                        │
└────┴──────────────┴────────────────────────────────────────┘
```

## 📄 Lisans

MIT License - Detaylar için [LICENSE](../LICENSE) dosyasına bakın.
