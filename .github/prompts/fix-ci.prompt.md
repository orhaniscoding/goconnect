---
mode: agent
---
**Görev:** CI kırmızı; önce kök sebep. Çıktı: PLAN → PATCHES → TESTS → COMMIT
- Log’u analiz et; ilk kırılan satırı ata.
- web-ui: typecheck/build hataları, lockfile uyumsuzluğu
- server/daemon: test -race, coverage gerilemesi
- Yama minimum ve atomik olsun; CI tekrar koştur.

