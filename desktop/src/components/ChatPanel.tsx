import { useState, useEffect, useRef } from 'react';
import { tauriApi, ChatMessage } from '../lib/tauri-api';
import { handleError } from '../lib/utils';
import { useToast } from './Toast';

interface ChatPanelProps {
    networkId: string;
    recipientId?: string;
    recipientName?: string;
}

export default function ChatPanel({ networkId, recipientId, recipientName }: ChatPanelProps) {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [newMessage, setNewMessage] = useState("");
    const [sending, setSending] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const toast = useToast();

    // Poll for messages
    useEffect(() => {
        loadMessages();
        const interval = setInterval(loadMessages, 3000);
        return () => clearInterval(interval);
    }, [networkId, recipientId]);

    // Scroll to bottom on new messages
    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const loadMessages = async () => {
        try {
            const msgs = await tauriApi.getMessages(networkId);

            // Client-side filtering for private chat mock
            // Note: Validation revealed backend proto lacks recipient_id, so this is purely UI simulation 
            // of how it WOULD work. Currently all messages are broadcast.
            // We can only filter by sender_id (peer_id) effectively.
            let filteredMsgs = msgs;
            if (recipientId) {
                // Determine if message is part of conversation with recipientId
                // For 'is_self' messages, we don't know who we sent it to without recipient_id in proto.
                // So we can only show messages FROM the recipient.
                filteredMsgs = msgs.filter(m => m.peer_id === recipientId);
                // Currently impossible to show "sent" private messages without update.
            }

            filteredMsgs.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
            setMessages(filteredMsgs);
        } catch (e) {
            console.error("Failed to load messages", e);
        }
    };

    const handleSend = async (e?: React.FormEvent) => {
        e?.preventDefault();
        if (!newMessage.trim() || sending) return;

        setSending(true);
        try {
            if (recipientId) {
                // MOCK: Private chat not supported by backend yet
                toast.error("Private messages not fully supported by daemon yet (Protocol Update Required)");
                // We won't send it to avoid polluting broadcast channel
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
            <div className="flex-1 overflow-y-auto p-4 space-y-3" aria-live="polite" aria-label="Chat History">
                {messages.length === 0 ? (
                    <div className="text-center text-gray-500 mt-10">
                        {recipientId ? "No private history available." : "No messages yet. Say hello!"}
                    </div>
                ) : (
                    messages.map(msg => (
                        <div key={msg.id} className={`flex flex-col ${msg.is_self ? 'items-end' : 'items-start'}`}>
                            <div className={`max-w-[70%] rounded-lg px-3 py-2 ${msg.is_self
                                ? 'bg-gc-primary text-white'
                                : 'bg-gc-dark-600 text-gray-200'
                                }`}>
                                <div className="text-sm">{msg.content}</div>
                            </div>
                            <div className="text-xs text-gray-500 mt-1">
                                {!msg.is_self && !recipientId && <span className="mr-2 font-bold">{msg.peer_id.substring(0, 8)}</span>}
                                {new Date(msg.timestamp).toLocaleTimeString()}
                            </div>
                        </div>
                    ))
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Input Area */}
            <form onSubmit={handleSend} className="p-3 bg-gc-dark-900 border-t border-gc-dark-700 flex gap-2">
                <input
                    className="flex-1 bg-gc-dark-800 border border-gc-dark-600 rounded px-3 py-2 text-white outline-none focus:border-gc-primary"
                    placeholder={`Type a message${recipientId ? ' (Mock)' : '...'}`}
                    value={newMessage}
                    onChange={e => setNewMessage(e.target.value)}
                    aria-label="Type a message"
                />
                <button
                    type="submit"
                    className="px-4 py-2 bg-gc-primary text-white rounded hover:bg-opacity-90 disabled:opacity-50"
                    disabled={sending || !newMessage.trim()}
                    aria-label="Send message"
                >
                    Send
                </button>
            </form>
        </div>
    );
}

// Add some colors to tailwind/css later if needed. Using bg-gc-primary assuming it exists in index.css from previous context (previous conversation mentioned creating one?), 
// but wait, I never created index.css. I am editing existing project.
// The `App.tsx` I viewed earlier used `bg-gc-primary`. So it must exist.
