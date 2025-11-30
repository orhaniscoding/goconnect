import { useState, useEffect, useCallback } from "react";
import * as api from "./api";

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
