import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAppStore } from '../store';
import { act } from 'react';

// Mock localStorage to verify persistence
const localStorageMock = (function () {
    let store: Record<string, string> = {};
    return {
        getItem: vi.fn((key: string) => store[key] || null),
        setItem: vi.fn((key: string, value: string) => {
            store[key] = value.toString();
        }),
        removeItem: vi.fn((key: string) => {
            delete store[key];
        }),
        clear: vi.fn(() => {
            store = {};
        }),
    };
})();

Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
});

describe('useAppStore', () => {
    beforeEach(() => {
        // Reset store before each test
        act(() => {
            useAppStore.setState({
                user: null,
                isLoggedIn: false,
                servers: [],
                networks: {},
                selectedServerId: null,
                selectedNetworkId: null,
            });
        });
        localStorageMock.clear();
    });

    it('should handle login and logout', () => {
        const user = { deviceId: 'dev1', username: 'tester' };

        act(() => useAppStore.getState().login(user));
        expect(useAppStore.getState().isLoggedIn).toBe(true);
        expect(useAppStore.getState().user).toEqual(user);

        act(() => useAppStore.getState().logout());
        expect(useAppStore.getState().isLoggedIn).toBe(false);
        expect(useAppStore.getState().user).toBeNull();
    });

    it('should add and remove servers', () => {
        const server = {
            id: 's1',
            name: 'Test Server',
            icon: 'T',
            isOwner: true,
            memberCount: 1
        };

        act(() => useAppStore.getState().addServer(server));
        expect(useAppStore.getState().servers).toHaveLength(1);
        expect(useAppStore.getState().servers[0]).toEqual(server);
        // Should initialize empty networks array
        expect(useAppStore.getState().networks['s1']).toEqual([]);

        act(() => useAppStore.getState().removeServer('s1'));
        expect(useAppStore.getState().servers).toHaveLength(0);
        expect(useAppStore.getState().networks['s1']).toBeUndefined();
    });

    it('should select server and network', () => {
        act(() => {
            useAppStore.getState().selectServer('s1');
        });
        expect(useAppStore.getState().selectedServerId).toBe('s1');

        act(() => {
            useAppStore.getState().selectNetwork('n1');
        });
        expect(useAppStore.getState().selectedNetworkId).toBe('n1');

        // Selecting a new server should reset network selection
        act(() => {
            useAppStore.getState().selectServer('s2');
        });
        expect(useAppStore.getState().selectedServerId).toBe('s2');
        expect(useAppStore.getState().selectedNetworkId).toBeNull();
    });
});
