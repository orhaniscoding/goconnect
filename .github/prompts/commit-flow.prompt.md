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
```prompt
------
mode: agent
------

# GoConnect — Commit→Push→PR→Post-merge Akış (Signed + Protected Branch Uyumlu)

Bu şablon, Copilot Chat’te `@prompt commit-flow` diyerek başlatabileceğin uçtan uca akışı tanımlar. Komut örnekleri PowerShell/Windows uyumludur. Repo kanonik şartnamesi `docs/TECH_SPEC.md`’dir. Main protected; PR, signed commits ve required checks zorunludur.

Bu şablon, Copilot Chat’te `@prompt commit-flow` diyerek başlatabileceğin uçtan uca akışı tanımlar. Komut örnekleri PowerShell/Windows uyumludur. Repo kanonik şartnamesi `docs/TECH_SPEC.md`’dir. Main protected; PR, signed commits ve required checks zorunludur.

## Hedefler

1) Çalışma dizinindeki değişiklikleri incele ve uygun Conventional Commit kategorilerini öner.
2) Değişiklikleri kullanıcı onayıyla stage et.
3) Signed commit(ler) oluştur (`git commit -S`) — mesajlar Conventional Commits’e uygun.
4) Hızlı smoke: minimal test/derleme doğrulaması.
5) Yoksa uygun isimli yeni branch oluştur: `feat/*`, `fix/*`, `docs/*`, `chore/*`.
6) Push ve PR linki üret.
7) PR açıklamasına kısa özet + kabul kriterleri + CI notu ekle.
8) Kullanıcı “Squash & Merge yaptım” dediğinde yerel/remote temizlik ve senkronizasyon adımlarını uygula.

## Kurallar

- Asla direkt `main`’e push önerme; protected branch üzerinden PR ile ilerle.
- Commit’ler signed olmalı (`git commit -S`).
- Conventional Commits tipleri: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`.
- PR öncesi hızlı smoke mutlaka koşulmalı; CI’ı kırmamak esastır.

## Çıktı Formatın (KATI)

Copilot yanıtını aşağıdaki başlıklarla üret:

### PLAN
- Yapılacaklar maddeler halinde (branch adı önerisi dahil).

### PATCHES
- (Varsa) dosya/diff özetleri ve önerilen commit başlıkları.

### TESTS
- Koşturulacak kısa komutlar (PowerShell blokları):

```powershell
go test ./server/... -run TestTrivial -count=1
go test ./client-daemon/... -run TestTrivial -count=1
cd web-ui
npm install
npm run typecheck
npm run build --no-lint
cd ..
```

### COMMIT
- Hazır çalıştırılabilir PowerShell komutları:

```powershell
git add .
git commit -S -m "<type>(<scope>): <summary>"
# Gerekirse mantıksal değişiklikler için birden fazla commit at
```

### BRANCH & PUSH

```powershell
git switch -c <recommended-branch-name>  # yoksa oluştur
git push -u origin <recommended-branch-name>
```

### PR
- Oluşan PR URL’sini yaz.
- PR açıklaması: kısa özet + kabul kriterleri + CI notu.

### POST-MERGE (Kullanıcı “Squash & Merge yaptım” dediğinde)

```powershell
git checkout main
git pull
git branch -d <recommended-branch-name>
git push origin --delete <recommended-branch-name>
# (opsiyonel)
git gc --prune=now
```

### Notlar / Kenar Durumlar

- “Protected branch” hatası: PR üzerinden ilerle; main’e direkt push yok.
- Signed commit hatası: GPG/SSH signing yönergesi ver.
- CRLF/LF uyarıları: `.gitattributes` öner (varsa dokunma).

---

## 1) Hızlı Smoke Ayrıntıları

Çekirdek doğrulamalar (hızlı):

```powershell
go test ./server/... -run TestTrivial -count=1
go test ./client-daemon/... -run TestTrivial -count=1
cd web-ui
npm run typecheck
npm run build --no-lint
cd ..
```

Gerekiyorsa `docs/API_EXAMPLES.http` ile 1–2 happy path istek.

---

## 2) “Squash & Merge yaptım” sonrası tek komut seti

PR’ı squash & merge ettikten sonra, her seferinde şu komutları çalıştır:

```powershell
git checkout main
git pull
git branch -d <recommended-branch-name>
git push origin --delete <recommended-branch-name>
git gc --prune=now
```

> Copilot’a “Squash & Merge yaptım” dediğinde, bu blokları kendisi de önerecek.

---

## 3) İsteğe bağlı: commit lint (otomatik kontrol)

İleride commit mesajlarını otomatik doğrulamak için:

- `commitlint` + `husky` ile `commit-msg` hook’u eklenebilir.
- GitHub Actions’ta `commitlint` adımı çalıştırılabilir.

> İstenirse hazır patch ve CI iş akışı eklenebilir.

`````

