import { useState, useEffect } from 'react';
import { tauriApi, Settings, NetworkInfo } from '../lib/tauri-api';
import { handleError } from '../lib/utils';
import { useToast } from './Toast';
import BannedMembersPanel from './BannedMembersPanel';

interface SettingsPanelProps {
    selectedNetwork?: NetworkInfo | null;
}

export default function SettingsPanel({ selectedNetwork = null }: SettingsPanelProps) {
    const [settings, setSettings] = useState<Settings | null>(null);
    const [loading, setLoading] = useState(false);
    const toast = useToast();

    useEffect(() => {
        loadSettings();
    }, []);

    const loadSettings = async () => {
        setLoading(true);
        try {
            const s = await tauriApi.getSettings();
            setSettings(s);
        } catch (e) {
            handleError(e, "Failed to load settings");
        } finally {
            setLoading(false);
        }
    };

    const handleToggle = async (key: keyof Settings) => {
        if (!settings) return;

        const newSettings = { ...settings, [key]: !settings[key] };
        setSettings(newSettings); // Optimistic update

        try {
            await tauriApi.updateSettings(newSettings);
            toast.success("Settings updated");
        } catch (e) {
            handleError(e, "Failed to update settings");
            setSettings(settings); // Revert
        }
    };

    const handleReset = async () => {
        if (!confirm("Are you sure you want to reset all settings to default?")) return;

        try {
            const defaults = await tauriApi.resetSettings();
            setSettings(defaults);
            toast.success("Settings reset to defaults");
        } catch (e) {
            handleError(e, "Failed to reset settings");
        }
    };

    if (loading && !settings) {
        return <div className="p-8 text-center text-gray-500">Loading settings...</div>;
    }

    if (!settings) {
        return <div className="p-8 text-center text-red-400">Failed to load settings.</div>;
    }

    return (
        <div className="flex flex-col h-full bg-gc-dark-800 rounded-lg border border-gc-dark-600 overflow-hidden">
            <div className="p-6 border-b border-gc-dark-700 flex justify-between items-center">
                <h2 className="text-xl font-bold text-white">Settings</h2>
                <button
                    onClick={handleReset}
                    className="text-xs text-red-400 hover:text-red-300 hover:underline"
                >
                    Reset to Defaults
                </button>
            </div>

            <div className="p-6 space-y-6">
                {/* Auto Connect */}
                <div className="flex items-center justify-between">
                    <div>
                        <div className="font-medium text-white">Auto Connect</div>
                        <div className="text-sm text-gray-400">Automatically connect to the network on startup</div>
                    </div>
                    <button
                        type="button"
                        role="switch"
                        aria-checked={settings.auto_connect}
                        aria-label="Toggle Auto Connect"
                        onClick={() => handleToggle('auto_connect')}
                        className={`w-12 h-6 rounded-full p-1 transition-colors duration-200 ease-in-out ${settings.auto_connect ? 'bg-gc-primary' : 'bg-gc-dark-600'}`}
                    >
                        <div className={`w-4 h-4 rounded-full bg-white shadow-sm transform transition-transform duration-200 ease-in-out ${settings.auto_connect ? 'translate-x-6' : 'translate-x-0'}`} />
                    </button>
                </div>

                {/* Notifications */}
                <div className="flex items-center justify-between">
                    <div>
                        <div className="font-medium text-white">Notifications</div>
                        <div className="text-sm text-gray-400">Show desktop notifications for new messages and transfers</div>
                    </div>
                    <button
                        type="button"
                        role="switch"
                        aria-checked={settings.notifications_enabled}
                        aria-label="Toggle Notifications"
                        onClick={() => handleToggle('notifications_enabled')}
                        className={`w-12 h-6 rounded-full p-1 transition-colors duration-200 ease-in-out ${settings.notifications_enabled ? 'bg-gc-primary' : 'bg-gc-dark-600'}`}
                    >
                        <div className={`w-4 h-4 rounded-full bg-white shadow-sm transform transition-transform duration-200 ease-in-out ${settings.notifications_enabled ? 'translate-x-6' : 'translate-x-0'}`} />
                    </button>
                </div>

                {/* Notification Sound */}
                {settings.notifications_enabled && (
                    <div className="flex items-center justify-between ml-4 border-l-2 border-gc-dark-600 pl-4">
                        <div>
                            <div className="font-medium text-white">Notification Sound</div>
                            <div className="text-sm text-gray-400">Play sound when notifications arrive</div>
                        </div>
                        <button
                            type="button"
                            role="switch"
                            aria-checked={settings.notification_sound ?? true}
                            aria-label="Toggle Notification Sound"
                            onClick={() => handleToggle('notification_sound' as keyof Settings)}
                            className={`w-12 h-6 rounded-full p-1 transition-colors duration-200 ease-in-out ${settings.notification_sound ?? true ? 'bg-gc-primary' : 'bg-gc-dark-600'}`}
                        >
                            <div className={`w-4 h-4 rounded-full bg-white shadow-sm transform transition-transform duration-200 ease-in-out ${settings.notification_sound ?? true ? 'translate-x-6' : 'translate-x-0'}`} />
                        </button>
                    </div>
                )}

                {/* Do Not Disturb */}
                <div className="flex items-center justify-between">
                    <div>
                        <div className="font-medium text-white">Do Not Disturb</div>
                        <div className="text-sm text-gray-400">Silence all notifications temporarily</div>
                    </div>
                    <button
                        type="button"
                        role="switch"
                        aria-checked={settings.do_not_disturb ?? false}
                        aria-label="Toggle Do Not Disturb"
                        onClick={() => handleToggle('do_not_disturb' as keyof Settings)}
                        className={`w-12 h-6 rounded-full p-1 transition-colors duration-200 ease-in-out ${settings.do_not_disturb ? 'bg-amber-500' : 'bg-gc-dark-600'}`}
                    >
                        <div className={`w-4 h-4 rounded-full bg-white shadow-sm transform transition-transform duration-200 ease-in-out ${settings.do_not_disturb ? 'translate-x-6' : 'translate-x-0'}`} />
                    </button>
                </div>

                {/* Start Minimized */}
                <div className="flex items-center justify-between">
                    <div>
                        <div className="font-medium text-white">Start Minimized</div>
                        <div className="text-sm text-gray-400">Launch application in the system tray</div>
                    </div>
                    <button
                        type="button"
                        role="switch"
                        aria-checked={settings.start_minimized}
                        aria-label="Toggle Start Minimized"
                        onClick={() => handleToggle('start_minimized')}
                        className={`w-12 h-6 rounded-full p-1 transition-colors duration-200 ease-in-out ${settings.start_minimized ? 'bg-gc-primary' : 'bg-gc-dark-600'}`}
                    >
                        <div className={`w-4 h-4 rounded-full bg-white shadow-sm transform transition-transform duration-200 ease-in-out ${settings.start_minimized ? 'translate-x-6' : 'translate-x-0'}`} />
                    </button>
                </div>

                {/* Banned Members - only shown when a network is selected */}
                {selectedNetwork && (
                    <div className="pt-6 border-t border-gc-dark-700">
                        <BannedMembersPanel network={selectedNetwork} />
                    </div>
                )}

                <div className="pt-6 border-t border-gc-dark-700">
                    <div className="text-xs text-gray-500 uppercase font-semibold mb-3">Updates</div>
                    <UpdateCheckerSection />
                </div>

                <div className="pt-6 border-t border-gc-dark-700">
                    <div className="text-xs text-gray-500 uppercase font-semibold mb-2">Application Info</div>
                    <div className="grid grid-cols-2 gap-2 text-sm">
                        <div className="text-gray-400">Version</div>
                        <div className="text-white text-right font-mono">v1.3.0</div>
                        <div className="text-gray-400">Daemon Status</div>
                        <div className="text-green-400 text-right">Connected via gRPC</div>
                    </div>
                </div>
            </div>
        </div>
    );
}

// Lazy-loaded UpdateChecker to avoid import errors in dev mode
function UpdateCheckerSection() {
    const [UpdateChecker, setUpdateChecker] = useState<React.ComponentType | null>(null);

    useEffect(() => {
        import('./UpdateChecker')
            .then((mod) => setUpdateChecker(() => mod.default))
            .catch(() => setUpdateChecker(null));
    }, []);

    if (!UpdateChecker) {
        return (
            <div className="text-sm text-gray-400">
                Update checking available in production build
            </div>
        );
    }

    return <UpdateChecker />;
}
