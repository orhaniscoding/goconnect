# Story 2.1: Device Identity Generation & Storage

## Story
As a Security Engineer,
I want the daemon to generate a stable, unique device identity (UUID + Keypair) and store it securely,
So that the device can be consistently identified and authenticated by the control plane.

## Acceptance Criteria
- [ ] **Given** a fresh installation
- [ ] **When** the daemon starts for the first time
- [ ] **Then** it generates a new Device UUID
- [ ] **And** it generates a WireGuard Keypair (Private/Public)
- [ ] **And** it stores these securely (e.g., file with restricted permissions or system keyring)
- [ ] **And** subsequent restarts load the *same* identity

## Tasks/Subtasks
- [ ] Analyze `cli/internal/identity` for existing logic.
- [ ] Verify key generation uses crypto-secure defaults (WireGuard standard).
- [ ] Implement/Verify secure storage mechanisms (permissions check for file storage).
- [ ] Ensure `idManager.LoadOrCreateIdentity()` satisfies strict security requirements.

## Dev Notes
- `cli/internal/identity/manager.go` likely already exists.
- We need to ensure it's not just writing world-readable JSON.
- Permissions should be `0600` for identity files.
- Consider moving sensitive parts to system keyring if not already done.

## File List
- cli/internal/identity/manager.go
- cli/internal/identity/storage.go (if exists)

## Change Log
- Initial creation.
- Implemented `LocalUUID` generation using `google/uuid`.
- Added backfill logic for existing identities.
- Updated unit tests to verify UUID persistence.

## Status
done
