import { useState, useEffect } from 'react';
import * as api from './lib/api';
import { SettingsModal } from './components/SettingsModal';
import { PeerList } from './components/PeerList';
import { useToast, Toaster } from './components/Toast';

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

// Onboarding step type
type OnboardingStep = 'welcome' | 'username' | 'choice' | 'create' | 'join' | 'done';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState("");
  const [usernameError, setUsernameError] = useState("");
  const [user, setUser] = useState<User | null>(null);

  // Onboarding state
  const [onboardingStep, setOnboardingStep] = useState<OnboardingStep>('welcome');
  const [isFirstTime, setIsFirstTime] = useState(true);

  const [servers, setServers] = useState<Server[]>([]);
  const [networks, setNetworks] = useState<Network[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);

  const [selectedServer, setSelectedServer] = useState<Server | null>(null);
  const [selectedNetwork, setSelectedNetwork] = useState<Network | null>(null);

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showCreateNetwork, setShowCreateNetwork] = useState(false);
  const [showServerSettings, setShowServerSettings] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);

  const [newServerName, setNewServerName] = useState("");
  const [newServerIcon, setNewServerIcon] = useState("🎮");
  const [inviteLink, setInviteLink] = useState("");
  const [newNetworkName, setNewNetworkName] = useState("");
  const [copiedInvite, setCopiedInvite] = useState(false);

  const toast = useToast();

  // Check if user has existing data on mount
  useEffect(() => {
    const savedUser = localStorage.getItem('goconnect_user');
    if (savedUser) {
      try {
        const userData = JSON.parse(savedUser);
        setUser(userData);
        setUsername(userData.username);
        setIsFirstTime(false);
        setOnboardingStep('done');
        setIsLoggedIn(true);
        loadServers();
      } catch {
        // Invalid saved data, start fresh
        localStorage.removeItem('goconnect_user');
      }
    }
  }, []);

  // Load invite link when modal opens
  useEffect(() => {
    if (showInviteModal && selectedServer && !inviteLinkValue) {
      loadInviteLink();
    }
  }, [showInviteModal, selectedServer]);


  const loadServers = async () => {
    const res = await api.getMyServers();
    if (res.data) {
      setServers(res.data.map(s => ({
        id: s.id,
        name: s.name,
        icon: s.icon || "🎮",
        description: s.description,
        isOwner: !!s.isOwner,
        memberCount: s.memberCount,
        unread: 0
      })));
    }
  };

  const handleSelectServer = async (server: Server) => {
    setSelectedServer(server);
    setSelectedNetwork(null);

    // Load networks
    const res = await api.listNetworks(server.id);
    if (res.data) {
      setNetworks(res.data.map(n => ({
        id: n.id,
        serverId: n.serverId,
        name: n.name,
        subnet: n.subnet,
        connected: !!n.connected,
        myIp: n.myIp,
        clients: [] // Clients loaded separately when network selected
      })));
    }

    // Channels are not yet implemented in the API
    // Will be added when chat feature is fully implemented
    setChannels([]);
  };

  const handleSelectNetwork = async (network: Network) => {
    setSelectedNetwork(network);
    
    // Load clients for the selected network
    if (selectedServer) {
      const clientsRes = await api.listNetworkClients(selectedServer.id, network.id);
      if (clientsRes.data) {
        const mappedClients: Client[] = clientsRes.data.map(c => ({
          id: c.id,
          name: c.username,
          ip: c.ip,
          status: c.status,
          isHost: c.isHost
        }));
        
        setNetworks(prev => prev.map(n => 
          n.id === network.id ? { ...n, clients: mappedClients } : n
        ));
      }
    }
  };

  const handleCreateServer = async () => {
    if (!newServerName.trim()) return;

    const res = await api.createServer({
      name: newServerName,
      icon: newServerIcon
    });

    if (res.data) {
      setShowCreateModal(false);
      setNewServerName("");
      loadServers();
      toast.success("Server created!");
    } else {
      toast.error(res.error || "Failed to create server");
    }
  };

  const handleJoinServer = async () => {
    if (!inviteLink.trim()) return;

    // Extract code from link if needed
    const code = inviteLink.split('/').pop() || inviteLink;

    const res = await api.joinServerByCode(code);
    if (res.data) {
      setShowJoinModal(false);
      setInviteLink("");
      loadServers();
      toast.success("Joined server!");
    } else {
      toast.error(res.error || "Failed to join server");
    }
  };

  const handleCreateNetwork = async () => {
    if (!selectedServer || !newNetworkName.trim()) return;

    const res = await api.createNetwork(selectedServer.id, {
      name: newNetworkName
    });

    if (res.data) {
      setShowCreateNetwork(false);
      setNewNetworkName("");
      handleSelectServer(selectedServer); // Reload networks
      toast.success("Network created!");
    } else {
      toast.error(res.error || "Failed to create network");
    }
  };

  const handleDeleteServer = async () => {
    if (!selectedServer) return;
    if (confirm("Are you sure you want to delete this server?")) {
      await api.deleteServer(selectedServer.id);
      setSelectedServer(null);
      loadServers();
      setShowServerSettings(false);
    }
  };

  const handleLeaveServer = async () => {
    if (!selectedServer) return;
    if (confirm("Are you sure you want to leave this server?")) {
      await api.leaveServer(selectedServer.id);
      setSelectedServer(null);
      loadServers();
      setShowServerSettings(false);
    }
  };

  const [inviteLinkValue, setInviteLinkValue] = useState("");

  const loadInviteLink = async () => {
    if (!selectedServer) {
      setInviteLinkValue("");
      return;
    }
    
    try {
      // Create or get server invite
      const inviteRes = await api.createServerInvite(selectedServer.id);
      if (inviteRes.data) {
        setInviteLinkValue(`gc://join.goconnect.io/${inviteRes.data.code}`);
      } else {
        setInviteLinkValue(`gc://join.goconnect.io/${selectedServer.id}`);
      }
    } catch (error) {
      console.error("Failed to get invite link:", error);
      setInviteLinkValue(`gc://join.goconnect.io/${selectedServer.id}`);
    }
  };

  const handleCopyInvite = async () => {
    if (!inviteLinkValue) {
      await loadInviteLink();
    }
    await navigator.clipboard.writeText(inviteLinkValue);
    setCopiedInvite(true);
    setTimeout(() => setCopiedInvite(false), 2000);
  };

  // Save user to localStorage when logged in
  const saveUserData = (userData: User) => {
    localStorage.setItem('goconnect_user', JSON.stringify(userData));
  };

  // Enhanced handleStart that saves user data
  const handleStartEnhanced = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim()) {
      setUsernameError("Username is required");
      return;
    }
    if (username.length < 2) {
      setUsernameError("Username must be at least 2 characters");
      return;
    }

    try {
      const response = await api.registerDevice(username);
      if (response.data) {
        const userData = { deviceId: response.data.deviceId, username };
        setUser(userData);
        saveUserData(userData);
        
        if (isFirstTime) {
          // First time user - show choice screen
          setOnboardingStep('choice');
        } else {
          // Returning user - go directly to app
          setIsLoggedIn(true);
          loadServers();
        }
      }
    } catch (error) {
      setUsernameError("Failed to register. Please try again.");
    }
  };

  // Onboarding: Welcome Screen
  if (!isLoggedIn && onboardingStep === 'welcome') {
    return (
      <div className="h-screen w-screen bg-gradient-to-br from-gc-dark-900 via-gc-dark-800 to-gc-dark-700 flex items-center justify-center">
        <div className="text-center animate-fade-in">
          <div className="text-8xl mb-6 animate-bounce">🔗</div>
          <h1 className="text-5xl font-bold text-white mb-4">GoConnect</h1>
          <p className="text-xl text-gray-400 mb-8">Virtual LAN made simple</p>
          
          <button
            onClick={() => setOnboardingStep('username')}
            className="px-8 py-4 bg-gc-primary hover:bg-gc-primary/80 text-white text-lg font-medium rounded-lg transition transform hover:scale-105 shadow-lg"
          >
            Get Started →
          </button>
          
          <p className="text-gray-500 text-sm mt-8">
            Play games, share files, and chat with friends<br />
            as if you were on the same network
          </p>
        </div>
      </div>
    );
  }

  // Onboarding: Username Screen
  if (!isLoggedIn && onboardingStep === 'username') {
    return (
      <div className="h-screen w-screen bg-gradient-to-br from-gc-dark-900 via-gc-dark-800 to-gc-dark-700 flex items-center justify-center">
        <div className="bg-gc-dark-800/80 backdrop-blur p-8 rounded-xl shadow-2xl w-[420px] border border-gc-dark-600">
          <div className="text-center mb-8">
            <div className="text-5xl mb-4">👤</div>
            <h2 className="text-2xl font-bold text-white mb-2">What should we call you?</h2>
            <p className="text-gray-400">This is how others will see you</p>
          </div>

          <form onSubmit={handleStartEnhanced} className="space-y-6">
            <div>
              <input
                type="text"
                value={username}
                onChange={(e) => { setUsername(e.target.value); setUsernameError(""); }}
                className="w-full px-4 py-3 bg-gc-dark-900 border border-gc-dark-600 rounded-lg text-white text-lg text-center focus:border-gc-primary focus:outline-none focus:ring-2 focus:ring-gc-primary/20"
                placeholder="Your username"
                autoFocus
                maxLength={20}
              />
              {usernameError && <div className="text-red-400 text-sm mt-2 text-center">{usernameError}</div>}
            </div>

            <button
              type="submit"
              className="w-full py-3 bg-gc-primary hover:bg-gc-primary/80 text-white font-medium rounded-lg transition"
            >
              Continue →
            </button>
          </form>

          <button
            onClick={() => setOnboardingStep('welcome')}
            className="w-full mt-4 text-gray-500 hover:text-gray-300 text-sm"
          >
            ← Back
          </button>
        </div>
      </div>
    );
  }

  // Onboarding: Choice Screen (Create or Join)
  if (!isLoggedIn && onboardingStep === 'choice') {
    return (
      <div className="h-screen w-screen bg-gradient-to-br from-gc-dark-900 via-gc-dark-800 to-gc-dark-700 flex items-center justify-center">
        <div className="text-center max-w-2xl mx-auto px-4">
          <div className="text-5xl mb-4">👋</div>
          <h2 className="text-3xl font-bold text-white mb-2">Welcome, {username}!</h2>
          <p className="text-gray-400 mb-10">What would you like to do?</p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Create Server Card */}
            <button
              onClick={() => {
                setOnboardingStep('done');
                setIsLoggedIn(true);
                loadServers();
                setShowCreateModal(true);
              }}
              className="bg-gc-dark-800/80 backdrop-blur p-8 rounded-xl border border-gc-dark-600 hover:border-gc-primary transition group text-left"
            >
              <div className="text-5xl mb-4 group-hover:scale-110 transition">🌐</div>
              <h3 className="text-xl font-bold text-white mb-2">Create a Server</h3>
              <p className="text-gray-400 text-sm">
                Start your own private network and invite friends to join
              </p>
            </button>

            {/* Join Server Card */}
            <button
              onClick={() => {
                setOnboardingStep('done');
                setIsLoggedIn(true);
                loadServers();
                setShowJoinModal(true);
              }}
              className="bg-gc-dark-800/80 backdrop-blur p-8 rounded-xl border border-gc-dark-600 hover:border-gc-primary transition group text-left"
            >
              <div className="text-5xl mb-4 group-hover:scale-110 transition">🔗</div>
              <h3 className="text-xl font-bold text-white mb-2">Join a Server</h3>
              <p className="text-gray-400 text-sm">
                Have an invite link? Join your friend's network instantly
              </p>
            </button>
          </div>

          <button
            onClick={() => {
              setOnboardingStep('done');
              setIsLoggedIn(true);
              loadServers();
            }}
            className="mt-8 text-gray-500 hover:text-gray-300 text-sm"
          >
            Skip for now →
          </button>
        </div>
      </div>
    );
  }

  // Legacy login screen (for returning users without saved data)
  if (!isLoggedIn) {
    return (
      <div className="h-screen w-screen bg-gc-dark-700 flex items-center justify-center">
        <div className="bg-gc-dark-800 p-8 rounded-lg shadow-xl w-[400px]">
          <div className="text-center mb-8">
            <div className="text-5xl mb-4">🔗</div>
            <h1 className="text-3xl font-bold text-white mb-2">GoConnect</h1>
            <p className="text-gray-400">Welcome back!</p>
          </div>

          <form onSubmit={handleStartEnhanced} className="space-y-4">
            <div>
              <label className="block text-sm text-gray-300 mb-1">Username</label>
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
              Continue
            </button>
          </form>
        </div>
      </div>
    );
  }

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
            {server.unread ? (
              <span className="absolute -bottom-1 -right-1 w-5 h-5 bg-red-500 rounded-full text-xs text-white flex items-center justify-center">
                {server.unread}
              </span>
            ) : null}
          </button>
        ))}

        <button
          onClick={() => setShowCreateModal(true)}
          className="w-12 h-12 bg-gc-dark-700 hover:bg-green-600 text-green-500 hover:text-white rounded-3xl hover:rounded-xl transition-all flex items-center justify-center text-2xl"
        >
          +
        </button>
      </div>

      {/* Network/Channel Sidebar */}
      <div className="w-60 bg-gc-dark-800 flex flex-col border-r border-gc-dark-900">
        {selectedServer ? (
          <>
            <div className="h-12 px-4 flex items-center justify-between shadow-sm cursor-pointer hover:bg-gc-dark-700 transition"
              onClick={() => setShowServerSettings(true)}>
              <h2 className="font-bold text-white truncate">{selectedServer.name}</h2>
              <span className="text-gray-400">▼</span>
            </div>

            <div className="flex-1 overflow-y-auto p-2 space-y-1">
              <div className="text-xs text-gray-400 uppercase tracking-wide px-2 py-2 flex items-center justify-between">
                <span>Networks</span>
                {selectedServer.isOwner && (
                  <button onClick={(e) => { e.stopPropagation(); setShowCreateNetwork(true); }} className="hover:text-white text-lg">+</button>
                )}
              </div>
              {networks.map((network) => (
                <button
                  key={network.id}
                  onClick={() => handleSelectNetwork(network)}
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
            </div>
          </>
        ) : (
          <div className="text-center text-gray-500 mt-8 px-4">
            <p>Select a server</p>
          </div>
        )}

        {/* User Bar */}
        <div className="h-14 bg-gc-dark-900 px-2 flex items-center gap-2">
          <div className="w-8 h-8 bg-gc-primary rounded-full flex items-center justify-center text-white">
            {user?.username?.[0]?.toUpperCase() || "U"}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm text-white truncate">{user?.username}</div>
            <div className="text-xs text-gray-400">Online</div>
          </div>
          <button onClick={() => setShowSettingsModal(true)} className="text-gray-400 hover:text-white p-1">
            ⚙️
          </button>
          <button onClick={() => { setIsLoggedIn(false); setUser(null); setServers([]); }} className="text-gray-400 hover:text-white p-1">
            🚪
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        <div className="h-12 px-4 flex items-center border-b border-gc-dark-900 bg-gc-dark-700">
          {selectedNetwork ? (
            <div className="flex items-center">
              <span className="text-white font-medium">{selectedNetwork.name}</span>
              <span className="mx-2 text-gray-500">|</span>
              <span className="text-gray-400 text-sm font-mono">{selectedNetwork.subnet}</span>
            </div>
          ) : (
            <div className="text-white font-medium">Dashboard</div>
          )}
          <div className="flex-1" />
        </div>

        <div className="flex-1 p-6 overflow-y-auto">
          {selectedNetwork ? (
            selectedNetwork.connected && selectedNetwork.clients.length > 0 ? (
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
            ) : (
              <div className="text-center text-gray-500 mt-20">
                <p>No clients connected</p>
              </div>
            )
          ) : selectedServer ? (
            <div className="text-center text-gray-400 mt-20">
              <div className="text-6xl mb-4">🌐</div>
              <h3 className="text-xl text-white mb-2">Welcome to {selectedServer.name}</h3>
              <p>Select a network from the sidebar</p>
            </div>
          ) : (
            <div className="space-y-6">
              <div className="text-center text-gray-400 mt-10">
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

              {/* P2P Status Section */}
              <div className="max-w-2xl mx-auto">
                <PeerList />
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Modals */}
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
                onClick={async () => {
                  setShowServerSettings(false);
                  setShowInviteModal(true);
                  await loadInviteLink();
                }}
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

      {showInviteModal && selectedServer && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowInviteModal(false)}>
          <div className="bg-gc-dark-800 rounded-lg p-6 w-[400px]" onClick={e => e.stopPropagation()}>
            <h2 className="text-xl font-bold text-white mb-2">Invite Friends</h2>
            <p className="text-gray-400 mb-4">Share this link to invite people to {selectedServer.name}</p>

            <div className="flex gap-2">
              <input
                type="text"
                value={inviteLinkValue}
                readOnly
                className="flex-1 px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white font-mono text-sm"
                placeholder="Loading invite link..."
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

      <SettingsModal isOpen={showSettingsModal} onClose={() => setShowSettingsModal(false)} />
      <Toaster />
    </div>
  );
}

export default App;
