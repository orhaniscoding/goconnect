// =============================================================================
// GoConnect Daemon API Client (via Tauri)
// =============================================================================
// Communicates with the local daemon via Tauri commands -> Rust gRPC client

import { invoke } from '@tauri-apps/api/core';

// =============================================================================
// TYPES (match Rust structs)
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
    status: 'pending' | 'active' | 'completed' | 'failed' | 'cancelled' | 'rejected';
    direction: 'upload' | 'download';
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
// API WRAPPER
// =============================================================================

export class DaemonApi {
    // =========================================================================
    // CONNECTION CHECK
    // =========================================================================

    /**
     * Check if the daemon is running and reachable
     */
    static async isRunning(): Promise<boolean> {
        try {
            return await invoke<boolean>('daemon_is_running');
        } catch {
            return false;
        }
    }

    // =========================================================================
    // DAEMON SERVICE
    // =========================================================================

    /**
     * Get the current daemon status
     */
    static async getStatus(): Promise<DaemonStatus> {
        return invoke<DaemonStatus>('daemon_get_status');
    }

    /**
     * Get daemon version information
     */
    static async getVersion(): Promise<VersionInfo> {
        return invoke<VersionInfo>('daemon_get_version');
    }

    // =========================================================================
    // NETWORK SERVICE
    // =========================================================================

    /**
     * Create a new network
     */
    static async createNetwork(name: string): Promise<NetworkInfo> {
        return invoke<NetworkInfo>('daemon_create_network', { name });
    }

    /**
     * Join an existing network via invite code
     */
    static async joinNetwork(inviteCode: string): Promise<NetworkInfo> {
        return invoke<NetworkInfo>('daemon_join_network', { inviteCode });
    }

    /**
     * List all networks the user is part of
     */
    static async listNetworks(): Promise<NetworkInfo[]> {
        return invoke<NetworkInfo[]>('daemon_list_networks');
    }

    /**
     * Leave a network
     */
    static async leaveNetwork(networkId: string): Promise<void> {
        return invoke<void>('daemon_leave_network', { networkId });
    }

    /**
     * Generate an invite code for a network
     */
    static async generateInvite(networkId: string): Promise<string> {
        return invoke<string>('daemon_generate_invite', { networkId });
    }

    // =========================================================================
    // PEER SERVICE
    // =========================================================================

    /**
     * Get list of peers in the current network
     */
    static async getPeers(): Promise<PeerInfo[]> {
        return invoke<PeerInfo[]>('daemon_get_peers');
    }

    /**
     * Kick a peer from a network (admin only)
     */
    static async kickPeer(networkId: string, peerId: string): Promise<void> {
        return invoke<void>('daemon_kick_peer', { networkId, peerId });
    }

    /**
     * Ban a peer from a network (admin only)
     */
    static async banPeer(networkId: string, peerId: string, reason?: string): Promise<void> {
        return invoke<void>('daemon_ban_peer', { networkId, peerId, reason: reason || '' });
    }

    /**
     * Unban a peer from a network (admin only)
     */
    static async unbanPeer(networkId: string, peerId: string): Promise<void> {
        return invoke<void>('daemon_unban_peer', { networkId, peerId });
    }

    // =========================================================================
    // SETTINGS SERVICE
    // =========================================================================

    /**
     * Get daemon settings
     */
    static async getSettings(): Promise<Settings> {
        return invoke<Settings>('daemon_get_settings');
    }

    /**
     * Update daemon settings
     */
    static async updateSettings(settings: Partial<Settings>): Promise<Settings> {
        return invoke<Settings>('daemon_update_settings', { settings });
    }

    /**
     * Reset daemon settings to defaults
     */
    static async resetSettings(): Promise<Settings> {
        return invoke<Settings>('daemon_reset_settings');
    }

    // =========================================================================
    // CHAT SERVICE
    // =========================================================================

    /**
     * Get chat messages with a peer
     */
    static async getMessages(peerId: string, limit?: number, before?: string): Promise<ChatMessage[]> {
        return invoke<ChatMessage[]>('daemon_get_messages', { peerId, limit, before });
    }

    /**
     * Send a chat message to a peer
     */
    static async sendMessage(peerId: string, content: string): Promise<void> {
        return invoke<void>('daemon_send_message', { peerId, content });
    }

    // =========================================================================
    // TRANSFER SERVICE
    // =========================================================================

    /**
     * List file transfers with optional filtering
     */
    static async listTransfers(status?: string, peerId?: string): Promise<TransferInfo[]> {
        return invoke<TransferInfo[]>('daemon_list_transfers', { status, peerId });
    }

    /**
     * Get transfer statistics
     */
    static async getTransferStats(): Promise<TransferStats> {
        return invoke<TransferStats>('daemon_get_transfer_stats');
    }

    /**
     * Cancel an active transfer
     */
    static async cancelTransfer(transferId: string): Promise<void> {
        return invoke<void>('daemon_cancel_transfer', { transferId });
    }

    /**
     * Reject an incoming transfer request
     */
    static async rejectTransfer(transferId: string): Promise<void> {
        return invoke<void>('daemon_reject_transfer', { transferId });
    }
}

// =============================================================================
// REACT HOOK HELPERS
// =============================================================================

/**
 * Poll daemon status at regular intervals
 * Usage: Import this in your React component and implement with useState/useEffect
 */
export function createDaemonStatusPoller(intervalMs: number = 5000) {
    let timer: ReturnType<typeof setInterval> | null = null;

    return {
        start(callback: (status: DaemonStatus) => void) {
            this.stop();
            const poll = async () => {
                try {
                    const status = await DaemonApi.getStatus();
                    callback(status);
                } catch (e) {
                    console.error('Failed to poll daemon status:', e);
                }
            };
            poll(); // Initial call
            timer = setInterval(poll, intervalMs);
        },
        stop() {
            if (timer) {
                clearInterval(timer);
                timer = null;
            }
        }
    };
}

// Default export for convenience
export default DaemonApi;
