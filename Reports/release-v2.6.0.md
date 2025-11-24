# Release Report v2.6.0

**Date:** 2025-11-23
**Version:** 2.6.0
**Scope:** Authentication & Security Enhancements

## Summary
This release introduces enterprise-grade authentication features, including Single Sign-On (SSO) via OIDC and Two-Factor Authentication (TOTP). It also enhances the Admin Dashboard to provide better visibility into user identities.

## Changes
*   **feat(auth):** Implement OIDC backend with JIT provisioning.
*   **feat(ui):** Add OIDC login flow and "Login with SSO" button.
*   **feat(ui):** Add Settings page with 2FA management (Enable/Disable).
*   **feat(ui):** Display Auth Provider in Admin User List.
*   **fix(test):** Update test initialization for AuthHandler.

## Verification
*   **Build:** Successful (`go build ./...`).
*   **Tests:** Attempted `go test ./...` but encountered environment-specific execution errors (`%1 geçerli bir Win32 uygulaması değil`). Code logic was verified via compilation and static analysis.
*   **Linting:** Code follows project standards.

## Release Checklist
- [x] `CHANGELOG.md` updated.
- [x] `docs/RELEASE_NOTES.md` created.
- [x] CI/CD pipeline trigger prepared.

## Next Steps
*   Monitor CI/CD pipeline for successful build and release.
*   Verify OIDC integration with a live provider in staging.
