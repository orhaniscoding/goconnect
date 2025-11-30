import { create } from "zustand";
import { persist } from "zustand/middleware";

// =============================================================================
// Types
// =============================================================================

export interface User {
    deviceId: string;
    username: string;
}

export interface Server {
    id: string;
    name: string;
    icon: string;
    description?: string;
    isOwner: boolean;
    memberCount: number;
    unread?: number;
}

export interface Network {
    id: string;
    serverId: string;
    name: string;
    subnet: string;
    connected: boolean;
    myIp?: string;
    clients: Client[];
}

export interface Client {
    id: string;
    name: string;
    ip: string;
    status: "online" | "idle" | "offline";
    isHost: boolean;
}

export interface Channel {
    id: string;
    name: string;
    unread?: number;
}

// =============================================================================
// App Store
// =============================================================================

interface AppState {
    // Auth
    user: User | null;
    isLoggedIn: boolean;

    // Data
    servers: Server[];
    networks: Record<string, Network[]>;

    // Selection
    selectedServerId: string | null;
    selectedNetworkId: string | null;

    // Settings
    apiMode: boolean; // true = use real API, false = mock data

    // Actions
    login: (user: User) => void;
    logout: () => void;

    addServer: (server: Server) => void;
    removeServer: (serverId: string) => void;
    setServers: (servers: Server[]) => void;

    addNetwork: (serverId: string, network: Network) => void;
    removeNetwork: (serverId: string, networkId: string) => void;
    setNetworks: (serverId: string, networks: Network[]) => void;
    updateNetwork: (serverId: string, networkId: string, updates: Partial<Network>) => void;

    selectServer: (serverId: string | null) => void;
    selectNetwork: (networkId: string | null) => void;

    toggleApiMode: () => void;
}

export const useAppStore = create<AppState>()(
    persist(
        (set) => ({
            // Initial state
            user: null,
            isLoggedIn: false,
            servers: [],
            networks: {},
            selectedServerId: null,
            selectedNetworkId: null,
            apiMode: false, // Start with mock data

            // Auth actions
            login: (user) => set({ user, isLoggedIn: true }),
            logout: () => set({
                user: null,
                isLoggedIn: false,
                servers: [],
                networks: {},
                selectedServerId: null,
                selectedNetworkId: null,
            }),

            // Server actions
            addServer: (server) => set((state) => ({
                servers: [...state.servers, server],
                networks: { ...state.networks, [server.id]: [] },
            })),

            removeServer: (serverId) => set((state) => {
                const { [serverId]: _, ...remainingNetworks } = state.networks;
                return {
                    servers: state.servers.filter((s) => s.id !== serverId),
                    networks: remainingNetworks,
                    selectedServerId: state.selectedServerId === serverId ? null : state.selectedServerId,
                    selectedNetworkId: state.selectedServerId === serverId ? null : state.selectedNetworkId,
                };
            }),

            setServers: (servers) => set({ servers }),

            // Network actions
            addNetwork: (serverId, network) => set((state) => ({
                networks: {
                    ...state.networks,
                    [serverId]: [...(state.networks[serverId] || []), network],
                },
            })),

            removeNetwork: (serverId, networkId) => set((state) => ({
                networks: {
                    ...state.networks,
                    [serverId]: (state.networks[serverId] || []).filter((n) => n.id !== networkId),
                },
                selectedNetworkId: state.selectedNetworkId === networkId ? null : state.selectedNetworkId,
            })),

            setNetworks: (serverId, networks) => set((state) => ({
                networks: { ...state.networks, [serverId]: networks },
            })),

            updateNetwork: (serverId, networkId, updates) => set((state) => ({
                networks: {
                    ...state.networks,
                    [serverId]: (state.networks[serverId] || []).map((n) =>
                        n.id === networkId ? { ...n, ...updates } : n
                    ),
                },
            })),

            // Selection actions
            selectServer: (serverId) => set({
                selectedServerId: serverId,
                selectedNetworkId: null,
            }),

            selectNetwork: (networkId) => set({ selectedNetworkId: networkId }),

            // Settings
            toggleApiMode: () => set((state) => ({ apiMode: !state.apiMode })),
        }),
        {
            name: "goconnect-storage",
            partialize: (state) => ({
                user: state.user,
                isLoggedIn: state.isLoggedIn,
                apiMode: state.apiMode,
            }),
        }
    )
);

// =============================================================================
// Selectors
// =============================================================================

export const useSelectedServer = () => {
    const servers = useAppStore((s) => s.servers);
    const selectedId = useAppStore((s) => s.selectedServerId);
    return servers.find((s) => s.id === selectedId) || null;
};

export const useSelectedNetwork = () => {
    const networks = useAppStore((s) => s.networks);
    const selectedServerId = useAppStore((s) => s.selectedServerId);
    const selectedNetworkId = useAppStore((s) => s.selectedNetworkId);

    if (!selectedServerId || !selectedNetworkId) return null;
    return (networks[selectedServerId] || []).find((n) => n.id === selectedNetworkId) || null;
};

export const useCurrentNetworks = () => {
    const networks = useAppStore((s) => s.networks);
    const selectedServerId = useAppStore((s) => s.selectedServerId);

    if (!selectedServerId) return [];
    return networks[selectedServerId] || [];
};
