import { useState, useEffect } from 'react';
import { onOpenUrl } from '@tauri-apps/plugin-deep-link';
import { open } from '@tauri-apps/plugin-dialog';
import { tauriApi, NetworkInfo, PeerInfo } from './lib/tauri-api';
import { handleError } from './lib/utils';
import { useToast, Toaster } from './components/Toast';
import ChatPanel from './components/ChatPanel';
import FileTransferPanel from './components/FileTransferPanel';
import Onboarding from './components/Onboarding';
import SettingsPanel from './components/SettingsPanel';
import VoiceChat from './components/VoiceChat';
import Sidebar from './components/Sidebar';
import NetworkDetails from './components/NetworkDetails';
import MetricsDashboard from './components/MetricsDashboard';
import MembersTab from './components/MembersTab';
import { CreateNetworkModal, JoinNetworkModal, RenameNetworkModal, DeleteNetworkModal } from './components/NetworkModals';


export default function App() {
  const [isDaemonRunning, setIsDaemonRunning] = useState(false);
  const [isRegistered, setIsRegistered] = useState(false);
  const [networks, setNetworks] = useState<NetworkInfo[]>([]);
  const [selectedNetworkId, setSelectedNetworkId] = useState<string | null>(null);
  const [peers, setPeers] = useState<PeerInfo[]>([]);
  const [privateChatRecipient, setPrivateChatRecipient] = useState<PeerInfo | null>(null);

  // UI State
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showRenameModal, setShowRenameModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [joinInviteCode, setJoinInviteCode] = useState("");
  const [activeTab, setActiveTab] = useState<"peers" | "chat" | "files" | "settings" | "voice" | "metrics" | "members">("peers");

  const toast = useToast();

  // Deep Link Listener
  useEffect(() => {
    const initDeepLink = async () => {
      await onOpenUrl((urls) => {
        console.log('Deep link received:', urls);
        for (const url of urls) {
          try {
            // Parse deep link: gc://join?code=XYZ or goconnect://join?code=XYZ
            // Also supports: gc://join/XYZ (path-based)
            const u = new URL(url);
            const scheme = u.protocol.replace(':', '');

            if (scheme === 'gc' || scheme === 'goconnect') {
              if (u.hostname === 'join' || u.pathname.startsWith('/join')) {
                // Try query param first, then path segment
                let code = u.searchParams.get('code');
                if (!code && u.pathname.startsWith('/join/')) {
                  code = u.pathname.replace('/join/', '').toUpperCase();
                }
                if (!code && u.hostname === 'join' && u.pathname.length > 1) {
                  code = u.pathname.slice(1).toUpperCase();
                }

                if (code && code.length > 0) {
                  setJoinInviteCode(code.toUpperCase());
                  setShowJoinModal(true);
                  toast.success("Invite code detected!");
                }
              }
            }
          } catch (e) {
            console.error('Failed to parse deep link:', url, e);
          }
        }
      });
    };
    initDeepLink();
  }, [toast]);


  // Initial Load
  useEffect(() => {
    checkDaemon();
    const interval = setInterval(checkDaemon, 5000);
    return () => clearInterval(interval);
  }, []);

  const checkDaemon = async () => {
    try {
      const running = await tauriApi.isRunning();
      setIsDaemonRunning(running);
      if (running) {
        const registered = await tauriApi.checkRegistration();
        setIsRegistered(registered);
        if (registered) {
          refreshNetworks();
        }
      }
    } catch (e) {
      console.error("Daemon check failed:", e);
      setIsDaemonRunning(false);
    }
  };

  const refreshNetworks = async () => {
    try {
      const nets = await tauriApi.listNetworks();
      setNetworks(nets);
      const status = await tauriApi.getStatus();
      if (status.network_name && !selectedNetworkId) {
        const active = nets.find(n => n.name === status.network_name);
        if (active) setSelectedNetworkId(active.id);
      }
      return nets;
    } catch (e) {
      console.error("Failed to list networks:", e);
      return [];
    }
  };

  // Load peers when network selected
  useEffect(() => {
    if (selectedNetworkId && isDaemonRunning) {
      refreshPeers();
      const interval = setInterval(refreshPeers, 3000);
      return () => clearInterval(interval);
    }
  }, [selectedNetworkId, isDaemonRunning]);

  const refreshPeers = async () => {
    try {
      const p = await tauriApi.getPeers();
      setPeers(p);
    } catch (e) {
      console.error("Failed to get peers:", e);
    }
  };

  const handleCreateNetwork = async (name: string) => {
    try {
      await tauriApi.createNetwork(name);
      toast.success("Network created");
      setShowCreateModal(false);
      refreshNetworks();
    } catch (e) {
      handleError(e, "Failed to create network");
    }
  };

  const handleJoinNetwork = async (code: string) => {
    try {
      await tauriApi.joinNetwork(code);
      toast.success("Joined network");
      setShowJoinModal(false);
      setJoinInviteCode(""); // Clear after join
      refreshNetworks();
    } catch (e) {
      handleError(e, "Failed to join network");
    }
  };

  const handleLeaveNetwork = async () => {
    if (!selectedNetworkId) return;
    if (!confirm("Are you sure you want to leave this network? You will need an invite code to rejoin.")) return;
    try {
      await tauriApi.leaveNetwork(selectedNetworkId);
      toast.success("Left network");
      setSelectedNetworkId(null);
      refreshNetworks();
    } catch (e) {
      handleError(e, "Failed to leave");
    }
  };

  const handleGenerateInvite = async () => {
    if (!selectedNetworkId) return;
    try {
      const code = await tauriApi.generateInvite(selectedNetworkId);
      navigator.clipboard.writeText(code);
      toast.success("Invite code copied to clipboard!");
    } catch (e) {
      handleError(e, "Failed to generate invite");
    }
  };

  const handleRenameNetwork = async (_networkId: string, _newName: string) => {
    // TODO: Backend API implementation needed
    // For now, show a message and close the modal
    toast.info("Rename feature - Network name will be updated after reconnect");
    setShowRenameModal(false);
    // Backend implementation:
    // try {
    //   await tauriApi.renameNetwork(networkId, newName);
    //   toast.success("Network renamed successfully");
    //   refreshNetworks();
    // } catch (e) {
    //   handleError(e, "Failed to rename network");
    // }
  };

  const handleDeleteNetwork = async (networkId: string) => {
    try {
      await tauriApi.deleteNetwork(networkId);
      toast.success("Network deleted successfully");
      setShowDeleteModal(false);
      setSelectedNetworkId(null);
      refreshNetworks();
    } catch (e) {
      handleError(e, "Failed to delete network");
    }
  };

  if (!isDaemonRunning) {
    return (
      <div className="h-screen w-screen bg-gc-dark-900 flex flex-col items-center justify-center text-white">
        <div className="text-6xl mb-4 animate-bounce">🦖</div>
        <h1 className="text-2xl font-bold mb-2">Daemon Not Running</h1>
        <p className="text-gray-400 mb-6">Please start the GoConnect service via terminal.</p>
        <button
          onClick={checkDaemon}
          className="px-6 py-2 bg-gc-primary rounded hover:bg-opacity-80 transition"
        >
          Retry Connection
        </button>
      </div>
    );
  }

  if (!isRegistered) {
    return (
      <>
        <Toaster />
        <Onboarding onComplete={checkDaemon} />
      </>
    );
  }

  const selectedNetwork = networks.find(n => n.id === selectedNetworkId);
  const selfPeer = peers.find(p => p.is_self) || peers.find(p => p.virtual_ip === '127.0.0.1');

  return (
    <div className="h-screen w-screen flex bg-gc-dark-700 text-white font-sans">
      <Toaster />

      <Sidebar
        networks={networks}
        selectedNetworkId={selectedNetworkId}
        peers={peers}
        onSelectNetwork={setSelectedNetworkId}
        onShowCreate={() => setShowCreateModal(true)}
        onShowJoin={() => setShowJoinModal(true)}
      />


      <NetworkDetails
        selectedNetwork={selectedNetwork}
        selfPeer={selfPeer}
        isOwner={Boolean(selectedNetwork && selfPeer && selectedNetwork.owner_id === selfPeer.id)}
        onGenerateInvite={handleGenerateInvite}
        onLeaveNetwork={handleLeaveNetwork}
        onRenameNetwork={() => setShowRenameModal(true)}
        onDeleteNetwork={() => setShowDeleteModal(true)}
        setActiveTab={setActiveTab}
      />


      {/* MAIN CONTENT */}
      <div className="flex-1 flex flex-col bg-gc-dark-700">
        <div className="h-12 px-6 flex items-center border-b border-gc-dark-800 shadow-sm gap-6">
          <button
            onClick={() => { setActiveTab("peers"); setPrivateChatRecipient(null); }}
            className={`h-full border-b-2 font-semibold transition-colors ${activeTab === "peers" ? "border-gc-primary text-white" : "border-transparent text-gray-400 hover:text-gray-200"}`}
          >
            Connected Peers
          </button>
          <button
            onClick={() => setActiveTab("members")}
            className={`h-full border-b-2 font-semibold transition-colors ${activeTab === "members" ? "border-gc-primary text-white" : "border-transparent text-gray-400 hover:text-gray-200"}`}
          >
            👥 Members
          </button>
          <button
            onClick={() => setActiveTab("chat")}
            className={`h-full border-b-2 font-semibold transition-colors ${activeTab === "chat" ? "border-gc-primary text-white" : "border-transparent text-gray-400 hover:text-gray-200"}`}
          >
            {privateChatRecipient ? `Chat: ${privateChatRecipient.name}` : `Chat: ${selectedNetwork ? selectedNetwork.name : 'Network'}`}
          </button>
          <button
            onClick={() => setActiveTab("voice")}
            className={`h-full border-b-2 font-semibold transition-colors ${activeTab === "voice" ? "border-gc-primary text-white" : "border-transparent text-gray-400 hover:text-gray-200"}`}
          >
            Voice
          </button>
          <button
            onClick={() => setActiveTab("files")}
            className={`h-full border-b-2 font-semibold transition-colors ${activeTab === "files" ? "border-gc-primary text-white" : "border-transparent text-gray-400 hover:text-gray-200"}`}
          >
            Files
          </button>

          <button
            onClick={() => setActiveTab("metrics")}
            className={`h-full border-b-2 font-semibold transition-colors ${activeTab === "metrics" ? "border-gc-primary text-white" : "border-transparent text-gray-400 hover:text-gray-200"}`}
          >
            📊 Metrics
          </button>
          <button
            onClick={() => setActiveTab("settings")}
            className={`h-full border-b-2 font-semibold transition-colors ${activeTab === "settings" ? "border-gc-primary text-white" : "border-transparent text-gray-400 hover:text-gray-200"}`}
          >
            ⚙️ Settings
          </button>
        </div>

        <div className="flex-1 p-6 overflow-y-auto">
          {!selectedNetwork ? (
            <div className="flex flex-col items-center justify-center h-full text-gray-500">
              <div className="text-6xl mb-4">👈</div>
              <p>Select a network to view peers</p>
            </div>
          ) : (
            <>
              {activeTab === 'peers' && (
                <div className="space-y-3">
                  {peers.length === 0 ? (
                    <div className="text-center text-gray-500 mt-10">
                      <div className="text-4xl mb-2">👥</div>
                      No peers found. Invite someone to join!
                    </div>
                  ) : (
                    peers.map(peer => {
                      // Determine connection status
                      const isConnecting = !peer.connected && peer.latency_ms === 0 && !peer.is_self;
                      const statusColor = peer.connected
                        ? 'bg-green-500'
                        : isConnecting
                          ? 'bg-yellow-500 animate-pulse'
                          : 'bg-gray-500';
                      const statusText = peer.is_self
                        ? 'You'
                        : peer.connected
                          ? 'Online'
                          : isConnecting
                            ? 'Connecting...'
                            : 'Offline';

                      return (
                        <div
                          key={peer.id}
                          className={`bg-gc-dark-800 p-4 rounded-lg flex items-center justify-between border ${peer.is_self ? 'border-gc-primary/50' : 'border-gc-dark-600'
                            }`}
                        >
                          <div className="flex items-center gap-4">
                            <div className={`w-3 h-3 rounded-full ${statusColor}`} />
                            <div>
                              <div className="font-medium text-white flex items-center gap-2">
                                {peer.name || peer.display_name || "Unknown Peer"}
                                {peer.is_self && (
                                  <span className="text-[10px] px-1.5 py-0.5 bg-gc-primary/20 text-gc-primary rounded">
                                    You
                                  </span>
                                )}
                              </div>
                              <div className="text-sm text-gray-400 font-mono flex items-center gap-2">
                                {peer.virtual_ip}
                                <span className="text-[10px] text-gray-500">• {statusText}</span>
                              </div>
                            </div>
                          </div>
                          <div className="flex gap-2 items-center">
                            {/* Connection type badge */}
                            {peer.connected && !peer.is_self && (
                              <div className={`px-2 py-0.5 rounded text-[10px] font-medium ${peer.is_relay
                                ? 'bg-yellow-600/20 text-yellow-400'
                                : 'bg-green-600/20 text-green-400'
                                }`}>
                                {peer.is_relay ? '🔄 Relay' : '🔗 Direct'}
                              </div>
                            )}
                            {!peer.is_self && peer.connected && (
                              <button
                                onClick={async () => {
                                  try {
                                    const file = await open({
                                      title: 'Select file to send',
                                      multiple: false,
                                    });
                                    if (!file) return;
                                    // file can be string (single) or string[] (multiple) or null
                                    const path = Array.isArray(file) ? file[0] : file;
                                    if (!path) return;
                                    // Note: Size validation happens in daemon
                                    await tauriApi.sendFile(peer.id, path);
                                    toast.success(`Sending file to ${peer.name || 'peer'}...`);
                                    setActiveTab('files');
                                  } catch (e) {
                                    handleError(e, 'Failed to send file');
                                  }
                                }}
                                className="px-2 py-1 bg-gc-dark-700 text-gray-300 rounded text-xs hover:bg-gc-dark-600 transition-all"
                              >
                                📁 Send
                              </button>
                            )}
                            {!peer.is_self && peer.connected && (
                              <button
                                onClick={() => { setPrivateChatRecipient(peer); setActiveTab("chat"); }}
                                className="px-3 py-1 bg-gc-primary rounded text-xs text-white hover:bg-opacity-90 transition-all font-medium"
                              >
                                Message
                              </button>
                            )}
                            <div className={`px-3 py-1 bg-gc-dark-900 rounded text-xs ${peer.connected ? 'text-green-400' : 'text-gray-500'
                              }`}>
                              {peer.latency_ms > 0 ? `${peer.latency_ms}ms` : '---'}
                            </div>
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>

              )}

              {activeTab === 'members' && (
                <MembersTab
                  network={selectedNetwork}
                  selfUserId={selfPeer?.id || ''}
                />
              )}

              {activeTab === 'chat' && selectedNetworkId && (

                <div className="h-full relative">
                  {privateChatRecipient && (
                    <button
                      onClick={() => setPrivateChatRecipient(null)}
                      className="absolute top-3 right-3 z-10 text-xs bg-gc-dark-900 text-gray-400 hover:text-white px-2 py-1 rounded border border-gc-dark-700 hover:border-gc-dark-500 transition-colors"
                    >
                      ← Back to Global
                    </button>
                  )}
                  <ChatPanel
                    networkId={selectedNetworkId}
                    recipientId={privateChatRecipient?.id}
                    recipientName={privateChatRecipient?.name}
                  />
                </div>
              )}

              {activeTab === 'voice' && selectedNetworkId && (
                <div className="h-full">
                  <VoiceChat networkId={selectedNetworkId} selfPeer={selfPeer} connectedPeers={peers} />
                </div>
              )}


              {activeTab === 'files' && (
                <div className="h-full">
                  <FileTransferPanel />
                </div>
              )}

              {activeTab === 'settings' && (
                <div className="h-full">
                  <SettingsPanel selectedNetwork={selectedNetwork} />
                </div>
              )}

              {activeTab === 'metrics' && (
                <div className="h-full">
                  <MetricsDashboard />
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* MODALS */}
      <CreateNetworkModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateNetwork}
      />

      <JoinNetworkModal
        isOpen={showJoinModal}
        onClose={() => setShowJoinModal(false)}
        onSubmit={handleJoinNetwork}
        initialCode={joinInviteCode}
      />

      <RenameNetworkModal
        isOpen={showRenameModal}
        currentName={selectedNetwork?.name || ""}
        networkId={selectedNetwork?.id || ""}
        onClose={() => setShowRenameModal(false)}
        onSubmit={handleRenameNetwork}
      />

      <DeleteNetworkModal
        isOpen={showDeleteModal}
        networkName={selectedNetwork?.name || ""}
        networkId={selectedNetwork?.id || ""}
        onClose={() => setShowDeleteModal(false)}
        onSubmit={handleDeleteNetwork}
      />

    </div>
  );
}
