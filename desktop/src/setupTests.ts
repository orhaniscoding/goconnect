import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock Tauri OS plugin
vi.mock('@tauri-apps/plugin-os', () => ({
    platform: () => 'linux',
    type: () => 'Linux',
    version: () => '1.0.0',
    arch: () => 'x64',
}));

// Mock window.matchMedia (needed for some UI libraries)
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(), // deprecated
        removeListener: vi.fn(), // deprecated
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
    })),
});
