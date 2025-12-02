// =============================================================================
// GoConnect API Client
// =============================================================================
// Terminology: Server (tenant) > Network > Client (device/member)

import { platform } from '@tauri-apps/plugin-os';

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

// User-friendly error messages
const ERROR_MESSAGES: Record<string, string> = {
    "Failed to fetch": "Cannot connect to server. Please check your internet connection.",
    "NetworkError": "Network error. Please check your connection.",
    "TypeError: Failed to fetch": "Server is not reachable. Make sure GoConnect server is running.",
    "401": "Authentication failed. Please log in again.",
    "403": "You don't have permission to do this.",
    "404": "The requested resource was not found.",
    "409": "This already exists. Try a different name.",
    "429": "Too many requests. Please wait a moment.",
    "500": "Server error. Please try again later.",
    "502": "Server is temporarily unavailable.",
    "503": "Service unavailable. Please try again later.",
};

function getUserFriendlyError(error: string | Error, statusCode?: number): string {
    const errorStr = error instanceof Error ? error.message : error;
    
    // Check for status code based messages
    if (statusCode && ERROR_MESSAGES[statusCode.toString()]) {
        return ERROR_MESSAGES[statusCode.toString()];
    }
    
    // Check for known error patterns
    for (const [pattern, message] of Object.entries(ERROR_MESSAGES)) {
        if (errorStr.includes(pattern)) {
            return message;
        }
    }
    
    // Return original if no match
    return errorStr || "Something went wrong. Please try again.";
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

        // Handle non-JSON responses
        const contentType = response.headers.get("content-type");
        if (!contentType || !contentType.includes("application/json")) {
            if (!response.ok) {
                return { error: getUserFriendlyError("", response.status) };
            }
            return { data: undefined as unknown as T };
        }

        const data = await response.json();

        if (!response.ok) {
            const errorMsg = data.message || data.error || "";
            return { error: getUserFriendlyError(errorMsg, response.status) };
        }

        return { data };
    } catch (error) {
        return { error: getUserFriendlyError(error instanceof Error ? error : String(error)) };
    }
}

// =============================================================================
// Auth / Device
// =============================================================================

export async function registerDevice(username: string) {
    // Generate device ID if not exists
    let deviceId = localStorage.getItem("gc_device_id");
    if (!deviceId) {
        deviceId = "dev_" + Math.random().toString(36).substr(2, 9);
        localStorage.setItem("gc_device_id", deviceId);
    }
    localStorage.setItem("gc_username", username);
    
    const result = await apiFetch<{ device_id: string; username: string }>("/devices/register", {
        method: "POST",
        body: JSON.stringify({
            device_id: deviceId,
            username,
            platform: platform(),
            hostname: "Desktop",
        }),
    });
    
    if (result.data) {
        return { 
            data: { 
                id: result.data.device_id,
                deviceId: result.data.device_id, 
                username: result.data.username 
            } as User
        };
    }
    return { error: result.error };
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

// =============================================================================
// Local Daemon Client (P2P Control)
// =============================================================================

const DAEMON_BASE = "http://localhost:12345";

export interface DaemonStatus {
    device: {
        registered: boolean;
        public_key: string;
        device_id: string;
    };
    peers: Record<string, PeerStatus>;
    wireguard: {
        public_key: string;
        peers_configured: number;
    };
}

export interface PeerStatus {
    endpoint: string;
    last_handshake: string; // Time string or "Never"
    allowed_ips: string[];
    latency_ms?: number;
    connection_state?: string; // "checking" | "connected" | "failed" | "disconnected" | "closed"
    connected?: boolean;
}

export async function getDaemonStatus() {
    try {
        const response = await fetch(`${DAEMON_BASE}/status`);
        if (!response.ok) throw new Error("Daemon unreachable");
        return await response.json() as DaemonStatus;
    } catch (error) {
        console.error("Daemon status check failed:", error);
        return null;
    }
}

export async function manualP2PConnect(peerId: string) {
    try {
        const response = await fetch(`${DAEMON_BASE}/p2p/connect`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ peer_id: peerId }),
        });
        if (!response.ok) {
            const err = await response.text();
            throw new Error(err || "Connection failed");
        }
        return await response.json();
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

export interface DaemonConfig {
    server: {
        url: string;
    };
    daemon: {
        listen_addr: string;
        local_port: number;
        health_check_interval: number;
    };
    wireguard: {
        interface_name: string;
    };
    identity: {
        path: string;
    };
    p2p: {
        enabled: boolean;
        stun_server: string;
    };
}

export async function getDaemonConfig() {
    try {
        const response = await fetch(`${DAEMON_BASE}/config`);
        if (!response.ok) throw new Error("Failed to fetch config");
        return await response.json() as DaemonConfig;
    } catch (error) {
        console.error("Get config failed:", error);
        return null;
    }
}

export async function updateDaemonConfig(config: { p2p_enabled?: boolean; stun_server?: string }) {
    try {
        const response = await fetch(`${DAEMON_BASE}/config`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(config),
        });
        if (!response.ok) throw new Error("Failed to update config");
        return await response.json();
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

export async function sendChatMessage(peerId: string, content: string) {
    try {
        const response = await fetch(`${DAEMON_BASE}/chat/send`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ peer_id: peerId, content }),
        });
        if (!response.ok) throw new Error("Failed to send message");
        return await response.json();
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

export interface FileTransferRequest {
    id: string;
    file_name: string;
    file_size: number;
}

export interface FileTransferSession {
    id: string;
    peer_id: string;
    file_path: string;
    file_name: string;
    file_size: number;
    sent_bytes: number;
    status: "pending" | "in_progress" | "completed" | "failed" | "cancelled";
    is_sender: boolean;
    error?: string;
}

export async function sendFileRequest(peerId: string, filePath: string) {
    try {
        const response = await fetch(`${DAEMON_BASE}/file/send`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ peer_id: peerId, file_path: filePath }),
        });
        if (!response.ok) throw new Error("Failed to send file request");
        return await response.json() as FileTransferSession;
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

export async function acceptFileRequest(request: FileTransferRequest, peerId: string, savePath: string) {
    try {
        const response = await fetch(`${DAEMON_BASE}/file/accept`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ request, peer_id: peerId, save_path: savePath }),
        });
        if (!response.ok) throw new Error("Failed to accept file");
        return await response.json();
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

export function subscribeToEvents(onMessage: (event: any) => void) {
    const eventSource = new EventSource(`${DAEMON_BASE}/events`);

    eventSource.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            onMessage(data);
        } catch (e) {
            console.error("Failed to parse SSE message:", e);
        }
    };

    eventSource.onerror = (err) => {
        console.error("EventSource failed:", err);
        // EventSource automatically retries, but we might want to handle specific errors
    };

    return () => {
        eventSource.close();
    };
}

export function isApiAvailable(): Promise<boolean> {
    return fetch(`${API_BASE}/health`)
        .then((r) => r.ok)
        .catch(() => false);
}
