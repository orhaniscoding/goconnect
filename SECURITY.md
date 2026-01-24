# 🛡️ Security Policy / Güvenlik Politikası

---

[English](#english) | [Türkçe](#türkçe)

---

## English

## 📋 Overview

**What is this document?**

This document explains how GoConnect handles security, how to report vulnerabilities, and what we do to keep you safe.

**Why does this matter?**

Security is everyone's responsibility. This document helps:
- **Users** understand how GoConnect protects their data
- **Developers** know how to report security issues responsibly
- **Researchers** learn about our vulnerability disclosure program

---

## 🔒 Security Principles

GoConnect follows these core security principles:

### 1. Zero Trust Architecture

**What does this mean?**

We never trust anything by default. Every connection, every request, every user is verified.

**Why is this important?**

If one component is compromised, the damage is limited.

**Examples:**
- ✅ Every API call requires authentication
- ✅ Every WebSocket connection is validated
- ✅ Every file upload is scanned
- ❌ No "trusted internal network" assumptions

### 2. Encryption Everywhere

**What we encrypt:**

| Data Type | Encryption Method | Why? |
|-----------|-------------------|------|
| **Network Traffic** | WireGuard (ChaCha20-Poly1305) | Peer-to-peer connections |
| **API Communication** | TLS 1.3 | Server-client communication |
| **Stored Passwords** | Argon2id | Prevents password theft |
| **JWT Tokens** | RS256 | Prevents token forgery |
| **Database** | Optional encryption at rest | Prevents data theft from server |

**Why so much encryption?**

If someone intercepts your traffic, steals your database, or compromises your server, they still can't read your data.

### 3. Least Privilege

**What this means:**

Every component has the minimum permissions needed to do its job.

**Examples:**
- CLI only needs network access → No file system access to other apps
- Desktop app only needs UI permissions → No system-level access
- Server only needs database access → No direct file system access

**Why?**

If the desktop app is hacked, the attacker can't access the CLI. If the server is hacked, the attacker can't access other services.

### 4. Defense in Depth

**What this means:**

Multiple layers of security. If one layer fails, others still protect you.

**Layers:**
1. **Encryption** - If traffic is intercepted, it's unreadable
2. **Authentication** - If encryption fails, attackers still can't impersonate users
3. **Authorization** - If authentication fails, attackers still can't access resources
4. **Rate Limiting** - If authorization fails, attackers still can't brute force
5. **Monitoring** - If all else fails, we detect and respond

---

## 🛡️ How GoConnect Protects You

### Network Security

#### WireGuard Encryption

**What is WireGuard?**

WireGuard is a modern VPN protocol used by militaries and corporations.

**How it works:**

```
Your Computer                          Friend's Computer
     │                                      │
     │  1. Exchange keys (Curve25519)       │
     │<------------------------------------->│
     │                                      │
     │  2. Derive session key               │
     │     (ChaCha20-Poly1305)              │
     │                                      │
     │  3. Encrypted traffic               │
     │<====================================>│
     │                                      │
     ✅ Even if intercepted, unreadable    │
```

**What algorithms are used?**

| Algorithm | Purpose | Key Size | Security Level |
|-----------|---------|----------|----------------|
| **ChaCha20** | Encryption | 256-bit | ~256-bit security |
| **Poly1305** | Authentication | 128-bit | Prevents tampering |
| **Curve25519** | Key Exchange | 256-bit | Ephemeral keys |
| **Blake2s** | Hashing | 256-bit | Fast, secure |

**Why these algorithms?**

- **Battle-tested**: Used in HTTPS, SSH, VPNs worldwide
- **Fast**: Minimal performance impact
- **Future-proof**: Quantum-resistant (somewhat)

**What this means for you:**

Even if someone records your GoConnect traffic, they cannot decrypt it. Even with a supercomputer, it would take billions of years.

### Authentication & Authorization

#### Password Security

**How we store passwords:**

We **never** store your actual password. Instead, we store a "hash" - a mathematical fingerprint.

**Process:**

```
Your Password: "mypassword123"
                    │
                    ▼
            Add Salt (random data)
                    │
                    ▼
          Argon2id Hash (100,000 iterations)
                    │
                    ▼
         Stored Hash: "$argon2id$v=19$m=4096,t=3,p=1$..."
```

**Why Argon2id?**

- **Memory-hard**: Requires lots of RAM to crack (expensive for attackers)
- **Slow**: Takes time to compute (slows down brute force)
- **Recommended**: OWASP, industry standard

**What this means:**

Even if someone steals our database, they cannot get your password. They would need billions of dollars of computing power to crack one password.

#### JWT Tokens

**What are JWTs?**

JSON Web Tokens - like digital ID cards that prove you're logged in.

**How they work:**

```
1. You log in → Server verifies password
2. Server creates JWT → Signs with private key
3. Server sends JWT → Your browser stores it
4. You send JWT with every request → Server verifies signature
5. If valid → Access granted
```

**Why are they secure?**

- **Signed**: Cannot be forged (without private key)
- **Stateless**: Server doesn't need to store sessions
- **Short-lived**: Expire quickly (reduces risk if stolen)
- **Refreshable**: Can get new tokens without password

**Token structure:**

```json
{
  "header": {
    "alg": "RS256",           // Signing algorithm
    "typ": "JWT"              // Token type
  },
  "payload": {
    "sub": "user123",         // User ID
    "exp": 1706457600,        // Expiration time
    "iat": 1706371200,        // Issued at
    "permissions": ["read", "write"]
  },
  "signature": "..."          // Cryptographic signature
}
```

### Data Protection

#### What We Collect

**Data we store:**

| Data | Purpose | Retention | Encryption |
|------|---------|-----------|------------|
| **Email** | Account identification | Forever | TLS in transit |
| **Password Hash** | Authentication | Forever | Argon2id |
| **Network Name** | Your networks | Forever | TLS in transit |
| **IP Address** | Network assignment | Until network deleted | WireGuard |
| **Chat Messages** | In-memory relay | Seconds (until delivered) | WireGuard |
| **File Transfers** | P2P routing | Seconds (until delivered) | WireGuard |

**Data we DON'T store:**

- ❌ Chat message history (deleted after delivery)
- ❌ File contents (P2P, never goes through server)
- ❌ Voice/video data (P2P WebRTC)
- ❌ Your private keys (stored locally on your device)

#### What We Share

**We never sell your data. Period.**

**Who we share with:**

| Third Party | What | Why |
|-------------|------|-----|
| **Nobody** | Nothing | N/A |

**Exception:**

- **Legal Requirements**: If required by law (court order), we comply
- **With Your Permission**: If you explicitly consent

---

## 🐛 Reporting Vulnerabilities

### How to Report

**❌ DO NOT:**
- Report vulnerabilities in public (GitHub Issues, Discord, etc.)
- Share proof-of-concept code publicly
- Attempt to exploit the vulnerability beyond what's necessary

**✅ DO:**
- Report vulnerabilities privately
- Provide enough information to reproduce
- Give us time to fix before disclosing

### Report a Vulnerability

**Email:** [security@goconnect.io](mailto:security@goconnect.io)

**What to include:**

1. **Description**: What is the vulnerability?
2. **Steps**: How do we reproduce it?
3. **Impact**: What can an attacker do?
4. **Proof**: Screenshots, logs, or PoC code (if safe)
5. **Version**: Which version of GoConnect?

**Example Report:**

```
Subject: SQL Injection in Network Search

Description:
The network search parameter is vulnerable to SQL injection.

Steps to Reproduce:
1. Create a network
2. Go to search
3. Enter: '; DROP TABLE networks; --
4. Click search

Impact:
An attacker can delete all networks in the database.

Version:
v1.2.0

Proof:
[Attachment: screenshot]

Suggested Fix:
Use prepared statements parameterized queries.
```

### Response Timeline

| Time | What Happens |
|------|--------------|
| **24 hours** | We acknowledge receipt |
| **3 days** | We investigate and ask questions |
| **7 days** | We provide fix estimate |
| **90 days** | We fix before public disclosure (if critical) |

### Coordinated Disclosure

**Process:**

1. **You report** → Send vulnerability details
2. **We confirm** → We verify within 24 hours
3. **We fix** → We develop and test patch
4. **We release** → We deploy update
5. **You disclose** → You can publicly discuss after fix is deployed

**Why wait?**

Disclosing before fix puts users at risk. Coordinated disclosure protects everyone.

### Reward

**What you get:**

- 🏆 **Hall of Fame**: Your name listed on our website
- 🎖️ **Security Researcher Badge**: Special badge on your GitHub profile
- 📢 **Public Recognition**: We credit you in release notes
- 🎁 **Swag**: GoConnect stickers/t-shirt (if available)

**Note:** We don't pay monetary bounties (yet). This may change in the future.

---

## 🚨 Severity Levels

We classify vulnerabilities by severity:

### 🔴 Critical (48 hours to fix)

**Examples:**
- Remote Code Execution (RCE)
- Full database access
- Ability to impersonate any user
- Complete system compromise

**Response:**
- Drop everything and fix immediately
- Release security update within 48 hours
- Notify all users to update

### 🟠 High (7 days to fix)

**Examples:**
- Access to other users' data (not all)
- Privilege escalation (normal → admin)
- Denial of Service (affects all users)
- Bypassing authentication

**Response:**
- Fix in next release
- Release security update within 7 days
- Notify affected users

### 🟡 Medium (30 days to fix)

**Examples:**
- Access to your own data you shouldn't see
- Minor privilege escalation
- DoS (affects some users)
- Information disclosure (non-sensitive)

**Response:**
- Fix in next planned release
- Include in release notes
- No urgent notification

### 🟢 Low (90 days to fix)

**Examples:**
- Minor information disclosure
- Missing security headers
- Weak password requirements (not exploitable)
- UI/UX security issues

**Response:**
- Fix when convenient
- Mention in release notes
- No special notification

---

## 🛡️ Best Practices for Users

### How to Stay Safe

#### 1. Keep Updated

**Why?**

Updates often contain security fixes.

**How?**

- **Desktop App**: Auto-updates (default on)
- **CLI**: Run `goconnect update` regularly
- **Self-Hosted**: Watch for releases

**Check your version:**

```bash
goconnect version
```

#### 2. Use Strong Passwords

**What is a strong password?**

- ✅ At least 12 characters
- ✅ Mix of uppercase, lowercase, numbers, symbols
- ✅ Not a dictionary word
- ✅ Not personal information (birthday, name, etc.)

**Example:**

❌ Bad: `password123`, `qwerty`, `myname1980`
✅ Good: `Tr0ub4dor&3Horse!-Battery`, `Correct-Horse-Battery-Staple`

**Better: Use a password manager**

- Bitwarden (free, open-source)
- KeePassXC (free, offline)
- 1Password (paid)

#### 3. Enable Two-Factor Authentication (Coming Soon)

**What is 2FA?**

Something you know (password) + Something you have (phone).

**Why?**

If someone steals your password, they still can't log in without your phone.

**Coming in v1.3.0**

#### 4. Only Join Trusted Networks

**Why?**

Network hosts can technically see your traffic (metadata, not content due to encryption).

**Best practices:**

- ✅ Only join networks from people you trust
- ✅ Leave networks you no longer use
- ✅ Check member list regularly
- ❌ Don't join public networks from strangers

#### 5. Self-Host for Sensitive Data

**When to self-host:**

- Sharing work documents
- Accessing home systems
- Healthcare/financial data
- Anything you wouldn't want on a public server

**Why?**

You control the server, the database, and the logs.

---

## 🔍 Security Audits

### Past Audits

**Have we been audited?**

Not yet by third-party firms. However:

- ✅ Code review by security experts
- ✅ Penetration testing by developers
- ✅ Dependency scanning (Dependabot, Snyk)
- ✅ Static analysis (golangci-lint, Semgrep)

**Planning:**

We plan to commission a professional security audit in Q2 2025.

### Dependency Scanning

**How we check dependencies:**

1. **GitHub Dependabot**: Automated PRs for vulnerabilities
2. **Snyk**: Container and dependency scanning
3. **Govulncheck**: Go-specific vulnerability checker

**How often:**

- Automated: Every commit
- Manual: Before every release

**Zero-Dependency Policy:**

Production binary has **zero external dependencies**. This reduces attack surface.

---

## 📜 Legal

### Responsible Disclosure Law

**Good news:** In many countries, security research is legal if you follow responsible disclosure.

**Legal protections:**

- **USA**: DMCA Section 1201 exemption for security research
- **EU**: GDPR doesn't apply if you don't access personal data
- **Others**: Check your local laws

**Our promise:**

- We will NEVER take legal action against responsible disclosure
- We will credit you for finding vulnerabilities
- We may even hire you (if you're interested!)

---

## 📞 Contact

### Security Team

**Email:** [security@goconnect.io](mailto:security@goconnect.io)

**PGP Key:** (Coming Soon)

**Response Time:** Within 24 hours

### Non-Security Issues

**For bugs that aren't security issues:**

- GitHub Issues: https://github.com/orhaniscoding/goconnect/issues
- Discord: (Coming soon)
- Email: [support@goconnect.io](mailto:support@goconnect.io)

---

## 🔗 Resources

- **OWASP**: https://owasp.org/ - Web security standards
- **CVE**: https://cve.mitre.org/ - Vulnerability database
- **WireGuard**: https://www.wireguard.com/ - WireGuard protocol
- **Go Security**: https://golang.org/security/ - Go language security

---

**Last Updated:** 2025-01-24
**Language:** English
**Version:** 1.0.0

---

## Türkçe

## 📋 Genel Bakış

**Bu belge nedir?**

Bu belge, GoConnect'in güvenliği nasıl ele aldığını, güvenlik açıklarının nasıl bildirileceğini ve sizi nasıl koruduğunu açıklar.

**Neden bu önemli?**

Güvenlik herkesin sorumluluğudur. Bu belge şunlara yardımcı olur:
- **Kullanıcılar** GoConnect'in verilerini nasıl koruduğunu anlamak için
- **Geliştiriciler** güvenlik sorunlarını nasıl sorumlu şekilde bildireceklerini bilmek için
- **Araştırmacılar** güvenlik açığı disclosure programımızı öğrenmek için

---

## 🔒 Güvenlik İlkeleri

GoConnect şu temel güvenlik ilkelerini takip eder:

### 1. Sıfır Güven Mimarisi (Zero Trust)

**Bu ne demek?**

Asla varsayılan olarak hiçbir şeye güvenmeyiz. Her bağlantı, her istek, her kullanıcı doğrulanır.

**Neden önemli?**

Bir bileşen tehlikeye girerse, zarar sınırlı kalır.

**Örnekler:**
- ✅ Her API çağrısı kimlik doğrulaması gerektirir
- ✅ Her WebSocket bağlantısı doğrulanır
- ✅ Her dosya yüklemesi taranır
- ❌ "güvenilir dahili ağ" varsayımları yok

### 2. Her Yerde Şifreleme

**Ne şifreliyoruz?**

| Veri Türü | Şifreleme Yöntemi | Neden? |
|-----------|-------------------|-------|
| **Ağ Trafiği** | WireGuard (ChaCha20-Poly1305) | Peer-to-peer bağlantılar |
| **API İletişimi** | TLS 1.3 | Sunucu-istemci iletişimi |
| **Saklanan Şifreler** | Argon2id | Şifre hırsızlığını önler |
| **JWT Tokenlar** | RS256 | Token sahteciliğini önler |
| **Veritabanı** | İsteğe bağlı şifreleme (rest) | Sunucudan veri hırsızlığını önler |

**Neden bu kadar şifreleme?**

Biri trafiğinizi dinlerse, veritabanınızı çalarsa veya sunucunuzu tehlikeye atsa bile, verilerinizi okuyamaz.

### 3. Minimum Yetki (Least Privilege)

**Bu ne demek?**

Her bileşen işini yapmak için gereken minimum izinlere sahiptir.

**Örnekler:**
- CLI sadece ağ erişimine ihtiyaç duyar → Diğer uygulamalara dosya sistemi erişimi yok
- Masaüstü uygulaması sadece UI izinlerine ihtiyaç duyar → Sistem seviyesi erişim yok
- Sunucu sadece veritabanı erişimine ihtiyaç duyar → Doğrudan dosya sistemi erişimi yok

**Neden?**

Masaüstü uygulaması hacklenirse, saldırgan CLI'ye erişemez. Sunucu hacklenirse, saldırgan diğer servislere erişemez.

### 4. Derinlikli Savunma

**Bu ne demek?**

Güvenliğin çoklu katmanları. Bir katman başarısız olursa, diğerleri sizi hâlâ korur.

**Katmanlar:**
1. **Şifreleme** - Trafik dinlenirse, okunamaz
2. **Kimlik Doğrulama** - Şifreleme başarısız olursa, saldırganlar yine de kullanıcıları taklit edemez
3. **Yetkilendirme** - Kimlik doğrulama başarısız olursa, saldırganlar yine de kaynaklara erişemez
4. **Hız Sınırlandırma** - Yetkilendirme başarısız olursa, saldırganlar yine de brute force yapamaz
5. **İzleme** - Hepsi başarısız olursa, biz tespit eder ve yanıt veririz

---

## 🛡️ GoConnect Sizi Nasıl Korur

### Ağ Güvenliği

#### WireGuard Şifrelemesi

**WireGuard nedir?**

Ordular ve şirketler tarafından kullanılan modern bir VPN protokolüdür.

**Nasıl çalışır?**

```
Bilgisayarınız                          Arkadaşınızın Bilgisayarı
     │                                      │
     │  1. Anahtar değişimi (Curve25519)    │
     │<------------------------------------->│
     │                                      │
     │  2. Oturum anahtarı türet              │
     │     (ChaCha20-Poly1305)              │
     │                                      │
     │  3. Şifrelenmiş trafik               │
     │<====================================>│
     │                                      │
     ✅ Dinlenirse bile, okunamaz           │
```

**Hangi algoritmalar kullanılır?**

| Algoritma | Amaç | Anahtar Boyutu | Güvenlik Seviyesi |
|-----------|------|----------------|-------------------|
| **ChaCha20** | Şifreleme | 256-bit | ~256-bit güvenlik |
| **Poly1305** | Kimlik Doğrulama | 128-bit | Değişikliği önler |
| **Curve25519** | Anahtar Değişimi | 256-bit | Geçici anahtarlar |
| **Blake2s** | Hashleme | 256-bit | Hızlı, güvenli |

**Neden bu algoritmalar?**

- **Test edilmiş**: HTTPS, SSH, VPN'lerde dünya çapında kullanılır
- **Hızlı**: Minimum performans etkisi
- **Geleceğe hazır** - Kuantum dirençli (bir dereceye kadar)

**Bu sizin için ne anlamına gelir?**

Biri GoConnect trafiğinizi kaydetse bile, şifresini çözemez. Bir süperbilgisayarları bile olsa, milyarlarca yıl sürer.

### Kimlik Doğrulama ve Yetkilendirme

#### Şifre Güvenliği

**Şifreleri nasıl saklıyoruz?**

Asla gerçek şifrenizi saklamayız. Bunun yerine, bir "hash" - matematiksel bir parmak izi saklarız.

**Süreç:**

```
Şifreniz: "sifrem123"
                    │
                    ▼
            Salt Ekle (rastgele veri)
                    │
                    ▼
          Argon2id Hash (100,000 iterasyon)
                    │
                    ▼
         Saklanan Hash: "$argon2id$v=19$m=4096,t=3,p=1$..."
```

**Neden Argon2id?**

- **Bellek-ağır**: Cracking için çok RAM gerektirir (saldırganlar için pahalı)
- **Yavaş**: Hesaplamak zaman alır (brute force'ı yavaşlatır)
- **Önerilen**: OWASP, endüstri standardı

**Bu ne anlama gelir:**

Biri veritabanımızı çalsa bile, şifrenizi alamazlar. Bir şifreyi kırmak için milyarlarca dolarlık işlem gücü gerekir.

#### JWT Tokenlar

**JWT'ler nedir?**

JSON Web Tokenlar - giriş yaptığınızı kanıtlayan dijital kimlik kartları gibi.

**Nasıl çalışırlar?**

```
1. Giriş yaparsınız → Sunucu şifreyi doğrular
2. Sunucu JWT oluşturur → Özel anahtarla imzalar
3. Sunucu JWT gönderir → Tarayıcınız saklar
4. Her istekte JWT gönderirsiniz → Sunucu imzayı doğrular
5. Geçerliyse → Erişim izni verilir
```

**Neden güvenli?**

- **İmzalanmış**: Özel anahtar olmadan sahtesi yapılamaz
- **Durumsuz**: Sunucu oturumları saklamak zorunda değil
- **Kısa ömürlü**: Hızlı expires (çalınsa risk azalır)
- **Yenilenebilir**: Şifre olmadan yeni token alınabilir

**Token yapısı:**

```json
{
  "header": {
    "alg": "RS256",           // İmzalama algoritması
    "typ": "JWT"              // Token tipi
  },
  "payload": {
    "sub": "user123",         // Kullanıcı ID
    "exp": 1706457600,        // Bitiş zamanı
    "iat": 1706371200,        // Verildiği zaman
    "permissions": ["read", "write"]
  },
  "signature": "..."          // Kriptografik imza
}
```

### Veri Koruma

#### Ne Topluyoruz

**Sakladığımız veriler:**

| Veri | Amaç | Saklama Süresi | Şifreleme |
|------|------|----------------|-----------|
| **E-posta** | Hesap kimliği | Sonsuz | İletimde TLS |
| **Şifre Hash** | Kimlik doğrulama | Sonsuz | Argon2id |
| **Ağ Adı** | Ağlarınız | Sonsuz | İletimde TLS |
| **IP Adresi** | Ağ ataması | Ağ silinene kadar | WireGuard |
| **Sohbet Mesajları** | Bellekte röle | Saniyeler (teslim edilene kadar) | WireGuard |
| **Dosya Transferleri** | P2P yönlendirme | Saniyeler (teslim edilene kadar) | WireGuard |

**Saklamadığımız veriler:**

- ❌ Sohbet mesajı geçmişi (teslimden sonra silinir)
- ❌ Dosya içerikleri (P2P, asla sunucudan geçmez)
- ❌ Ses/video verileri (P2P WebRTC)
- ❌ Özel anahtarlarınız (cihazınızda yerel olarak saklanır)

#### Ne Paylaşıyoruz

**Verilerinizi ASLA satmayız. Nokta.**

**Kiminle paylaşıyoruz:**

| Üçüncü Taraf | Ne | Neden |
|-------------|------|------|
| **Kimse** | Hiçbir şey | Yok |

**İstisna:**

- **Yasal Gereklilikler** - Yasalar gereği talep edilirse (mahkeme emri), uyarırız
- **İzninizle** - Açıkça izin verirseniz

---

## 🐛 Güvenlik Açığı Bildirme

### Nasıl Bildirilir

**❌ YAPMAYIN:**
- Güvenlik açıklarını herkese açık yerlerde bildirin (GitHub Issues, Discord vb.)
- Proof-of-concept kodunu herkese açık paylaşın
- Gereğinden fazla açığı sömürmeye çalışın

**✅ YAPIN:**
- Güvenlik açıklarını özel olarak bildirin
- Yeniden üretmek için yeterli bilgi sağlayın
- Halka açıklamadan önce düzeltmemiz için bizi bekleyin

### Güvenlik Açığı Bildir

**E-posta:** [security@goconnect.io](mailto:security@goconnect.io)

**Ne dahil edilmeli:**

1. **Açıklama**: Güvenlik açığı nedir?
2. **Adımlar**: Nasıl yeniden üretilir?
3. **Etki**: Bir saldırgan ne yapabilir?
4. **Kanıt**: Ekran görüntüleri, loglar veya PoC kodu (güvenliyse)
5. **Sürüm**: Hangi GoConnect sürümü?

**Örnek Bildirim:**

```
Konu: Ağ Aramasında SQL Enjeksiyonu

Açıklama:
Ağ arama parametresi SQL enjeksiyonuna karşı savunmasız.

Yeniden Üretme Adımları:
1. Bir ağ oluşturun
2. Aramaya gidin
3. Şunu girin: '; DROP TABLE networks; --
4. Arama'ya tıklayın

Etki:
Bir saldırgan veritabanındaki tüm ağları silebilir.

Sürüm:
v1.2.0

Kanıt:
[Ek: ekran görüntüsü]

Önerilen Çözüm:
Prepared statements ve parametreli sorgular kullanın.
```

### Yanıt Zaman Çizelgesi

| Zaman | Ne Olur |
|------|---------|
| **24 saat** | Bildirimi aldığımızı onaylarız |
| **3 gün** | İnceler ve sorular sorarız |
| **7 gün** | Düzeltme tahmini sunarız |
| **90 gün** | Halka açıklamadan önce düzeltiriz (kritikse) |

### Koordineli İfşa

**Süreç:**

1. **Siz bildirirsiniz** → Güvenlik açığı detaylarını gönderirsiniz
2. **Biz onaylarız** → 24 saat içinde doğrularız
3. **Biz düzeltiriz** → Yama geliştirir ve test ederiz
4. **Biz yayınlarız** → Güncellemeyi deploy ederiz
5. **Siz ifşa edersiniz** → Yayından sonra herkese açıkça konuşabilirsiniz

**Neden bekleyelim?**

Düzeltmeden önce ifşa, kullanıcıları riske atar. Koordineli ifşa herkesi korur.

### Ödül

**Ne alırsınız:**

- 🏆 **Onur Listesi**: İsminiz web sitemizde listelenir
- 🎖️ **Güvenlik Araştırmacısı Rozeti**: GitHub profilinizde özel rozet
- 📢 **Kamuoyu Tanınması**: Sürüm notlarında size kredit veririz
- 🎁 **Swag**: GoConnect stickerleri/t-shirt (mümkünse)

**Not:** Para ödülü vermiyoruz (şimdilik). Bu gelecekte değişebilir.

---

## 🚨 Ciddiyet Seviyeleri

Güvenlik açıklarını ciddiyete göre sınıflandırıyoruz:

### 🔴 Kritik (48 saat içinde düzeltme)

**Örnekler:**
- Uzaktan Kod Çalıştırma (RCE)
- Tüm veritabanına erişim
- Herhangi bir kullanıcının kimliğine bürünme
- Tam sistem kontrolü

**Yanıt:**
- Her şeyi bırakıp hemen düzelt
- 48 saat içinde güvenlik güncellemesi yayınla
- Tüm kullanıcıları güncellemeye çağır

### 🟠 Yüksek (7 gün içinde düzeltme)

**Örnekler:**
- Diğer kullanıcıların verilerine erişim (hepsi değil)
- Yetki yükseltme (normal → admin)
- Hizmet Reddi (DoS) (tüm kullanıcıları etkiler)
- Kimlik doğrulamayı atlatma

**Yanıt:**
- Sonraki sürümde düzelt
- 7 gün içinde güvenlik güncellemesi yayınla
- Etkilen kullanıcıları bilgilendir

### 🟡 Orta (30 gün içinde düzeltme)

**Örnekler:**
- Görmemeniz gereken kendi verinize erişim
- Küçük yetki yükseltmesi
- DoS (bazı kullanıcıları etkiler)
- Bilgi ifşası (hassas olmayan)

**Yanıt:**
- Planlanmış sonraki sürümde düzelt
- Sürüm notlarına dahil et
- Acil bildirim yok

### 🟢 Düşük (90 gün içinde düzeltme)

**Örnekler:**
- Küçük bilgi ifşası
- Eksik güvenlik başlıkları
- Zayıf şifre gereksinimleri (sömürülemez)
- UI/UX güvenlik sorunları

**Yanıt:**
- Uygun olduğunda düzelt
- Sürüm notlarında bahset
- Özel bildirim yok

---

## 🛡️ Kullanıcılar İçin En İyi Uygulamalar

### Güvende Kalma

#### 1. Güncel Tutun

**Neden?**

Güncellemeler genellikle güvenlik düzeltmeleri içerir.

**Nasıl?**

- **Masaüstü Uygulaması**: Otomatik güncellemeler (varsayılan açık)
- **CLI**: Düzenli olarak `goconnect update` çalıştırın
- **Self-Hosted**: Sürümleri takip edin

**Sürümünüzü kontrol edin:**

```bash
goconnect version
```

#### 2. Güçlü Şifreler Kullanın

**Güçlü şifre nedir?**

- ✅ En az 12 karakter
- ✅ Büyük harf, küçük harf, sayı, sembol karışımı
- ✅ Sözlük kelimesi değil
- ✅ Kişisel bilgi değil (doğum günü, isim vb.)

**Örnek:**

❌ Kötü: `sifre123`, `qwerty`, `isim1980`
✅ İyi: `Tr0ub4dor&3At!-Pil`, `Dogu-Akrep-Pil-Sabah`

**Daha iyi: Bir şifre yöneticisi kullanın**

- Bitwarden (ücretsiz, açık kaynak)
- KeePassXC (ücretsiz, çevrimdışı)
- 1Password (ücretli)

#### 3. İki Faktörlü Kimlik Doğrulama Etkinleştirin (Yakında)

**2FA nedir?**

Bildikleriniz bir şey (şifre) + Sahip olduğunuz bir şey (telefon).

**Neden?**

Biri şifrenizi çalsa bile, telefonunuz olmadan giriş yapamazlar.

**v1.3.0'da geliyor**

#### 4. Sadece Güvenilir Ağlara Katılın

**Neden?**

Ağ sahipleri teknik olarak trafiğinizi görebilir (içerik değil, metadata - şifreleme yüzünden).

**En iyi uygulamalar:**

- ✅ Sadece güvendiğiniz kişilerden ağlara katılın
- ✅ Artık kullanmadığınız ağlardan ayrılın
- ✅ Üye listesini düzenli kontrol edin
- ❌ Yabancılardan gelen herkese açık ağlara katılmayın

#### 5. Hassas Veriler İçin Self-Host

**Ne zaman self-host?**

- İş belgelerini paylaşırken
- Ev sistemlerine erişirken
- Sağlık/finans verileri
- Herkese açık sunucuda istemediğiniz her şey

**Neden?**

Sunucuyu, veritabanını ve logları siz kontrol edersiniz.

---

## 🔍 Güvenlik Denetimleri

### Geçmiş Denetimler

**Denetlendik mi?**

Henüz üçüncü taraf firmalar tarafından değil. Ancak:

- ✅ Güvenlik uzmanları tarafından kod incelemesi
- ✅ Geliştiriciler tarafından penetration testing
- ✅ Bağımlılık taraması (Dependabot, Snyk)
- ✅ Statik analiz (golangci-lint, Semgrep)

**Planlama:**

2025 Q2'de profesyonel bir güvenlik denetimi planlıyoruz.

### Bağımlılık Taraması

**Bağımlılıkları nasıl kontrol ediyoruz?**

1. **GitHub Dependabot**: Güvenlik açıkları için otomatik PR'ler
2. **Snyk**: Konteyner ve bağımlılık taraması
3. **Govulncheck**: Go'ya özel vulnerability checker

**Ne sıklıkta:**

- Otomatik: Her commit
- Manuel: Her sürüm öncesi

**Sıfır Bağımlılık Politikası:**

Production binary'de **sıfır dış bağımlılık** var. Bu saldırı yüzeyini azaltır.

---

## 📜 Yasal

### Sorumlu Disclosure Yasası

**İyi haber:** Birçok ülkede, sorumlu disclosure'i takip ederseniz güvenlik araştırması yasaldır.

**Yasal korumalar:**

- **ABD**: Güvenlik araştırması için DMCA Bölüm 1201 istisnası
- **AB**: Kişisel veriye erişmezseniz GDPR geçmez
- **Diğerleri**: Yerel yasalarınızı kontrol edin

**Sözümüz:**

- Sorumlu disclosure nedeniyle ASLA yasal işlem yapmayız
- Güvenlik açıklarını bulduğunuz için size kredit veririz
- Hatta işe alabiliriz (eğer ilginiz varsa!)

---

## 📞 İletişim

### Güvenlik Ekibi

**E-posta:** [security@goconnect.io](mailto:security@goconnect.io)

**PGP Anahtarı:** (Çok Yakında)

**Yanıt Süresi:** 24 saat içinde

### Güvenlik Olmayan Sorunlar

**Güvenlik olmayan hatalar için:**

- GitHub Issues: https://github.com/orhaniscoding/goconnect/issues
- Discord: (Çok yakında)
- E-posta: [support@goconnect.io](mailto:support@goconnect.io)

---

## 🔗 Kaynaklar

- **OWASP**: https://owasp.org/ - Web güvenlik standartları
- **CVE**: https://cve.mitre.org/ - Güvenlik açığı veritabanı
- **WireGuard**: https://www.wireguard.com/ - WireGuard protokolü
- **Go Security**: https://golang.org/security/ - Go dili güvenliği

---

**Son Güncelleme:** 2025-01-24
**Dil:** Türkçe
**Sürüm:** 1.0.0
