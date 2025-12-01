import React, { useState, useEffect, useRef } from 'react';
import { sendChatMessage, subscribeToEvents, sendFileRequest, acceptFileRequest, FileTransferRequest, FileTransferSession } from '../lib/api';
import { Send, X, MessageSquare, Paperclip, FileText, Download } from 'lucide-react';
import { open } from '@tauri-apps/plugin-dialog'; // For file picker
import { useToast } from './Toast';

interface ChatMessage {
    from: string;
    content: string;
    time: string;
    isSelf?: boolean;
    type?: 'text' | 'file_request' | 'file_progress';
    fileData?: any;
}

interface ChatWindowProps {
    peerId: string;
    peerName?: string;
    onClose: () => void;
}

export const ChatWindow: React.FC<ChatWindowProps> = ({ peerId, peerName, onClose }) => {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [input, setInput] = useState("");
    const [transfers, setTransfers] = useState<Record<string, FileTransferSession>>({});
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const toast = useToast();

    useEffect(() => {
        // Subscribe to incoming messages
        const unsubscribe = subscribeToEvents((event) => {
            if (event.type === 'chat_message') {
                const msg = event.data;
                // Only show messages from this peer
                // Note: In a real app we'd want to store all messages globally
                if (msg.From === peerId) {
                    setMessages(prev => [...prev, {
                        from: msg.From,
                        content: msg.Content,
                        time: msg.Time,
                        isSelf: false
                    }]);
                }
            } else if (event.type === 'file_request') {
                const data = event.data;
                if (data.sender_id === peerId) {
                    setMessages(prev => [...prev, {
                        from: peerId,
                        content: "Sent a file request",
                        time: new Date().toISOString(),
                        isSelf: false,
                        type: 'file_request',
                        fileData: data.request
                    }]);
                }
            } else if (event.type === 'file_progress') {
                const session = event.data as FileTransferSession;
                if (session.peer_id === peerId) {
                    setTransfers(prev => ({ ...prev, [session.id]: session }));
                    if (session.status === 'completed') {
                        toast.success(`File transfer completed: ${session.file_name}`);
                    } else if (session.status === 'failed') {
                        toast.error(`File transfer failed: ${session.file_name}`);
                    }
                }
            }
        });

        return () => {
            unsubscribe();
        };
    }, [peerId]);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const handleSend = async () => {
        if (!input.trim()) return;

        const content = input;
        setInput(""); // Optimistic clear

        // Add to local list immediately
        const now = new Date().toISOString();
        setMessages(prev => [...prev, {
            from: "me",
            content: content,
            time: now,
            isSelf: true
        }]);

        const res = await sendChatMessage(peerId, content);
        if (res.error) {
            console.error("Failed to send message:", res.error);
            toast.error("Failed to send message");
        }
    };

    const handleFileSelect = async () => {
        try {
            const selected = await open({
                multiple: false,
            });

            if (selected && typeof selected === 'string') {
                const res = await sendFileRequest(peerId, selected);
                if (res.error) {
                    console.error("Failed to send file:", res.error);
                    toast.error("Failed to send file request");
                } else {
                    // Add optimistic message
                    setMessages(prev => [...prev, {
                        from: "me",
                        content: `Sending file: ${selected.split(/[\\/]/).pop()}`,
                        time: new Date().toISOString(),
                        isSelf: true,
                        type: 'text' // Or a special type for sent files
                    }]);
                    toast.success("File request sent");
                }
            }
        } catch (err) {
            console.error("File selection failed:", err);
            toast.error("Failed to select file");
        }
    };

    const handleAcceptFile = async (req: FileTransferRequest) => {
        try {
            const savePath = await open({
                directory: true,
                multiple: false,
            });

            if (savePath && typeof savePath === 'string') {
                // Construct full path (mocking filename append for now, ideally user picks full path or dir)
                const fullPath = `${savePath}/${req.file_name}`;
                await acceptFileRequest(req, peerId, fullPath);
                toast.info("Starting download...");
            }
        } catch (err) {
            console.error("Accept file failed:", err);
            toast.error("Failed to accept file");
        }
    };

    return (
        <div className="fixed bottom-4 right-4 w-80 h-96 bg-gray-900 border border-gray-700 rounded-lg shadow-xl flex flex-col z-50">
            {/* Header */}
            <div className="flex items-center justify-between p-3 border-b border-gray-700 bg-gray-800 rounded-t-lg">
                <div className="flex items-center gap-2">
                    <MessageSquare size={18} className="text-blue-400" />
                    <span className="font-medium text-white">{peerName || peerId}</span>
                </div>
                <button onClick={onClose} className="text-gray-400 hover:text-white">
                    <X size={18} />
                </button>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3 bg-gray-900">
                {messages.length === 0 && (
                    <div className="text-center text-gray-500 text-sm mt-10">
                        Start a conversation with {peerName || "peer"}
                    </div>
                )}
                {messages.map((msg, i) => (
                    <div
                        key={i}
                        className={`flex flex-col ${msg.isSelf ? 'items-end' : 'items-start'}`}
                    >
                        <div
                            className={`max-w-[80%] px-3 py-2 rounded-lg text-sm ${msg.isSelf
                                ? 'bg-blue-600 text-white rounded-br-none'
                                : 'bg-gray-700 text-gray-200 rounded-bl-none'
                                }`}
                        >
                            {msg.type === 'file_request' && msg.fileData ? (
                                <div className="flex flex-col gap-2">
                                    <div className="flex items-center gap-2 font-medium">
                                        <FileText size={16} />
                                        {msg.fileData.file_name}
                                    </div>
                                    <div className="text-xs opacity-75">
                                        {(msg.fileData.file_size / 1024).toFixed(1)} KB
                                    </div>
                                    <button
                                        onClick={() => handleAcceptFile(msg.fileData)}
                                        className="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-xs flex items-center gap-1 mt-1 w-fit"
                                    >
                                        <Download size={12} /> Accept
                                    </button>
                                </div>
                            ) : (
                                msg.content
                            )}
                        </div>
                        <span className="text-[10px] text-gray-500 mt-1">
                            {new Date(msg.time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </span>
                    </div>
                ))}
                <div ref={messagesEndRef} />

                {/* Active Transfers Overlay */}
                {Object.values(transfers).filter(t => t.status === 'in_progress' || t.status === 'pending' || t.status === 'failed').length > 0 && (
                    <div className="absolute bottom-20 left-4 right-4 bg-gray-800 p-2 rounded border border-gray-600 shadow-lg max-h-40 overflow-y-auto">
                        {Object.values(transfers).filter(t => t.status === 'in_progress' || t.status === 'pending' || t.status === 'failed').map(t => (
                            <div key={t.id} className="text-xs text-white mb-2 last:mb-0">
                                <div className="flex justify-between mb-1 items-center">
                                    <span className="truncate max-w-[60%]">{t.file_name}</span>
                                    <div className="flex items-center gap-2">
                                        <span className={t.status === 'failed' ? 'text-red-400' : 'text-gray-300'}>
                                            {t.status === 'failed' ? 'Failed' : `${Math.round((t.sent_bytes / t.file_size) * 100)}%`}
                                        </span>
                                        <button
                                            onClick={() => setTransfers(prev => {
                                                const next = { ...prev };
                                                delete next[t.id];
                                                return next;
                                            })}
                                            className="text-gray-400 hover:text-white"
                                        >
                                            <X size={12} />
                                        </button>
                                    </div>
                                </div>
                                <div className="w-full bg-gray-700 rounded-full h-1.5">
                                    <div
                                        className={`h-1.5 rounded-full transition-all duration-300 ${t.status === 'failed' ? 'bg-red-500' : 'bg-blue-500'}`}
                                        style={{ width: `${(t.sent_bytes / t.file_size) * 100}%` }}
                                    ></div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* Input */}
            <div className="p-3 border-t border-gray-700 bg-gray-800 rounded-b-lg">
                <div className="flex gap-2">
                    <button
                        onClick={handleFileSelect}
                        className="text-gray-400 hover:text-white p-2 rounded transition-colors"
                        title="Send File"
                    >
                        <Paperclip size={18} />
                    </button>
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleSend()}
                        placeholder="Type a message..."
                        className="flex-1 bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
                    />
                    <button
                        onClick={handleSend}
                        className="bg-blue-600 hover:bg-blue-700 text-white p-2 rounded transition-colors"
                    >
                        <Send size={18} />
                    </button>
                </div>
            </div>
        </div>
    );
};
