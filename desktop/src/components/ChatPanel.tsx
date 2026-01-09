import { useState, useEffect, useRef, useCallback } from 'react';
import { tauriApi, ChatMessage } from '../lib/tauri-api';
import { handleError, formatMarkdown, canEditMessage, formatMessageTime } from '../lib/utils';
import { useToast } from './Toast';
// Notification import ready for future use: import { notifyNewMessage } from '../lib/notifications';



interface ChatPanelProps {
    networkId: string;
    recipientId?: string;
    recipientName?: string;
}

export default function ChatPanel({ networkId, recipientId, recipientName }: ChatPanelProps) {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [newMessage, setNewMessage] = useState("");
    const [sending, setSending] = useState(false);
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [editContent, setEditContent] = useState("");
    const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

    const messagesContainerRef = useRef<HTMLDivElement>(null);
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const toast = useToast();

    // Initial load
    useEffect(() => {
        loadMessages();
        const interval = setInterval(loadMessages, 3000);
        return () => clearInterval(interval);
    }, [networkId, recipientId]);

    // Scroll to bottom on initial load and new messages (not when loading older)
    useEffect(() => {
        if (!loading) {
            messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
        }
    }, [messages.length]);

    const loadMessages = useCallback(async (before?: string) => {
        try {
            const msgs = await tauriApi.getMessages(networkId, 50, before);

            let filteredMsgs = msgs;
            if (recipientId) {
                filteredMsgs = msgs.filter(m => m.peer_id === recipientId);
            }

            filteredMsgs.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

            if (before) {
                // Prepend older messages
                setMessages(prev => [...filteredMsgs, ...prev]);
                setHasMore(filteredMsgs.length === 50);
            } else {
                setMessages(filteredMsgs);
            }
        } catch (e) {
            console.error("Failed to load messages", e);
        }
    }, [networkId, recipientId]);

    // Handle scroll for infinite scroll
    const handleScroll = useCallback(async (e: React.UIEvent<HTMLDivElement>) => {
        const { scrollTop } = e.currentTarget;

        if (scrollTop === 0 && hasMore && !loading && messages.length > 0) {
            setLoading(true);
            const oldestMessage = messages[0];
            await loadMessages(oldestMessage.id);
            setLoading(false);
        }
    }, [hasMore, loading, messages, loadMessages]);

    const handleSend = async (e?: React.FormEvent) => {
        e?.preventDefault();
        if (!newMessage.trim() || sending) return;

        setSending(true);
        try {
            if (recipientId) {
                toast.error("Private messages not fully supported yet");
            } else {
                await tauriApi.sendMessage(networkId, newMessage.trim());
                setNewMessage("");
                loadMessages();
            }
        } catch (e) {
            handleError(e, "Failed to send");
        } finally {
            setSending(false);
        }
    };

    const handleEdit = async (messageId: string) => {
        if (!editContent.trim()) return;

        try {
            await tauriApi.editMessage(messageId, editContent.trim());
            setEditingId(null);
            setEditContent("");
            loadMessages();
            toast.success("Message edited");
        } catch (e) {
            handleError(e, "Failed to edit");
        }
    };

    const handleDelete = async (messageId: string) => {
        try {
            await tauriApi.deleteMessage(messageId);
            setDeleteConfirmId(null);
            loadMessages();
            toast.success("Message deleted");
        } catch (e) {
            handleError(e, "Failed to delete");
        }
    };

    const startEdit = (msg: ChatMessage) => {
        setEditingId(msg.id);
        setEditContent(msg.content);
    };

    const cancelEdit = () => {
        setEditingId(null);
        setEditContent("");
    };

    return (
        <div className="flex flex-col h-full bg-gc-dark-800 rounded-lg border border-gc-dark-600 overflow-hidden">
            {/* Header for Private Chat */}
            {recipientId && (
                <div className="p-3 bg-gc-dark-900 border-b border-gc-dark-700 font-bold text-white flex items-center gap-2">
                    <span className="text-gray-400">Chat with</span>
                    <span className="text-gc-primary">{recipientName || recipientId.substring(0, 8)}</span>
                </div>
            )}

            {/* Messages Area */}
            <div
                ref={messagesContainerRef}
                onScroll={handleScroll}
                className="flex-1 overflow-y-auto p-4 space-y-3"
                aria-live="polite"
                aria-label="Chat History"
            >
                {/* Loading indicator for older messages */}
                {loading && (
                    <div className="text-center py-2 text-gray-500 text-sm animate-pulse">
                        Loading older messages...
                    </div>
                )}

                {/* No more messages indicator */}
                {!hasMore && messages.length > 0 && (
                    <div className="text-center py-2 text-gray-600 text-xs">
                        — Beginning of chat —
                    </div>
                )}

                {messages.length === 0 && !loading ? (
                    <div className="text-center text-gray-500 mt-10">
                        {recipientId ? "No private history available." : "No messages yet. Say hello!"}
                    </div>
                ) : (
                    messages.map(msg => (
                        <div
                            key={msg.id}
                            className={`message-container flex flex-col ${msg.is_self ? 'items-end' : 'items-start'} group`}
                        >
                            {/* Sender name for others */}
                            {!msg.is_self && !recipientId && (
                                <div className="text-xs text-gray-400 mb-1 ml-1 font-medium">
                                    {msg.peer_name || msg.peer_id.substring(0, 8) + '...'}
                                </div>
                            )}

                            <div className="flex items-start gap-2">
                                {/* Message bubble */}
                                {editingId === msg.id ? (
                                    // Edit mode
                                    <div className="flex flex-col gap-2 w-full max-w-[70%]">
                                        <textarea
                                            className="bg-gc-dark-700 border border-gc-dark-500 rounded px-3 py-2 text-white text-sm resize-none"
                                            value={editContent}
                                            onChange={e => setEditContent(e.target.value)}
                                            rows={3}
                                            autoFocus
                                        />
                                        <div className="flex gap-2 justify-end">
                                            <button
                                                onClick={cancelEdit}
                                                className="px-3 py-1 text-xs text-gray-400 hover:text-white"
                                            >
                                                Cancel
                                            </button>
                                            <button
                                                onClick={() => handleEdit(msg.id)}
                                                className="px-3 py-1 text-xs bg-gc-primary text-white rounded hover:bg-opacity-90"
                                            >
                                                Save
                                            </button>
                                        </div>
                                    </div>
                                ) : (
                                    // Normal display
                                    <>
                                        <div className={`max-w-[70%] rounded-lg px-3 py-2 ${msg.is_deleted
                                            ? 'bg-gc-dark-700 text-gray-500 italic'
                                            : msg.is_self
                                                ? 'bg-gc-primary text-white'
                                                : 'bg-gc-dark-600 text-gray-200'
                                            }`}>
                                            {msg.is_deleted ? (
                                                <span className="text-sm">[Message deleted]</span>
                                            ) : (
                                                <div
                                                    className="text-sm"
                                                    dangerouslySetInnerHTML={{ __html: formatMarkdown(msg.content) }}
                                                />
                                            )}
                                        </div>

                                        {/* Action buttons */}
                                        {msg.is_self && !msg.is_deleted && (
                                            <div className="message-actions flex gap-1">
                                                {canEditMessage(msg.timestamp) && (
                                                    <button
                                                        onClick={() => startEdit(msg)}
                                                        className="p-1 text-gray-500 hover:text-white text-xs"
                                                        title="Edit message"
                                                    >
                                                        ✏️
                                                    </button>
                                                )}
                                                <button
                                                    onClick={() => setDeleteConfirmId(msg.id)}
                                                    className="p-1 text-gray-500 hover:text-red-400 text-xs"
                                                    title="Delete message"
                                                >
                                                    🗑️
                                                </button>
                                            </div>
                                        )}
                                    </>
                                )}
                            </div>

                            {/* Timestamp and edited indicator */}
                            <div className="text-xs text-gray-500 mt-1">
                                {formatMessageTime(msg.timestamp)}
                                {msg.is_edited && <span className="edited-indicator">(edited)</span>}
                            </div>
                        </div>
                    ))
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Delete Confirmation Modal */}
            {deleteConfirmId && (
                <div className="absolute inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-gc-dark-800 border border-gc-dark-600 rounded-lg p-4 max-w-sm mx-4">
                        <h3 className="text-white font-bold mb-2">Delete Message?</h3>
                        <p className="text-gray-400 text-sm mb-4">
                            This action cannot be undone.
                        </p>
                        <div className="flex gap-2 justify-end">
                            <button
                                onClick={() => setDeleteConfirmId(null)}
                                className="px-4 py-2 text-gray-400 hover:text-white"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={() => handleDelete(deleteConfirmId)}
                                className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
                            >
                                Delete
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Input Area */}
            <form onSubmit={handleSend} className="p-3 bg-gc-dark-900 border-t border-gc-dark-700 flex gap-2">
                <input
                    className="flex-1 bg-gc-dark-800 border border-gc-dark-600 rounded px-3 py-2 text-white outline-none focus:border-gc-primary"
                    placeholder={recipientId ? "Private messages coming soon..." : "Type a message... (supports **bold**, *italic*, `code`)"}
                    value={newMessage}
                    onChange={e => setNewMessage(e.target.value)}
                    disabled={!!recipientId}
                    aria-label="Type a message"
                />
                <button
                    type="submit"
                    className="px-4 py-2 bg-gc-primary text-white rounded hover:bg-opacity-90 disabled:opacity-50"
                    disabled={sending || !newMessage.trim() || !!recipientId}
                    aria-label="Send message"
                >
                    Send
                </button>
            </form>
        </div>
    );
}
