import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { PeerList } from '../PeerList';
import * as api from '../../lib/api';

// Mock API
vi.mock('../../lib/api', () => ({
    getDaemonStatus: vi.fn(),
    manualP2PConnect: vi.fn(),
    sendChatMessage: vi.fn(),
    subscribeToEvents: vi.fn(() => () => { }),
    getDaemonConfig: vi.fn(),
}));

// Mock Child Components
vi.mock('../ChatWindow', () => ({
    ChatWindow: () => <div data-testid="chat-window">Chat Window</div>
}));

// Mock Toast hook
vi.mock('../Toast', () => ({
    useToast: () => ({
        success: vi.fn(),
        error: vi.fn(),
        warning: vi.fn(),
        info: vi.fn(),
    })
}));

describe('PeerList Component', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders loading state initially', async () => {
        (api.getDaemonStatus as any).mockResolvedValue(null);
        render(<PeerList />);
        expect(screen.getByText(/Waiting for local daemon/i)).toBeInTheDocument();
    });

    it('renders peers when status is available', async () => {
        (api.getDaemonStatus as any).mockResolvedValue({
            peers: {
                'peer-1': {
                    endpoint: '1.2.3.4:51820',
                    connected: true,
                    latency_ms: 15,
                    connection_state: 'connected',
                    allowed_ips: ['10.0.0.2/32']
                },
                'peer-2': {
                    connected: false,
                    connection_state: 'failed',
                    allowed_ips: []
                }
            }
        });

        render(<PeerList />);

        // Wait for loading to disappear
        await waitFor(() => {
            expect(screen.queryByText(/Waiting for local daemon/i)).not.toBeInTheDocument();
        });

        expect(screen.getByText('peer-1...')).toBeInTheDocument();
        expect(screen.getByText('15ms')).toBeInTheDocument();
        expect(screen.getByText('peer-2...')).toBeInTheDocument();
        expect(screen.getByText('failed')).toBeInTheDocument();
    });
});
