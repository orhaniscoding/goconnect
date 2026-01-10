import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { vi, describe, beforeEach, test, expect } from 'vitest';
import VoiceChat from '../VoiceChat';
import { PeerInfo } from '../../lib/tauri-api';

// Mock tauri-api
const mockStore = {
    signals: [] as any[]
};

vi.mock('../../lib/tauri-api', () => ({
    tauriApi: {
        getVoiceSignals: vi.fn().mockImplementation(() => Promise.resolve([...mockStore.signals])),
        sendVoiceSignal: vi.fn().mockImplementation((signal) => {
            mockStore.signals.push(signal);
            return Promise.resolve();
        })
    }
}));

describe('VoiceChat Component', () => {
    const mockSelfPeer: PeerInfo = {
        id: 'peer-1',
        name: 'Test Peer',
        display_name: 'Test Peer',
        virtual_ip: '10.0.0.2',
        connected: true,
        is_relay: false,
        latency_ms: 10,
        is_self: true
    };

    const mockConnectedPeers: PeerInfo[] = [
        mockSelfPeer,
        {
            id: 'peer-2',
            name: 'Other Peer',
            display_name: 'Other Peer',
            virtual_ip: '10.0.0.3',
            connected: true,
            is_relay: false,
            latency_ms: 15,
            is_self: false
        }
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        mockStore.signals = [];
    });

    test('renders initial state correctly', () => {
        render(<VoiceChat networkId="net-1" selfPeer={mockSelfPeer} connectedPeers={mockConnectedPeers} />);
        expect(screen.getByRole('heading', { name: /Voice Channel/i })).toBeInTheDocument();
        expect(screen.getByText('Join Voice')).toBeInTheDocument();
        expect(screen.getByText('No activity')).toBeInTheDocument();
    });

    test('can join and leave voice channel', () => {
        render(<VoiceChat networkId="net-1" selfPeer={mockSelfPeer} connectedPeers={mockConnectedPeers} />);

        // Join
        fireEvent.click(screen.getByText('Join Voice'));
        expect(screen.getByText('Leave Voice')).toBeInTheDocument();
        expect(screen.getByText(/Joined Voice Channel/)).toBeInTheDocument();

        // Leave
        fireEvent.click(screen.getByText('Leave Voice'));
        expect(screen.getByText('Join Voice')).toBeInTheDocument();
        expect(screen.getByText(/Left Voice Channel/)).toBeInTheDocument();
    });

    test('can ping/send signal when in call', async () => {
        render(<VoiceChat networkId="net-1" selfPeer={mockSelfPeer} connectedPeers={mockConnectedPeers} />);

        fireEvent.click(screen.getByText('Join Voice'));

        const pingBtn = screen.getByText('Ping');
        fireEvent.click(pingBtn);

        await waitFor(() => {
            expect(screen.getByText(/Signal dispatched to backend/)).toBeInTheDocument();
        });
    });
});

