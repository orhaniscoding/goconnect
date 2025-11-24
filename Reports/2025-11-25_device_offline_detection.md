# Device Offline Detection & Audit Log Fix

**Date:** 2025-11-25
**Author:** GitHub Copilot

## 🚀 Summary
Implemented a background worker to detect and mark offline devices, and fixed a critical bug in the Audit Logs handler.

## ✨ Features

### 1. Device Offline Detection
- **Problem:** Devices were never marked as "Offline" unless explicitly disconnected. Stale devices remained "Online" indefinitely.
- **Solution:**
    - Added `GetStaleDevices` method to `DeviceRepository` (InMemory & Postgres).
    - Implemented `StartOfflineDetection` in `DeviceService` which runs a periodic check (every 30s).
    - Devices unseen for >2 minutes are marked as inactive and a `device.offline` notification is sent via WebSocket.
    - Integrated the worker into `main.go`.

### 2. Audit Log Fix
- **Problem:** `AuditListHandler` was attempting to cast a non-existent `claims` object from the context, causing 401 errors or potential panics.
- **Solution:** Updated the handler to correctly retrieve `tenant_id` directly from the context (set by `AuthMiddleware`).

## 🛠 Technical Details

### Repository Changes
- **Interface:** Added `GetStaleDevices(ctx, threshold)` to `DeviceRepository`.
- **Postgres:** Implemented efficient query: `WHERE active = TRUE AND last_seen < $1 AND disabled_at IS NULL`.
- **InMemory:** Implemented filtering logic matching Postgres behavior.

### Service Changes
- **DeviceService:** Added `StartOfflineDetection` loop.
- **Main:** Started the detection goroutine with `30s` interval and `2m` threshold.

## 🧪 Testing
- **Unit Tests:** Added `TestDeviceRepository_GetStaleDevices` to verify filtering logic (Active, Inactive, Disabled, Stale).
- **Integration:** Verified `go test ./internal/repository/...` passes.

## 📝 Next Steps
- Verify the "Device Offline" notification in the frontend (Web UI).
