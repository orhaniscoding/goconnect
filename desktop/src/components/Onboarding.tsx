import { useState } from 'react';
import { tauriApi } from '../lib/tauri-api';
import { handleError } from '../lib/utils';
import { useToast } from './Toast';

interface Props {
    onComplete: () => void;
}

export default function Onboarding({ onComplete }: Props) {
    const [token, setToken] = useState("");
    const [name, setName] = useState("");
    const [loading, setLoading] = useState(false);
    const toast = useToast();

    const handleConnect = async () => {
        if (!token.trim()) {
            toast.error("Please enter a valid token");
            return;
        }

        setLoading(true);
        try {
            await tauriApi.register(token.trim(), name.trim() || undefined);
            toast.success("Device registered successfully!");
            onComplete();
        } catch (e) {
            handleError(e, "Registration failed");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="h-screen w-screen bg-gc-dark-900 flex flex-col items-center justify-center text-white relative overflow-hidden">
            {/* Background blobs */}
            <div className="absolute top-[-20%] left-[-10%] w-[500px] h-[500px] bg-gc-primary/20 rounded-full blur-[100px]" />
            <div className="absolute bottom-[-20%] right-[-10%] w-[500px] h-[500px] bg-blue-600/20 rounded-full blur-[100px]" />

            <div className="z-10 w-full max-w-md p-8 bg-gc-dark-800/80 backdrop-blur-xl rounded-2xl border border-gc-dark-600 shadow-2xl flex flex-col items-center">
                <div className="w-16 h-16 bg-gc-primary rounded-2xl flex items-center justify-center text-3xl font-bold mb-6 shadow-lg shadow-gc-primary/30">
                    GC
                </div>

                <h1 className="text-2xl font-bold mb-2">Welcome to GoConnect</h1>
                <p className="text-gray-400 text-center mb-8">
                    Secure, peer-to-peer networking made simple. <br />
                    Please authenticate this device to continue.
                </p>

                <div className="w-full space-y-4">
                    <div>
                        <label htmlFor="token-input" className="block text-xs font-semibold uppercase text-gray-500 mb-2">Device Token</label>
                        <input
                            id="token-input"
                            aria-label="Device Token"
                            value={token}
                            onChange={(e) => setToken(e.target.value)}
                            placeholder="eyJhbGciOiJIUzI1NiIsIn..."
                            className="w-full bg-gc-dark-900 border border-gc-dark-700 rounded-lg px-4 py-3 text-white placeholder-gray-600 focus:border-gc-primary focus:ring-1 focus:ring-gc-primary outline-none transition-all font-mono text-sm"
                        />
                    </div>

                    <div>
                        <label htmlFor="name-input" className="block text-xs font-semibold uppercase text-gray-500 mb-2">Device Name (Optional)</label>
                        <input
                            id="name-input"
                            aria-label="Device Name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="My Desktop"
                            className="w-full bg-gc-dark-900 border border-gc-dark-700 rounded-lg px-4 py-3 text-white placeholder-gray-600 focus:border-gc-primary focus:ring-1 focus:ring-gc-primary outline-none transition-all font-sans text-sm"
                        />
                    </div>

                    <button
                        onClick={handleConnect}
                        disabled={loading}
                        aria-busy={loading}
                        className={`w-full py-3 rounded-lg font-bold text-white transition-all transform active:scale-95 shadow-lg
                            ${loading ? 'bg-gc-dark-600 cursor-not-allowed' : 'bg-gc-primary hover:bg-opacity-90 hover:shadow-gc-primary/40'}
                        `}
                    >
                        {loading ? 'Connecting...' : 'Connect Device'}
                    </button>
                </div>

                <div className="mt-8 text-xs text-gray-500 text-center">
                    <p>Don't have a token?</p>
                    <a
                        href="https://dashboard.goconnect.io"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-gc-primary hover:underline"
                    >
                        Log in to the Web Dashboard
                    </a>
                </div>
            </div>

            <div className="absolute bottom-4 text-xs text-gray-600">
                GoConnect Client v0.1.0
            </div>
        </div>
    );
}
