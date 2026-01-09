import { isPermissionGranted, requestPermission, sendNotification } from '@tauri-apps/plugin-notification';

/**
 * NotificationType for categorizing notifications
 */
export type NotificationType = 'message' | 'voice' | 'transfer' | 'member' | 'system';

/**
 * NotificationOptions for sending notifications
 */
export interface NotificationOptions {
    type: NotificationType;
    title: string;
    body: string;
    icon?: string;
}

// Track permission state
let permissionGranted = false;

/**
 * Initialize notifications and request permission if needed
 */
export async function initNotifications(): Promise<boolean> {
    try {
        permissionGranted = await isPermissionGranted();

        if (!permissionGranted) {
            const permission = await requestPermission();
            permissionGranted = permission === 'granted';
        }

        return permissionGranted;
    } catch (e) {
        console.warn('Notifications not available:', e);
        return false;
    }
}

/**
 * Send a desktop notification
 */
export async function sendDesktopNotification(options: NotificationOptions): Promise<void> {
    // Check if notifications are enabled in settings
    const settings = getNotificationSettings();

    if (!settings.enabled) return;
    if (!settings.types[options.type]) return;

    // Don't notify if window is focused (optional behavior)
    if (settings.onlyWhenMinimized && document.hasFocus()) return;

    try {
        if (!permissionGranted) {
            await initNotifications();
        }

        if (permissionGranted) {
            await sendNotification({
                title: options.title,
                body: options.body,
            });
        }
    } catch (e) {
        console.warn('Failed to send notification:', e);
    }
}

/**
 * Settings storage for notification preferences
 */
interface NotificationSettings {
    enabled: boolean;
    onlyWhenMinimized: boolean;
    types: {
        message: boolean;
        voice: boolean;
        transfer: boolean;
        member: boolean;
        system: boolean;
    };
}

const DEFAULT_SETTINGS: NotificationSettings = {
    enabled: true,
    onlyWhenMinimized: false,
    types: {
        message: true,
        voice: true,
        transfer: true,
        member: true,
        system: true,
    },
};

const STORAGE_KEY = 'goconnect_notification_settings';

/**
 * Get notification settings from localStorage
 */
export function getNotificationSettings(): NotificationSettings {
    try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored) {
            return { ...DEFAULT_SETTINGS, ...JSON.parse(stored) };
        }
    } catch (e) {
        console.warn('Failed to load notification settings:', e);
    }
    return DEFAULT_SETTINGS;
}

/**
 * Save notification settings to localStorage
 */
export function saveNotificationSettings(settings: Partial<NotificationSettings>): void {
    try {
        const current = getNotificationSettings();
        const updated = { ...current, ...settings };
        localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
    } catch (e) {
        console.warn('Failed to save notification settings:', e);
    }
}

/**
 * Convenience method for message notifications
 */
export function notifyNewMessage(senderName: string, preview: string, networkName?: string): void {
    sendDesktopNotification({
        type: 'message',
        title: networkName ? `${networkName} - New Message` : 'New Message',
        body: `${senderName}: ${preview.substring(0, 100)}`,
    });
}

/**
 * Convenience method for voice call notifications
 */
export function notifyVoiceCall(peerName: string, action: 'joined' | 'left'): void {
    sendDesktopNotification({
        type: 'voice',
        title: 'Voice Channel',
        body: `${peerName} ${action} the voice channel`,
    });
}

/**
 * Convenience method for file transfer notifications
 */
export function notifyTransfer(peerName: string, fileName: string, status: 'incoming' | 'complete' | 'failed'): void {
    const messages = {
        incoming: `${peerName} wants to send you: ${fileName}`,
        complete: `Transfer complete: ${fileName}`,
        failed: `Transfer failed: ${fileName}`,
    };

    sendDesktopNotification({
        type: 'transfer',
        title: 'File Transfer',
        body: messages[status],
    });
}

/**
 * Convenience method for member notifications
 */
export function notifyMember(memberName: string, action: 'joined' | 'left', networkName?: string): void {
    sendDesktopNotification({
        type: 'member',
        title: networkName || 'Network',
        body: `${memberName} ${action} the network`,
    });
}
