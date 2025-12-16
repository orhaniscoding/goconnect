import { useState, useEffect } from 'react';

interface CreateModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (name: string) => Promise<void>;
}

export function CreateNetworkModal({ isOpen, onClose, onSubmit }: CreateModalProps) {
    const [name, setName] = useState("");

    // Reset name when modal opens
    useEffect(() => {
        if (isOpen) setName("");
    }, [isOpen]);

    if (!isOpen) return null;

    const handleSubmit = async () => {
        if (!name.trim()) return;
        await onSubmit(name);
    };

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                <h3 className="text-xl font-bold mb-4 text-white">Create Network</h3>
                <input
                    className="w-full bg-gc-dark-900 border border-gc-dark-700 rounded p-2 mb-4 text-white focus:border-gc-primary outline-none"
                    placeholder="Network Name"
                    value={name}
                    onChange={e => setName(e.target.value)}
                    autoFocus
                    onKeyDown={e => {
                        if (e.key === 'Enter') handleSubmit();
                        if (e.key === 'Escape') onClose();
                    }}
                />
                <div className="flex gap-2 justify-end">
                    <button onClick={onClose} className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300">Cancel</button>
                    <button onClick={handleSubmit} className="px-4 py-2 bg-gc-primary rounded text-white hover:bg-opacity-90">Create</button>
                </div>
            </div>
        </div>
    );
}

interface JoinModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (code: string) => Promise<void>;
}

export function JoinNetworkModal({ isOpen, onClose, onSubmit }: JoinModalProps) {
    const [code, setCode] = useState("");

    // Reset code when modal opens
    useEffect(() => {
        if (isOpen) setCode("");
    }, [isOpen]);

    if (!isOpen) return null;

    const handleSubmit = async () => {
        if (!code.trim()) return;
        await onSubmit(code);
    };

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gc-dark-800 p-6 rounded-lg w-96 shadow-xl border border-gc-dark-600">
                <h3 className="text-xl font-bold mb-4 text-white">Join Network</h3>
                <input
                    className="w-full bg-gc-dark-900 border border-gc-dark-700 rounded p-2 mb-4 text-white focus:border-gc-primary outline-none"
                    placeholder="Invite Code"
                    value={code}
                    onChange={e => setCode(e.target.value)}
                    autoFocus
                    onKeyDown={e => {
                        if (e.key === 'Enter') handleSubmit();
                        if (e.key === 'Escape') onClose();
                    }}
                />
                <div className="flex gap-2 justify-end">
                    <button onClick={onClose} className="px-4 py-2 hover:bg-gc-dark-700 rounded text-gray-300">Cancel</button>
                    <button onClick={handleSubmit} className="px-4 py-2 bg-gc-primary rounded text-white hover:bg-opacity-90">Join</button>
                </div>
            </div>
        </div>
    );
}
