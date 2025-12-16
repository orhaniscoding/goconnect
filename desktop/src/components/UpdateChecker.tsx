import { useState, useEffect } from 'react';
import { check } from '@tauri-apps/plugin-updater';
import { relaunch } from '@tauri-apps/plugin-process';

interface UpdateInfo {
    version: string;
    date?: string;
    body?: string;
}

export default function UpdateChecker() {
    const [checking, setChecking] = useState(false);
    const [updateAvailable, setUpdateAvailable] = useState<UpdateInfo | null>(null);
    const [downloading, setDownloading] = useState(false);
    const [progress, setProgress] = useState(0);
    const [error, setError] = useState<string | null>(null);

    const checkForUpdates = async () => {
        setChecking(true);
        setError(null);

        try {
            const update = await check();
            if (update) {
                setUpdateAvailable({
                    version: update.version,
                    date: update.date,
                    body: update.body,
                });
            } else {
                setUpdateAvailable(null);
            }
        } catch (e) {
            console.error('Update check failed:', e);
            setError('Failed to check for updates');
        } finally {
            setChecking(false);
        }
    };

    const downloadAndInstall = async () => {
        if (!updateAvailable) return;

        setDownloading(true);
        setProgress(0);

        try {
            const update = await check();
            if (update) {
                let downloaded = 0;
                let contentLength = 0;

                await update.downloadAndInstall((event) => {
                    switch (event.event) {
                        case 'Started':
                            contentLength = event.data.contentLength || 0;
                            break;
                        case 'Progress':
                            downloaded += event.data.chunkLength;
                            if (contentLength > 0) {
                                setProgress((downloaded / contentLength) * 100);
                            }
                            break;
                        case 'Finished':
                            setProgress(100);
                            break;
                    }
                });

                // Prompt for relaunch
                await relaunch();
            }
        } catch (e) {
            console.error('Update failed:', e);
            setError('Failed to install update');
        } finally {
            setDownloading(false);
        }
    };

    // Check on component mount
    useEffect(() => {
        checkForUpdates();
    }, []);

    if (checking) {
        return (
            <div className="bg-gc-dark-700 rounded-lg p-4 border border-gc-dark-600">
                <div className="flex items-center gap-3">
                    <div className="animate-spin">⏳</div>
                    <span className="text-gray-400">Checking for updates...</span>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-red-500/10 rounded-lg p-4 border border-red-500/30">
                <div className="flex items-center justify-between">
                    <span className="text-red-400">{error}</span>
                    <button
                        onClick={checkForUpdates}
                        className="text-sm text-red-400 hover:text-red-300 underline"
                    >
                        Retry
                    </button>
                </div>
            </div>
        );
    }

    if (!updateAvailable) {
        return (
            <div className="bg-gc-dark-700 rounded-lg p-4 border border-gc-dark-600">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <span className="text-green-400">✓</span>
                        <span className="text-gray-300">You're up to date!</span>
                    </div>
                    <button
                        onClick={checkForUpdates}
                        className="text-sm text-gc-primary hover:text-gc-primary/80"
                    >
                        Check again
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="bg-gc-primary/10 rounded-lg p-4 border border-gc-primary/30">
            <div className="flex items-center justify-between mb-3">
                <div>
                    <div className="flex items-center gap-2">
                        <span className="text-gc-primary text-lg">🆕</span>
                        <span className="text-white font-medium">
                            Update Available: v{updateAvailable.version}
                        </span>
                    </div>
                    {updateAvailable.date && (
                        <div className="text-xs text-gray-400 mt-1">
                            Released: {new Date(updateAvailable.date).toLocaleDateString()}
                        </div>
                    )}
                </div>

                {!downloading ? (
                    <button
                        onClick={downloadAndInstall}
                        className="px-4 py-2 bg-gc-primary text-white rounded hover:bg-gc-primary/80 transition font-medium"
                    >
                        Install Update
                    </button>
                ) : (
                    <div className="text-sm text-gray-400">
                        Downloading... {progress.toFixed(0)}%
                    </div>
                )}
            </div>

            {downloading && (
                <div className="w-full bg-gc-dark-600 rounded-full h-2">
                    <div
                        className="bg-gc-primary h-2 rounded-full transition-all duration-300"
                        style={{ width: `${progress}%` }}
                    />
                </div>
            )}

            {updateAvailable.body && (
                <details className="mt-3">
                    <summary className="text-sm text-gray-400 cursor-pointer hover:text-gray-300">
                        Release Notes
                    </summary>
                    <div className="mt-2 text-sm text-gray-300 bg-gc-dark-800 p-3 rounded max-h-32 overflow-y-auto">
                        {updateAvailable.body}
                    </div>
                </details>
            )}
        </div>
    );
}
