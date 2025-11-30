import { useState, useEffect } from "react";

// Types
interface Tenant {
  id: string;
  name: string;
  icon?: string;
  unread?: number;
}

interface Network {
  id: string;
  name: string;
  subnet: string;
  connected: boolean;
  memberCount: number;
}

interface User {
  id: string;
  email: string;
  name: string;
}

// Mock data - will be replaced with API calls
const mockTenants: Tenant[] = [
  { id: "1", name: "Gaming Server", icon: "🎮", unread: 3 },
  { id: "2", name: "Work VPN", icon: "💼" },
  { id: "3", name: "Friends Network", icon: "👥", unread: 1 },
];

const mockNetworks: Network[] = [
  { id: "1", name: "Minecraft LAN", subnet: "10.0.1.0/24", connected: true, memberCount: 5 },
  { id: "2", name: "File Sharing", subnet: "10.0.2.0/24", connected: false, memberCount: 3 },
  { id: "3", name: "Game Servers", subnet: "10.0.3.0/24", connected: false, memberCount: 8 },
];

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null);
  const [networks, setNetworks] = useState<Network[]>([]);
  const [selectedNetwork, setSelectedNetwork] = useState<Network | null>(null);
  const [daemonStatus, setDaemonStatus] = useState<"connected" | "disconnected" | "connecting">("disconnected");

  // Login form
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [serverUrl, setServerUrl] = useState("http://localhost:8081");
  const [loginError, setLoginError] = useState("");

  // Check daemon status
  useEffect(() => {
    const checkDaemon = async () => {
      try {
        const response = await fetch("http://127.0.0.1:12345/status");
        if (response.ok) {
          setDaemonStatus("connected");
        }
      } catch {
        setDaemonStatus("disconnected");
      }
    };

    checkDaemon();
    const interval = setInterval(checkDaemon, 5000);
    return () => clearInterval(interval);
  }, []);

  // Handle login
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginError("");

    try {
      const response = await fetch(`${serverUrl}/api/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        throw new Error("Invalid credentials");
      }

      const data = await response.json();
      setUser(data.user);
      localStorage.setItem("token", data.access_token);
      localStorage.setItem("serverUrl", serverUrl);
      setIsLoggedIn(true);

      // Load tenants
      setTenants(mockTenants); // TODO: Replace with API call
    } catch (err) {
      setLoginError("Login failed. Check your credentials and server URL.");
    }
  };

  // Handle tenant selection
  const handleSelectTenant = (tenant: Tenant) => {
    setSelectedTenant(tenant);
    setNetworks(mockNetworks); // TODO: Replace with API call
    setSelectedNetwork(null);
  };

  // Handle network connection toggle
  const handleToggleConnection = async (network: Network) => {
    if (daemonStatus !== "connected") {
      alert("Daemon is not running. Please start the GoConnect daemon.");
      return;
    }

    // TODO: Call daemon API to connect/disconnect
    setNetworks(networks.map(n =>
      n.id === network.id ? { ...n, connected: !n.connected } : n
    ));
  };

  // Login Screen
  if (!isLoggedIn) {
    return (
      <div className="h-screen w-screen bg-gc-dark-700 flex items-center justify-center">
        <div className="bg-gc-dark-800 p-8 rounded-lg shadow-xl w-[400px]">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-white mb-2">GoConnect</h1>
            <p className="text-gray-400">Virtual LAN made simple</p>
          </div>

          <form onSubmit={handleLogin} className="space-y-4">
            <div>
              <label className="block text-sm text-gray-300 mb-1">Server URL</label>
              <input
                type="url"
                value={serverUrl}
                onChange={(e) => setServerUrl(e.target.value)}
                className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none"
                placeholder="http://localhost:8081"
              />
            </div>

            <div>
              <label className="block text-sm text-gray-300 mb-1">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none"
                placeholder="you@example.com"
                required
              />
            </div>

            <div>
              <label className="block text-sm text-gray-300 mb-1">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-3 py-2 bg-gc-dark-900 border border-gc-dark-600 rounded text-white focus:border-gc-primary focus:outline-none"
                placeholder="••••••••"
                required
              />
            </div>

            {loginError && (
              <div className="text-gc-red text-sm">{loginError}</div>
            )}

            <button
              type="submit"
              className="w-full py-2 bg-gc-primary hover:bg-gc-primary/80 text-white font-medium rounded transition"
            >
              Login
            </button>

            <p className="text-center text-gray-400 text-sm">
              Don't have an account?{" "}
              <a href="#" className="text-gc-primary hover:underline">Register</a>
            </p>
          </form>

          {/* Daemon Status */}
          <div className="mt-6 pt-4 border-t border-gc-dark-600">
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-400">Daemon Status:</span>
              <span className={`flex items-center gap-1 ${daemonStatus === "connected" ? "text-gc-green" : "text-gc-red"
                }`}>
                <span className={`w-2 h-2 rounded-full ${daemonStatus === "connected" ? "bg-gc-green" : "bg-gc-red"
                  }`}></span>
                {daemonStatus === "connected" ? "Running" : "Not Running"}
              </span>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Main App Screen
  return (
    <div className="h-screen w-screen flex bg-gc-dark-700">
      {/* Server/Tenant Sidebar (Left) */}
      <div className="w-[72px] bg-gc-dark-900 flex flex-col items-center py-3 gap-2">
        {/* Home Button */}
        <button className="w-12 h-12 bg-gc-primary rounded-2xl hover:rounded-xl transition-all flex items-center justify-center text-white text-xl">
          🏠
        </button>

        <div className="w-8 h-[2px] bg-gc-dark-600 rounded-full my-1"></div>

        {/* Tenant List */}
        {tenants.map((tenant) => (
          <button
            key={tenant.id}
            onClick={() => handleSelectTenant(tenant)}
            className={`relative w-12 h-12 rounded-2xl hover:rounded-xl transition-all flex items-center justify-center text-2xl ${selectedTenant?.id === tenant.id
              ? "bg-gc-primary rounded-xl"
              : "bg-gc-dark-700 hover:bg-gc-primary"
              }`}
            title={tenant.name}
          >
            {tenant.icon || tenant.name[0]}
            {tenant.unread && (
              <span className="absolute -bottom-1 -right-1 w-5 h-5 bg-gc-red rounded-full text-xs text-white flex items-center justify-center">
                {tenant.unread}
              </span>
            )}
          </button>
        ))}

        {/* Add Server Button */}
        <button className="w-12 h-12 bg-gc-dark-700 rounded-2xl hover:rounded-xl hover:bg-gc-green transition-all flex items-center justify-center text-gc-green hover:text-white text-2xl">
          +
        </button>

        {/* Bottom spacer */}
        <div className="flex-1"></div>

        {/* User Avatar */}
        <button className="w-12 h-12 bg-gc-dark-700 rounded-full flex items-center justify-center text-xl">
          👤
        </button>
      </div>

      {/* Channel/Network Sidebar */}
      <div className="w-60 bg-gc-dark-800 flex flex-col">
        {/* Server Header */}
        <div className="h-12 px-4 flex items-center border-b border-gc-dark-900 shadow">
          <h2 className="font-semibold text-white truncate">
            {selectedTenant?.name || "Select a Server"}
          </h2>
        </div>

        {/* Network List */}
        <div className="flex-1 overflow-y-auto p-2">
          {selectedTenant ? (
            <>
              <div className="text-xs text-gray-400 uppercase tracking-wide px-2 py-2">
                Networks
              </div>
              {networks.map((network) => (
                <div
                  key={network.id}
                  onClick={() => setSelectedNetwork(network)}
                  className={`flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer ${selectedNetwork?.id === network.id
                    ? "bg-gc-dark-600 text-white"
                    : "text-gray-400 hover:text-white hover:bg-gc-dark-700"
                    }`}
                >
                  <span className={`w-2 h-2 rounded-full ${network.connected ? "bg-gc-green" : "bg-gray-500"
                    }`}></span>
                  <span className="flex-1 truncate">{network.name}</span>
                  <span className="text-xs text-gray-500">{network.memberCount}</span>
                </div>
              ))}

              <div className="text-xs text-gray-400 uppercase tracking-wide px-2 py-2 mt-4">
                Chat Channels
              </div>
              <div className="flex items-center gap-2 px-2 py-1.5 rounded text-gray-400 hover:text-white hover:bg-gc-dark-700 cursor-pointer">
                <span>#</span>
                <span>general</span>
              </div>
              <div className="flex items-center gap-2 px-2 py-1.5 rounded text-gray-400 hover:text-white hover:bg-gc-dark-700 cursor-pointer">
                <span>#</span>
                <span>help</span>
              </div>
            </>
          ) : (
            <div className="text-center text-gray-500 mt-8">
              Select a server from the left
            </div>
          )}
        </div>

        {/* User Panel */}
        <div className="h-14 bg-gc-dark-900 px-2 flex items-center gap-2">
          <div className="w-8 h-8 bg-gc-primary rounded-full flex items-center justify-center text-white">
            {user?.name?.[0] || "U"}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm text-white truncate">{user?.name || "User"}</div>
            <div className="text-xs text-gray-400 truncate">{user?.email}</div>
          </div>
          <button
            onClick={() => setIsLoggedIn(false)}
            className="text-gray-400 hover:text-white p-1"
            title="Logout"
          >
            🚪
          </button>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="h-12 px-4 flex items-center border-b border-gc-dark-900 shadow bg-gc-dark-700">
          {selectedNetwork ? (
            <>
              <span className="text-white font-medium">{selectedNetwork.name}</span>
              <span className="mx-2 text-gray-500">|</span>
              <span className="text-gray-400 text-sm">{selectedNetwork.subnet}</span>
              <div className="flex-1"></div>
              <button
                onClick={() => handleToggleConnection(selectedNetwork)}
                className={`px-4 py-1.5 rounded text-sm font-medium transition ${selectedNetwork.connected
                  ? "bg-gc-red hover:bg-gc-red/80 text-white"
                  : "bg-gc-green hover:bg-gc-green/80 text-white"
                  }`}
              >
                {selectedNetwork.connected ? "Disconnect" : "Connect"}
              </button>
            </>
          ) : (
            <span className="text-gray-400">Select a network</span>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 p-6 overflow-y-auto">
          {selectedNetwork ? (
            <div className="space-y-6">
              {/* Connection Status Card */}
              <div className="bg-gc-dark-800 rounded-lg p-4">
                <h3 className="text-white font-medium mb-3">Connection Status</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <div className="text-gray-400 text-sm">Status</div>
                    <div className={`font-medium ${selectedNetwork.connected ? "text-gc-green" : "text-gray-500"
                      }`}>
                      {selectedNetwork.connected ? "🟢 Connected" : "⚪ Disconnected"}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-400 text-sm">Your IP</div>
                    <div className="text-white font-mono">
                      {selectedNetwork.connected ? "10.0.1.15" : "-"}
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-400 text-sm">Subnet</div>
                    <div className="text-white font-mono">{selectedNetwork.subnet}</div>
                  </div>
                  <div>
                    <div className="text-gray-400 text-sm">Members Online</div>
                    <div className="text-white">{selectedNetwork.memberCount}</div>
                  </div>
                </div>
              </div>

              {/* Online Members */}
              <div className="bg-gc-dark-800 rounded-lg p-4">
                <h3 className="text-white font-medium mb-3">Online Members</h3>
                <div className="space-y-2">
                  {[
                    { name: "Alice", ip: "10.0.1.2", status: "online" },
                    { name: "Bob", ip: "10.0.1.5", status: "online" },
                    { name: "Charlie", ip: "10.0.1.8", status: "idle" },
                  ].map((member, i) => (
                    <div key={i} className="flex items-center gap-3 p-2 rounded hover:bg-gc-dark-700">
                      <div className="relative">
                        <div className="w-8 h-8 bg-gc-primary rounded-full flex items-center justify-center text-white">
                          {member.name[0]}
                        </div>
                        <span className={`absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-gc-dark-800 ${member.status === "online" ? "bg-gc-green" : "bg-gc-yellow"
                          }`}></span>
                      </div>
                      <div className="flex-1">
                        <div className="text-white text-sm">{member.name}</div>
                        <div className="text-gray-400 text-xs font-mono">{member.ip}</div>
                      </div>
                      <button className="text-gray-400 hover:text-white text-sm">
                        Ping
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ) : selectedTenant ? (
            <div className="text-center text-gray-400 mt-20">
              <div className="text-6xl mb-4">🌐</div>
              <h3 className="text-xl text-white mb-2">Welcome to {selectedTenant.name}</h3>
              <p>Select a network from the sidebar to connect</p>
            </div>
          ) : (
            <div className="text-center text-gray-400 mt-20">
              <div className="text-6xl mb-4">👋</div>
              <h3 className="text-xl text-white mb-2">Welcome to GoConnect</h3>
              <p>Select a server from the left sidebar to get started</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
