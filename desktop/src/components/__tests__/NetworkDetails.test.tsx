import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NetworkDetails from '../NetworkDetails';
import { NetworkInfo, PeerInfo } from '../../lib/tauri-api';

const mockNetwork: NetworkInfo = { id: 'net-1', name: 'Alpha Corp', invite_code: 'abc-123', owner_id: 'self-1' };
const mockSelfPeer: PeerInfo = {
    id: 'self-1',
    name: 'My Device',
    display_name: 'My Device',
    virtual_ip: '10.0.0.1',
    connected: true,
    is_relay: false,
    latency_ms: 0,
    is_self: true
};

describe('NetworkDetails', () => {
    it('renders placeholder when no network selected', () => {
        render(
            <NetworkDetails
                selectedNetwork={undefined}
                selfPeer={mockSelfPeer}
                isOwner={false}
                onGenerateInvite={vi.fn()}
                onLeaveNetwork={vi.fn()}
                onRenameNetwork={vi.fn()}
                onDeleteNetwork={vi.fn()}
                setActiveTab={vi.fn()}
            />
        );

        expect(screen.getByText('Select Network')).toBeInTheDocument();
        expect(screen.queryByText('General')).not.toBeInTheDocument();
    });

    it('renders network details when selected', () => {
        render(
            <NetworkDetails
                selectedNetwork={mockNetwork}
                selfPeer={mockSelfPeer}
                isOwner={false}
                onGenerateInvite={vi.fn()}
                onLeaveNetwork={vi.fn()}
                onRenameNetwork={vi.fn()}
                onDeleteNetwork={vi.fn()}
                setActiveTab={vi.fn()}
            />
        );

        expect(screen.getByText('Alpha Corp')).toBeInTheDocument();
        expect(screen.getByText('General')).toBeInTheDocument();
        expect(screen.getByText(/Peers/)).toBeInTheDocument();
        expect(screen.getByText(/Invite/)).toBeInTheDocument();
        expect(screen.getByText(/Leave Network/)).toBeInTheDocument();
    });

    it('triggers actions', () => {
        const onInvite = vi.fn();
        const onLeave = vi.fn();
        const onTab = vi.fn();

        render(
            <NetworkDetails
                selectedNetwork={mockNetwork}
                selfPeer={mockSelfPeer}
                isOwner={false}
                onGenerateInvite={onInvite}
                onLeaveNetwork={onLeave}
                onRenameNetwork={vi.fn()}
                onDeleteNetwork={vi.fn()}
                setActiveTab={onTab}
            />
        );

        fireEvent.click(screen.getByText(/Invite/));
        expect(onInvite).toHaveBeenCalled();

        fireEvent.click(screen.getByText(/Leave Network/));
        expect(onLeave).toHaveBeenCalled();

        fireEvent.click(screen.getByText(/Peers/));
        expect(onTab).toHaveBeenCalledWith('peers');
    });

    it('renders self peer info footer', () => {
        render(
            <NetworkDetails
                selectedNetwork={mockNetwork}
                selfPeer={mockSelfPeer}
                isOwner={false}
                onGenerateInvite={vi.fn()}
                onLeaveNetwork={vi.fn()}
                onRenameNetwork={vi.fn()}
                onDeleteNetwork={vi.fn()}
                setActiveTab={vi.fn()}
            />
        );

        expect(screen.getByText('My Device')).toBeInTheDocument();
        expect(screen.getByTitle('10.0.0.1')).toBeInTheDocument();
    });

    it('triggers settings tab', () => {
        const onTab = vi.fn();
        render(
            <NetworkDetails
                selectedNetwork={mockNetwork}
                selfPeer={mockSelfPeer}
                isOwner={false}
                onGenerateInvite={vi.fn()}
                onLeaveNetwork={vi.fn()}
                onRenameNetwork={vi.fn()}
                onDeleteNetwork={vi.fn()}
                setActiveTab={onTab}
            />
        );

        fireEvent.click(screen.getByTitle('Settings'));
        expect(onTab).toHaveBeenCalledWith('settings');
    });

    it('shows owner management section when isOwner is true', () => {
        render(
            <NetworkDetails
                selectedNetwork={mockNetwork}
                selfPeer={mockSelfPeer}
                isOwner={true}
                onGenerateInvite={vi.fn()}
                onLeaveNetwork={vi.fn()}
                onRenameNetwork={vi.fn()}
                onDeleteNetwork={vi.fn()}
                setActiveTab={vi.fn()}
            />
        );

        expect(screen.getByText('Management')).toBeInTheDocument();
        expect(screen.getByText(/Manage Network/)).toBeInTheDocument();
    });
});

