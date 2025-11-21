# Admin Dashboard Guide

The Admin Dashboard is a centralized interface for system administrators to manage the GoConnect instance. It provides a high-level overview of the system status and allows for the management of users, tenants, networks, and devices.

## Accessing the Admin Panel

To access the Admin Panel:
1. Log in to the GoConnect Web UI.
2. If your user account has the `is_admin` flag set to `true`, you will see an "Admin Panel" link in the navigation menu or dashboard.
3. Navigate to `/admin`.

> **Note:** Only users with the `is_admin` privilege can access this page. Attempts to access it by non-admin users will result in an "Access Denied" message.

## Features

The Admin Dashboard is divided into several tabs:

### 1. 📊 Statistics
This is the landing tab, providing a snapshot of the system's health and usage.
- **Total Users:** Count of registered users.
- **Total Tenants:** Count of active tenants.
- **Total Networks:** Count of networks created.
- **Total Devices:** Count of registered devices.
- **Active Connections:** (If supported) Number of currently active connections.
- **Messages Today:** Number of chat messages sent in the last 24 hours.

### 2. 👥 User Management
Allows administrators to view and manage all users in the system.
- **List Users:** View a list of users with their Email, Name, Role, and Tenant ID.
- **Toggle Admin:** Click the "Toggle Admin" button to grant or revoke administrator privileges for a user.
- **Delete User:** Click the "Delete" button to permanently remove a user from the system. **Warning:** This action is irreversible.

### 3. 🏢 Tenant Management
Allows administrators to oversee the tenants (organizations) in the system.
- **List Tenants:** View a list of tenants with their Name and ID.
- **Delete Tenant:** Click the "Delete" button to remove a tenant. **Warning:** This will likely cascade and delete associated resources (depending on database configuration).

### 4. 🌐 Network Management
Provides a system-wide view of all networks.
- **List Networks:** View all networks across all tenants.
- **Details:** Displays Network Name, CIDR, Tenant ID, and Creation Date.
- **Note:** Currently, this view is read-only.

### 5. 💻 Device Management
Allows administrators to monitor and manage devices connected to the system.
- **List Devices:** View all devices registered in the system.
- **Details:** Displays Device Name, IP Address, Public Key (truncated), Tenant ID, and Last Seen timestamp.
- **Delete Device:** Click the "Delete" button to remove a device. This is useful for cleaning up stale or compromised devices.

### 6. 📜 Audit Logs
A chronological log of important system events for security and compliance.
- **List Logs:** View recent system actions.
- **Columns:**
  - **Time:** When the event occurred.
  - **Action:** The type of event (e.g., `USER_CREATED`, `DEVICE_DELETED`).
  - **Actor:** The user ID who performed the action.
  - **Object:** The ID of the object being acted upon.
  - **Details:** JSON details about the event.

## Security Considerations
- The Admin Panel exposes sensitive system information. Ensure that only trusted individuals are granted admin privileges.
- All administrative actions (like deleting users or devices) should be performed with caution.
- Check the **Audit Logs** regularly to monitor for suspicious administrative activity.
