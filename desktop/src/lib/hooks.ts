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
// Tenants Hook
// =============================================================================

export function useTenants() {
    const [tenants, setTenants] = useState<api.Tenant[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchTenants = useCallback(async () => {
        setLoading(true);
        setError(null);
        const result = await api.getMyTenants();
        if (result.error) {
            setError(result.error);
        } else {
            setTenants(result.data || []);
        }
        setLoading(false);
    }, []);

    const createTenant = useCallback(async (name: string, icon?: string) => {
        const result = await api.createTenant({ name, icon });
        if (result.data) {
            setTenants((prev) => [...prev, result.data!]);
        }
        return result;
    }, []);

    const deleteTenant = useCallback(async (tenantId: string) => {
        const result = await api.deleteTenant(tenantId);
        if (!result.error) {
            setTenants((prev) => prev.filter((t) => t.id !== tenantId));
        }
        return result;
    }, []);

    const joinByCode = useCallback(async (code: string) => {
        const result = await api.joinTenantByCode(code);
        if (result.data) {
            setTenants((prev) => [...prev, result.data!]);
        }
        return result;
    }, []);

    const leaveTenant = useCallback(async (tenantId: string) => {
        const result = await api.leaveTenant(tenantId);
        if (!result.error) {
            setTenants((prev) => prev.filter((t) => t.id !== tenantId));
        }
        return result;
    }, []);

    useEffect(() => {
        fetchTenants();
    }, [fetchTenants]);

    return {
        tenants,
        loading,
        error,
        refresh: fetchTenants,
        createTenant,
        deleteTenant,
        joinByCode,
        leaveTenant,
    };
}

// =============================================================================
// Networks Hook
// =============================================================================

export function useNetworks(tenantId: string | null) {
    const [networks, setNetworks] = useState<api.Network[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchNetworks = useCallback(async () => {
        if (!tenantId) {
            setNetworks([]);
            return;
        }
        setLoading(true);
        setError(null);
        const result = await api.listNetworks(tenantId);
        if (result.error) {
            setError(result.error);
        } else {
            setNetworks(result.data || []);
        }
        setLoading(false);
    }, [tenantId]);

    const createNetwork = useCallback(
        async (name: string) => {
            if (!tenantId) return { error: "No tenant selected" };
            const result = await api.createNetwork(tenantId, { name });
            if (result.data) {
                setNetworks((prev) => [...prev, result.data!]);
            }
            return result;
        },
        [tenantId]
    );

    const deleteNetwork = useCallback(
        async (networkId: string) => {
            if (!tenantId) return { error: "No tenant selected" };
            const result = await api.deleteNetwork(tenantId, networkId);
            if (!result.error) {
                setNetworks((prev) => prev.filter((n) => n.id !== networkId));
            }
            return result;
        },
        [tenantId]
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
// Network Members Hook
// =============================================================================

export function useNetworkMembers(tenantId: string | null, networkId: string | null) {
    const [members, setMembers] = useState<api.NetworkMember[]>([]);
    const [loading, setLoading] = useState(false);

    const fetchMembers = useCallback(async () => {
        if (!tenantId || !networkId) {
            setMembers([]);
            return;
        }
        setLoading(true);
        const result = await api.listNetworkMembers(tenantId, networkId);
        if (result.data) {
            setMembers(result.data);
        }
        setLoading(false);
    }, [tenantId, networkId]);

    useEffect(() => {
        fetchMembers();
        // Refresh members every 10 seconds when connected
        const interval = setInterval(fetchMembers, 10000);
        return () => clearInterval(interval);
    }, [fetchMembers]);

    return { members, loading, refresh: fetchMembers };
}

// =============================================================================
// Invite Hook
// =============================================================================

export function useTenantInvite(tenantId: string | null) {
    const [invite, setInvite] = useState<api.TenantInvite | null>(null);
    const [loading, setLoading] = useState(false);

    const createInvite = useCallback(async () => {
        if (!tenantId) return;
        setLoading(true);
        const result = await api.createTenantInvite(tenantId, {
            max_uses: 0, // Unlimited
            expires_in: 604800, // 7 days
        });
        if (result.data) {
            setInvite(result.data);
        }
        setLoading(false);
        return result;
    }, [tenantId]);

    return { invite, loading, createInvite };
}
