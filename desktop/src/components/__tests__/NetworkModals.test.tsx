import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { CreateNetworkModal, JoinNetworkModal } from '../NetworkModals';

describe('NetworkModals', () => {
    describe('CreateNetworkModal', () => {
        it('does not render when closed', () => {
            render(<CreateNetworkModal isOpen={false} onClose={vi.fn()} onSubmit={vi.fn()} />);
            expect(screen.queryByText('Create Network')).not.toBeInTheDocument();
        });

        it('renders and handles input', () => {
            const onSubmit = vi.fn();
            render(<CreateNetworkModal isOpen={true} onClose={vi.fn()} onSubmit={onSubmit} />);

            expect(screen.getByText('Create Network', { selector: 'h3' })).toBeInTheDocument();

            const input = screen.getByPlaceholderText('Network Name');
            fireEvent.change(input, { target: { value: 'New Net' } });

            fireEvent.click(screen.getByText('Create', { selector: 'button' }));
            expect(onSubmit).toHaveBeenCalledWith('New Net');
        });

        it('handles enter key', () => {
            const onSubmit = vi.fn();
            render(<CreateNetworkModal isOpen={true} onClose={vi.fn()} onSubmit={onSubmit} />);

            const input = screen.getByPlaceholderText('Network Name');
            fireEvent.change(input, { target: { value: 'Enter Net' } });
            fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

            expect(onSubmit).toHaveBeenCalledWith('Enter Net');
        });

        it('does not submit empty name', () => {
            const onSubmit = vi.fn();
            render(<CreateNetworkModal isOpen={true} onClose={vi.fn()} onSubmit={onSubmit} />);

            fireEvent.click(screen.getByText('Create', { selector: 'button' }));
            expect(onSubmit).not.toHaveBeenCalled();
        });
    });

    describe('JoinNetworkModal', () => {
        it('does not render when closed', () => {
            render(<JoinNetworkModal isOpen={false} onClose={vi.fn()} onSubmit={vi.fn()} />);
            expect(screen.queryByText('Join Network')).not.toBeInTheDocument();
        });

        it('renders and handles input', () => {
            const onSubmit = vi.fn();
            render(<JoinNetworkModal isOpen={true} onClose={vi.fn()} onSubmit={onSubmit} />);

            expect(screen.getByText('Join Network', { selector: 'h3' })).toBeInTheDocument();

            const input = screen.getByPlaceholderText('Invite Code');
            fireEvent.change(input, { target: { value: 'ABC-123' } });

            fireEvent.click(screen.getByText('Join', { selector: 'button' }));
            expect(onSubmit).toHaveBeenCalledWith('ABC-123');
        });
    });
});
