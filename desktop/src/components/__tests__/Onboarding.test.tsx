import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import Onboarding from '../Onboarding';
import { tauriApi } from '../../lib/tauri-api';

// Mock tauriApi
vi.mock('../../lib/tauri-api', () => ({
    tauriApi: {
        register: vi.fn(),
    }
}));

// Mock Toast hook
vi.mock('../Toast', () => ({
    useToast: () => ({
        success: vi.fn(),
        error: vi.fn(),
    }),
}));

describe('Onboarding Component', () => {
    it('renders correctly', () => {
        render(<Onboarding onComplete={() => { }} />);
        expect(screen.getByText('Welcome to GoConnect')).toBeInTheDocument();
        expect(screen.getByPlaceholderText('eyJhbGciOiJIUzI1NiIsIn...')).toBeInTheDocument();
        expect(screen.getByPlaceholderText('My Desktop')).toBeInTheDocument();
    });

    it('validates empty token', async () => {
        const onComplete = vi.fn();
        render(<Onboarding onComplete={onComplete} />);

        fireEvent.click(screen.getByText('Connect Device'));

        expect(tauriApi.register).not.toHaveBeenCalled();
        expect(onComplete).not.toHaveBeenCalled();
    });

    it('submits valid token and name', async () => {
        const onComplete = vi.fn();
        (tauriApi.register as any).mockResolvedValue(undefined);

        render(<Onboarding onComplete={onComplete} />);

        fireEvent.change(screen.getByPlaceholderText('eyJhbGciOiJIUzI1NiIsIn...'), { target: { value: 'valid-token' } });
        fireEvent.change(screen.getByPlaceholderText('My Desktop'), { target: { value: 'Test Device' } });

        fireEvent.click(screen.getByText('Connect Device'));

        await waitFor(() => {
            expect(tauriApi.register).toHaveBeenCalledWith('valid-token', 'Test Device');
            expect(onComplete).toHaveBeenCalled();
        });
    });

    it('submits valid token without name', async () => {
        const onComplete = vi.fn();
        (tauriApi.register as any).mockResolvedValue(undefined);

        render(<Onboarding onComplete={onComplete} />);

        fireEvent.change(screen.getByPlaceholderText('eyJhbGciOiJIUzI1NiIsIn...'), { target: { value: 'valid-token' } });

        fireEvent.click(screen.getByText('Connect Device'));

        await waitFor(() => {
            expect(tauriApi.register).toHaveBeenCalledWith('valid-token', undefined);
            expect(onComplete).toHaveBeenCalled();
        });
    });
});
