import { useState, useEffect, useCallback } from "react";
import * as api from "./api";
import DaemonApi, { DaemonStatus, Settings, PeerInfo, TransferInfo, TransferStats, ChatMessage } from "./daemon";

// =============================================================================
// Connection Status Hook
// =============================================================================

export function useApiStatus() {
    const [isOnline, setIsOnline] = useState(false);
    const [checking, setChecking] = useState(true);

    useEffect(() => {
        const check = async () => {
            setChecking(true);
            const available = await api.isApiAvailable();
            setIsOnline(available);
            setChecking(false);
        };

        check();
        const interval = setInterval(check, 30000); // Check every 30s
        return () => clearInterval(interval);
    }, []);

    return { isOnline, checking };
}

// =============================================================================
// Servers Hook
// =============================================================================

export function useServers() {
    const [servers, setServers] = useState<api.Server[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchServers = useCallback(async () => {
        setLoading(true);
        setError(null);
        const result = await api.getMyServers();
        if (result.error) {
            setError(result.error);
        } else {
            setServers(result.data || []);
        }
        setLoading(false);
    }, []);

    const createServer = useCallback(async (name: string, icon?: string) => {
        const result = await api.createServer({ name, icon });
        if (result.data) {
            setServers((prev) => [...prev, result.data!]);
        }
        return result;
    }, []);

    const deleteServer = useCallback(async (serverId: string) => {
        const result = await api.deleteServer(serverId);
        if (!result.error) {
            setServers((prev) => prev.filter((s) => s.id !== serverId));
        }
        return result;
    }, []);

    const joinByCode = useCallback(async (code: string) => {
        const result = await api.joinServerByCode(code);
        if (result.data) {
            setServers((prev) => [...prev, result.data!]);
        }
        return result;
    }, []);

    const leaveServer = useCallback(async (serverId: string) => {
        const result = await api.leaveServer(serverId);
        if (!result.error) {
            setServers((prev) => prev.filter((s) => s.id !== serverId));
        }
        return result;
    }, []);

    useEffect(() => {
        fetchServers();
    }, [fetchServers]);

    return {
        servers,
        loading,
        error,
        refresh: fetchServers,
        createServer,
        deleteServer,
        joinByCode,
        leaveServer,
    };
}

// =============================================================================
// Networks Hook
// =============================================================================

export function useNetworks(serverId: string | null) {
    const [networks, setNetworks] = useState<api.Network[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchNetworks = useCallback(async () => {
        if (!serverId) {
            setNetworks([]);
            return;
        }
        setLoading(true);
        setError(null);
        const result = await api.listNetworks(serverId);
        if (result.error) {
            setError(result.error);
        } else {
            setNetworks(result.data || []);
        }
        setLoading(false);
    }, [serverId]);

    const createNetwork = useCallback(
        async (name: string) => {
            if (!serverId) return { error: "No server selected" };
            const result = await api.createNetwork(serverId, { name });
            if (result.data) {
                setNetworks((prev) => [...prev, result.data!]);
            }
            return result;
        },
        [serverId]
    );

    const deleteNetwork = useCallback(
        async (networkId: string) => {
            if (!serverId) return { error: "No server selected" };
            const result = await api.deleteNetwork(serverId, networkId);
            if (!result.error) {
                setNetworks((prev) => prev.filter((n) => n.id !== networkId));
            }
            return result;
        },
        [serverId]
    );

    useEffect(() => {
        fetchNetworks();
    }, [fetchNetworks]);

    return {
        networks,
        loading,
        error,
        refresh: fetchNetworks,
        createNetwork,
        deleteNetwork,
    };
}

// =============================================================================
// Network Clients Hook
// =============================================================================

export function useNetworkClients(serverId: string | null, networkId: string | null) {
    const [clients, setClients] = useState<api.Client[]>([]);
    const [loading, setLoading] = useState(false);

    const fetchClients = useCallback(async () => {
        if (!serverId || !networkId) {
            setClients([]);
            return;
        }
        setLoading(true);
        const result = await api.listNetworkClients(serverId, networkId);
        if (result.data) {
            setClients(result.data);
        }
        setLoading(false);
    }, [serverId, networkId]);

    useEffect(() => {
        fetchClients();
        // Refresh clients every 10 seconds when connected
        const interval = setInterval(fetchClients, 10000);
        return () => clearInterval(interval);
    }, [fetchClients]);

    return { clients, loading, refresh: fetchClients };
}

// =============================================================================
// Server Invite Hook
// =============================================================================

export function useServerInvite(serverId: string | null) {
    const [invite, setInvite] = useState<api.ServerInvite | null>(null);
    const [loading, setLoading] = useState(false);

    const createInvite = useCallback(async () => {
        if (!serverId) return;
        setLoading(true);
        const result = await api.createServerInvite(serverId, {
            maxUses: 0, // Unlimited
            expiresIn: 604800, // 7 days
        });
        if (result.data) {
            setInvite(result.data);
        }
        setLoading(false);
        return result;
    }, [serverId]);

    return { invite, loading, createInvite };
}

// =============================================================================
// Daemon Status Hook
// =============================================================================

export function useDaemonStatus(pollingInterval = 5000) {
    const [status, setStatus] = useState<DaemonStatus | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchStatus = useCallback(async () => {
        try {
            const isRunning = await DaemonApi.isRunning();
            setIsConnected(isRunning);

            if (isRunning) {
                const daemonStatus = await DaemonApi.getStatus();
                setStatus(daemonStatus);
                setError(null);
            } else {
                setStatus(null);
                setError("Daemon not running");
            }
        } catch (e) {
            setIsConnected(false);
            setError(e instanceof Error ? e.message : "Unknown error");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchStatus();
        const interval = setInterval(fetchStatus, pollingInterval);
        return () => clearInterval(interval);
    }, [fetchStatus, pollingInterval]);

    return { status, isConnected, loading, error, refresh: fetchStatus };
}

// =============================================================================
// Daemon Peers Hook
// =============================================================================

export function useDaemonPeers(pollingInterval = 10000) {
    const [peers, setPeers] = useState<PeerInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchPeers = useCallback(async () => {
        try {
            const peerList = await DaemonApi.getPeers();
            setPeers(peerList);
            setError(null);
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to fetch peers");
        } finally {
            setLoading(false);
        }
    }, []);

    const kickPeer = useCallback(async (networkId: string, peerId: string) => {
        try {
            await DaemonApi.kickPeer(networkId, peerId);
            await fetchPeers(); // Refresh
            return { success: true };
        } catch (e) {
            return { error: e instanceof Error ? e.message : "Failed to kick peer" };
        }
    }, [fetchPeers]);

    const banPeer = useCallback(async (networkId: string, peerId: string, reason?: string) => {
        try {
            await DaemonApi.banPeer(networkId, peerId, reason);
            await fetchPeers();
            return { success: true };
        } catch (e) {
            return { error: e instanceof Error ? e.message : "Failed to ban peer" };
        }
    }, [fetchPeers]);

    const unbanPeer = useCallback(async (networkId: string, peerId: string) => {
        try {
            await DaemonApi.unbanPeer(networkId, peerId);
            await fetchPeers();
            return { success: true };
        } catch (e) {
            return { error: e instanceof Error ? e.message : "Failed to unban peer" };
        }
    }, [fetchPeers]);

    useEffect(() => {
        fetchPeers();
        const interval = setInterval(fetchPeers, pollingInterval);
        return () => clearInterval(interval);
    }, [fetchPeers, pollingInterval]);

    return { peers, loading, error, refresh: fetchPeers, kickPeer, banPeer, unbanPeer };
}

// =============================================================================
// Daemon Settings Hook
// =============================================================================

export function useDaemonSettings() {
    const [settings, setSettings] = useState<Settings | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchSettings = useCallback(async () => {
        try {
            const s = await DaemonApi.getSettings();
            setSettings(s);
            setError(null);
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to fetch settings");
        } finally {
            setLoading(false);
        }
    }, []);

    const updateSettings = useCallback(async (newSettings: Partial<Settings>) => {
        setSaving(true);
        try {
            const updated = await DaemonApi.updateSettings(newSettings);
            setSettings(updated);
            setError(null);
            return { success: true };
        } catch (e) {
            const msg = e instanceof Error ? e.message : "Failed to update settings";
            setError(msg);
            return { error: msg };
        } finally {
            setSaving(false);
        }
    }, []);

    const resetSettings = useCallback(async () => {
        setSaving(true);
        try {
            const defaults = await DaemonApi.resetSettings();
            setSettings(defaults);
            setError(null);
            return { success: true };
        } catch (e) {
            const msg = e instanceof Error ? e.message : "Failed to reset settings";
            setError(msg);
            return { error: msg };
        } finally {
            setSaving(false);
        }
    }, []);

    useEffect(() => {
        fetchSettings();
    }, [fetchSettings]);

    return { settings, loading, saving, error, updateSettings, resetSettings, refresh: fetchSettings };
}

// =============================================================================
// Daemon Transfers Hook
// =============================================================================

export function useDaemonTransfers(pollingInterval = 2000) {
    const [transfers, setTransfers] = useState<TransferInfo[]>([]);
    const [stats, setStats] = useState<TransferStats | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchTransfers = useCallback(async () => {
        try {
            const [transferList, transferStats] = await Promise.all([
                DaemonApi.listTransfers(),
                DaemonApi.getTransferStats(),
            ]);
            setTransfers(transferList);
            setStats(transferStats);
            setError(null);
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to fetch transfers");
        } finally {
            setLoading(false);
        }
    }, []);

    const cancelTransfer = useCallback(async (transferId: string) => {
        try {
            await DaemonApi.cancelTransfer(transferId);
            await fetchTransfers();
            return { success: true };
        } catch (e) {
            return { error: e instanceof Error ? e.message : "Failed to cancel transfer" };
        }
    }, [fetchTransfers]);

    const rejectTransfer = useCallback(async (transferId: string) => {
        try {
            await DaemonApi.rejectTransfer(transferId);
            await fetchTransfers();
            return { success: true };
        } catch (e) {
            return { error: e instanceof Error ? e.message : "Failed to reject transfer" };
        }
    }, [fetchTransfers]);

    useEffect(() => {
        fetchTransfers();
        const interval = setInterval(fetchTransfers, pollingInterval);
        return () => clearInterval(interval);
    }, [fetchTransfers, pollingInterval]);

    return { transfers, stats, loading, error, refresh: fetchTransfers, cancelTransfer, rejectTransfer };
}

// =============================================================================
// Daemon Chat Hook
// =============================================================================

export function useDaemonChat(peerId: string | null) {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [loading, setLoading] = useState(true);
    const [sending, setSending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchMessages = useCallback(async (limit = 50, before?: string) => {
        if (!peerId) {
            setMessages([]);
            setLoading(false);
            return;
        }

        try {
            const msgs = await DaemonApi.getMessages(peerId, limit, before);
            if (before) {
                // Prepend older messages
                setMessages(prev => [...msgs, ...prev]);
            } else {
                setMessages(msgs);
            }
            setError(null);
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to fetch messages");
        } finally {
            setLoading(false);
        }
    }, [peerId]);

    const sendMessage = useCallback(async (content: string) => {
        if (!peerId || !content.trim()) return { error: "Invalid message" };

        setSending(true);
        try {
            await DaemonApi.sendMessage(peerId, content);
            // Optimistically add message
            const newMsg: ChatMessage = {
                id: Date.now().toString(),
                peer_id: peerId,
                content,
                timestamp: new Date().toISOString(),
                is_self: true,
            };
            setMessages(prev => [...prev, newMsg]);
            setError(null);
            return { success: true };
        } catch (e) {
            const msg = e instanceof Error ? e.message : "Failed to send message";
            setError(msg);
            return { error: msg };
        } finally {
            setSending(false);
        }
    }, [peerId]);

    const loadOlderMessages = useCallback(async () => {
        if (messages.length === 0) return;
        const oldestId = messages[0]?.id;
        await fetchMessages(50, oldestId);
    }, [messages, fetchMessages]);

    useEffect(() => {
        fetchMessages();
    }, [fetchMessages]);

    return { messages, loading, sending, error, sendMessage, loadOlderMessages, refresh: fetchMessages };
}
