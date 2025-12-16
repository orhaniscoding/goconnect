import { NetworkInfo } from '../lib/tauri-api';

interface SidebarProps {
    networks: NetworkInfo[];
    selectedNetworkId: string | null;
    onSelectNetwork: (id: string) => void;
    onShowCreate: () => void;
    onShowJoin: () => void;
}

export default function Sidebar({
    networks,
    selectedNetworkId,
    onSelectNetwork,
    onShowCreate,
    onShowJoin
}: SidebarProps) {
    return (
        <div className="w-[72px] bg-gc-dark-900 flex flex-col items-center py-3 gap-2 border-r border-gc-dark-800">
            <div className="mb-2">
                <div className="w-10 h-10 bg-gc-primary rounded-xl flex items-center justify-center text-xl font-bold" aria-label="GoConnect Logo">
                    GC
                </div>
            </div>

            <div className="w-8 h-[2px] bg-gc-dark-600 rounded-full my-1" />

            {networks.map(net => (
                <button
                    key={net.id}
                    onClick={() => onSelectNetwork(net.id)}
                    className={`w-12 h-12 rounded-2xl transition-all flex items-center justify-center text-xl font-bold
                    ${selectedNetworkId === net.id ? 'bg-gc-primary text-white' : 'bg-gc-dark-800 text-gray-400 hover:bg-gc-dark-700 hover:text-white'}`}
                    title={net.name}
                    aria-label={`Select network ${net.name}`}
                >
                    {net.name.substring(0, 2).toUpperCase()}
                </button>
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
