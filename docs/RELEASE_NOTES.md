# Release Notes v2.6.0

## 🚀 New Features

### Authentication & Security
- **OIDC (SSO) Support:** Users can now log in using OpenID Connect providers (e.g., Google, GitHub, Keycloak).
- **Just-In-Time (JIT) Provisioning:** New users logging in via SSO are automatically created and assigned a default tenant.
- **Two-Factor Authentication (2FA):** Added TOTP-based 2FA support. Users can enable/disable 2FA from their settings.
- **Admin Visibility:** Admin dashboard now displays the authentication provider (SSO vs Local) for each user.

### User Interface
- **Settings Page:** New settings page for user profile and security management.
- **Login Page:** Added "Login with SSO" button.
- **Admin Dashboard:** Enhanced user list with provider badges.

## 🐛 Bug Fixes
- Fixed `NewAuthHandler` initialization in tests.

## 🔧 Technical Details
- **Backend:** Integrated `coreos/go-oidc` for OIDC flow.
- **Frontend:** Added `qrcode` library for 2FA setup.
- **Database:** Updated `User` model to support `auth_provider` and `external_id`.

## 📦 Upgrade Instructions
No database migrations are required for this release (schema changes are handled in code/NoSQL-like usage for now, or assumed compatible).
Ensure `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, and `OIDC_REDIRECT_URL` environment variables are set if using SSO.
