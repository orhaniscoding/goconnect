import { invoke } from '@tauri-apps/api/core';

// =============================================================================
// Type Definitions (Must match Rust structs in commands.rs)
// =============================================================================

export interface DaemonStatus {
    connected: boolean;
    virtual_ip: string;
    active_peers: number;
    network_name: string;
}

export interface VersionInfo {
    version: string;
    build_date: string;
    commit: string;
    go_version: string;
    os: string;
    arch: string;
}

export interface NetworkInfo {
    id: string;
    name: string;
    invite_code: string;
}

export interface PeerInfo {
    id: string;
    name: string;
    display_name: string;
    virtual_ip: string;
    connected: boolean;
    is_relay: boolean;
    latency_ms: number;
    is_self: boolean;
}

export interface Settings {
    auto_connect: boolean;
    start_minimized: boolean;
    notifications_enabled: boolean;
    log_level: string;
}

export interface ChatMessage {
    id: string;
    peer_id: string;
    content: string;
    timestamp: string;
    is_self: boolean;
}

export interface TransferInfo {
    id: string;
    peer_id: string;
    file_name: string;
    file_size: number;
    transferred: number;
    status: string;
    direction: string;
    error?: string;
}

export interface TransferStats {
    total_uploads: number;
    total_downloads: number;
    active_transfers: number;
    completed_transfers: number;
    failed_transfers: number;
    total_bytes_sent: number;
    total_bytes_received: number;
}

// =============================================================================
// API Wrapper
// =============================================================================

export const tauriApi = {
    // Daemon
    getStatus: () => invoke<DaemonStatus>('daemon_get_status'),
    getVersion: () => invoke<VersionInfo>('daemon_get_version'),
    isRunning: () => invoke<boolean>('daemon_is_running'),

    // Networks
    createNetwork: (name: string) => invoke<NetworkInfo>('daemon_create_network', { name }),
    joinNetwork: (inviteCode: string) => invoke<NetworkInfo>('daemon_join_network', { inviteCode }), // Rust uses snake_case arg names usually, but tauri converts camelCase to snake_case automatically? Check commands.rs. 
    // Wait, in commands.rs: `invite_code: String`. Tauri automatically maps camelCase JS args to snake_case Rust args.
    listNetworks: () => invoke<NetworkInfo[]>('daemon_list_networks'),
    leaveNetwork: (networkId: string) => invoke<void>('daemon_leave_network', { networkId }),
    generateInvite: (networkId: string) => invoke<string>('daemon_generate_invite', { networkId }),

    // Peers
    getPeers: () => invoke<PeerInfo[]>('daemon_get_peers'),
    kickPeer: (networkId: string, peerId: string) => invoke<void>('daemon_kick_peer', { networkId, peerId }),
    banPeer: (networkId: string, peerId: string, reason: string) => invoke<void>('daemon_ban_peer', { networkId, peerId, reason }),
    unbanPeer: (networkId: string, peerId: string) => invoke<void>('daemon_unban_peer', { networkId, peerId }),

    // Settings
    getSettings: () => invoke<Settings>('daemon_get_settings'),
    updateSettings: (settings: Settings) => invoke<Settings>('daemon_update_settings', { settings }),
    resetSettings: () => invoke<Settings>('daemon_reset_settings'),

    // Chat
    getMessages: (networkId: string, limit?: number, before?: string) => invoke<ChatMessage[]>('daemon_get_messages', { networkId, limit, before }),
    sendMessage: (networkId: string, content: string) => invoke<void>('daemon_send_message', { networkId, content }),

    // Transfers
    listTransfers: (status?: string, peerId?: string) => invoke<TransferInfo[]>('daemon_list_transfers', { status, peerId }),
    async getTransferStats(): Promise<TransferStats> {
        return await invoke('daemon_get_transfer_stats');
    },

    // HTTP Fallbacks for Registration (RPC not available)
    async checkRegistration(): Promise<boolean> {
        try {
            const res = await fetch('http://localhost:34100/status');
            if (!res.ok) return false;
            const data = await res.json();
            return !!data.device?.registered;
        } catch (e) {
            console.error("Failed to check registration", e);
            return false;
        }
    },

    async register(token: string, name?: string): Promise<void> {
        const res = await fetch('http://localhost:34100/register', {
            method: 'POST',
            body: JSON.stringify({ token, name }),
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(txt || 'Registration failed');
        }
    },
    cancelTransfer: (transferId: string) => invoke<void>('daemon_cancel_transfer', { transferId }),
    rejectTransfer: (transferId: string) => invoke<void>('daemon_reject_transfer', { transferId }),
    sendFile: (peerId: string, filePath: string) => invoke<string>('daemon_send_file', { peerId, filePath }),
    acceptTransfer: (transferId: string, savePath: string) => invoke<void>('daemon_accept_transfer', { transferId, savePath }),
};
