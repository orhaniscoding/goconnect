import { useState, useEffect } from "react";

// =============================================================================
// Types
// =============================================================================

interface Server {
  id: string;
  name: string;
  icon: string;
  description?: string;
  isOwner: boolean;
  memberCount: number;
  unread?: number;
}

interface Network {
  id: string;
  serverId: string;
  name: string;
  subnet: string;
  connected: boolean;
  myIp?: string;
  clients: Client[];
}

interface Client {
  id: string;
  name: string;
  ip: string;
  status: "online" | "idle" | "offline";
  isHost: boolean;
}

interface Channel {
  id: string;
  name: string;
  unread?: number;
}

interface User {
  deviceId: string;
  username: string;
}

// =============================================================================
// Mock Data
// =============================================================================

const mockServers: Server[] = [
  { id: "1", name: "Gaming Squad", icon: "🎮", isOwner: true, memberCount: 12, unread: 3 },
  { id: "2", name: "Work Team", icon: "💼", isOwner: false, memberCount: 8 },
  { id: "3", name: "Friends", icon: "👥", isOwner: false, memberCount: 5, unread: 1 },
];

const mockNetworks: Record<string, Network[]> = {
  "1": [
    {
      id: "n1", serverId: "1", name: "Minecraft LAN", subnet: "10.0.1.0/24",
      connected: true, myIp: "10.0.1.5",
      clients: [
        { id: "c1", name: "Alice", ip: "10.0.1.1", status: "online", isHost: true },
        { id: "c2", name: "Bob", ip: "10.0.1.2", status: "online", isHost: false },
        { id: "c3", name: "You", ip: "10.0.1.5", status: "online", isHost: false },
      ]
    },
    { id: "n2", serverId: "1", name: "Valorant Party", subnet: "10.0.2.0/24", connected: false, clients: [] },
  ],
  "2": [
    { id: "n3", serverId: "2", name: "Office VPN", subnet: "10.1.0.0/24", connected: false, clients: [] },
  ],
  "3": [
    { id: "n4", serverId: "3", name: "Movie Night", subnet: "10.2.0.0/24", connected: false, clients: [] },
  ],
};

const mockChannels: Channel[] = [
  { id: "c1", name: "general", unread: 2 },
  { id: "c2", name: "announcements" },
];

// =============================================================================
// App Component
// =============================================================================

