import { useState, useEffect } from 'react';
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
import { CreateNetworkModal, JoinNetworkModal } from './components/NetworkModals';

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
  const [activeTab, setActiveTab] = useState<"peers" | "chat" | "files" | "settings" | "voice" | "metrics">("peers");

  const toast = useToast();

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
      refreshNetworks();
    } catch (e) {
      handleError(e, "Failed to join network");
    }
  };

  const handleLeaveNetwork = async () => {
    if (!selectedNetworkId) return;
    if (!confirm("Start leaving network?")) return;
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
        onSelectNetwork={setSelectedNetworkId}
        onShowCreate={() => setShowCreateModal(true)}
        onShowJoin={() => setShowJoinModal(true)}
      />

      <NetworkDetails
        selectedNetwork={selectedNetwork}
        selfPeer={selfPeer}
        onGenerateInvite={handleGenerateInvite}
        onLeaveNetwork={handleLeaveNetwork}
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
                      No peers found. Invite someone!
                    </div>
                  ) : (
                    peers.map(peer => (
                      <div key={peer.id} className="bg-gc-dark-800 p-4 rounded-lg flex items-center justify-between border border-gc-dark-600">
                        <div className="flex items-center gap-4">
                          <div className={`w-3 h-3 rounded-full ${peer.connected ? 'bg-green-500' : 'bg-gray-500'}`} />
                          <div>
                            <div className="font-medium text-white">{peer.name || peer.display_name || "Unknown Peer"}</div>
                            <div className="text-sm text-gray-400 font-mono">{peer.virtual_ip}</div>
                          </div>
                        </div>
                        <div className="flex gap-2">
                          {!peer.is_self && (
                            <button
                              onClick={() => { setPrivateChatRecipient(peer); setActiveTab("chat"); }}
                              className="px-3 py-1 bg-gc-primary rounded text-xs text-white hover:bg-opacity-90 transition-all font-medium"
                            >
                              Message
                            </button>
                          )}
                          <div className="px-3 py-1 bg-gc-dark-900 rounded text-xs text-gray-400">
                            {peer.latency_ms > 0 ? `${peer.latency_ms}ms` : '---'}
                          </div>
                        </div>
                      </div>
                    ))
                  )}
                </div>
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
                  <VoiceChat networkId={selectedNetworkId} selfPeer={selfPeer} />
                </div>
              )}

              {activeTab === 'files' && (
                <div className="h-full">
                  <FileTransferPanel />
                </div>
              )}

              {activeTab === 'settings' && (
                <div className="h-full">
                  <SettingsPanel />
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
      />

    </div>
  );
}
