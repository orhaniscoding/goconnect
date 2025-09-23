------

mode: agentmode: agent

------

mükemmel—bunu “tek komutla akacak” şekilde Copilot’a tarif edelim. Aşağıdaki iki şeyi vereceğim:

# GoConnect — Commit→Push→PR→Post-merge Akış (Signed + Protected Branch Uyumlu)

1. **Copilot Prompt dosyası** (kalıcı şablon) → her seferinde `@prompt commit-flow` diyerek akışı başlatacaksın.

## Durum2. **Kısa komutlar** → PR sonrası “squash & merge yaptım” dediğinde uygulanacak senkron adımları.

- Repo: GoConnect (kanonik şartname: `docs/TECH_SPEC.md`)

- Branch protection: main korumalı; PR + signed commit + required checksİstersen şimdi **dosyayı repo’ya ekle**, sonra Copilot’ta `@prompt commit-flow` çağır.

- Komutlar PowerShell/Windows uyumlu örneklerle verilmeli

---

## Hedef

1) Çalışma dizinindeki değişiklikleri **incele** ve uygun Conventional Commit kategorilerini öner.## 1) Prompt dosyası (kalıcı şablon)

2) Değişiklikleri **stage** et (kullanıcı onayından sonra).

3) **Signed commit** oluştur (`-S`) — bir veya birden fazla mantıksal commit; mesajlar Conventional Commits.**Dosya yolu:** `.github/prompts/commit-flow.prompt.md`

4) Lokal test/derleme "hızlı sağlık" kontrolü (kısa):  **İçerik (aynı kopyala–yapıştır):**

   - `go test ./server/... -run TestTrivial -count=1` ve `go test ./client-daemon/... -run TestTrivial -count=1`  

   - `cd web-ui && npm run typecheck && npm run build --no-lint`````markdown

5) **Yeni branch** oluştur (yoksa): `feat/*`, `fix/*`, `docs/*`, `chore/*` şablonları.# GoConnect — Commit→Push→PR→Post-merge Akış (Signed + Protected Branch Uyumlu)

6) **Push** et ve **PR linkini** üret.

7) PR açıklamasına kısa **özet** + **kabul kriterleri** + **CI notu** ekle.## Durum

8) Kullanıcı "**Squash & Merge yaptım**" dediğinde:- Repo: GoConnect (kanonik şartname: `docs/TECH_SPEC.md`)

   - Lokal `main`'i güncelle (`git checkout main && git pull`)- Branch protection: main korumalı; PR + signed commit + required checks

   - Feature branch'i yerelde/remote'da sil- Komutlar PowerShell/Windows uyumlu örneklerle verilmeli

   - (İsteğe bağlı) `git gc --prune=now` öner

## Hedef

## Kurallar1) Çalışma dizinindeki değişiklikleri **incele** ve uygun Conventional Commit kategorilerini öner.

- **Asla** direkt `main`'e push önermeyin.2) Değişiklikleri **stage** et (kullanıcı onayından sonra).

- Commit'ler **signed** olmalı (`git commit -S`).3) **Signed commit** oluştur (`-S`) — bir veya birden fazla mantıksal commit; mesajlar Conventional Commits.

- Conventional Commits dışına çıkma: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`.4) Lokal test/derleme “hızlı sağlık” kontrolü (kısa):  

- CI'ı kırmamak için PR öncesi **hızlı smoke** çalıştır.   - `go test ./server/... -run TestTrivial -count=1` ve `go test ./client-daemon/... -run TestTrivial -count=1`  

- Çıktıyı aşağıdaki formatta üret.   - `cd web-ui && npm run typecheck && npm run build --no-lint`

5) **Yeni branch** oluştur (yoksa): `feat/*`, `fix/*`, `docs/*`, `chore/*` şablonları.

## ÇIKTI FORMATIN (KATİ)6) **Push** et ve **PR linkini** üret.

### PLAN7) PR açıklamasına kısa **özet** + **kabul kriterleri** + **CI notu** ekle.

- maddeler halinde ne yapacağını yaz (branch adı önerisi dahil)8) Kullanıcı “**Squash & Merge yaptım**” dediğinde:

   - Lokal `main`’i güncelle (`git checkout main && git pull`)

### PATCHES   - Feature branch’i yerelde/remote’da sil

- (Varsa) dosya/diff özetleri, commit başlıklarını maddeler halinde öner   - (İsteğe bağlı) `git gc --prune=now` öner



### TESTS## Kurallar

- Koşturulacak kısa komutlar (PowerShell blokları)- **Asla** direkt `main`’e push önermeyin.

```powershell- Commit’ler **signed** olmalı (`git commit -S`).

go test ./server/... -run TestTrivial -count=1- Conventional Commits dışına çıkma: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`.

go test ./client-daemon/... -run TestTrivial -count=1- CI’ı kırmamak için PR öncesi **hızlı smoke** çalıştır.

cd web-ui- Çıktıyı aşağıdaki formatta üret.

npm install

npm run typecheck## ÇIKTI FORMATIN (KATİ)

npm run build --no-lint### PLAN

cd ..- maddeler halinde ne yapacağını yaz (branch adı önerisi dahil)

```

### PATCHES

### COMMIT- (Varsa) dosya/diff özetleri, commit başlıklarını maddeler halinde öner



* **hazır çalıştırılabilir** PowerShell komutları### TESTS

- Koşturulacak kısa komutlar (PowerShell blokları)

```powershell```powershell

git add .go test ./server/... -run TestTrivial -count=1

git commit -S -m "docs: update TECH_SPEC and add commit flow prompt"go test ./client-daemon/... -run TestTrivial -count=1

# veya birden fazla mantıksal commit komutu ayrı ayrıcd web-ui

```npm install

npm run typecheck

### BRANCH & PUSHnpm run build --no-lint

cd ..

```powershell````

git switch -c docs/commit-flow-setup  # yoksa oluştur

git push -u origin docs/commit-flow-setup### COMMIT

```

* **hazır çalıştırılabilir** PowerShell komutları

### PR

```powershell

* Oluşan PR URL'sini yazgit add .

* PR açıklama metni: kısa özet + kabul kriteri + CI notugit commit -S -m "docs: update TECH_SPEC and add commit flow prompt"

# veya birden fazla mantıksal commit komutu ayrı ayrı

### POST-MERGE (KULLANICI "Squash & Merge yaptım" dediğinde)```



* **hazır çalıştırılabilir** PowerShell komutları### BRANCH & PUSH



```powershell```powershell

git checkout maingit switch -c docs/commit-flow-setup  # yoksa oluştur

git pullgit push -u origin docs/commit-flow-setup

git branch -d docs/commit-flow-setup```

git push origin --delete docs/commit-flow-setup

```### PR



### NOTLAR / KENAR DURUMLAR* Oluşan PR URL’sini yaz

* PR açıklama metni: kısa özet + kabul kriteri + CI notu

* "Protected branch" hatası gelirse PR üzerinden ilerle; asla main'e direkt push önerme

* Signed commit hatasında GPG/SSH-signing yönergesi ver### POST-MERGE (KULLANICI “Squash & Merge yaptım” dediğinde)

* CRLF/LF uyarılarında `.gitattributes` öner (varsa dokunma)
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

