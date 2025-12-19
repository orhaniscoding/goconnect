import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import App from '../App';
import { tauriApi } from '../lib/tauri-api';

// Mock dependencies
vi.mock('../lib/tauri-api', () => ({
    tauriApi: {
        isRunning: vi.fn(),
        checkRegistration: vi.fn(),
        listNetworks: vi.fn(),
        getStatus: vi.fn(),
        getPeers: vi.fn(),
        createNetwork: vi.fn(),
        joinNetwork: vi.fn(),
        leaveNetwork: vi.fn(),
        generateInvite: vi.fn(),
    }
}));

vi.mock('@tauri-apps/plugin-deep-link', () => ({
    onOpenUrl: vi.fn().mockResolvedValue(undefined),
}));

// Mock Child Components to simplify integration test
vi.mock('../components/FileTransferPanel', () => ({ default: () => <div data-testid="file-transfer-panel">Files Panel</div> }));
vi.mock('../components/ChatPanel', () => ({ default: () => <div data-testid="chat-panel">Chat Panel</div> }));
vi.mock('../components/SettingsPanel', () => ({ default: () => <div data-testid="settings-panel">Settings Panel</div> }));
vi.mock('../components/VoiceChat', () => ({ default: () => <div data-testid="voice-chat">Voice Chat</div> }));
vi.mock('../components/Onboarding', () => ({ default: ({ onComplete }: any) => <button onClick={onComplete}>Complete Onboarding</button> }));
vi.mock('../components/Toast', () => ({
    useToast: () => ({ success: vi.fn(), error: vi.fn() }),
    Toaster: () => <div>Toaster</div>
}));

describe('App Integration', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        // Default happy path
        (tauriApi.isRunning as any).mockResolvedValue(true);
        (tauriApi.checkRegistration as any).mockResolvedValue(true);
        (tauriApi.listNetworks as any).mockResolvedValue([{ id: 'net1', name: 'Test Net', invite_code: '123' }]);
        (tauriApi.getStatus as any).mockResolvedValue({ network_name: 'Test Net' });
        (tauriApi.getPeers as any).mockResolvedValue([{ id: 'peer1', name: 'Peer 1', is_self: false }]);
    });

    it('renders loading/daemon not running state', async () => {
        (tauriApi.isRunning as any).mockResolvedValue(false);
        render(<App />);
        await waitFor(() => expect(screen.getByText('Daemon Not Running')).toBeInTheDocument());
    });

    it('renders onboarding if not registered', async () => {
        (tauriApi.checkRegistration as any).mockResolvedValue(false);
        render(<App />);
        await waitFor(() => expect(screen.getByText('Complete Onboarding')).toBeInTheDocument());
    });

    it('renders main dashboard when running and registered', async () => {
        render(<App />);
        await waitFor(() => expect(screen.getByText('Test Net'.substring(0, 2).toUpperCase())).toBeInTheDocument());
        expect(screen.getByText('Connected Peers')).toBeInTheDocument();
    });

    it('switches tabs', async () => {
        render(<App />);
        await waitFor(() => screen.getByText('Connected Peers'));

        fireEvent.click(screen.getByText('Chat: Test Net'));
        expect(screen.getByTestId('chat-panel')).toBeInTheDocument();

        fireEvent.click(screen.getByText('Files'));
        expect(screen.getByTestId('file-transfer-panel')).toBeInTheDocument();

        fireEvent.click(screen.getByTitle('Settings'));
        expect(screen.getByTestId('settings-panel')).toBeInTheDocument();

        fireEvent.click(screen.getByText('Voice'));
        expect(screen.getByTestId('voice-chat')).toBeInTheDocument();
    });

    it('opens create network modal', async () => {
        render(<App />);
        await waitFor(() => screen.getByTitle('Create Network'));

        fireEvent.click(screen.getByTitle('Create Network'));
        expect(screen.getByText('Create Network', { selector: 'h3' })).toBeInTheDocument();
    });

    it('invokes refresh on network selection', async () => {
        render(<App />);
        await waitFor(() => screen.getByTitle('Test Net'));

        // Mock getPeers for specific network
        (tauriApi.getPeers as any).mockResolvedValue([{ id: 'p2', name: 'P2' }]);

        // Click the network button
        const netBtn = screen.getByTitle('Test Net');
        fireEvent.click(netBtn);

        await waitFor(() => expect(tauriApi.getPeers).toHaveBeenCalled());
    });
});
