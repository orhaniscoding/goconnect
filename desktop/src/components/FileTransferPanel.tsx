import { useState, useEffect } from 'react';
import { tauriApi, TransferInfo, TransferStats } from '../lib/tauri-api';
import { handleError } from '../lib/utils';
import { save } from '@tauri-apps/plugin-dialog';
import { useToast } from './Toast';

export default function FileTransferPanel() {
    const [transfers, setTransfers] = useState<TransferInfo[]>([]);
    const [stats, setStats] = useState<TransferStats | null>(null);
    const toast = useToast();

    useEffect(() => {
        loadTransfers();
        const interval = setInterval(loadTransfers, 2000);
        return () => clearInterval(interval);
    }, []);

    const loadTransfers = async () => {
        try {
            const list = await tauriApi.listTransfers();
            // Sort by status (active first) then reverse id (newest first)
            list.sort((a, b) => {
                const score = (s: string) => s === 'active' || s === 'pending' ? 2 : 1;
                if (score(a.status) !== score(b.status)) return score(b.status) - score(a.status);
                return b.id.localeCompare(a.id);
            });
            setTransfers(list);

            const s = await tauriApi.getTransferStats();
            setStats(s);
        } catch (e) {
            console.error("Failed to load transfers", e);
        }
    };

    const handleAccept = async (id: string) => {
        try {
            const transfer = transfers.find(t => t.id === id);
            const defaultName = transfer ? transfer.file_name : 'download';

            const path = await save({
                defaultPath: defaultName,
                title: 'Save File',
            });

            if (!path) return; // User cancelled

            await tauriApi.acceptTransfer(id, path);
            toast.success("Transfer accepted");
            loadTransfers();
        } catch (e) {
            handleError(e, "Failed to accept");
        }
    };

    const handleReject = async (id: string) => {
        try {
            await tauriApi.rejectTransfer(id);
            toast.info("Transfer rejected");
            loadTransfers();
        } catch (e) {
            handleError(e, "Failed to reject");
        }
    };

    const handleCancel = async (id: string) => {
        try {
            await tauriApi.cancelTransfer(id);
            toast.info("Transfer cancelled");
            loadTransfers();
        } catch (e) {
            handleError(e, "Failed to cancel");
        }
    };

    const formatBytes = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return (
        <div className="flex flex-col h-full bg-gc-dark-800 rounded-lg border border-gc-dark-600 overflow-hidden">
            {/* Stats Header */}
            {stats && (
                <div className="p-4 bg-gc-dark-900 border-b border-gc-dark-700 grid grid-cols-4 gap-4 text-center">
                    <div>
                        <div className="text-xs text-gray-500 uppercase">Active</div>
                        <div className="text-xl font-bold text-blue-400">{stats.active_transfers}</div>
                    </div>
                    <div>
                        <div className="text-xs text-gray-500 uppercase">Up</div>
                        <div className="text-xl font-bold text-green-400">{formatBytes(stats.total_bytes_sent)}</div>
                    </div>
                    <div>
                        <div className="text-xs text-gray-500 uppercase">Down</div>
                        <div className="text-xl font-bold text-purple-400">{formatBytes(stats.total_bytes_received)}</div>
                    </div>
                    <div>
                        <div className="text-xs text-gray-500 uppercase">Completed</div>
                        <div className="text-xl font-bold text-gray-400">{stats.completed_transfers}</div>
                    </div>
                </div>
            )}

            {/* List */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3" role="list" aria-label="Transfer List">
                {transfers.length === 0 ? (
                    <div className="text-center text-gray-500 mt-10">
                        No active or recent transfers.
                    </div>
                ) : (
                    transfers.map(t => (
                        <div key={t.id} role="listitem" className="bg-gc-dark-700 p-3 rounded border border-gc-dark-600 flex items-center justify-between">
                            <div className="flex-1 min-w-0 mr-4">
                                <div className="flex items-center gap-2 mb-1">
                                    <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${t.direction === 'upload' ? 'bg-blue-900 text-blue-200' : 'bg-green-900 text-green-200'
                                        }`} aria-label={`Direction: ${t.direction}`}>
                                        {t.direction.toUpperCase()}
                                    </span>
                                    <span className="font-medium truncate text-white" title={t.file_name}>{t.file_name}</span>
                                </div>
                                <div className="text-xs text-gray-400 flex justify-between">
                                    <span>Peer: {t.peer_id.substring(0, 8)}</span>
                                    <span>{formatBytes(t.transferred)} / {formatBytes(t.file_size)}</span>
                                </div>
                                {/* Progress Bar */}
                                <div className="w-full bg-gc-dark-900 h-1.5 mt-2 rounded-full overflow-hidden" role="progressbar" aria-valuenow={Math.min(100, (t.transferred / t.file_size) * 100)} aria-valuemin={0} aria-valuemax={100}>
                                    <div
                                        className={`h-full ${t.status === 'failed' ? 'bg-red-500' : 'bg-gc-primary'}`}
                                        style={{ width: `${Math.min(100, (t.transferred / t.file_size) * 100)}%` }}
                                    />
                                </div>
                            </div>

                            <div className="flex flex-col gap-2 items-end">
                                <div className={`text-xs font-bold ${t.status === 'active' ? 'text-blue-400' :
                                    t.status === 'completed' ? 'text-green-500' :
                                        t.status === 'failed' ? 'text-red-500' : 'text-gray-500'
                                    }`} aria-label={`Status: ${t.status}`}>
                                    {t.status.toUpperCase()}
                                </div>

                                {t.status === 'pending' && t.direction === 'download' && (
                                    <div className="flex gap-2">
                                        <button onClick={() => handleAccept(t.id)} className="px-2 py-1 bg-green-600 hover:bg-green-500 text-white text-xs rounded" aria-label="Accept Transfer">
                                            Accept
                                        </button>
                                        <button onClick={() => handleReject(t.id)} className="px-2 py-1 bg-red-600 hover:bg-red-500 text-white text-xs rounded" aria-label="Reject Transfer">
                                            Reject
                                        </button>
                                    </div>
                                )}
                                {(t.status === 'active' || t.status === 'pending') && (
                                    <button onClick={() => handleCancel(t.id)} className="text-xs text-red-400 hover:text-red-300 underline" aria-label="Cancel Transfer">
                                        Cancel
                                    </button>
                                )}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
