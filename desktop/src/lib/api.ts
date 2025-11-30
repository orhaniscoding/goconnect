// =============================================================================
// GoConnect API Client
// =============================================================================
// Terminology: Server (tenant) > Network > Client (device/member)

const API_BASE = "http://localhost:8081/api/v1";

// Types - UI uses Server/Network/Client terminology
export interface User {
    id: string;
    deviceId: string;
    username: string;
}

// Server = Tenant in backend
export interface Server {
    id: string;
    name: string;
    description?: string;
    icon?: string;
    visibility: "public" | "unlisted" | "private";
    accessType: "open" | "password" | "invite_only";
    memberCount: number;
    createdAt: string;
    isOwner?: boolean;
}

export interface Network {
    id: string;
    serverId: string;
    name: string;
    subnet: string;
    visibility: "public" | "private";
    joinPolicy: "open" | "invite" | "approval";
    memberCount?: number;
    connected?: boolean;
    myIp?: string;
}

// Client = Member/Device in backend
export interface Client {
    id: string;
    userId: string;
    username: string;
    ip: string;
    status: "online" | "idle" | "offline";
    isHost: boolean;
    lastSeen: string;
}

export interface ServerInvite {
    id: string;
    code: string;
    maxUses: number;
    useCount: number;
    expiresAt?: string;
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
// Servers (backend: tenants)
// =============================================================================

export async function listServers() {
    const result = await apiFetch<any[]>("/tenants");
    return { ...result, data: result.data?.map(mapServerFromApi) };
}

export async function getMyServers() {
    const result = await apiFetch<any[]>("/tenants/me");
    return { ...result, data: result.data?.map(mapServerFromApi) };
}

export async function createServer(data: {
    name: string;
    description?: string;
    icon?: string;
    visibility?: "public" | "unlisted" | "private";
    accessType?: "open" | "password" | "invite_only";
}) {
    const result = await apiFetch<any>("/tenants", {
        method: "POST",
        body: JSON.stringify({
            name: data.name,
            description: data.description || "",
            visibility: data.visibility || "private",
            access_type: data.accessType || "invite_only",
        }),
    });
    return { ...result, data: result.data ? mapServerFromApi(result.data) : undefined };
}

export async function getServer(serverId: string) {
    const result = await apiFetch<any>(`/tenants/${serverId}`);
    return { ...result, data: result.data ? mapServerFromApi(result.data) : undefined };
}

export async function deleteServer(serverId: string) {
    return apiFetch<void>(`/tenants/${serverId}`, {
        method: "DELETE",
    });
}

export async function joinServerByCode(code: string) {
    const result = await apiFetch<any>("/tenants/join", {
        method: "POST",
        body: JSON.stringify({ code }),
    });
    return { ...result, data: result.data ? mapServerFromApi(result.data) : undefined };
}

export async function leaveServer(serverId: string) {
    return apiFetch<void>(`/tenants/${serverId}/leave`, {
        method: "POST",
    });
}

// Map backend tenant to UI server
function mapServerFromApi(data: any): Server {
    return {
        id: data.id,
        name: data.name,
        description: data.description,
        icon: data.icon,
        visibility: data.visibility,
        accessType: data.access_type,
        memberCount: data.member_count,
        createdAt: data.created_at,
        isOwner: data.is_owner,
    };
}

// =============================================================================
// Server Invites
// =============================================================================

export async function createServerInvite(
    serverId: string,
    data?: { maxUses?: number; expiresIn?: number }
) {
    const result = await apiFetch<any>(`/tenants/${serverId}/invites`, {
        method: "POST",
        body: JSON.stringify({
            max_uses: data?.maxUses,
            expires_in: data?.expiresIn,
        }),
    });
    return { ...result, data: result.data ? mapInviteFromApi(result.data) : undefined };
}

export async function listServerInvites(serverId: string) {
    const result = await apiFetch<any[]>(`/tenants/${serverId}/invites`);
    return { ...result, data: result.data?.map(mapInviteFromApi) };
}

export async function revokeServerInvite(serverId: string, inviteId: string) {
    return apiFetch<void>(`/tenants/${serverId}/invites/${inviteId}`, {
        method: "DELETE",
    });
}

function mapInviteFromApi(data: any): ServerInvite {
    return {
        id: data.id,
        code: data.code,
        maxUses: data.max_uses,
        useCount: data.use_count,
        expiresAt: data.expires_at,
    };
}

// =============================================================================
// Networks
// =============================================================================

export async function listNetworks(serverId: string) {
    const result = await apiFetch<any[]>(`/tenants/${serverId}/networks`);
    return { ...result, data: result.data?.map(mapNetworkFromApi) };
}

export async function createNetwork(
    serverId: string,
    data: {
        name: string;
        subnet?: string;
        visibility?: "public" | "private";
        joinPolicy?: "open" | "invite" | "approval";
    }
) {
    const result = await apiFetch<any>(`/tenants/${serverId}/networks`, {
        method: "POST",
        body: JSON.stringify({
            name: data.name,
            cidr: data.subnet || "10.0.0.0/24",
            visibility: data.visibility || "private",
            join_policy: data.joinPolicy || "open",
        }),
    });
    return { ...result, data: result.data ? mapNetworkFromApi(result.data) : undefined };
}

export async function getNetwork(serverId: string, networkId: string) {
    const result = await apiFetch<any>(`/tenants/${serverId}/networks/${networkId}`);
    return { ...result, data: result.data ? mapNetworkFromApi(result.data) : undefined };
}

export async function deleteNetwork(serverId: string, networkId: string) {
    return apiFetch<void>(`/tenants/${serverId}/networks/${networkId}`, {
        method: "DELETE",
    });
}

export async function joinNetwork(serverId: string, networkId: string) {
    return apiFetch<{ ip: string }>(`/tenants/${serverId}/networks/${networkId}/join`, {
        method: "POST",
    });
}

export async function leaveNetwork(serverId: string, networkId: string) {
    return apiFetch<void>(`/tenants/${serverId}/networks/${networkId}/leave`, {
        method: "POST",
    });
}

export async function listNetworkClients(serverId: string, networkId: string) {
    const result = await apiFetch<any[]>(`/tenants/${serverId}/networks/${networkId}/members`);
    return { ...result, data: result.data?.map(mapClientFromApi) };
}

function mapNetworkFromApi(data: any): Network {
    return {
        id: data.id,
        serverId: data.tenant_id,
        name: data.name,
        subnet: data.cidr,
        visibility: data.visibility,
        joinPolicy: data.join_policy,
        memberCount: data.member_count,
    };
}

function mapClientFromApi(data: any): Client {
    return {
        id: data.id,
        userId: data.user_id,
        username: data.username,
        ip: data.ip,
        status: data.status,
        isHost: data.is_host,
        lastSeen: data.last_seen,
    };
}

// =============================================================================
// Connection (WireGuard)
// =============================================================================

export async function getNetworkConfig(serverId: string, networkId: string) {
    return apiFetch<{
        interface: {
            privateKey?: string;
            addresses: string[];
            dns?: string[];
            mtu?: number;
        };
        peers: Array<{
            publicKey: string;
            endpoint?: string;
            allowedIps: string[];
            persistentKeepalive?: number;
        }>;
    }>(`/tenants/${serverId}/networks/${networkId}/config`);
}

// =============================================================================
// Utility
// =============================================================================

export function isApiAvailable(): Promise<boolean> {
    return fetch(`${API_BASE}/health`)
        .then((r) => r.ok)
        .catch(() => false);
}
