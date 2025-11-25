# SSO & 2FA Design

## 1. OIDC (OpenID Connect) Integration

GoConnect supports SSO via OpenID Connect (OIDC). This allows users to log in using providers like Google, GitHub, GitLab, or any standard OIDC provider.

### Configuration
The following environment variables are required:
- `OIDC_ISSUER`: The issuer URL (e.g., `https://accounts.google.com`)
- `OIDC_CLIENT_ID`: The client ID
- `OIDC_CLIENT_SECRET`: The client secret
- `OIDC_REDIRECT_URL`: The callback URL (e.g., `https://api.goconnect.example/v1/auth/oidc/callback`)

### Flow
1. **Initiate Login:**
   - Client requests `GET /v1/auth/oidc/login?provider=google`
   - Server generates a state parameter (stored in cookie/redis) and redirects to Provider's Authorization URL.

2. **Callback:**
   - Provider redirects user back to `GET /v1/auth/oidc/callback?code=...&state=...`
   - Server validates `state`.
   - Server exchanges `code` for `id_token` and `access_token`.
   - Server verifies `id_token`.
   - Server extracts user info (email, sub, name).

3. **User Provisioning:**
   - If user exists (by email), log them in.
   - If user does not exist, create a new user (JIT provisioning) with `auth_provider='oidc'`.

4. **Session:**
   - Server issues standard GoConnect JWT (Access + Refresh tokens).
   - Redirects user to Web UI with tokens (e.g., `https://app.goconnect.example/auth/callback?access_token=...`).

## 2. Two-Factor Authentication (TOTP)

Implemented using `github.com/pquerna/otp/totp`.

### Flow
1. **Enable:**
   - User requests `POST /v1/auth/2fa/generate`.
   - Server returns Secret + QR Code URL.
   - User scans QR and sends code to `POST /v1/auth/2fa/enable`.
   - Server verifies code and sets `two_fa_enabled=true` in DB.

2. **Login:**
   - If user has 2FA enabled, `POST /v1/auth/login` returns `403 Forbidden` with error `ERR_2FA_REQUIRED`.
   - Client prompts for 2FA code.
   - Client sends `POST /v1/auth/login` with `code` field.
   - Server verifies code and issues JWT.

### Recovery Codes

Recovery codes provide a backup method to access your account if you lose access to your authenticator app.

#### Implementation Details
- **8 one-time use codes** are generated per user
- Format: `XXXXX-XXXXX` (10 alphanumeric chars with dash)
- Character set excludes confusing characters: No `O`, `0`, `I`, `1`
- Codes are **Argon2id hashed** before storage (same as passwords)
- Each code can only be used **once** - it's removed after use

#### API Endpoints

##### Generate Recovery Codes
```http
POST /v1/auth/2fa/recovery-codes
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "code": "123456"  // Current TOTP code to verify
}
```

Response:
```json
{
  "data": {
    "codes": [
      "ABCDE-FGHIJ",
      "KLMNP-QRSTU",
      "VWXYZ-23456",
      "789AB-CDEFG",
      "HJKLM-NPQRS",
      "TUVWX-YZ234",
      "56789-ABCDE",
      "FGHJK-LMNPQ"
    ]
  },
  "message": "Recovery codes generated. Store these safely - they will only be shown once!"
}
```

##### Get Remaining Code Count
```http
GET /v1/auth/2fa/recovery-codes/count
Authorization: Bearer {access_token}
```

Response:
```json
{
  "data": {
    "remaining_codes": 6
  }
}
```

##### Login with Recovery Code
```http
POST /v1/auth/2fa/recovery
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "userpassword",
  "recovery_code": "ABCDE-FGHIJ"
}
```

Response: Same as regular login response with JWT tokens.

#### User Flow

1. **Initial Setup:**
   - User enables 2FA via authenticator app
   - User generates recovery codes from Settings
   - User downloads/stores codes securely (shown only once!)

2. **Lost Authenticator:**
   - On login page, user clicks "Use recovery code instead"
   - User enters email, password, and one recovery code
   - Code is validated and consumed (one-time use)
   - User logs in successfully
   - System recommends generating new recovery codes

3. **Regenerating Codes:**
   - User can regenerate codes anytime from Settings
   - Requires current TOTP code to verify identity
   - Old codes are **invalidated** when new ones are generated

#### Security Considerations
- Codes are hashed with Argon2id (OWASP recommended)
- Rate limiting applies to recovery code attempts
- Codes cannot be recovered - if all codes are used, user must contact admin
- Consider storing codes in a password manager or physical safe
