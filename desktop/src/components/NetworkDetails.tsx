import { NetworkInfo, PeerInfo } from '../lib/tauri-api';

interface NetworkDetailsProps {
    selectedNetwork: NetworkInfo | undefined;
    selfPeer: PeerInfo | undefined;
    onGenerateInvite: () => void;
    onLeaveNetwork: () => void;
    setActiveTab: (tab: "peers" | "chat" | "files" | "settings" | "voice") => void;
}

export default function NetworkDetails({
    selectedNetwork,
    selfPeer,
    onGenerateInvite,
    onLeaveNetwork,
    setActiveTab
}: NetworkDetailsProps) {
    return (
        <div className="w-60 bg-gc-dark-800 flex flex-col border-r border-gc-dark-900">
            <div className="h-12 px-4 flex items-center justify-between shadow-sm bg-gc-dark-800/50">
                <h2 className="font-bold truncate text-white">
                    {selectedNetwork ? selectedNetwork.name : "Select Network"}
                </h2>
            </div>

            <div className="flex-1 p-2 space-y-1">
                {selectedNetwork && (
                    <>
                        <div className="px-2 py-1 text-xs text-gray-500 uppercase font-semibold">General</div>
                        <button
                            onClick={() => setActiveTab("peers")}
                            className="w-full text-left px-3 py-2 rounded bg-gc-dark-700 text-white font-medium flex items-center gap-2"
                        >
                            <span>👥</span> Peers
                        </button>
                        <button
                            onClick={onGenerateInvite}
                            className="w-full text-left px-3 py-2 rounded hover:bg-gc-dark-700 text-gray-300 flex items-center gap-2"
                        >
                            <span>📨</span> Invite
                        </button>
                        <button
                            onClick={onLeaveNetwork}
                            className="w-full text-left px-3 py-2 rounded hover:bg-red-900/30 text-red-400 flex items-center gap-2 mt-4"
                        >
                            <span>🚪</span> Leave
                        </button>
                    </>
                )}
            </div>

            <div className="p-3 bg-gc-dark-900/50 text-xs text-gray-500 text-center flex items-center justify-between">
                {selfPeer ? (
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]"></div>
                        <div className="overflow-hidden">
                            <div className="font-bold truncate text-white">{selfPeer.name}</div>
                            <div className="text-[10px] font-mono text-gray-500 truncate" title={selfPeer.virtual_ip}>{selfPeer.virtual_ip}</div>
                        </div>
                    </div>
                ) : (
                    <div className="text-center opacity-50">GoConnect v0.3.0</div>
                )}
                <button
                    onClick={() => setActiveTab('settings')}
                    className="hover:text-white transition-colors p-1 text-base"
                    title="Settings"
                    aria-label="Settings"
                >
                    ⚙️
                </button>
            </div>
        </div>
    );
}
