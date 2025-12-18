import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import SettingsPanel from '../SettingsPanel';
import { tauriApi, Settings } from '../../lib/tauri-api';

vi.mock('../../lib/tauri-api', () => ({
    tauriApi: {
        getSettings: vi.fn(),
        updateSettings: vi.fn(),
        resetSettings: vi.fn(),
    }
}));

vi.mock('../Toast', () => ({
    useToast: () => ({
        success: vi.fn(),
        error: vi.fn(),
    }),
}));

const mockSettings: Settings = {
    auto_connect: false,
    start_minimized: false,
    notifications_enabled: false,
    log_level: 'info'
};

describe('SettingsPanel', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        (tauriApi.getSettings as any).mockResolvedValue(mockSettings);
    });

    it('renders and loads settings', async () => {
        render(<SettingsPanel />);
        expect(screen.getByText('Loading settings...')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.getByText('Settings')).toBeInTheDocument();
            expect(screen.getByText('Auto Connect')).toBeInTheDocument();
        });
    });

    it('toggles settings', async () => {
        (tauriApi.updateSettings as any).mockResolvedValue({ ...mockSettings, auto_connect: true });

        render(<SettingsPanel />);
        await waitFor(() => screen.getByText('Auto Connect'));

        // Select all toggle buttons
        const toggles = screen.getAllByRole('switch');
        expect(toggles.length).toBe(4);

        fireEvent.click(toggles[0]); // Auto Connect

        await waitFor(() => {
            expect(tauriApi.updateSettings).toHaveBeenCalledWith(expect.objectContaining({ auto_connect: true }));
        });
    });

    it('resets settings', async () => {
        vi.spyOn(window, 'confirm').mockImplementation(() => true);
        (tauriApi.resetSettings as any).mockResolvedValue(mockSettings);

        render(<SettingsPanel />);
        await waitFor(() => screen.getByText('Reset to Defaults'));

        fireEvent.click(screen.getByText('Reset to Defaults'));

        expect(window.confirm).toHaveBeenCalled();
        await waitFor(() => {
            expect(tauriApi.resetSettings).toHaveBeenCalled();
        });
    });
});
