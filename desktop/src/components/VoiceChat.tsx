import { useState, useEffect, useRef, useCallback } from 'react';
import { tauriApi, PeerInfo, VoiceSignal } from '../lib/tauri-api';
import { useToast } from './Toast';
// Notification import ready for future use: import { notifyVoiceCall } from '../lib/notifications';



interface Props {
    networkId: string;
    selfPeer: PeerInfo | undefined;
    connectedPeers: PeerInfo[];
}

interface PeerConnection {
    peerId: string;
    peerName: string;
    connection: RTCPeerConnection;
    audioElement: HTMLAudioElement;
    isMuted: boolean;
}

export default function VoiceChat({ networkId, selfPeer, connectedPeers }: Props) {
    const [isInCall, setIsInCall] = useState(false);
    const [isMuted, setIsMuted] = useState(false);
    const [localStream, setLocalStream] = useState<MediaStream | null>(null);
    const [peerConnections, setPeerConnections] = useState<Map<string, PeerConnection>>(new Map());
    const [logs, setLogs] = useState<string[]>([]);

    const toast = useToast();
    const pollIntervalRef = useRef<number | null>(null);

    const addLog = useCallback((msg: string) => {
        setLogs(prev => [...prev, `${new Date().toLocaleTimeString()} - ${msg}`].slice(-15));
    }, []);

    // Cleanup on unmount
    useEffect(() => {
        return () => {
            if (pollIntervalRef.current) {
                clearInterval(pollIntervalRef.current);
            }
            cleanupCall();
        };
    }, []);

    // Poll for signals when in call
    useEffect(() => {
        if (!isInCall || !selfPeer) return;

        const pollSignals = async () => {
            try {
                const signals = await tauriApi.getVoiceSignals(networkId);
                for (const signal of signals) {
                    await handleIncomingSignal(signal);
                }
            } catch (e) {
                console.error("Signal polling error", e);
            }
        };

        pollIntervalRef.current = window.setInterval(pollSignals, 1000);
        return () => {
            if (pollIntervalRef.current) {
                clearInterval(pollIntervalRef.current);
            }
        };
    }, [isInCall, networkId, selfPeer]);

    const handleIncomingSignal = async (signal: VoiceSignal) => {
        if (!selfPeer || signal.target_id !== selfPeer.id) return;

        const senderId = signal.sender_id;
        let peerConn = peerConnections.get(senderId);

        if (signal.type === 'offer') {
            addLog(`Received offer from ${senderId.substring(0, 8)}`);

            // Create peer connection if doesn't exist
            if (!peerConn) {
                peerConn = await createPeerConnection(senderId, signal.sender_id);
            }

            if (signal.sdp) {
                await peerConn.connection.setRemoteDescription(new RTCSessionDescription(signal.sdp));
                const answer = await peerConn.connection.createAnswer();
                await peerConn.connection.setLocalDescription(answer);

                // Send answer
                await tauriApi.sendVoiceSignal({
                    type: 'answer',
                    sender_id: selfPeer.id,
                    target_id: senderId,
                    network_id: networkId,
                    sdp: answer,
                });
                addLog(`Sent answer to ${senderId.substring(0, 8)}`);
            }
        } else if (signal.type === 'answer' && peerConn && signal.sdp) {
            addLog(`Received answer from ${senderId.substring(0, 8)}`);
            await peerConn.connection.setRemoteDescription(new RTCSessionDescription(signal.sdp));
        } else if (signal.type === 'candidate' && peerConn && signal.candidate) {
            await peerConn.connection.addIceCandidate(new RTCIceCandidate(signal.candidate));
        }
    };

    const createPeerConnection = async (peerId: string, peerName: string): Promise<PeerConnection> => {
        const config: RTCConfiguration = {
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun.cloudflare.com:3478' },
            ],
        };

        const connection = new RTCPeerConnection(config);
        const audioElement = new Audio();
        audioElement.autoplay = true;

        // Add local stream tracks
        if (localStream) {
            localStream.getTracks().forEach(track => {
                connection.addTrack(track, localStream);
            });
        }

        // Handle remote stream
        connection.ontrack = (event) => {
            addLog(`Audio connected with ${peerId.substring(0, 8)}`);
            audioElement.srcObject = event.streams[0];
        };

        // Handle ICE candidates
        connection.onicecandidate = async (event) => {
            if (event.candidate && selfPeer) {
                await tauriApi.sendVoiceSignal({
                    type: 'candidate',
                    sender_id: selfPeer.id,
                    target_id: peerId,
                    network_id: networkId,
                    candidate: event.candidate.toJSON(),
                });
            }
        };

        connection.onconnectionstatechange = () => {
            addLog(`Connection to ${peerId.substring(0, 8)}: ${connection.connectionState}`);
        };

        const peerConn: PeerConnection = {
            peerId,
            peerName,
            connection,
            audioElement,
            isMuted: false,
        };

        setPeerConnections(prev => new Map(prev).set(peerId, peerConn));
        return peerConn;
    };

    const handleJoinVoice = async () => {
        if (!selfPeer) {
            toast.error("Not connected to network");
            return;
        }

        try {
            // Request microphone access
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
            setLocalStream(stream);
            setIsInCall(true);
            addLog("Joined voice channel");
            addLog("Microphone enabled");

            // Create offers for all connected peers
            for (const peer of connectedPeers) {
                if (peer.id === selfPeer.id || !peer.connected) continue;

                const peerConn = await createPeerConnection(peer.id, peer.name || peer.id);
                const offer = await peerConn.connection.createOffer();
                await peerConn.connection.setLocalDescription(offer);

                await tauriApi.sendVoiceSignal({
                    type: 'offer',
                    sender_id: selfPeer.id,
                    target_id: peer.id,
                    network_id: networkId,
                    sdp: offer,
                });
                addLog(`Sent offer to ${peer.name || peer.id.substring(0, 8)}`);
            }
        } catch (e) {
            toast.error("Failed to access microphone");
            console.error("getUserMedia error:", e);
            addLog("ERROR: Microphone access denied");
        }
    };

    const handleLeaveVoice = () => {
        cleanupCall();
        setIsInCall(false);
        addLog("Left voice channel");
    };

    const cleanupCall = () => {
        // Stop local stream
        if (localStream) {
            localStream.getTracks().forEach(track => track.stop());
            setLocalStream(null);
        }

        // Close all peer connections
        peerConnections.forEach(pc => {
            pc.connection.close();
            pc.audioElement.srcObject = null;
        });
        setPeerConnections(new Map());
    };

    const toggleMute = () => {
        if (localStream) {
            localStream.getAudioTracks().forEach(track => {
                track.enabled = isMuted; // Toggle
            });
            setIsMuted(!isMuted);
            addLog(isMuted ? "Unmuted" : "Muted");
        }
    };

    return (
        <div className="flex flex-col h-full bg-gc-dark-800 rounded-lg border border-gc-dark-600 overflow-hidden">
            {/* Header */}
            <div className="p-4 border-b border-gc-dark-700 bg-gc-dark-900 flex justify-between items-center">
                <h3 className="font-bold text-white flex items-center gap-2">
                    <span>🎙️</span> Voice Channel
                </h3>
                <div className="flex items-center gap-2">
                    <span className="text-xs text-gray-500">{peerConnections.size} connected</span>
                    <div
                        className={`w-3 h-3 rounded-full ${isInCall ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}
                        title={isInCall ? "Voice Connected" : "Voice Disconnected"}
                    />
                </div>
            </div>

            {/* Participants */}
            {isInCall && (
                <div className="p-3 border-b border-gc-dark-700 bg-gc-dark-850">
                    <div className="text-xs text-gray-500 mb-2">In Call</div>
                    <div className="flex flex-wrap gap-2">
                        {/* Self */}
                        <div className={`px-2 py-1 rounded text-xs ${isMuted ? 'bg-red-900/50 text-red-300' : 'bg-green-900/50 text-green-300'}`}>
                            {selfPeer?.name || 'You'} {isMuted ? '🔇' : '🎤'}
                        </div>
                        {/* Peers */}
                        {Array.from(peerConnections.values()).map(pc => (
                            <div key={pc.peerId} className="px-2 py-1 rounded text-xs bg-gc-dark-700 text-gray-300">
                                {pc.peerName || pc.peerId.substring(0, 8)} 🔊
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Logs */}
            <div className="flex-1 p-4 overflow-y-auto space-y-1 font-mono text-xs text-gray-500">
                {logs.length === 0 && <div className="text-center italic opacity-50 mt-10">No activity</div>}
                {logs.map((log, i) => (
                    <div key={i}>{log}</div>
                ))}
            </div>

            {/* Controls */}
            <div className="p-4 border-t border-gc-dark-700 bg-gc-dark-900 flex gap-2">
                {!isInCall ? (
                    <button
                        onClick={handleJoinVoice}
                        className="flex-1 bg-green-600 hover:bg-green-700 text-white py-2 rounded font-bold transition-colors"
                    >
                        🎤 Join Voice
                    </button>
                ) : (
                    <>
                        <button
                            onClick={toggleMute}
                            className={`px-4 py-2 rounded font-medium transition-colors ${isMuted
                                ? 'bg-red-600 hover:bg-red-700 text-white'
                                : 'bg-gc-dark-700 hover:bg-gc-dark-600 text-gray-300'
                                }`}
                        >
                            {isMuted ? '🔇 Unmute' : '🎤 Mute'}
                        </button>
                        <button
                            onClick={handleLeaveVoice}
                            className="flex-1 bg-red-600 hover:bg-red-700 text-white py-2 rounded font-bold transition-colors"
                        >
                            📴 Leave Voice
                        </button>
                    </>
                )}
            </div>
        </div>
    );
}
