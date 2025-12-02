# CLI HTTP Client Implementation Notes

**Tarih:** 2025-01-22  
**Durum:** ✅ Dokümante Edildi

---

## 📝 Notlar

### Daemon-Specific Operations (gRPC Only)

Aşağıdaki metodlar daemon'a özel operasyonlar olduğu için sadece gRPC üzerinden çalışır:

1. **LeaveNetwork** - Daemon'un network state'ini yönetir
2. **GetPeers** - Daemon'un peer listesini döndürür
3. **SendChatMessage** - Daemon üzerinden chat mesajı gönderir
4. **SendFile** - Daemon üzerinden dosya transferi başlatır
5. **GetSettings** - Daemon ayarlarını getirir
6. **UpdateSettings** - Daemon ayarlarını günceller

### Neden HTTP API'de Yok?

Bu metodlar daemon'un kendi internal state'ini yönetir ve daemon'un HTTP API'sinde expose edilmemiştir. Bunun yerine:

- **gRPC IPC** kullanılır (Unix socket / Named pipe)
- Daemon'un kendi HTTP API'si (`http://localhost:12345`) sadece basit health check ve status için kullanılır
- Network operations için server HTTP API'si (`api.Client`) kullanılır

### Mevcut Durum

- ✅ `unified_client.go` metodları gRPC-only olarak işaretlendi
- ✅ Hata mesajları açıklayıcı hale getirildi
- ✅ Yorumlar eklendi

### Gelecek İyileştirmeler

Eğer bu metodların HTTP üzerinden de çalışması istenirse:

1. Daemon'un HTTP API'sine bu endpoint'ler eklenebilir
2. Veya `api.Client` kullanılarak server HTTP API'si üzerinden implement edilebilir

---

**Son Güncelleme:** 2025-01-22

