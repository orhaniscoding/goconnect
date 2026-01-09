import { NetworkInfo, PeerInfo } from '../lib/tauri-api';

interface SidebarProps {
    networks: NetworkInfo[];
    selectedNetworkId: string | null;
    peers: PeerInfo[];
    onSelectNetwork: (id: string) => void;
    onShowCreate: () => void;
    onShowJoin: () => void;
}

export default function Sidebar({
    networks,
    selectedNetworkId,
    peers,
    onSelectNetwork,
    onShowCreate,
    onShowJoin
}: SidebarProps) {
    // Count online peers (excluding self)
    const onlinePeerCount = peers.filter(p => p.connected && !p.is_self).length;

    return (
        <div className="w-[72px] bg-gc-dark-900 flex flex-col items-center py-3 gap-2 border-r border-gc-dark-800">
            <div className="mb-2">
                <div className="w-10 h-10 bg-gc-primary rounded-xl flex items-center justify-center text-xl font-bold" aria-label="GoConnect Logo">
                    GC
                </div>
            </div>

            <div className="w-8 h-[2px] bg-gc-dark-600 rounded-full my-1" />

            {networks.map(net => (
                <div key={net.id} className="relative">
                    <button
                        onClick={() => onSelectNetwork(net.id)}
                        className={`w-12 h-12 rounded-2xl transition-all flex items-center justify-center text-xl font-bold
                        ${selectedNetworkId === net.id ? 'bg-gc-primary text-white' : 'bg-gc-dark-800 text-gray-400 hover:bg-gc-dark-700 hover:text-white'}`}
                        title={net.name}
                        aria-label={`Select network ${net.name}`}
                    >
                        {net.name.substring(0, 2).toUpperCase()}
                    </button>
                    {/* Peer count badge - shown for selected network */}
                    {selectedNetworkId === net.id && onlinePeerCount > 0 && (
                        <div className="absolute -bottom-1 -right-1 w-5 h-5 bg-green-500 rounded-full flex items-center justify-center text-[10px] font-bold text-white shadow-lg">
                            {onlinePeerCount > 9 ? '9+' : onlinePeerCount}
                        </div>
                    )}
                </div>
            ))}

            <button
                onClick={onShowCreate}
                className="w-12 h-12 bg-gc-dark-800 text-green-500 hover:bg-green-600 hover:text-white rounded-3xl hover:rounded-xl transition-all flex items-center justify-center text-2xl"
                title="Create Network"
                aria-label="Create Network"
            >
                +
            </button>
            <button
                onClick={onShowJoin}
                className="w-12 h-12 bg-gc-dark-800 text-blue-500 hover:bg-blue-600 hover:text-white rounded-3xl hover:rounded-xl transition-all flex items-center justify-center text-xl"
                title="Join Network"
                aria-label="Join Network"
            >
                🔗
            </button>
        </div>
    );
}

