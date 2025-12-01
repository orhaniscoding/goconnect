import { useState, useEffect, useRef } from "react";
import * as api from "../lib/api";
import { useToast } from "./Toast";
import { MessageSquare } from "lucide-react";
import { ChatWindow } from "./ChatWindow";

export function PeerList() {
    const [status, setStatus] = useState<api.DaemonStatus | null>(null);
    const [connecting, setConnecting] = useState<string | null>(null);
    const [activeChat, setActiveChat] = useState<string | null>(null);
    const toast = useToast();
    const prevPeersRef = useRef<Record<string, api.PeerStatus>>({});

    const fetchStatus = async () => {
        const res = await api.getDaemonStatus();
        if (res) {
            // Check for new connections or disconnections
            const prevPeers = prevPeersRef.current;
            const currentPeers = res.peers || {};

            // Check for new peers (connected)
            Object.keys(currentPeers).forEach(id => {
                if (!prevPeers[id] && currentPeers[id].endpoint) {
                    toast.success(`Connected to peer ${id.slice(0, 8)}`);
                }
            });

            // Check for disconnected peers (removed or lost endpoint)
            Object.keys(prevPeers).forEach(id => {
                if (!currentPeers[id] || (prevPeers[id].endpoint && !currentPeers[id].endpoint)) {
                    toast.warning(`Lost connection to peer ${id.slice(0, 8)}`);
                }
            });

            // Check for high latency
            Object.entries(currentPeers).forEach(([id, peer]) => {
                if (peer.latency_ms && peer.latency_ms > 200 && (!prevPeers[id] || (prevPeers[id].latency_ms || 0) <= 200)) {
                    toast.warning(`High latency detected for peer ${id.slice(0, 8)} (${peer.latency_ms}ms)`);
                }
            });

            setStatus(res);
            prevPeersRef.current = currentPeers;
        }
    };

    useEffect(() => {
        fetchStatus();
        const interval = setInterval(fetchStatus, 5000);
        return () => clearInterval(interval);
    }, []);

    const handleConnect = async (peerId: string) => {
        setConnecting(peerId);
        const res = await api.manualP2PConnect(peerId);
        setConnecting(null);
        if (res.error) {
            toast.error("Connection failed: " + res.error);
        } else {
            toast.info("Connection initiated...");
            fetchStatus();
        }
    };

    if (!status) {
        return (
            <div className="bg-gc-dark-800 rounded-lg p-4 animate-pulse border border-gc-dark-700">
                <div className="flex justify-between items-center mb-4">
                    <div className="h-4 bg-gc-dark-700 rounded w-1/4"></div>
                    <div className="h-4 bg-gc-dark-700 rounded w-1/6"></div>
                </div>
                <div className="space-y-3">
                    <div className="h-10 bg-gc-dark-700 rounded"></div>
                    <div className="h-10 bg-gc-dark-700 rounded"></div>
                </div>
                <p className="text-xs text-gray-500 mt-3 flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-yellow-500 animate-pulse"></span>
                    Waiting for local daemon (port 12345)...
                </p>
            </div>
        );
    }

    const peers = status.peers || {};

    return (
        <div className="bg-gc-dark-800 rounded-lg p-4 border border-gc-dark-700 relative">
            <h3 className="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">P2P Mesh Status</h3>

            {Object.keys(peers).length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                    <div className="text-2xl mb-2">🕸️</div>
                    <p>No active P2P connections</p>
                </div>
            ) : (
                <div className="space-y-2">
                    {Object.entries(peers).map(([id, peer]) => {
                        const isConnected = peer.connected || !!peer.endpoint;
                        const state = peer.connection_state || (isConnected ? "connected" : "disconnected");
                        const latency = peer.latency_ms || 0;

                        let stateColor = "bg-gray-500";
                        if (state === "connected") stateColor = "bg-green-500";
                        else if (state === "checking") stateColor = "bg-yellow-500 animate-pulse";
                        else if (state === "failed") stateColor = "bg-red-500";

                        let latencyColor = "text-green-400";
                        if (latency > 200) latencyColor = "text-red-400";
                        else if (latency > 50) latencyColor = "text-yellow-400";

                        return (
                            <div key={id} className="flex items-center justify-between p-3 bg-gc-dark-900 rounded-lg border border-gc-dark-700">
                                <div className="flex items-center gap-3">
                                    <div className={`w-2 h-2 rounded-full ${stateColor}`} title={state} />
                                    <div>
                                        <div className="text-white font-mono text-sm">{id.slice(0, 8)}...</div>
                                        <div className="text-xs text-gray-500 capitalize">
                                            {state} {isConnected && peer.endpoint ? `(${peer.endpoint})` : ''}
                                        </div>
                                    </div>
                                </div>

                                <div className="flex items-center gap-3">
                                    {isConnected && (
                                        <>
                                            <div className={`text-xs font-mono ${latencyColor} bg-gc-dark-800 px-2 py-1 rounded`}>
                                                {latency}ms
                                            </div>
                                            <button
                                                onClick={() => setActiveChat(id)}
                                                className="p-1.5 text-gray-400 hover:text-white hover:bg-gray-700 rounded transition"
                                                title="Chat"
                                            >
                                                <MessageSquare size={16} />
                                            </button>
                                        </>
                                    )}

                                    {!isConnected && state !== "checking" && (
                                        <button
                                            onClick={() => handleConnect(id)}
                                            disabled={connecting === id}
                                            className="text-xs bg-gc-primary hover:bg-gc-primary/80 text-white px-3 py-1 rounded transition disabled:opacity-50"
                                        >
                                            {connecting === id ? 'Connecting...' : 'Connect'}
                                        </button>
                                    )}
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}

            {activeChat && (
                <ChatWindow
                    peerId={activeChat}
                    peerName={activeChat.slice(0, 8)}
                    onClose={() => setActiveChat(null)}
                />
            )}
        </div>
    );
}
