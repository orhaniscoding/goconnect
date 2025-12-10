import { render, screen, act } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Toaster, useToast } from '../Toast';

// Helper component to trigger toast
const TriggerToast = ({ msg }: { msg: string }) => {
    const toast = useToast();
    return (
        <button onClick={() => toast.success(msg)}>
            Trigger
        </button>
    );
};

describe('Toast Component', () => {
    it('renders toaster without crashing', () => {
        render(<Toaster />);
    });

    it('displays toast message when triggered', async () => {
        render(
            <>
                <Toaster />
                <TriggerToast msg="Test Message" />
            </>
        );

        const button = screen.getByText('Trigger');
        act(() => {
            button.click();
        });

        // Depending on sonner implementation, we might need to wait
        expect(await screen.findByText('Test Message')).toBeInTheDocument();
    });
});
