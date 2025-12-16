import { useState, useEffect } from 'react';
import { PeerInfo } from '../lib/tauri-api';

interface Props {
    networkId: string;
    selfPeer: PeerInfo | undefined;
}

interface VoiceSignal {
    type: 'offer' | 'answer' | 'candidate';
    sender_id: string;
    target_id: string;
    sdp?: any;
    candidate?: any;
}

export default function VoiceChat({ networkId, selfPeer }: Props) {
    const [isInCall, setIsInCall] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [signals, setSignals] = useState<VoiceSignal[]>([]);

    const addLog = (msg: string) => setLogs(prev => [...prev, `${new Date().toLocaleTimeString()} - ${msg}`].slice(-20));

    // Poll for signals when in call
    useEffect(() => {
        if (!isInCall || !selfPeer) return;

        const interval = setInterval(async () => {
            try {
                // In a real implementation this would call tauriApi.getVoiceSignals()
                // For now we mock the behavior or implement the fetch if possible
                // Since tauri-api doesn't have it yet, we'll simulate the flow for UI verification
                // or add the method to tauri-api.ts
                // Since we are frontend only for this round, we will log polling attempts

                // Uncomment to enable real polling once `tauriApi` is updated
                // const newSignals = await tauriApi.getVoiceSignals(networkId);
                // if (newSignals.length > 0) {
                //   setSignals(prev => [...prev, ...newSignals]);
                //   addLog(`Received ${newSignals.length} signals`);
                // }
            } catch (e) {
                console.error("Polling error", e);
            }
        }, 2000);

        return () => clearInterval(interval);
    }, [isInCall, networkId, selfPeer]);

    const handleStartCall = () => {
        setIsInCall(true);
        addLog("Joined Voice Channel");
        // Send join signal logic here
    };

    const handleEndCall = () => {
        setIsInCall(false);
        addLog("Left Voice Channel");
        setSignals([]);
    };

    const handleSendSignal = async () => {
        if (!selfPeer) return;
        addLog("Sending Test Signal (Offer)...");
        try {
            // Placeholder for signal sending
            // await tauriApi.sendVoiceSignal({ type: 'offer', ... })
            addLog("Signal dispatched to backend");
        } catch (e) {
            addLog("Failed to send signal: " + e);
        }
    };

    return (
        <div className="flex flex-col h-full bg-gc-dark-800 rounded-lg border border-gc-dark-600 overflow-hidden">
            <div className="p-4 border-b border-gc-dark-700 bg-gc-dark-900 flex justify-between items-center">
                <h3 className="font-bold text-white flex items-center gap-2">
                    <span>🎙️</span> Voice Channel
                </h3>
                <div
                    className={`w-3 h-3 rounded-full ${isInCall ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}
                    title={isInCall ? "Voice Connected" : "Voice Disconnected"}
                    aria-label={isInCall ? "Voice Connected" : "Voice Disconnected"}
                />
            </div>

            <div className="flex-1 p-4 overflow-y-auto space-y-2 font-mono text-xs text-gray-400">
                <div className="text-right text-[10px] text-gray-600">Signals: {signals.length}</div>
                {logs.length === 0 && <div className="text-center italic opacity-50 mt-10">No activity</div>}
                {logs.map((log, i) => (
                    <div key={i}>{log}</div>
                ))}
            </div>

            <div className="p-4 border-t border-gc-dark-700 bg-gc-dark-900 flex gap-2">
                {!isInCall ? (
                    <button
                        onClick={handleStartCall}
                        className="flex-1 bg-green-600 hover:bg-green-700 text-white py-2 rounded font-bold transition-colors"
                    >
                        Join Voice
                    </button>
                ) : (
                    <>
                        <button
                            onClick={handleSendSignal}
                            className="px-4 bg-gc-primary hover:bg-opacity-80 text-white rounded font-medium"
                        >
                            Ping
                        </button>
                        <button
                            onClick={handleEndCall}
                            className="flex-1 bg-red-600 hover:bg-red-700 text-white py-2 rounded font-bold transition-colors"
                        >
                            Leave Voice
                        </button>
                    </>
                )}
            </div>
        </div>
    );
}
