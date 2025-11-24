# Release v2.7.0 Report

**Date:** 2025-11-25
**Version:** v2.7.0

## 🚀 Summary
This release introduces significant improvements to authentication security, client daemon capabilities, and frontend internationalization.

## ✨ Key Features

### 1. Token Blacklisting (Redis)
- Implemented JWT blacklisting using Redis.
- `Logout` now invalidates the refresh token immediately.
- `Refresh` endpoint checks the blacklist before issuing new tokens.
- Added `RedisConfig` to server configuration.

### 2. Client Daemon Improvements
- Enhanced OS detection logic in `internal/system/info.go`.
- Improved Linux DNS configuration using `resolvconf`.
- Better handling of network interface configuration.

### 3. Frontend Internationalization (i18n)
- Refactored Login and Register pages to use the `useT` hook.
- Improved error message localization.
- Standardized translation keys.

## 🛠 Technical Details
- **Backend:** Go 1.24, `go-redis/v9`.
- **Frontend:** Next.js 15, `next-intl` (custom implementation).
- **Database:** PostgreSQL (Schema unchanged).

## 🧪 Testing
- All backend unit tests passed (`go test ./...`).
- Verified Redis integration for token blacklisting.
- Verified frontend build (implicit via CI).

## 📝 Notes
- Redis is now a recommended dependency for production deployments to enable token revocation.
- Without Redis, the blacklist feature gracefully degrades (is disabled).
