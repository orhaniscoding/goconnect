import RNWg from 'react-native-wireguard-vpn';
import apiClient from './api';

export interface WireGuardConfig {
    interface: {
        privateKey: string;
        address: string;
        dns?: string;
    };
    peers: {
        publicKey: string;
        allowedIps: string;
        endpoint: string;
    }[];
}

class WireGuardService {
    private isConnected = false;

    async connect(networkId: string): Promise<boolean> {
        try {
            // 1. Get Config from API
            // Note: The API currently returns a file download for 'Generate Config'
            // We might need an endpoint that returns JSON, or we parse the file response.
            // For MVP, assuming we refactored backend or can parse the text response.
            // Let's assume we fetch the text config.
            const response = await apiClient.post(`/networks/${networkId}/config`, {}, {
                responseType: 'text' // We expect the .conf file content
            });

            const configText = response.data;

            // 2. Start VPN
            // RNWg usually takes the config as a string
            // NOTE: react-native-wireguard-vpn API might vary, assuming standard `connect(config)`
            // Checking library signature (mocked assumption based on standard libs)
            await (RNWg as any).connect(configText);

            this.isConnected = true;
            return true;
        } catch (error) {
            console.error('VPN Connect Failed:', error);
            return false;
        }
    }

    async disconnect(): Promise<void> {
        try {
            await (RNWg as any).disconnect();
            this.isConnected = false;
        } catch (error) {
            console.error('VPN Disconnect Failed:', error);
        }
    }

    getStatus(): boolean {
        return this.isConnected;
    }
}

export default new WireGuardService();
