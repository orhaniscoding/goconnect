import { useState, useEffect } from 'react';
import { useToast } from './Toast';
import * as api from '../lib/api';
import { RefreshCw, Activity, Wifi } from 'lucide-react';

interface SettingsModalProps {
    isOpen: boolean;
    onClose: () => void;
}

export function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
    const toast = useToast();
    const [enableP2P, setEnableP2P] = useState(true);
    const [stunServer, setStunServer] = useState('');

    // Diagnostics state
    const [peers, setPeers] = useState<Record<string, api.PeerStatus>>({});
    const [selectedPeerId, setSelectedPeerId] = useState<string>('');
    const [testResult, setTestResult] = useState<string | null>(null);
    const [testing, setTesting] = useState(false);

    const [loading, setLoading] = useState(false);

    // Load settings and peers on mount
    useEffect(() => {
        if (!isOpen) return;

        const loadSettings = async () => {
            const config = await api.getDaemonConfig();
            if (config) {
                setEnableP2P(config.p2p.enabled);
                setStunServer(config.p2p.stun_server || '');
            }

            refreshPeers();
        };
        loadSettings();
    }, [isOpen]);

    const refreshPeers = async () => {
        const status = await api.getDaemonStatus();
        if (status && status.peers) {
            setPeers(status.peers);
            // Select first peer if none selected
            if (!selectedPeerId && Object.keys(status.peers).length > 0) {
                setSelectedPeerId(Object.keys(status.peers)[0]);
            }
        }
    };

    const handleSave = async () => {
        setLoading(true);
        // Save P2P setting to backend
        const res = await api.updateDaemonConfig({
            p2p_enabled: enableP2P,
            stun_server: stunServer
        });

        setLoading(false);

        if (res.error) {
            toast.error("Failed to save settings: " + res.error);
        } else {
            toast.success("Settings saved successfully");
            onClose();
        }
    };

    const handleTestConnection = async () => {
        if (!selectedPeerId) return;

        setTesting(true);
        setTestResult("Initiating connection check...");

        try {
            const res = await api.manualP2PConnect(selectedPeerId);
            if (res.error) {
                setTestResult(`Connection failed: ${res.error}`);
            } else {
                setTestResult("Connection check initiated. Waiting for handshake...");
                // Poll for status update
                setTimeout(async () => {
                    const status = await api.getDaemonStatus();
                    if (status && status.peers && status.peers[selectedPeerId]) {
                        const peer = status.peers[selectedPeerId];
                        const latency = peer.latency_ms ? `${peer.latency_ms}ms` : 'N/A';
                        setTestResult(`Status: ${peer.connection_state || 'Unknown'}\nLatency: ${latency}\nLast Handshake: ${peer.last_handshake}`);
                    } else {
                        setTestResult("Peer not found in status after check.");
                    }
                    setTesting(false);
                }, 2000);
            }
        } catch (err) {
            setTestResult(`Error: ${err}`);
            setTesting(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={onClose}>
            <div className="bg-gc-dark-800 rounded-lg p-6 w-[550px] shadow-2xl border border-gc-dark-700 max-h-[90vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-bold text-white">Settings</h2>
                    <button onClick={onClose} className="text-gray-400 hover:text-white">✕</button>
                </div>

                <div className="space-y-8">
                    {/* P2P Settings */}
                    <div className="space-y-4">
                        <h3 className="text-sm font-medium text-gray-400 uppercase tracking-wider flex items-center gap-2">
                            <Wifi size={16} /> P2P Networking
                        </h3>

                        <div className="flex items-center justify-between p-4 bg-gc-dark-900 rounded-lg border border-gc-dark-700">
                            <div>
                                <div className="text-white font-medium">Enable P2P Mesh</div>
                                <div className="text-sm text-gray-500">Allow direct connections to other peers</div>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                                <input
                                    type="checkbox"
                                    className="sr-only peer"
                                    checked={enableP2P}
                                    onChange={(e) => setEnableP2P(e.target.checked)}
                                />
                                <div className="w-11 h-6 bg-gc-dark-600 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-gc-primary"></div>
                            </label>
                        </div>

                        <div>
                            <label className="block text-sm text-gray-300 mb-1">Custom STUN/TURN Server</label>
                            <input
                                type="text"
                                value={stunServer}
                                onChange={(e) => setStunServer(e.target.value)}
                                className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none placeholder-gray-600"
                                placeholder="stun:stun.l.google.com:19302"
                            />
                            <p className="text-xs text-gray-500 mt-1">Leave empty to use default servers</p>
                        </div>
                    </div>

                    {/* Connection Diagnostics */}
                    <div className="space-y-4 pt-4 border-t border-gc-dark-700">
                        <div className="flex items-center justify-between">
                            <h3 className="text-sm font-medium text-gray-400 uppercase tracking-wider flex items-center gap-2">
                                <Activity size={16} /> Connection Diagnostics
                            </h3>
                            <button
                                onClick={refreshPeers}
                                className="text-xs text-blue-400 hover:text-blue-300 flex items-center gap-1"
                            >
                                <RefreshCw size={12} /> Refresh Peers
                            </button>
                        </div>

                        <div className="bg-gc-dark-900 p-4 rounded-lg border border-gc-dark-700 space-y-3">
                            <div>
                                <label className="block text-xs text-gray-400 mb-1">Select Peer to Test</label>
                                <select
                                    value={selectedPeerId}
                                    onChange={(e) => {
                                        setSelectedPeerId(e.target.value);
                                        setTestResult(null);
                                    }}
                                    className="w-full px-3 py-2 bg-gc-dark-800 border border-gc-dark-600 rounded text-white text-sm focus:outline-none"
                                >
                                    <option value="">-- Select a Peer --</option>
                                    {Object.entries(peers).map(([id, peer]) => (
                                        <option key={id} value={id}>
                                            {id.substring(0, 8)}... ({peer.endpoint || 'Unknown Endpoint'})
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <button
                                onClick={handleTestConnection}
                                disabled={!selectedPeerId || testing}
                                className="w-full py-2 bg-gc-dark-700 hover:bg-gc-dark-600 text-white rounded text-sm transition disabled:opacity-50"
                            >
                                {testing ? 'Testing Connection...' : 'Test Connection'}
                            </button>

                            {testResult && (
                                <div className="mt-2 p-3 bg-black/30 rounded border border-gc-dark-700 font-mono text-xs text-gray-300 whitespace-pre-wrap">
                                    {testResult}
                                </div>
                            )}
                        </div>
                    </div>

                    {/* About Section */}
                    <div className="pt-6 border-t border-gc-dark-700">
                        <div className="flex items-center justify-between text-sm">
                            <span className="text-gray-400">GoConnect Desktop</span>
                            <span className="text-gray-500 font-mono">v2.28.2</span>
                        </div>
                    </div>
                </div>

                <div className="flex gap-3 mt-8">
                    <button
                        onClick={onClose}
                        className="flex-1 py-2 bg-gc-dark-700 hover:bg-gc-dark-600 text-white rounded transition"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={loading}
                        className="flex-1 py-2 bg-gc-primary hover:bg-gc-primary/80 text-white rounded transition font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {loading ? 'Saving...' : 'Save Changes'}
                    </button>
                </div>
            </div>
        </div>
    );
}
