import { useState, useEffect, useCallback } from 'react';

// =============================================================================
// Validation Utilities
// =============================================================================

/**
 * Validates a network name according to the spec rules:
 * - 3-50 characters after trimming
 * - Alphanumeric, spaces, hyphens, underscores only
 * - No leading/trailing hyphens or underscores
 */
export function validateNetworkName(name: string): { valid: boolean; error?: string; sanitized: string } {
    // Trim and collapse multiple spaces
    const sanitized = name.trim().replace(/\s+/g, ' ');

    if (sanitized.length < 3) {
        return { valid: false, error: 'Network name must be at least 3 characters', sanitized };
    }

    if (sanitized.length > 50) {
        return { valid: false, error: 'Network name must be at most 50 characters', sanitized };
    }

    // Check for leading/trailing special chars
    if (/^[-_]/.test(sanitized) || /[-_]$/.test(sanitized)) {
        return { valid: false, error: 'Network name cannot start or end with hyphens or underscores', sanitized };
    }

    // Check for valid characters only
    if (!/^[a-zA-Z0-9][a-zA-Z0-9 _-]*[a-zA-Z0-9]$/.test(sanitized) && !/^[a-zA-Z0-9]{1,2}$/.test(sanitized)) {
        return { valid: false, error: 'Network name can only contain letters, numbers, spaces, hyphens, and underscores', sanitized };
    }

    return { valid: true, sanitized };
}

// =============================================================================
// Create Network Modal
// =============================================================================

interface CreateModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (name: string) => Promise<{ inviteCode?: string } | void>;
}

