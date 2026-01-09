import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { CreateNetworkModal, JoinNetworkModal, RenameNetworkModal, DeleteNetworkModal, validateNetworkName } from '../NetworkModals';

describe('NetworkModals', () => {
    describe('validateNetworkName', () => {
        it('validates valid names correctly', () => {
            expect(validateNetworkName('My Network').valid).toBe(true);
            expect(validateNetworkName('Gaming-LAN-Server').valid).toBe(true);
            expect(validateNetworkName('ABC').valid).toBe(true);
        });

        it('rejects names that are too short', () => {
            const result = validateNetworkName('AB');
            expect(result.valid).toBe(false);
            expect(result.error).toContain('at least 3 characters');
        });

        it('rejects names that are too long', () => {
            const result = validateNetworkName('A'.repeat(51));
            expect(result.valid).toBe(false);
            expect(result.error).toContain('at most 50 characters');
        });

        it('rejects names with leading/trailing special chars', () => {
            expect(validateNetworkName('-Network').valid).toBe(false);
            expect(validateNetworkName('Network-').valid).toBe(false);
            expect(validateNetworkName('_Network').valid).toBe(false);
        });

        it('sanitizes whitespace', () => {
            const result = validateNetworkName('  My   Network  ');
            expect(result.sanitized).toBe('My Network');
        });
    });

    describe('CreateNetworkModal', () => {
        it('does not render when closed', () => {
            render(<CreateNetworkModal isOpen={false} onClose={vi.fn()} onSubmit={vi.fn()} />);
            expect(screen.queryByText('Create Network')).not.toBeInTheDocument();
        });

        it('renders and handles valid input', async () => {
            const onSubmit = vi.fn().mockResolvedValue({});
            render(<CreateNetworkModal isOpen={true} onClose={vi.fn()} onSubmit={onSubmit} />);

            expect(screen.getByText('Create Network', { selector: 'h3' })).toBeInTheDocument();

            const input = screen.getByPlaceholderText(/Network Name/);
            fireEvent.change(input, { target: { value: 'Valid Name' } });

            fireEvent.click(screen.getByRole('button', { name: /Create/i }));

            await waitFor(() => {
                expect(onSubmit).toHaveBeenCalledWith('Valid Name');
            });
        });

        it('shows validation error for short name', async () => {
            const onSubmit = vi.fn();
            render(<CreateNetworkModal isOpen={true} onClose={vi.fn()} onSubmit={onSubmit} />);

            const input = screen.getByPlaceholderText(/Network Name/);
            fireEvent.change(input, { target: { value: 'AB' } });
            fireEvent.click(screen.getByRole('button', { name: /Create/i }));

            await waitFor(() => {
                expect(screen.getByText(/at least 3 characters/)).toBeInTheDocument();
            });
            expect(onSubmit).not.toHaveBeenCalled();
        });

        it('shows invite code after successful creation', async () => {
            const onSubmit = vi.fn().mockResolvedValue({ inviteCode: 'ABC12XYZ' });
            render(<CreateNetworkModal isOpen={true} onClose={vi.fn()} onSubmit={onSubmit} />);

            const input = screen.getByPlaceholderText(/Network Name/);
            fireEvent.change(input, { target: { value: 'Test Network' } });
            fireEvent.click(screen.getByRole('button', { name: /Create/i }));

            await waitFor(() => {
                expect(screen.getByText('ABC12XYZ')).toBeInTheDocument();
                expect(screen.getByText('Network Created!')).toBeInTheDocument();
            });
        });
    });

    describe('JoinNetworkModal', () => {
        it('does not render when closed', () => {
            render(<JoinNetworkModal isOpen={false} onClose={vi.fn()} onSubmit={vi.fn()} />);
            expect(screen.queryByText('Join Network')).not.toBeInTheDocument();
        });

        it('renders and handles input', async () => {
            const onSubmit = vi.fn().mockResolvedValue(undefined);
            const onClose = vi.fn();
            render(<JoinNetworkModal isOpen={true} onClose={onClose} onSubmit={onSubmit} />);

            expect(screen.getByText('Join Network', { selector: 'h3' })).toBeInTheDocument();

            const input = screen.getByPlaceholderText(/INVITE CODE/i);
            fireEvent.change(input, { target: { value: 'ABC123' } });

            fireEvent.click(screen.getByRole('button', { name: /Join/i }));

            await waitFor(() => {
                expect(onSubmit).toHaveBeenCalledWith('ABC123');
            });
        });
    });

    describe('RenameNetworkModal', () => {
        it('does not render when closed', () => {
            render(
                <RenameNetworkModal
                    isOpen={false}
                    currentName="Old Name"
                    networkId="net1"
                    onClose={vi.fn()}
                    onSubmit={vi.fn()}
                />
            );
            expect(screen.queryByText('Rename Network')).not.toBeInTheDocument();
        });

        it('pre-fills with current name', () => {
            render(
                <RenameNetworkModal
                    isOpen={true}
                    currentName="My Network"
                    networkId="net1"
                    onClose={vi.fn()}
                    onSubmit={vi.fn()}
                />
            );
            expect(screen.getByDisplayValue('My Network')).toBeInTheDocument();
        });

        it('validates new name and submits', async () => {
            const onSubmit = vi.fn().mockResolvedValue(undefined);
            const onClose = vi.fn();
            render(
                <RenameNetworkModal
                    isOpen={true}
                    currentName="Old Name"
                    networkId="net1"
                    onClose={onClose}
                    onSubmit={onSubmit}
                />
            );

            const input = screen.getByDisplayValue('Old Name');
            fireEvent.change(input, { target: { value: 'New Name' } });
            fireEvent.click(screen.getByRole('button', { name: /Save/i }));

            await waitFor(() => {
                expect(onSubmit).toHaveBeenCalledWith('net1', 'New Name');
            });
        });
    });

    describe('DeleteNetworkModal', () => {
        it('does not render when closed', () => {
            render(
                <DeleteNetworkModal
                    isOpen={false}
                    networkName="Test Network"
                    networkId="net1"
                    onClose={vi.fn()}
                    onSubmit={vi.fn()}
                />
            );
            expect(screen.queryByText('Delete Network')).not.toBeInTheDocument();
        });

        it('requires typing network name to confirm', async () => {
            const onSubmit = vi.fn();
            render(
                <DeleteNetworkModal
                    isOpen={true}
                    networkName="Test Network"
                    networkId="net1"
                    onClose={vi.fn()}
                    onSubmit={onSubmit}
                />
            );

            // Button should be disabled initially
            const deleteBtn = screen.getByRole('button', { name: /Delete Network/i });
            expect(deleteBtn).toBeDisabled();

            // Type incorrect name
            const input = screen.getByPlaceholderText('Test Network');
            fireEvent.change(input, { target: { value: 'Wrong Name' } });
            expect(deleteBtn).toBeDisabled();

            // Type correct name
            fireEvent.change(input, { target: { value: 'Test Network' } });
            expect(deleteBtn).not.toBeDisabled();
        });

        it('calls onSubmit when confirmed correctly', async () => {
            const onSubmit = vi.fn().mockResolvedValue(undefined);
            const onClose = vi.fn();
            render(
                <DeleteNetworkModal
                    isOpen={true}
                    networkName="My Net"
                    networkId="net123"
                    onClose={onClose}
                    onSubmit={onSubmit}
                />
            );

            const input = screen.getByPlaceholderText('My Net');
            fireEvent.change(input, { target: { value: 'My Net' } });
            fireEvent.click(screen.getByRole('button', { name: /Delete Network/i }));

            await waitFor(() => {
                expect(onSubmit).toHaveBeenCalledWith('net123');
            });
        });
    });
});
