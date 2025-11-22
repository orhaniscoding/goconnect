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

### Recovery Codes (Planned)
- Generate 10 one-time use codes when enabling 2FA.