function App() {
  // Auth
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [user, setUser] = useState<User | null>(null);

  // Data
  const [servers, setServers] = useState<Server[]>([]);
  const [selectedServer, setSelectedServer] = useState<Server | null>(null);
  const [networks, setNetworks] = useState<Network[]>([]);
  const [selectedNetwork, setSelectedNetwork] = useState<Network | null>(null);
  const [channels] = useState<Channel[]>(mockChannels);

  // UI
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showCreateNetwork, setShowCreateNetwork] = useState(false);
  const [showServerSettings, setShowServerSettings] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [copiedInvite, setCopiedInvite] = useState(false);

  // Create Server form
  const [newServerName, setNewServerName] = useState("");
  const [newServerIcon, setNewServerIcon] = useState("🎮");

  // Join Server form
  const [inviteLink, setInviteLink] = useState("");

  // Create Network form
  const [newNetworkName, setNewNetworkName] = useState("");

  // Welcome form
  const [username, setUsername] = useState("");
  const [usernameError, setUsernameError] = useState("");

  // Auto-restore session on mount
  useEffect(() => {
    const savedUsername = localStorage.getItem("gc_username");
    const savedDeviceId = localStorage.getItem("gc_device_id");

    if (savedUsername && savedDeviceId) {
      setUser({ deviceId: savedDeviceId, username: savedUsername });
      setServers(mockServers);
      setIsLoggedIn(true);
    }
  }, []);

  // Generate or get device ID
  const getDeviceId = (): string => {
    let deviceId = localStorage.getItem("gc_device_id");
    if (!deviceId) {
      deviceId = crypto.randomUUID();
      localStorage.setItem("gc_device_id", deviceId);
    }
    return deviceId;
  };

  // Handle start
  const handleStart = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = username.trim();

    if (trimmed.length < 2) {
      setUsernameError("Username must be at least 2 characters");
      return;
    }
    if (trimmed.length > 20) {
      setUsernameError("Username must be less than 20 characters");
      return;
    }

    const deviceId = getDeviceId();
    setUser({ deviceId, username: trimmed });
    localStorage.setItem("gc_username", trimmed);
    setServers(mockServers);
    setIsLoggedIn(true);
  };

  // Handle server selection
  const handleSelectServer = (server: Server) => {
    setSelectedServer(server);
    setNetworks(mockNetworks[server.id] || []);
    setSelectedNetwork(null);
  };

  // Handle network toggle
  const handleToggleNetwork = (network: Network) => {
    setNetworks(prev => prev.map(n =>
      n.id === network.id
        ? { ...n, connected: !n.connected, myIp: !n.connected ? "10.0.1.5" : undefined }
        : n
    ));
    if (!network.connected) {
      setSelectedNetwork({ ...network, connected: true, myIp: "10.0.1.5" });
    }
  };

  // Handle create server
  const handleCreateServer = () => {
    if (!newServerName.trim()) return;

    const newServer: Server = {
      id: crypto.randomUUID(),
      name: newServerName.trim(),
      icon: newServerIcon,
      isOwner: true,
      memberCount: 1,
    };

    setServers(prev => [...prev, newServer]);
    mockNetworks[newServer.id] = [];
    setSelectedServer(newServer);
    setNetworks([]);
    setNewServerName("");
    setNewServerIcon("🎮");
    setShowCreateModal(false);
  };

  // Handle join server
  const handleJoinServer = () => {
    if (!inviteLink.trim()) return;

    // Mock: Create a joined server
    const newServer: Server = {
      id: crypto.randomUUID(),
      name: "Joined Server",
      icon: "🔗",
      isOwner: false,
      memberCount: 5,
    };

    setServers(prev => [...prev, newServer]);
    mockNetworks[newServer.id] = [];
    setInviteLink("");
    setShowJoinModal(false);
  };

  // Handle create network
  const handleCreateNetwork = () => {
    if (!newNetworkName.trim() || !selectedServer) return;

    const newNetwork: Network = {
      id: crypto.randomUUID(),
      serverId: selectedServer.id,
      name: newNetworkName.trim(),
      subnet: `10.${Math.floor(Math.random() * 255)}.0.0/24`,
      connected: false,
      clients: [],
    };

    setNetworks(prev => [...prev, newNetwork]);
    mockNetworks[selectedServer.id] = [...(mockNetworks[selectedServer.id] || []), newNetwork];
    setNewNetworkName("");
    setShowCreateNetwork(false);
  };

  // Handle delete server
  const handleDeleteServer = () => {
    if (!selectedServer) return;
    setServers(prev => prev.filter(s => s.id !== selectedServer.id));
    delete mockNetworks[selectedServer.id];
    setSelectedServer(null);
    setNetworks([]);
    setSelectedNetwork(null);
    setShowServerSettings(false);
  };

  // Handle leave server
  const handleLeaveServer = () => {
    if (!selectedServer) return;
    setServers(prev => prev.filter(s => s.id !== selectedServer.id));
    setSelectedServer(null);
    setNetworks([]);
    setSelectedNetwork(null);
    setShowServerSettings(false);
  };

  // Generate invite link
  const getInviteLink = () => {
    if (!selectedServer) return "";
    return `gc://join/${selectedServer.id.slice(0, 8)}`;
  };

  // Copy invite link
  const handleCopyInvite = async () => {
    const link = getInviteLink();
    await navigator.clipboard.writeText(link);
    setCopiedInvite(true);
    setTimeout(() => setCopiedInvite(false), 2000);
  };

  // ==========================================================================
  // Login Screen
  // ==========================================================================

  if (!isLoggedIn) {
    return (
      <div className="h-screen w-screen bg-gc-dark-700 flex items-center justify-center">
        <div className="bg-gc-dark-800 p-8 rounded-lg shadow-xl w-[400px]">
          <div className="text-center mb-8">
            <div className="text-5xl mb-4">🔗</div>
            <h1 className="text-3xl font-bold text-white mb-2">GoConnect</h1>
            <p className="text-gray-400">Virtual LAN made simple</p>
          </div>

          <form onSubmit={handleStart} className="space-y-4">
            <div>
              <label className="block text-sm text-gray-300 mb-1">Choose a username</label>
              <input
                type="text"
                value={username}
                onChange={(e) => { setUsername(e.target.value); setUsernameError(""); }}
                className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none"
                placeholder="Enter your username"
                autoFocus
                maxLength={20}
              />
            </div>

            {usernameError && <div className="text-red-400 text-sm">{usernameError}</div>}

            <button
              type="submit"
              className="w-full py-3 bg-gc-primary hover:bg-gc-primary/80 text-white font-medium rounded transition"
            >
              Get Started
            </button>

            <p className="text-center text-gray-500 text-xs">
              Your device ID is stored locally for identification
            </p>
          </form>
        </div>
      </div>
    );
  }

  // ==========================================================================
  // Main App
  // ==========================================================================

  return (
    <div className="h-screen w-screen flex bg-gc-dark-700">
      {/* Server Sidebar */}
      <div className="w-[72px] bg-gc-dark-900 flex flex-col items-center py-3 gap-2">
        <button
          onClick={() => { setSelectedServer(null); setSelectedNetwork(null); }}
          className={`w-12 h-12 rounded-2xl hover:rounded-xl transition-all flex items-center justify-center text-xl
            ${!selectedServer ? 'bg-gc-primary rounded-xl' : 'bg-gc-dark-700 hover:bg-gc-primary'}`}
        >
          🏠
        </button>

        <div className="w-8 h-[2px] bg-gc-dark-600 rounded-full my-1" />

        {servers.map((server) => (
          <button
            key={server.id}
            onClick={() => handleSelectServer(server)}
            className={`relative w-12 h-12 rounded-2xl hover:rounded-xl transition-all flex items-center justify-center text-2xl
              ${selectedServer?.id === server.id ? 'bg-gc-primary rounded-xl' : 'bg-gc-dark-700 hover:bg-gc-primary'}`}
            title={server.name}
          >
            {server.icon}
            {server.unread && (
              <span className="absolute -bottom-1 -right-1 w-5 h-5 bg-red-500 rounded-full text-xs text-white flex items-center justify-center">
                {server.unread}
              </span>
            )}
          </button>
        ))}

        <button
          onClick={() => setShowCreateModal(true)}
          className="w-12 h-12 bg-gc-dark-700 rounded-2xl hover:rounded-xl hover:bg-green-600 transition-all flex items-center justify-center text-green-500 hover:text-white text-2xl"
        >
          +
        </button>

        <div className="flex-1" />

        <button className="w-12 h-12 bg-gc-dark-700 rounded-full flex items-center justify-center text-xl">
          👤
        </button>
      </div>

      {/* Channel Sidebar */}
      <div className="w-60 bg-gc-dark-800 flex flex-col">
        <div className="h-12 px-4 flex items-center border-b border-gc-dark-900 shadow cursor-pointer hover:bg-gc-dark-700" onClick={() => selectedServer && setShowServerSettings(true)}>
          <h2 className="font-semibold text-white truncate flex-1">
            {selectedServer?.name || "Home"}
          </h2>
          {selectedServer?.isOwner && (
            <span className="ml-2 text-xs bg-gc-primary/20 text-gc-primary px-2 py-0.5 rounded">Owner</span>
          )}
          {selectedServer && <span className="ml-2 text-gray-400">⚙️</span>}
        </div>

        <div className="flex-1 overflow-y-auto p-2">
          {selectedServer ? (
            <>
              <div className="text-xs text-gray-400 uppercase tracking-wide px-2 py-2 flex items-center justify-between">
                <span>Networks</span>
                {selectedServer?.isOwner && (
                  <button onClick={() => setShowCreateNetwork(true)} className="hover:text-white text-lg">+</button>
                )}
              </div>
              {networks.map((network) => (
                <button
                  key={network.id}
                  onClick={() => setSelectedNetwork(network)}
                  className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-left
                    ${selectedNetwork?.id === network.id ? 'bg-gc-dark-600 text-white' : 'text-gray-400 hover:text-white hover:bg-gc-dark-700'}`}
                >
                  <span className={`w-2 h-2 rounded-full ${network.connected ? 'bg-green-500' : 'bg-gray-500'}`} />
                  <span className="flex-1 truncate">{network.name}</span>
                  <span className="text-xs text-gray-500">{network.clients.length}</span>
                </button>
              ))}

              <div className="text-xs text-gray-400 uppercase tracking-wide px-2 py-2 mt-4">Chat</div>
              {channels.map((channel) => (
                <button
                  key={channel.id}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-left text-gray-400 hover:text-white hover:bg-gc-dark-700"
                >
                  <span>#</span>
                  <span className="flex-1">{channel.name}</span>
                </button>
              ))}
            </>
          ) : (
            <div className="text-center text-gray-500 mt-8 px-4">
              <p>Select a server</p>
            </div>
          )}
        </div>

        <div className="h-14 bg-gc-dark-900 px-2 flex items-center gap-2">
          <div className="w-8 h-8 bg-gc-primary rounded-full flex items-center justify-center text-white">
            {user?.username?.[0]?.toUpperCase() || "U"}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm text-white truncate">{user?.username}</div>
            <div className="text-xs text-gray-400">Online</div>
          </div>
          <button onClick={() => { setIsLoggedIn(false); setUser(null); setServers([]); }} className="text-gray-400 hover:text-white p-1">
            🚪
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        <div className="h-12 px-4 flex items-center border-b border-gc-dark-900 bg-gc-dark-700">
          {selectedNetwork ? (
            <>
              <span className="text-white font-medium">{selectedNetwork.name}</span>
              <span className="mx-2 text-gray-500">|</span>
              <span className="text-gray-400 text-sm font-mono">{selectedNetwork.subnet}</span>
              <div className="flex-1" />
              <button
                onClick={() => handleToggleNetwork(selectedNetwork)}
                className={`px-4 py-1.5 rounded text-sm font-medium transition
                  ${selectedNetwork.connected ? 'bg-red-500 hover:bg-red-600' : 'bg-green-500 hover:bg-green-600'} text-white`}
              >
                {selectedNetwork.connected ? "Disconnect" : "Connect"}
              </button>
            </>
          ) : (
            <span className="text-gray-400">{selectedServer ? "Select a network" : "Welcome to GoConnect"}</span>
          )}
        </div>

        <div className="flex-1 p-6 overflow-y-auto">
          {selectedNetwork ? (
            <div className="space-y-6 max-w-3xl">
              <div className="bg-gc-dark-800 rounded-lg p-4">
                <h3 className="text-white font-medium mb-4">Connection Status</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <div className="text-gray-400 text-sm">Status</div>
                    <div className={selectedNetwork.connected ? 'text-green-400' : 'text-gray-500'}>
                      {selectedNetwork.connected ? '🟢 Connected' : '⚪ Disconnected'}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-400 text-sm">Your IP</div>
                    <div className="text-white font-mono">{selectedNetwork.myIp || '-'}</div>
                  </div>
                  <div>
                    <div className="text-gray-400 text-sm">Subnet</div>
                    <div className="text-white font-mono">{selectedNetwork.subnet}</div>
                  </div>
                  <div>
                    <div className="text-gray-400 text-sm">Clients</div>
                    <div className="text-white">{selectedNetwork.clients.length} online</div>
                  </div>
                </div>
              </div>

              {selectedNetwork.connected && selectedNetwork.clients.length > 0 && (
                <div className="bg-gc-dark-800 rounded-lg p-4">
                  <h3 className="text-white font-medium mb-4">Online Clients</h3>
                  <div className="space-y-2">
                    {selectedNetwork.clients.map((client) => (
                      <div key={client.id} className="flex items-center gap-3 p-2 rounded hover:bg-gc-dark-700">
                        <div className="relative">
                          <div className="w-10 h-10 bg-gc-primary rounded-full flex items-center justify-center text-white">
                            {client.name[0]}
                          </div>
                          <span className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-gc-dark-800
                            ${client.status === 'online' ? 'bg-green-500' : 'bg-yellow-500'}`} />
                        </div>
                        <div className="flex-1">
                          <div className="text-white flex items-center gap-2">
                            {client.name}
                            {client.isHost && <span className="text-xs bg-yellow-500/20 text-yellow-400 px-1.5 py-0.5 rounded">Host</span>}
                          </div>
                          <div className="text-gray-400 text-sm font-mono">{client.ip}</div>
                        </div>
                        <button className="px-3 py-1 text-sm text-gray-400 hover:text-white hover:bg-gc-dark-600 rounded">Ping</button>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : selectedServer ? (
            <div className="text-center text-gray-400 mt-20">
              <div className="text-6xl mb-4">🌐</div>
              <h3 className="text-xl text-white mb-2">Welcome to {selectedServer.name}</h3>
              <p>Select a network from the sidebar</p>
            </div>
          ) : (
            <div className="text-center text-gray-400 mt-20">
              <div className="text-6xl mb-4">👋</div>
              <h3 className="text-xl text-white mb-2">Welcome to GoConnect</h3>
              <p className="mb-8">Create or join a server to get started</p>
              <div className="flex gap-4 justify-center">
                <button onClick={() => setShowCreateModal(true)} className="px-6 py-3 bg-gc-primary hover:bg-gc-primary/80 text-white rounded-lg font-medium">
                  Create Server
                </button>
                <button onClick={() => setShowJoinModal(true)} className="px-6 py-3 bg-gc-dark-600 hover:bg-gc-dark-500 text-white rounded-lg font-medium">
                  Join Server
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Create Server Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreateModal(false)}>
          <div className="bg-gc-dark-800 rounded-lg p-6 w-[400px]" onClick={e => e.stopPropagation()}>
            <h2 className="text-xl font-bold text-white mb-4">Create a Server</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-gray-300 mb-1">Server Name</label>
                <input
                  type="text"
                  value={newServerName}
                  onChange={(e) => setNewServerName(e.target.value)}
                  className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none"
                  placeholder="My Awesome Server"
                  autoFocus
                />
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">Icon</label>
                <div className="flex gap-2 flex-wrap">
                  {["🎮", "💼", "👥", "🎵", "📚", "⚽", "🌐", "💻", "🎬", "🏠"].map((icon) => (
                    <button
                      key={icon}
                      type="button"
                      onClick={() => setNewServerIcon(icon)}
                      className={`w-10 h-10 rounded text-xl transition ${newServerIcon === icon ? 'bg-gc-primary' : 'bg-gc-dark-700 hover:bg-gc-dark-600'}`}
                    >
                      {icon}
                    </button>
                  ))}
                </div>
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => setShowCreateModal(false)} className="flex-1 py-2 bg-gc-dark-600 hover:bg-gc-dark-500 text-white rounded">Cancel</button>
              <button onClick={handleCreateServer} disabled={!newServerName.trim()} className="flex-1 py-2 bg-gc-primary hover:bg-gc-primary/80 text-white rounded disabled:opacity-50 disabled:cursor-not-allowed">Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Join Server Modal */}
      {showJoinModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowJoinModal(false)}>
          <div className="bg-gc-dark-800 rounded-lg p-6 w-[400px]" onClick={e => e.stopPropagation()}>
            <h2 className="text-xl font-bold text-white mb-4">Join a Server</h2>
            <div>
              <label className="block text-sm text-gray-300 mb-1">Invite Link</label>
              <input
                type="text"
                value={inviteLink}
                onChange={(e) => setInviteLink(e.target.value)}
                className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none"
                placeholder="gc://join.goconnect.io/abc123"
                autoFocus
              />
            </div>
            <p className="text-gray-500 text-sm mt-2">Enter an invite link or code to join</p>
            <div className="flex gap-3 mt-6">
              <button onClick={() => setShowJoinModal(false)} className="flex-1 py-2 bg-gc-dark-600 hover:bg-gc-dark-500 text-white rounded">Cancel</button>
              <button onClick={handleJoinServer} disabled={!inviteLink.trim()} className="flex-1 py-2 bg-gc-primary hover:bg-gc-primary/80 text-white rounded disabled:opacity-50 disabled:cursor-not-allowed">Join</button>
            </div>
          </div>
        </div>
      )}

      {/* Create Network Modal */}
      {showCreateNetwork && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreateNetwork(false)}>
          <div className="bg-gc-dark-800 rounded-lg p-6 w-[400px]" onClick={e => e.stopPropagation()}>
            <h2 className="text-xl font-bold text-white mb-4">Create a Network</h2>
            <div>
              <label className="block text-sm text-gray-300 mb-1">Network Name</label>
              <input
                type="text"
                value={newNetworkName}
                onChange={(e) => setNewNetworkName(e.target.value)}
                className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none"
                placeholder="Gaming LAN, File Share, etc."
                autoFocus
              />
            </div>
            <p className="text-gray-500 text-sm mt-2">A subnet will be automatically assigned</p>
            <div className="flex gap-3 mt-6">
              <button onClick={() => setShowCreateNetwork(false)} className="flex-1 py-2 bg-gc-dark-600 hover:bg-gc-dark-500 text-white rounded">Cancel</button>
              <button onClick={handleCreateNetwork} disabled={!newNetworkName.trim()} className="flex-1 py-2 bg-gc-primary hover:bg-gc-primary/80 text-white rounded disabled:opacity-50 disabled:cursor-not-allowed">Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Server Settings Modal */}
      {showServerSettings && selectedServer && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowServerSettings(false)}>
          <div className="bg-gc-dark-800 rounded-lg p-6 w-[400px]" onClick={e => e.stopPropagation()}>
            <div className="flex items-center gap-3 mb-6">
              <span className="text-4xl">{selectedServer.icon}</span>
              <div>
                <h2 className="text-xl font-bold text-white">{selectedServer.name}</h2>
                <p className="text-gray-400 text-sm">{selectedServer.memberCount} members</p>
              </div>
            </div>

            <div className="space-y-3">
              <button
                onClick={() => { setShowServerSettings(false); setShowInviteModal(true); }}
                className="w-full flex items-center gap-3 px-4 py-3 bg-gc-dark-700 hover:bg-gc-dark-600 rounded text-left"
              >
                <span>🔗</span>
                <div>
                  <div className="text-white">Invite People</div>
                  <div className="text-gray-400 text-sm">Share the invite link</div>
                </div>
              </button>

              {selectedServer.isOwner ? (
                <button
                  onClick={handleDeleteServer}
                  className="w-full flex items-center gap-3 px-4 py-3 bg-red-500/10 hover:bg-red-500/20 rounded text-left text-red-400"
                >
                  <span>🗑️</span>
                  <div>
                    <div>Delete Server</div>
                    <div className="text-red-400/70 text-sm">This cannot be undone</div>
                  </div>
                </button>
              ) : (
                <button
                  onClick={handleLeaveServer}
                  className="w-full flex items-center gap-3 px-4 py-3 bg-red-500/10 hover:bg-red-500/20 rounded text-left text-red-400"
                >
                  <span>🚪</span>
                  <div>
                    <div>Leave Server</div>
                    <div className="text-red-400/70 text-sm">You can rejoin with an invite</div>
                  </div>
                </button>
              )}
            </div>

            <button
              onClick={() => setShowServerSettings(false)}
              className="w-full mt-4 py-2 bg-gc-dark-600 hover:bg-gc-dark-500 text-white rounded"
            >
              Close
            </button>
          </div>
        </div>
      )}

      {/* Invite Modal */}
      {showInviteModal && selectedServer && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowInviteModal(false)}>
          <div className="bg-gc-dark-800 rounded-lg p-6 w-[400px]" onClick={e => e.stopPropagation()}>
            <h2 className="text-xl font-bold text-white mb-2">Invite Friends</h2>
            <p className="text-gray-400 mb-4">Share this link to invite people to {selectedServer.name}</p>

            <div className="flex gap-2">
              <input
                type="text"
                value={getInviteLink()}
                readOnly
                className="flex-1 px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white font-mono text-sm"
              />
              <button
                onClick={handleCopyInvite}
                className={`px-4 py-2 rounded font-medium transition ${copiedInvite ? 'bg-green-500 text-white' : 'bg-gc-primary hover:bg-gc-primary/80 text-white'}`}
              >
                {copiedInvite ? '✓ Copied' : 'Copy'}
              </button>
            </div>

            <p className="text-gray-500 text-xs mt-3">Link expires in 7 days</p>

            <button
              onClick={() => setShowInviteModal(false)}
              className="w-full mt-4 py-2 bg-gc-dark-600 hover:bg-gc-dark-500 text-white rounded"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;