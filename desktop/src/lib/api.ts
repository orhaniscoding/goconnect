// =============================================================================
// GoConnect API Client
// =============================================================================

const API_BASE = "http://localhost:8081/api/v1";

// Types matching backend domain
export interface User {
    id: string;
    deviceId: string;
    username: string;
}

export interface Tenant {
    id: string;
    name: string;
    description?: string;
    icon?: string;
    visibility: "public" | "unlisted" | "private";
    access_type: "open" | "password" | "invite_only";
    member_count: number;
    created_at: string;
    is_owner?: boolean;
}

export interface Network {
    id: string;
    tenant_id: string;
    name: string;
    cidr: string;
    visibility: "public" | "private";
    join_policy: "open" | "invite" | "approval";
    member_count?: number;
    connected?: boolean;
    my_ip?: string;
}

export interface NetworkMember {
    id: string;
    user_id: string;
    username: string;
    ip: string;
    status: "online" | "idle" | "offline";
    is_host: boolean;
    last_seen: string;
}

export interface TenantInvite {
    id: string;
    code: string;
    max_uses: number;
    use_count: number;
    expires_at?: string;
}

// API Response wrapper
interface ApiResponse<T> {
    data?: T;
    error?: string;
    message?: string;
}

// Get stored auth
function getAuth(): { deviceId: string; username: string } | null {
    const deviceId = localStorage.getItem("gc_device_id");
    const username = localStorage.getItem("gc_username");
    if (deviceId && username) {
        return { deviceId, username };
    }
    return null;
}

// Base fetch with auth headers
async function apiFetch<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<ApiResponse<T>> {
    const auth = getAuth();

    const headers: HeadersInit = {
        "Content-Type": "application/json",
        ...(auth && { "X-Device-ID": auth.deviceId }),
        ...(options.headers || {}),
    };

    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            headers,
        });

        const data = await response.json();

        if (!response.ok) {
            return { error: data.message || data.error || "Request failed" };
        }

        return { data };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

// =============================================================================
// Auth / Device
// =============================================================================

export async function registerDevice(username: string, deviceId: string) {
    return apiFetch<User>("/devices/register", {
        method: "POST",
        body: JSON.stringify({
            device_id: deviceId,
            username,
            platform: "windows", // TODO: detect platform
            hostname: "Desktop",
        }),
    });
}

export async function heartbeat() {
    return apiFetch<void>("/devices/heartbeat", {
        method: "POST",
    });
}

// =============================================================================
// Tenants (Servers)
// =============================================================================

export async function listTenants() {
    return apiFetch<Tenant[]>("/tenants");
}

export async function getMyTenants() {
    return apiFetch<Tenant[]>("/tenants/me");
}

export async function createTenant(data: {
    name: string;
    description?: string;
    icon?: string;
    visibility?: "public" | "unlisted" | "private";
    access_type?: "open" | "password" | "invite_only";
}) {
    return apiFetch<Tenant>("/tenants", {
        method: "POST",
        body: JSON.stringify({
            name: data.name,
            description: data.description || "",
            visibility: data.visibility || "private",
            access_type: data.access_type || "invite_only",
        }),
    });
}

export async function getTenant(tenantId: string) {
    return apiFetch<Tenant>(`/tenants/${tenantId}`);
}

export async function deleteTenant(tenantId: string) {
    return apiFetch<void>(`/tenants/${tenantId}`, {
        method: "DELETE",
    });
}

export async function joinTenantByCode(code: string) {
    return apiFetch<Tenant>("/tenants/join", {
        method: "POST",
        body: JSON.stringify({ code }),
    });
}

export async function leaveTenant(tenantId: string) {
    return apiFetch<void>(`/tenants/${tenantId}/leave`, {
        method: "POST",
    });
}

// =============================================================================
// Tenant Invites
// =============================================================================

export async function createTenantInvite(
    tenantId: string,
    data?: { max_uses?: number; expires_in?: number }
) {
    return apiFetch<TenantInvite>(`/tenants/${tenantId}/invites`, {
        method: "POST",
        body: JSON.stringify(data || {}),
    });
}

export async function listTenantInvites(tenantId: string) {
    return apiFetch<TenantInvite[]>(`/tenants/${tenantId}/invites`);
}

export async function revokeTenantInvite(tenantId: string, inviteId: string) {
    return apiFetch<void>(`/tenants/${tenantId}/invites/${inviteId}`, {
        method: "DELETE",
    });
}

// =============================================================================
// Networks
// =============================================================================

export async function listNetworks(tenantId: string) {
    return apiFetch<Network[]>(`/tenants/${tenantId}/networks`);
}

export async function createNetwork(
    tenantId: string,
    data: {
        name: string;
        cidr?: string;
        visibility?: "public" | "private";
        join_policy?: "open" | "invite" | "approval";
    }
) {
    return apiFetch<Network>(`/tenants/${tenantId}/networks`, {
        method: "POST",
        body: JSON.stringify({
            name: data.name,
            cidr: data.cidr || "10.0.0.0/24",
            visibility: data.visibility || "private",
            join_policy: data.join_policy || "open",
        }),
    });
}

export async function getNetwork(tenantId: string, networkId: string) {
    return apiFetch<Network>(`/tenants/${tenantId}/networks/${networkId}`);
}

export async function deleteNetwork(tenantId: string, networkId: string) {
    return apiFetch<void>(`/tenants/${tenantId}/networks/${networkId}`, {
        method: "DELETE",
    });
}

export async function joinNetwork(tenantId: string, networkId: string) {
    return apiFetch<{ ip: string }>(`/tenants/${tenantId}/networks/${networkId}/join`, {
        method: "POST",
    });
}

export async function leaveNetwork(tenantId: string, networkId: string) {
    return apiFetch<void>(`/tenants/${tenantId}/networks/${networkId}/leave`, {
        method: "POST",
    });
}

export async function listNetworkMembers(tenantId: string, networkId: string) {
    return apiFetch<NetworkMember[]>(`/tenants/${tenantId}/networks/${networkId}/members`);
}

// =============================================================================
// Connection (WireGuard)
// =============================================================================

export async function getNetworkConfig(tenantId: string, networkId: string) {
    return apiFetch<{
        interface: {
            private_key?: string;
            addresses: string[];
            dns?: string[];
            mtu?: number;
        };
        peers: Array<{
            public_key: string;
            endpoint?: string;
            allowed_ips: string[];
            persistent_keepalive?: number;
        }>;
    }>(`/tenants/${tenantId}/networks/${networkId}/config`);
}

// =============================================================================
// Utility
// =============================================================================

export function isApiAvailable(): Promise<boolean> {
    return fetch(`${API_BASE}/health`)
        .then((r) => r.ok)
        .catch(() => false);
}
