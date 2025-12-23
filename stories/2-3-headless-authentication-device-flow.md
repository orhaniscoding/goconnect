# Story 2.3: Headless Authentication Device Flow

## Story
As a Server Administrator,
I want to authenticate the GoConnect daemon on a headless server using a separate device (laptop/phone),
So that I can connect my server to the network without needing a GUI or browser on the server itself.

## Acceptance Criteria
- [ ] **Given** a headless server with the GoConnect daemon installed
- [ ] **When** I run `goconnect login`
- [ ] **Then** the CLI displays a verification URL and a User Code (e.g., "Visit https://goconnect.io/activate and enter code ABCD-1234")
- [ ] **And** it polls the API (or waits for SSE) for authentication completion
- [ ] **When** I complete the flow on my browser
- [ ] **Then** the CLI receives the Auth Token
- [ ] **And** securely stores it in the system keyring
- [ ] **And** the daemon connects automatically.

## Tasks/Subtasks
- [ ] Check `api/client.go` for "Device Flow" or "Poll Auth" endpoints.
- [ ] Implement `StartDeviceLogin()`, `PollLoginStatus()` in API client if missing.
- [ ] Implement `goconnect login` command in CLI.
- [ ] Integrate with `keyring.StoreAuthToken()`.

## Dev Notes
- This is standard OAuth2 Device Flow (RFC 8628).
- Daemon API needs to support:
    - `POST /auth/device/code` -> returns `device_code`, `user_code`, `verification_uri`
    - `POST /auth/device/token` -> exchange `device_code` for `access_token` (poll or blocking)
- If Server doesn't support this yet, we might need to mock it or implement a simplified "Copy Token" flow as a temporary fallback (User logs in on browser -> generates token -> `goconnect login --token <token>`).
- **Wait**, check if we already have `cli/cmd/auth.go` or similar.

## File List
- cli/cmd/login.go
- client-daemon/internal/api/client.go

## Change Log
- Initial creation.

## Status
ready-for-dev
