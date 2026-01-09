import { useState, useEffect, useCallback } from 'react';
import { tauriApi, MemberInfo, NetworkInfo } from '../lib/tauri-api';

interface BannedMembersPanelProps {
    network: NetworkInfo | null;
}

export default function BannedMembersPanel({ network }: BannedMembersPanelProps) {
    const [bannedMembers, setBannedMembers] = useState<MemberInfo[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [unbanTarget, setUnbanTarget] = useState<MemberInfo | null>(null);

    const loadBannedMembers = useCallback(async () => {
        if (!network) return;

        try {
            setIsLoading(true);
            const members = await tauriApi.getBannedMembers(network.id);
            setBannedMembers(members);
        } catch (err) {
            console.error('Failed to load banned members:', err);
        } finally {
            setIsLoading(false);
        }
    }, [network]);

    useEffect(() => {
        loadBannedMembers();
    }, [loadBannedMembers]);

    const handleUnban = async () => {
        if (!unbanTarget || !network) return;

        try {
            await tauriApi.unbanPeer(network.id, unbanTarget.id);
            setUnbanTarget(null);
            loadBannedMembers();
        } catch (err) {
            console.error('Failed to unban member:', err);
        }
    };

    if (!network) {
        return (
            <div className="text-gray-500 text-sm">
                Select a network to view banned members
            </div>
        );
    }

    return (
        <div className="space-y-3">
            <h3 className="text-sm font-medium text-gray-300 flex items-center gap-2">
                🚫 Banned Members
                {bannedMembers.length > 0 && (
                    <span className="text-xs bg-red-600/20 text-red-400 px-2 py-0.5 rounded">
                        {bannedMembers.length}
                    </span>
                )}
            </h3>

            {isLoading ? (
                <div className="text-gray-400 text-sm">Loading...</div>
            ) : bannedMembers.length === 0 ? (
                <div className="text-gray-500 text-sm py-4 text-center bg-gc-dark-800 rounded-lg border border-gc-dark-600">
                    No banned members
                </div>
            ) : (
                <div className="space-y-2">
                    {bannedMembers.map(member => (
                        <div
                            key={member.id}
                            className="bg-gc-dark-800 p-3 rounded-lg border border-red-900/30"
                        >
                            <div className="flex items-center justify-between">
                                <div>
                                    <div className="font-medium text-white text-sm">
                                        {member.display_name || member.name}
                                    </div>
                                    <div className="text-xs text-gray-400">
                                        {member.ban_reason && (
                                            <span className="text-red-400">
                                                Reason: {member.ban_reason}
                                            </span>
                                        )}
                                        {member.banned_at && (
                                            <span className="ml-2">
                                                • Banned {new Date(member.banned_at).toLocaleDateString()}
                                            </span>
                                        )}
                                    </div>
                                </div>
                                <button
                                    onClick={() => setUnbanTarget(member)}
                                    className="px-2 py-1 text-xs bg-gray-600/20 text-gray-400 rounded hover:bg-green-600/20 hover:text-green-400 transition"
                                >
                                    Unban
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* Unban Confirmation Modal */}
            {unbanTarget && (
                <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
                    <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                        <h3 className="text-lg font-bold mb-2 text-white">Unban Member?</h3>
                        <p className="text-gray-300 text-sm mb-4">
                            <strong>{unbanTarget.display_name || unbanTarget.name}</strong> will be removed
                            from the banned list and can rejoin "{network.name}" with an invite code.
                        </p>
                        <div className="flex gap-2 justify-end">
                            <button
                                onClick={() => setUnbanTarget(null)}
                                className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleUnban}
                                className="px-4 py-2 bg-green-600 rounded text-white hover:bg-green-700"
                            >
                                Unban Member
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
