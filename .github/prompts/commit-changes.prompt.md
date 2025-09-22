---
mode: agent
---
**Görev:** Stage edilmiş değişiklikler için Conventional Commit + signed commit öner.  
- Commit mesajı → tek satır summary + opsiyonel body.  
- Category: feat, fix, docs, chore, refactor, test.  
- Mesajda JIRA/Ticket yok → sadece repo context.  
- Çıktı formatı: tek satır `git commit -S -m "..."`.
- Sonra: `git push -u origin <branch>` komutu ekle.
