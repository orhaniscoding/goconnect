import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { vi, describe, beforeEach, test, expect } from 'vitest';
import FileTransferPanel from '../FileTransferPanel';
import { tauriApi, TransferInfo, TransferStats } from '../../lib/tauri-api';
import { save } from '@tauri-apps/plugin-dialog';

// Mock tauri-api
vi.mock('../../lib/tauri-api', () => ({
    tauriApi: {
        listTransfers: vi.fn(),
        getTransferStats: vi.fn(),
        acceptTransfer: vi.fn(),
        rejectTransfer: vi.fn(),
        cancelTransfer: vi.fn(),
    }
}));

// Mock dialog plugin
vi.mock('@tauri-apps/plugin-dialog', () => ({
    save: vi.fn(),
}));

// Mock Toast
vi.mock('../Toast', () => ({
    useToast: () => ({
        success: vi.fn(),
        error: vi.fn(),
        info: vi.fn(),
    }),
}));

describe('FileTransferPanel', () => {
    const mockTransfers: TransferInfo[] = [
        {
            id: 't1',
            peer_id: 'peer-1',
            file_name: 'test.txt',
            file_size: 1024,
            transferred: 512,
            status: 'pending',
            direction: 'download',
        },
        {
            id: 't2',
            peer_id: 'peer-2',
            file_name: 'image.png',
            file_size: 2048,
            transferred: 2048,
            status: 'completed',
            direction: 'upload',
        }
    ];

    const mockStats: TransferStats = {
        total_uploads: 1,
        total_downloads: 1,
        active_transfers: 1,
        completed_transfers: 1,
        failed_transfers: 0,
        total_bytes_sent: 2048,
        total_bytes_received: 512,
    };

    beforeEach(() => {
        vi.clearAllMocks();
        (tauriApi.listTransfers as any).mockResolvedValue([]);
        (tauriApi.getTransferStats as any).mockResolvedValue(mockStats);
    });

    test('renders initial empty state', async () => {
        render(<FileTransferPanel />);
        await waitFor(() => {
            expect(screen.getByText('No active or recent transfers.')).toBeInTheDocument();
        });
    });

    test('renders transfers and stats', async () => {
        (tauriApi.listTransfers as any).mockResolvedValue(mockTransfers);

        render(<FileTransferPanel />);

        await waitFor(() => {
            expect(screen.getByText('test.txt')).toBeInTheDocument();
            expect(screen.getByText('image.png')).toBeInTheDocument();
            // Stats
            expect(screen.getByText('DOWNLOAD')).toBeInTheDocument();
            expect(screen.getByText('UPLOAD')).toBeInTheDocument();
        });
    });

    test('handles accept transfer', async () => {
        (tauriApi.listTransfers as any).mockResolvedValue([mockTransfers[0]]);
        (save as any).mockResolvedValue('/path/to/save/test.txt');

        render(<FileTransferPanel />);
        await waitFor(() => screen.getByText('Accept'));

        fireEvent.click(screen.getByText('Accept'));

        await waitFor(() => {
            expect(save).toHaveBeenCalled();
            expect(tauriApi.acceptTransfer).toHaveBeenCalledWith('t1', '/path/to/save/test.txt');
        });
    });

    test('handles reject transfer', async () => {
        (tauriApi.listTransfers as any).mockResolvedValue([mockTransfers[0]]);

        render(<FileTransferPanel />);
        await waitFor(() => screen.getByText('Reject'));

        fireEvent.click(screen.getByText('Reject'));

        await waitFor(() => {
            expect(tauriApi.rejectTransfer).toHaveBeenCalledWith('t1');
        });
    });

    test('handles cancel transfer', async () => {
        const activeTransfer = { ...mockTransfers[0], status: 'active' };
        (tauriApi.listTransfers as any).mockResolvedValue([activeTransfer]);

        render(<FileTransferPanel />);
        await waitFor(() => screen.getByText('Cancel'));

        fireEvent.click(screen.getByText('Cancel'));

        await waitFor(() => {
            expect(tauriApi.cancelTransfer).toHaveBeenCalledWith('t1');
        });
    });
});
