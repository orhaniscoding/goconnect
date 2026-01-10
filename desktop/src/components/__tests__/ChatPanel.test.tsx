import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ChatPanel from '../ChatPanel';
import { tauriApi, ChatMessage } from '../../lib/tauri-api';

vi.mock('../../lib/tauri-api', () => ({
    tauriApi: {
        getMessages: vi.fn(),
        sendMessage: vi.fn(),
    }
}));

vi.mock('../Toast', () => ({
    useToast: () => ({
        success: vi.fn(),
        error: vi.fn(),
    }),
}));

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();

const mockMessages: ChatMessage[] = [
    { id: '1', peer_id: 'peer1', peer_name: 'Peer One', content: 'Hello Public', timestamp: new Date().toISOString(), is_self: false },
    { id: '2', peer_id: 'peer2', peer_name: 'Peer Two', content: 'Hello Private', timestamp: new Date().toISOString(), is_self: false },
    { id: '3', peer_id: 'self', peer_name: 'Me', content: 'My Msg', timestamp: new Date().toISOString(), is_self: true },
];

describe('ChatPanel', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        (tauriApi.getMessages as any).mockResolvedValue(mockMessages);
    });

    it('renders global chat', async () => {
        render(<ChatPanel networkId="net1" />);
        await waitFor(() => {
            expect(screen.getByText('Hello Public')).toBeInTheDocument();
            expect(screen.getByText('Hello Private')).toBeInTheDocument();
        });
        expect(screen.queryByText('Chat with')).not.toBeInTheDocument();
    });

    it('sends message in global chat', async () => {
        render(<ChatPanel networkId="net1" />);

        fireEvent.change(screen.getByPlaceholderText('Type a message...'), { target: { value: 'New Msg' } });
        fireEvent.click(screen.getByText('Send'));

        await waitFor(() => {
            expect(tauriApi.sendMessage).toHaveBeenCalledWith('net1', 'New Msg');
        });
    });

    it('renders private chat and filters messages', async () => {
        render(<ChatPanel networkId="net1" recipientId="peer2" recipientName="Peer Two" />);

        await waitFor(() => {
            expect(screen.getByText('Chat with')).toBeInTheDocument();
            expect(screen.getByText('Peer Two')).toBeInTheDocument();

            // Should show peer2 message
            expect(screen.getByText('Hello Private')).toBeInTheDocument();

            // Should NOT show peer1 message (filtered out)
            expect(screen.queryByText('Hello Public')).not.toBeInTheDocument();
        });
    });

    it('shows warning when sending in private chat', async () => {
        render(<ChatPanel networkId="net1" recipientId="peer2" />);

        fireEvent.change(screen.getByPlaceholderText('Type a message (Mock)'), { target: { value: 'Private Msg' } });
        fireEvent.click(screen.getByText('Send'));

        await waitFor(() => {
            expect(tauriApi.sendMessage).not.toHaveBeenCalled();
        });
    });
});
