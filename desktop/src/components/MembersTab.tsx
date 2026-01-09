import { useState, useEffect, useCallback } from 'react';
import { tauriApi, MemberInfo, MemberRole, NetworkInfo } from '../lib/tauri-api';

// =============================================================================
// Role Badge Component
// =============================================================================

function RoleBadge({ role }: { role: MemberRole }) {
    const config = {
        owner: { icon: '👑', label: 'Owner', className: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30' },
        admin: { icon: '⚡', label: 'Admin', className: 'bg-blue-500/20 text-blue-400 border-blue-500/30' },
        member: { icon: '👤', label: 'Member', className: 'bg-gray-500/20 text-gray-400 border-gray-500/30' },
    };
    const { icon, label, className } = config[role];

    return (
        <span className={`text-[10px] px-1.5 py-0.5 rounded border ${className}`}>
            {icon} {label}
        </span>
    );
}

// =============================================================================
// Member Card Component
// =============================================================================

interface MemberCardProps {
    member: MemberInfo;
    viewerRole: MemberRole;
    onPromote: (member: MemberInfo) => void;
    onDemote: (member: MemberInfo) => void;
    onKick: (member: MemberInfo) => void;
    onBan: (member: MemberInfo) => void;
}

function MemberCard({ member, viewerRole, onPromote, onDemote, onKick, onBan }: MemberCardProps) {
    const isOwner = viewerRole === 'owner';
    const isAdmin = viewerRole === 'admin';
    const canManage = isOwner || isAdmin;
    const canPromote = isOwner && member.role === 'member';
    const canDemote = isOwner && member.role === 'admin';
    const canKick = canManage && member.role !== 'owner' && member.user_id !== 'self';
    const canBan = isOwner && member.role !== 'owner';

    const joinedDate = new Date(member.joined_at);
    const timeAgo = getRelativeTime(joinedDate);

    return (
        <div className={`bg-gc-dark-800 p-4 rounded-lg border ${member.role === 'owner' ? 'border-yellow-500/30' : 'border-gc-dark-600'
            } flex items-center justify-between`}>
            <div className="flex items-center gap-4">
                {/* Online status indicator */}
                <div className={`w-3 h-3 rounded-full ${member.is_online ? 'bg-green-500' : 'bg-gray-500'}`} />
                <div>
                    <div className="font-medium text-white flex items-center gap-2">
                        {member.display_name || member.name}
                        <RoleBadge role={member.role} />
                    </div>
                    <div className="text-sm text-gray-400 flex items-center gap-2">
                        <span className="text-[10px] text-gray-500">Joined {timeAgo}</span>
                        <span className="text-[10px] text-gray-500">• {member.is_online ? 'Online' : 'Offline'}</span>
                    </div>
                </div>
            </div>

            {/* Action buttons */}
            {canManage && member.role !== 'owner' && (
                <div className="flex gap-2">
                    {canPromote && (
                        <button
                            onClick={() => onPromote(member)}
                            className="px-2 py-1 text-xs bg-blue-600/20 text-blue-400 rounded hover:bg-blue-600/30 transition"
                            title="Promote to Admin"
                        >
                            ⚡ Promote
                        </button>
                    )}
                    {canDemote && (
                        <button
                            onClick={() => onDemote(member)}
                            className="px-2 py-1 text-xs bg-gray-600/20 text-gray-400 rounded hover:bg-gray-600/30 transition"
                            title="Demote to Member"
                        >
                            👤 Demote
                        </button>
                    )}
                    {canKick && (
                        <button
                            onClick={() => onKick(member)}
                            className="px-2 py-1 text-xs bg-orange-600/20 text-orange-400 rounded hover:bg-orange-600/30 transition"
                            title="Remove from network"
                        >
                            Remove
                        </button>
                    )}
                    {canBan && (
                        <button
                            onClick={() => onBan(member)}
                            className="px-2 py-1 text-xs bg-red-600/20 text-red-400 rounded hover:bg-red-600/30 transition"
                            title="Ban from network"
                        >
                            Ban
                        </button>
                    )}
                </div>
            )}
        </div>
    );
}

// =============================================================================
// Pending Request Card Component
// =============================================================================

interface PendingCardProps {
    member: MemberInfo;
    onApprove: (member: MemberInfo) => void;
    onReject: (member: MemberInfo) => void;
}

function PendingRequestCard({ member, onApprove, onReject }: PendingCardProps) {
    const requestDate = new Date(member.joined_at);
    const timeAgo = getRelativeTime(requestDate);

    return (
        <div className="bg-gc-dark-800 p-4 rounded-lg border border-yellow-500/30 flex items-center justify-between">
            <div className="flex items-center gap-4">
                <div className="w-3 h-3 rounded-full bg-yellow-500 animate-pulse" />
                <div>
                    <div className="font-medium text-white">
                        {member.display_name || member.name}
                    </div>
                    <div className="text-sm text-gray-400">
                        <span className="text-[10px] text-yellow-400">Requested {timeAgo}</span>
                    </div>
                </div>
            </div>

            <div className="flex gap-2">
                <button
                    onClick={() => onApprove(member)}
                    className="px-3 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700 transition font-medium"
                >
                    ✓ Approve
                </button>
                <button
                    onClick={() => onReject(member)}
                    className="px-3 py-1 text-xs bg-red-600/20 text-red-400 rounded hover:bg-red-600/30 transition"
                >
                    ✕ Reject
                </button>
            </div>
        </div>
    );
}

// =============================================================================
// Confirmation Modals
// =============================================================================

interface ConfirmModalProps {
    isOpen: boolean;
    title: string;
    message: string;
    confirmLabel: string;
    confirmColor?: string;
    children?: React.ReactNode;
    onConfirm: () => void;
    onCancel: () => void;
}

function ConfirmModal({ isOpen, title, message, confirmLabel, confirmColor = 'bg-red-600', children, onConfirm, onCancel }: ConfirmModalProps) {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                <h3 className="text-xl font-bold mb-2 text-white">{title}</h3>
                <p className="text-gray-300 text-sm mb-4">{message}</p>
                {children}
                <div className="flex gap-2 justify-end mt-4">
                    <button
                        onClick={onCancel}
                        className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={onConfirm}
                        className={`px-4 py-2 ${confirmColor} rounded text-white hover:opacity-90`}
                    >
                        {confirmLabel}
                    </button>
                </div>
            </div>
        </div>
    );
}

