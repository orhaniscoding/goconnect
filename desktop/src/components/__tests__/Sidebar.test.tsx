import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import Sidebar from '../Sidebar';
import { NetworkInfo } from '../../lib/tauri-api';

const mockNetworks: NetworkInfo[] = [
    { id: 'net-1', name: 'Alpha Corp', invite_code: 'abc-123' },
    { id: 'net-2', name: 'Beta Team', invite_code: 'xyz-789' },
];

describe('Sidebar', () => {
    it('renders network list', () => {
        render(
            <Sidebar
                networks={mockNetworks}
                selectedNetworkId={null}
                onSelectNetwork={vi.fn()}
                onShowCreate={vi.fn()}
                onShowJoin={vi.fn()}
            />
        );

        expect(screen.getByText('AL')).toBeInTheDocument(); // Alpha Corp initials
        expect(screen.getByText('BE')).toBeInTheDocument(); // Beta Team initials
        expect(screen.getByTitle('Alpha Corp')).toBeInTheDocument();
        expect(screen.getByTitle('Beta Team')).toBeInTheDocument();
    });

    it('highlights selected network', () => {
        render(
            <Sidebar
                networks={mockNetworks}
                selectedNetworkId="net-1"
                onSelectNetwork={vi.fn()}
                onShowCreate={vi.fn()}
                onShowJoin={vi.fn()}
            />
        );

        const btnAlpha = screen.getByTitle('Alpha Corp');
        const btnBeta = screen.getByTitle('Beta Team');

        expect(btnAlpha).toHaveClass('bg-gc-primary');
        expect(btnBeta).not.toHaveClass('bg-gc-primary');
    });

    it('triggers selection on click', () => {
        const onSelect = vi.fn();
        render(
            <Sidebar
                networks={mockNetworks}
                selectedNetworkId={null}
                onSelectNetwork={onSelect}
                onShowCreate={vi.fn()}
                onShowJoin={vi.fn()}
            />
        );

        fireEvent.click(screen.getByTitle('Beta Team'));
        expect(onSelect).toHaveBeenCalledWith('net-2');
    });

    it('triggers create and join actions', () => {
        const onCreate = vi.fn();
        const onJoin = vi.fn();
        render(
            <Sidebar
                networks={mockNetworks}
                selectedNetworkId={null}
                onSelectNetwork={vi.fn()}
                onShowCreate={onCreate}
                onShowJoin={onJoin}
            />
        );

        fireEvent.click(screen.getByTitle('Create Network'));
        expect(onCreate).toHaveBeenCalled();

        fireEvent.click(screen.getByTitle('Join Network'));
        expect(onJoin).toHaveBeenCalled();
    });
});
