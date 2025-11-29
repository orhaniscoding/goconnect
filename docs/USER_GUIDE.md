# GoConnect User Guide

## Getting Started

GoConnect makes it easy to create virtual LANs for gaming, file sharing, and remote access. Think of it as "Discord for networks."

### Basic Concepts

- **Tenant**: Your community or organization (like a Discord server)
- **Network**: A virtual LAN within your tenant (like a Discord voice channel)
- **Client**: The software that runs on your computer to connect you

## Installation

### Windows
1. Download `goconnect-daemon.exe` from the releases page
2. Run the installer or extract to a folder
3. Double-click to start the daemon

### macOS
1. Download `goconnect-daemon-macos` from the releases page
2. Open the downloaded file
3. Follow the installation prompts

### Linux
1. Download `goconnect-daemon-linux` from the releases page
2. Make it executable: `chmod +x goconnect-daemon-linux`
3. Run: `./goconnect-daemon-linux`

## Joining a Network

### Method 1: Invite Link (Easiest)
1. Click the invite link shared by your friend
2. The web interface will open automatically
3. Create an account or login
4. Click "Join Network"

### Method 2: Server URL
1. Open your web browser
2. Go to the server URL (e.g., `https://goconnect.example.com`)
3. Create an account or login
4. Browse available networks and click "Join"

### Method 3: Invite Code
1. Open the web interface
2. Click "Join with Code"
3. Enter the invite code
4. Follow the prompts

## Using GoConnect for Gaming

### Minecraft LAN
1. Host starts Minecraft in LAN mode
2. All players join the same GoConnect network
3. Players connect to the host's LAN game as if on the same router
4. No port forwarding required!

### Other LAN Games
- Most games with LAN multiplayer support work the same way
- The game sees all players as if they're on the same local network
- No special configuration needed

## Managing Your Connection

### Web Interface
The web interface shows:
- **Network Status**: Connected networks and your IP addresses
- **Online Members**: Who else is currently connected
- **Chat**: Communicate with other network members
- **Settings**: Configure your devices and preferences

### Client Daemon
The daemon runs in the background and:
- Maintains your network connections
- Automatically reconnects if disconnected
- Manages WireGuard VPN interfaces
- Handles IP address allocation

## Network Features

### Chat
- **Tenant Chat**: Talk to everyone in your tenant
- **Network Chat**: Talk to people in specific networks
- **File Sharing**: Share files with chat members
- **Online Status**: See who's currently online

### Device Management
- **Multiple Devices**: Connect from multiple computers
- **Device Names**: Give your devices recognizable names
- **Auto-Discovery**: Devices automatically find each other

### Network Administration
- **Create Networks**: Tenants can create multiple networks
- **Access Control**: Set who can join your networks
- **Member Management**: Kick or ban members if needed
- **Invite System**: Generate invite links or codes

## Troubleshooting

### Connection Issues
1. **Check Daemon Status**: Make sure the daemon is running
2. **Verify Server URL**: Ensure you're connecting to the right server
3. **Check Firewall**: Allow GoConnect through your firewall
4. **Restart Daemon**: Stop and restart the client daemon

### Performance Issues
1. **Check Network Speed**: Test your internet connection
2. **Reduce Network Members**: Large networks may impact performance
3. **Update Software**: Ensure you're using the latest version
4. **Check Server Status**: Verify the server is operational

### Common Questions

**Q: Can I join multiple networks?**
A: Yes! You can be connected to multiple networks simultaneously.

**Q: Do I need to configure my router?**
A: No. GoConnect handles all networking automatically.

**Q: Is my data secure?**
A: Yes. All traffic is encrypted with WireGuard military-grade encryption.

**Q: Can I use this on mobile devices?**
A: Mobile support is coming soon! Currently Windows, macOS, and Linux are supported.

## Privacy and Security

### Data Protection
- **End-to-End Encryption**: All traffic is encrypted between devices
- **No Tracking**: We don't monitor or log your network traffic
- **Local Processing**: Your data stays on your devices when possible

### Account Security
- **Strong Passwords**: Use unique, strong passwords
- **Two-Factor Authentication**: Enable 2FA if available
- **Secure Connections**: Always use HTTPS for web interface

### Best Practices
- **Share Invite Links Carefully**: Only share with trusted people
- **Regular Updates**: Keep your client software updated
- **Network Hygiene**: Remove unknown devices from your networks

## Advanced Usage

### Power User Features
- **Multiple Tenants**: Join different organizations
- **Network Bridging**: Connect multiple networks together
- **Custom Configurations**: Advanced networking options
- **API Access**: Programmatic control for automation

### Development and Testing
- **Local Networks**: Create networks for development
- **Testing Environments**: Isolated networks for testing
- **API Documentation**: Available for developers

## Getting Help

### Community Support
- **Discord Server**: Join our community Discord
- **GitHub Issues**: Report bugs and request features
- **Documentation**: Check our online documentation

### Contact Support
- **Email Support**: support@goconnect.example.com
- **Bug Reports**: Use GitHub issues for bug reports
- **Feature Requests**: Suggest improvements on GitHub

---

**Happy gaming and networking!** 🎮🌐
