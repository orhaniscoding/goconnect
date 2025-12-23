const { withAndroidManifest, withInfoPlist, withEntitlementsPlists } = require('@expo/config-plugins');

const withWireGuard = (config) => {
    // Android Configuration
    config = withAndroidManifest(config, async (config) => {
        const androidManifest = config.modResults;
        const mainApplication = androidManifest.manifest.application[0];

        // Add permissions if not present
        const permissions = [
            'android.permission.FOREGROUND_SERVICE',
            'android.permission.BIND_VPN_SERVICE'
        ];

        // Note: BIND_VPN_SERVICE usually goes on the service definition, 
        // but React Native libraries often handle the service declaration in their own AAR.
        // We just ensure the permission is requested.

        if (!androidManifest.manifest['uses-permission']) {
            androidManifest.manifest['uses-permission'] = [];
        }

        permissions.forEach(permission => {
            if (!androidManifest.manifest['uses-permission'].some(p => p.$['android:name'] === permission)) {
                androidManifest.manifest['uses-permission'].push({
                    $: { 'android:name': permission }
                });
            }
        });

        return config;
    });

    // iOS Configuration
    config = withInfoPlist(config, async (config) => {
        if (!config.modResults.UIBackgroundModes) {
            config.modResults.UIBackgroundModes = [];
        }
        if (!config.modResults.UIBackgroundModes.includes('network-authentication')) {
            config.modResults.UIBackgroundModes.push('network-authentication');
        }
        // WireGuard usually needs 'packet-tunnel' via Network Extension, 
        // which requires a separate target. Expo config plugins can't easily creating targets yet without 'eas-cli-local-build-plugin' or complex mods.
        // For MVP, we document that iOS requires manual Xcode capabilities or a specialized plugin.
        // adding NSCameraUsageDescription just in case QR scan is needed later.
        return config;
    });

    return config;
};

module.exports = withWireGuard;
