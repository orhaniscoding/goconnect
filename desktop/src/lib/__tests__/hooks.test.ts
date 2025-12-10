import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useServers, useNetworks } from '../hooks';
import * as api from '../api';

// Mock API
vi.mock('../api', () => ({
    getMyServers: vi.fn(),
    createServer: vi.fn(),
    deleteServer: vi.fn(),
    joinServerByCode: vi.fn(),
    leaveServer: vi.fn(),

    listNetworks: vi.fn(),
    createNetwork: vi.fn(),
    deleteNetwork: vi.fn(),
}));

describe('useServers Hook', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should fetch servers on mount', async () => {
        const mockServers = [{ id: 's1', name: 'Server 1', isOwner: true, memberCount: 1 }];
        (api.getMyServers as any).mockResolvedValue({ data: mockServers });

        const { result } = renderHook(() => useServers());

        expect(result.current.loading).toBe(true);

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.servers).toEqual(mockServers);
        expect(result.current.error).toBeNull();
    });

    it('should handle fetch errors', async () => {
        (api.getMyServers as any).mockResolvedValue({ error: 'Failed to fetch' });

        const { result } = renderHook(() => useServers());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.servers).toEqual([]);
        expect(result.current.error).toBe('Failed to fetch');
    });

    it('should create a server', async () => {
        (api.getMyServers as any).mockResolvedValue({ data: [] });
        const { result } = renderHook(() => useServers());

        // Wait for initial load
        await waitFor(() => expect(result.current.loading).toBe(false));

        const newServer = { id: 's2', name: 'New Server', isOwner: true, memberCount: 1 };
        (api.createServer as any).mockResolvedValue({ data: newServer });

        await act(async () => {
            await result.current.createServer('New Server');
        });

        // State update might be async even after await, but usually act() handles it. 
        // If it still fails, we use waitFor
        await waitFor(() => {
            expect(result.current.servers).toContainEqual(newServer);
        });
    });
});

describe('useNetworks Hook', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should not fetch if serverId is null', async () => {
        const { result } = renderHook(() => useNetworks(null));
        expect(api.listNetworks).not.toHaveBeenCalled();
        expect(result.current.networks).toEqual([]);
    });

    it('should fetch networks when serverId provided', async () => {
        const mockNetworks = [{ id: 'n1', name: 'Net 1' }];
        (api.listNetworks as any).mockResolvedValue({ data: mockNetworks });

        const { result } = renderHook(() => useNetworks('s1'));

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(api.listNetworks).toHaveBeenCalledWith('s1');
        expect(result.current.networks).toEqual(mockNetworks);
    });
});