export function CreateNetworkModal({ isOpen, onClose, onSubmit }: CreateModalProps) {
    const [name, setName] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [createdInviteCode, setCreatedInviteCode] = useState<string | null>(null);
    const [copied, setCopied] = useState(false);

    // Reset state when modal opens
    useEffect(() => {
        if (isOpen) {
            setName("");
            setError(null);
            setIsLoading(false);
            setCreatedInviteCode(null);
            setCopied(false);
        }
    }, [isOpen]);

    const handleNameChange = useCallback((value: string) => {
        setName(value);
        // Clear error when user starts typing
        if (error) setError(null);
    }, [error]);

    const handleValidation = useCallback(() => {
        const result = validateNetworkName(name);
        if (!result.valid) {
            setError(result.error || 'Invalid network name');
            return false;
        }
        return true;
    }, [name]);

    const handleSubmit = async () => {
        if (isLoading) return;

        if (!handleValidation()) return;

        setIsLoading(true);
        setError(null);

        try {
            const result = await onSubmit(name.trim());
            if (result?.inviteCode) {
                setCreatedInviteCode(result.inviteCode);
            } else {
                onClose();
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to create network');
        } finally {
            setIsLoading(false);
        }
    };

    const handleCopyInvite = async () => {
        if (!createdInviteCode) return;
        const deepLink = `gc://join?code=${createdInviteCode}`;
        try {
            await navigator.clipboard.writeText(deepLink);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch {
            // Fallback for older browsers
            console.error('Failed to copy to clipboard');
        }
    };

    if (!isOpen) return null;

    // Show invite code after successful creation
    if (createdInviteCode) {
        return (
            <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
                <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                    <h3 className="text-xl font-bold mb-2 text-white flex items-center gap-2">
                        <span className="text-green-400">✓</span> Network Created!
                    </h3>
                    <p className="text-gray-400 mb-4 text-sm">Share this invite code with your friends:</p>

                    <div className="bg-gc-dark-900 border border-gc-dark-700 rounded p-4 mb-4">
                        <div className="text-2xl font-mono font-bold text-center text-gc-primary tracking-widest mb-2">
                            {createdInviteCode}
                        </div>
                        <div className="text-xs text-gray-500 text-center font-mono">
                            gc://join?code={createdInviteCode}
                        </div>
                    </div>

                    <div className="flex gap-2 justify-end">
                        <button
                            onClick={handleCopyInvite}
                            className="px-4 py-2 bg-gc-dark-700 hover:bg-gc-dark-600 rounded text-white flex items-center gap-2"
                        >
                            {copied ? '✓ Copied!' : '📋 Copy Link'}
                        </button>
                        <button
                            onClick={onClose}
                            className="px-4 py-2 bg-gc-primary rounded text-white hover:bg-opacity-90"
                        >
                            Done
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                <h3 className="text-xl font-bold mb-4 text-white">Create Network</h3>

                <div className="mb-4">
                    <input
                        className={`w-full bg-gc-dark-900 border rounded p-2 text-white focus:outline-none ${error
                            ? 'border-red-500 focus:border-red-500'
                            : 'border-gc-dark-700 focus:border-gc-primary'
                            }`}
                        placeholder="Network Name (3-50 characters)"
                        value={name}
                        onChange={e => handleNameChange(e.target.value)}
                        disabled={isLoading}
                        autoFocus
                        maxLength={51}
                        onKeyDown={e => {
                            if (e.key === 'Enter' && !isLoading) handleSubmit();
                            if (e.key === 'Escape') onClose();
                        }}
                    />
                    {error && (
                        <p className="text-red-400 text-xs mt-1">{error}</p>
                    )}
                    <p className="text-gray-500 text-xs mt-1">
                        {name.trim().length}/50 characters
                    </p>
                </div>

                <div className="flex gap-2 justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300"
                        disabled={isLoading}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        className="px-4 py-2 bg-gc-primary rounded text-white hover:bg-opacity-90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                        disabled={isLoading || name.trim().length < 3}
                    >
                        {isLoading && (
                            <span className="animate-spin">⟳</span>
                        )}
                        {isLoading ? 'Creating...' : 'Create'}
                    </button>
                </div>
            </div>
        </div>
    );
}

// =============================================================================
// Join Network Modal
// =============================================================================

interface JoinModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (code: string) => Promise<void>;
    initialCode?: string;
}

/**
 * Maps API error messages to user-friendly messages
 */
function mapJoinError(error: string): string {
    const errorLower = error.toLowerCase();

    if (errorLower.includes('invalid') || errorLower.includes('expired')) {
        return 'This invite code is invalid or has expired. Please check and try again.';
    }
    if (errorLower.includes('already') && errorLower.includes('member')) {
        return "You're already a member of this network.";
    }
    if (errorLower.includes('capacity') || errorLower.includes('maximum')) {
        return 'This network has reached its maximum number of members.';
    }
    if (errorLower.includes('deleted')) {
        return 'This network has been deleted by its owner.';
    }
    if (errorLower.includes('banned')) {
        return "You've been banned from this network.";
    }
    if (errorLower.includes('rate') || errorLower.includes('too many')) {
        return 'Too many join attempts. Please wait a moment.';
    }
    if (errorLower.includes('not found')) {
        return 'Network not found. The invite code may be incorrect.';
    }

    return error || 'Failed to join network';
}

export function JoinNetworkModal({ isOpen, onClose, onSubmit, initialCode }: JoinModalProps) {
    const [code, setCode] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [pasteSuccess, setPasteSuccess] = useState(false);

    // Reset or set code when modal opens
    useEffect(() => {
        if (isOpen) {
            setCode((initialCode || "").toUpperCase());
            setIsLoading(false);
            setError(null);
            setPasteSuccess(false);
        }
    }, [isOpen, initialCode]);

    const handlePasteFromClipboard = async () => {
        try {
            const text = await navigator.clipboard.readText();
            // Extract code from deep link if pasted
            let extractedCode = text.trim();

            // Handle gc://join?code=XYZ format
            if (extractedCode.includes('gc://') || extractedCode.includes('goconnect://')) {
                try {
                    const url = new URL(extractedCode);
                    const codeParam = url.searchParams.get('code');
                    if (codeParam) {
                        extractedCode = codeParam;
                    } else if (url.pathname.length > 1) {
                        extractedCode = url.pathname.replace(/^\/+/, '');
                    }
                } catch {
                    // Not a valid URL, use as-is
                }
            }

            // Clean and uppercase
            extractedCode = extractedCode.replace(/[^a-zA-Z0-9]/g, '').toUpperCase().slice(0, 8);

            if (extractedCode.length > 0) {
                setCode(extractedCode);
                setError(null);
                setPasteSuccess(true);
                setTimeout(() => setPasteSuccess(false), 1500);
            }
        } catch {
            console.error('Failed to read clipboard');
        }
    };

    const handleSubmit = async () => {
        if (!code.trim() || isLoading) return;

        // Basic validation
        if (code.trim().length < 6) {
            setError('Invite code must be at least 6 characters');
            return;
        }

        setIsLoading(true);
        setError(null);

        try {
            await onSubmit(code.trim().toUpperCase());
            onClose();
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err);
            setError(mapJoinError(errorMessage));
        } finally {
            setIsLoading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                <h3 className="text-xl font-bold mb-4 text-white">Join Network</h3>

                <p className="text-gray-400 text-sm mb-4">
                    Enter the 8-character invite code shared with you:
                </p>

                <div className="mb-4">
                    {/* Paste Button */}
                    <button
                        onClick={handlePasteFromClipboard}
                        className={`w-full mb-3 px-4 py-2 rounded border transition-colors flex items-center justify-center gap-2 text-sm ${pasteSuccess
                                ? 'bg-green-900/30 border-green-500 text-green-400'
                                : 'bg-gc-dark-900 border-gc-dark-700 hover:border-gc-dark-600 text-gray-300 hover:text-white'
                            }`}
                        disabled={isLoading}
                    >
                        {pasteSuccess ? (
                            <>✓ Pasted!</>
                        ) : (
                            <>📋 Paste from Clipboard</>
                        )}
                    </button>

                    {/* Code Input */}
                    <input
                        className={`w-full bg-gc-dark-900 border rounded p-3 text-white focus:outline-none font-mono uppercase tracking-[0.3em] text-center text-xl ${error
                                ? 'border-red-500 focus:border-red-500'
                                : 'border-gc-dark-700 focus:border-gc-primary'
                            }`}
                        placeholder="ABC12XYZ"
                        value={code}
                        onChange={e => {
                            const cleaned = e.target.value.replace(/[^a-zA-Z0-9]/g, '').toUpperCase();
                            setCode(cleaned);
                            if (error) setError(null);
                        }}
                        disabled={isLoading}
                        autoFocus
                        maxLength={8}
                        onKeyDown={e => {
                            if (e.key === 'Enter' && !isLoading && code.length >= 6) handleSubmit();
                            if (e.key === 'Escape') onClose();
                        }}
                    />

                    {error && (
                        <p className="text-red-400 text-xs mt-2 text-center">{error}</p>
                    )}

                    <p className="text-gray-500 text-xs mt-2 text-center">
                        {code.length}/8 characters
                    </p>
                </div>

                <div className="flex gap-2 justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300"
                        disabled={isLoading}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        className="px-4 py-2 bg-gc-primary rounded text-white hover:bg-opacity-90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                        disabled={isLoading || code.length < 6}
                    >
                        {isLoading && <span className="animate-spin">⟳</span>}
                        {isLoading ? 'Joining...' : 'Join Network'}
                    </button>
                </div>
            </div>
        </div>
    );
}


// =============================================================================
// Rename Network Modal
// =============================================================================

interface RenameModalProps {
    isOpen: boolean;
    currentName: string;
    networkId: string;
    onClose: () => void;
    onSubmit: (networkId: string, newName: string) => Promise<void>;
}

export function RenameNetworkModal({ isOpen, currentName, networkId, onClose, onSubmit }: RenameModalProps) {
    const [name, setName] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);

    // Reset state when modal opens
    useEffect(() => {
        if (isOpen) {
            setName(currentName);
            setError(null);
            setIsLoading(false);
        }
    }, [isOpen, currentName]);

    const handleValidation = useCallback(() => {
        const result = validateNetworkName(name);
        if (!result.valid) {
            setError(result.error || 'Invalid network name');
            return false;
        }
        if (result.sanitized === currentName) {
            setError('New name must be different from current name');
            return false;
        }
        return true;
    }, [name, currentName]);

    const handleSubmit = async () => {
        if (isLoading) return;
        if (!handleValidation()) return;

        setIsLoading(true);
        setError(null);

        try {
            await onSubmit(networkId, name.trim());
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to rename network');
        } finally {
            setIsLoading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                <h3 className="text-xl font-bold mb-4 text-white">Rename Network</h3>

                <div className="mb-4">
                    <input
                        className={`w-full bg-gc-dark-900 border rounded p-2 text-white focus:outline-none ${error
                            ? 'border-red-500 focus:border-red-500'
                            : 'border-gc-dark-700 focus:border-gc-primary'
                            }`}
                        placeholder="New Network Name"
                        value={name}
                        onChange={e => {
                            setName(e.target.value);
                            if (error) setError(null);
                        }}
                        disabled={isLoading}
                        autoFocus
                        maxLength={51}
                        onKeyDown={e => {
                            if (e.key === 'Enter' && !isLoading) handleSubmit();
                            if (e.key === 'Escape') onClose();
                        }}
                    />
                    {error && (
                        <p className="text-red-400 text-xs mt-1">{error}</p>
                    )}
                </div>

                <div className="flex gap-2 justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300"
                        disabled={isLoading}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        className="px-4 py-2 bg-gc-primary rounded text-white hover:bg-opacity-90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                        disabled={isLoading || name.trim() === currentName}
                    >
                        {isLoading && <span className="animate-spin">⟳</span>}
                        {isLoading ? 'Saving...' : 'Save'}
                    </button>
                </div>
            </div>
        </div>
    );
}

// =============================================================================
// Delete Network Modal
// =============================================================================

interface DeleteModalProps {
    isOpen: boolean;
    networkName: string;
    networkId: string;
    onClose: () => void;
    onSubmit: (networkId: string) => Promise<void>;
}

export function DeleteNetworkModal({ isOpen, networkName, networkId, onClose, onSubmit }: DeleteModalProps) {
    const [confirmName, setConfirmName] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Reset state when modal opens
    useEffect(() => {
        if (isOpen) {
            setConfirmName("");
            setIsLoading(false);
            setError(null);
        }
    }, [isOpen]);

    const isConfirmed = confirmName.toLowerCase() === networkName.toLowerCase();

    const handleSubmit = async () => {
        if (!isConfirmed || isLoading) return;

        setIsLoading(true);
        setError(null);

        try {
            await onSubmit(networkId);
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to delete network');
        } finally {
            setIsLoading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-red-500/30">
                <h3 className="text-xl font-bold mb-2 text-red-400 flex items-center gap-2">
                    ⚠️ Delete Network
                </h3>

                <div className="mb-4 text-gray-300 text-sm">
                    <p className="mb-2">
                        This will <strong className="text-red-400">permanently delete</strong> the network
                        <strong className="text-white"> "{networkName}"</strong> and:
                    </p>
                    <ul className="list-disc list-inside text-gray-400 text-xs space-y-1 ml-2">
                        <li>Remove all members from the network</li>
                        <li>Revoke all invite codes</li>
                        <li>Delete all IP allocations</li>
                        <li>This action cannot be undone</li>
                    </ul>
                </div>

                <div className="mb-4">
                    <label className="text-gray-400 text-xs mb-1 block">
                        Type <strong className="text-white">{networkName}</strong> to confirm:
                    </label>
                    <input
                        className={`w-full bg-gc-dark-900 border rounded p-2 text-white focus:outline-none ${error
                            ? 'border-red-500'
                            : isConfirmed
                                ? 'border-red-500 focus:border-red-500'
                                : 'border-gc-dark-700 focus:border-gc-dark-600'
                            }`}
                        placeholder={networkName}
                        value={confirmName}
                        onChange={e => {
                            setConfirmName(e.target.value);
                            if (error) setError(null);
                        }}
                        disabled={isLoading}
                        autoFocus
                        onKeyDown={e => {
                            if (e.key === 'Enter' && isConfirmed && !isLoading) handleSubmit();
                            if (e.key === 'Escape') onClose();
                        }}
                    />
                    {error && (
                        <p className="text-red-400 text-xs mt-1">{error}</p>
                    )}
                </div>

                <div className="flex gap-2 justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300"
                        disabled={isLoading}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        className="px-4 py-2 bg-red-600 rounded text-white hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                        disabled={!isConfirmed || isLoading}
                    >
                        {isLoading && <span className="animate-spin">⟳</span>}
                        {isLoading ? 'Deleting...' : 'Delete Network'}
                    </button>
                </div>
            </div>
        </div>
    );
}
