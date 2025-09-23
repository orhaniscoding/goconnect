---
mode: agent
---
mükemmel—bunu “tek komutla akacak” şekilde Copilot’a tarif edelim. Aşağıdaki iki şeyi vereceğim:

1. **Copilot Prompt dosyası** (kalıcı şablon) → her seferinde `@prompt commit-flow` diyerek akışı başlatacaksın.
2. **Kısa komutlar** → PR sonrası “squash & merge yaptım” dediğinde uygulanacak senkron adımları.

İstersen şimdi **dosyayı repo’ya ekle**, sonra Copilot’ta `@prompt commit-flow` çağır.

---

## 1) Prompt dosyası (kalıcı şablon)

**Dosya yolu:** `.github/prompts/commit-flow.prompt.md`
**İçerik (aynı kopyala–yapıştır):**

````markdown
# GoConnect — Commit→Push→PR→Post-merge Akış (Signed + Protected Branch Uyumlu)

## Durum
- Repo: GoConnect (kanonik şartname: `docs/TECH_SPEC.md`)
- Branch protection: main korumalı; PR + signed commit + required checks
- Komutlar PowerShell/Windows uyumlu örneklerle verilmeli

## Hedef
1) Çalışma dizinindeki değişiklikleri **incele** ve uygun Conventional Commit kategorilerini öner.
2) Değişiklikleri **stage** et (kullanıcı onayından sonra).
3) **Signed commit** oluştur (`-S`) — bir veya birden fazla mantıksal commit; mesajlar Conventional Commits.
4) Lokal test/derleme “hızlı sağlık” kontrolü (kısa):  
   - `go test ./server/... -run TestTrivial -count=1` ve `go test ./client-daemon/... -run TestTrivial -count=1`  
   - `cd web-ui && npm run typecheck && npm run build --no-lint`
5) **Yeni branch** oluştur (yoksa): `feat/*`, `fix/*`, `docs/*`, `chore/*` şablonları.
6) **Push** et ve **PR linkini** üret.
7) PR açıklamasına kısa **özet** + **kabul kriterleri** + **CI notu** ekle.
8) Kullanıcı “**Squash & Merge yaptım**” dediğinde:
   - Lokal `main`’i güncelle (`git checkout main && git pull`)
   - Feature branch’i yerelde/remote’da sil
   - (İsteğe bağlı) `git gc --prune=now` öner

## Kurallar
- **Asla** direkt `main`’e push önermeyin.
- Commit’ler **signed** olmalı (`git commit -S`).
- Conventional Commits dışına çıkma: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`.
- CI’ı kırmamak için PR öncesi **hızlı smoke** çalıştır.
- Çıktıyı aşağıdaki formatta üret.

## ÇIKTI FORMATIN (KATİ)
### PLAN
- maddeler halinde ne yapacağını yaz (branch adı önerisi dahil)

### PATCHES
- (Varsa) dosya/diff özetleri, commit başlıklarını maddeler halinde öner

### TESTS
- Koşturulacak kısa komutlar (PowerShell blokları)
```powershell
go test ./server/... -run TestTrivial -count=1
go test ./client-daemon/... -run TestTrivial -count=1
cd web-ui
npm install
npm run typecheck
npm run build --no-lint
cd ..
````

### COMMIT

* **hazır çalıştırılabilir** PowerShell komutları

```powershell
git add .
git commit -S -m "docs: update TECH_SPEC and add commit flow prompt"
# veya birden fazla mantıksal commit komutu ayrı ayrı
```

### BRANCH & PUSH

```powershell
git switch -c docs/commit-flow-setup  # yoksa oluştur
git push -u origin docs/commit-flow-setup
```

### PR

* Oluşan PR URL’sini yaz
* PR açıklama metni: kısa özet + kabul kriteri + CI notu

### POST-MERGE (KULLANICI “Squash & Merge yaptım” dediğinde)

* **hazır çalıştırılabilir** PowerShell komutları

```powershell
git checkout main
git pull
git branch -d docs/commit-flow-setup
git push origin --delete docs/commit-flow-setup
```

### NOTLAR / KENAR DURUMLAR

* “Protected branch” hatası gelirse PR üzerinden ilerle; asla main’e direkt push önerme
* Signed commit hatasında GPG/SSH-signing yönergesi ver
* CRLF/LF uyarılarında `.gitattributes` öner (varsa dokunma)

````

**Eklemek için (aynı branch’te):**
```powershell
mkdir -Force .github\prompts
notepad ".github\prompts\commit-flow.prompt.md"  # yukarıdaki metni yapıştır/kaydet
git add .github\prompts\commit-flow.prompt.md
git commit -S -m "docs: add Copilot commit-flow prompt (signed commit + PR + post-merge)"
git push
````

> Artık Copilot Chat’te **`@prompt commit-flow`** diyerek bu akışı çalıştırabilirsin.

---

## 2) “Squash & Merge yaptım” sonrası **tek komut seti**

PR’ı squash & merge ettikten sonra, **her seferinde** şu komutları çalıştır:

```powershell
# main’i güncelle
git checkout main
git pull

# feature branch’i yerelde sil (adı örnek)
git branch -d docs/commit-flow-setup

# remote branch’i sil
git push origin --delete docs/commit-flow-setup

# (opsiyonel) temizlik
git gc --prune=now
```

> Copilot’a “Squash & Merge yaptım” dediğinde, bu blokları **kendisi** de önerecek (prompt böyle tarif ediyoruz).

---

## 3) İsteğe bağlı: commit lint (otomatik kontrol)

İleride “commit mesajı hatalı” PR’larını engellemek istersen:

* `commitlint` + `husky` kurup `commit-msg` hook’u ekleyebiliriz.
* GitHub Actions’ta `commitlint` adımı koşup uygunsuz mesajları kırmızı yapabiliriz.

> İstersen bu kurulum için de hazır patch dosyası gönderirim.

---

