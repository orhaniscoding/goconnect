import { toast, Toaster as SonnerToaster } from 'sonner';

export const Toaster = () => (
    <SonnerToaster
        position="bottom-right"
        theme="dark"
        toastOptions={{
            style: {
                background: '#1f2937', // gc-dark-800
                border: '1px solid #374151', // gc-dark-700
                color: '#fff',
            },
        }}
    />
);

// Helper hook for components to easily show toasts
// Keeping the same interface as before for compatibility
export const useToast = () => {
    return {
        success: (msg: string) => toast.success(msg),
        error: (msg: string) => toast.error(msg),
        info: (msg: string) => toast.info(msg),
        warning: (msg: string) => toast.warning(msg),
        // Expose raw toast for custom needs
        raw: toast,
    };
};