// =============================================================================
// Main MembersTab Component
// =============================================================================

interface MembersTabProps {
    network: NetworkInfo | undefined;
    selfUserId: string;
}

export default function MembersTab({ network, selfUserId }: MembersTabProps) {
    const [members, setMembers] = useState<MemberInfo[]>([]);
    const [pendingMembers, setPendingMembers] = useState<MemberInfo[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [error, setError] = useState<string | null>(null);

    // Modal states
    const [kickTarget, setKickTarget] = useState<MemberInfo | null>(null);
    const [banTarget, setBanTarget] = useState<MemberInfo | null>(null);
    const [banReason, setBanReason] = useState('');

    // Get viewer's role
    const selfMember = members.find(m => m.user_id === selfUserId);
    const viewerRole: MemberRole = selfMember?.role || 'member';

    const loadMembers = useCallback(async () => {
        if (!network) return;

        try {
            setIsLoading(true);
            setError(null);

            const [approved, pending] = await Promise.all([
                tauriApi.listMembers(network.id, 'approved'),
                tauriApi.listMembers(network.id, 'pending'),
            ]);

            setMembers(approved);
            setPendingMembers(pending);
        } catch (err) {
            console.error('Failed to load members:', err);
            setError('Failed to load members');
        } finally {
            setIsLoading(false);
        }
    }, [network]);

    useEffect(() => {
        loadMembers();
    }, [loadMembers]);

    // Filter members by search
    const filteredMembers = members.filter(m =>
        m.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        m.display_name?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    // Sort: owner first, then admins, then members
    const sortedMembers = [...filteredMembers].sort((a, b) => {
        const roleOrder = { owner: 0, admin: 1, member: 2 };
        return roleOrder[a.role] - roleOrder[b.role];
    });

    // Handlers
    const handlePromote = async (member: MemberInfo) => {
        try {
            await tauriApi.promoteMember(network!.id, member.id);
            loadMembers();
        } catch (err) {
            console.error('Failed to promote member:', err);
        }
    };

    const handleDemote = async (member: MemberInfo) => {
        try {
            await tauriApi.demoteMember(network!.id, member.id);
            loadMembers();
        } catch (err) {
            console.error('Failed to demote member:', err);
        }
    };

    const handleKickConfirm = async () => {
        if (!kickTarget || !network) return;
        try {
            await tauriApi.kickPeer(network.id, kickTarget.id);
            setKickTarget(null);
            loadMembers();
        } catch (err) {
            console.error('Failed to kick member:', err);
        }
    };

    const handleBanConfirm = async () => {
        if (!banTarget || !network || !banReason.trim()) return;
        try {
            await tauriApi.banPeer(network.id, banTarget.id, banReason);
            setBanTarget(null);
            setBanReason('');
            loadMembers();
        } catch (err) {
            console.error('Failed to ban member:', err);
        }
    };

    const handleApprove = async (member: MemberInfo) => {
        if (!network) return;
        try {
            await tauriApi.approveMember(network.id, member.id);
            loadMembers();
        } catch (err) {
            console.error('Failed to approve member:', err);
        }
    };

    const handleReject = async (member: MemberInfo) => {
        if (!network) return;
        try {
            await tauriApi.rejectMember(network.id, member.id);
            loadMembers();
        } catch (err) {
            console.error('Failed to reject member:', err);
        }
    };

    if (!network) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-gray-500">
                <div className="text-4xl mb-2">👥</div>
                <p>Select a network to view members</p>
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full text-gray-400">
                <span className="animate-spin mr-2">⟳</span> Loading members...
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-red-400">
                <p>{error}</p>
                <button onClick={loadMembers} className="mt-2 text-sm underline">Try again</button>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {/* Header with search */}
            <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-white">
                    Members ({members.length})
                    {pendingMembers.length > 0 && (
                        <span className="ml-2 text-sm text-yellow-400">• {pendingMembers.length} pending</span>
                    )}
                </h2>
                <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search members..."
                    className="px-3 py-1.5 bg-gc-dark-900 border border-gc-dark-700 rounded text-sm text-white placeholder-gray-500 focus:outline-none focus:border-gc-primary w-48"
                />
            </div>

            {/* Pending requests section */}
            {pendingMembers.length > 0 && (viewerRole === 'owner' || viewerRole === 'admin') && (
                <div className="space-y-2">
                    <h3 className="text-sm font-medium text-yellow-400 flex items-center gap-2">
                        <span className="animate-pulse">🔔</span> Pending Join Requests
                    </h3>
                    {pendingMembers.map(member => (
                        <PendingRequestCard
                            key={member.id}
                            member={member}
                            onApprove={handleApprove}
                            onReject={handleReject}
                        />
                    ))}
                </div>
            )}

            {/* Members list */}
            <div className="space-y-2">
                {sortedMembers.length === 0 ? (
                    <div className="text-center text-gray-500 py-8">
                        {searchQuery ? 'No members match your search' : 'No members found'}
                    </div>
                ) : (
                    sortedMembers.map(member => (
                        <MemberCard
                            key={member.id}
                            member={member}
                            viewerRole={viewerRole}
                            onPromote={handlePromote}
                            onDemote={handleDemote}
                            onKick={(m) => setKickTarget(m)}
                            onBan={(m) => setBanTarget(m)}
                        />
                    ))
                )}
            </div>

            {/* Kick Confirmation Modal */}
            <ConfirmModal
                isOpen={!!kickTarget}
                title="Remove Member from Network?"
                message={`${kickTarget?.display_name || kickTarget?.name} will be removed from "${network.name}". They can rejoin if they have a valid invite code.`}
                confirmLabel="Remove Member"
                confirmColor="bg-orange-600"
                onConfirm={handleKickConfirm}
                onCancel={() => setKickTarget(null)}
            />

            {/* Ban Confirmation Modal */}
            <ConfirmModal
                isOpen={!!banTarget}
                title="Ban Member from Network?"
                message={`${banTarget?.display_name || banTarget?.name} will be banned from "${network.name}". They cannot rejoin until unbanned.`}
                confirmLabel="Ban Member"
                confirmColor="bg-red-600"
                onConfirm={handleBanConfirm}
                onCancel={() => { setBanTarget(null); setBanReason(''); }}
            >
                <div className="mt-2">
                    <label className="text-xs text-gray-400">Ban reason (required)</label>
                    <input
                        type="text"
                        value={banReason}
                        onChange={(e) => setBanReason(e.target.value)}
                        placeholder="Enter reason for ban..."
                        className="w-full mt-1 px-3 py-2 bg-gc-dark-900 border border-gc-dark-700 rounded text-sm text-white placeholder-gray-500 focus:outline-none focus:border-red-500"
                        autoFocus
                    />
                </div>
            </ConfirmModal>
        </div>
    );
}

// =============================================================================
// Helper Functions
// =============================================================================

function getRelativeTime(date: Date): string {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffMins > 0) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
    return 'just now';
}
