# 📂 P2P File Sharing with GoConnect

GoConnect allows you to send files directly to other peers in your network using high-speed, encrypted peer-to-peer connections.

## How it Works

*   **Direct**: Files are sent directly from your device to the peer (not stored on our servers).
*   **Encrypted**: All data travels through the secure WireGuard tunnel.
*   **Fast**: Transfers happen at your local network or direct internet speed.

## 📤 Sending a File

1.  Open the **Desktop App**.
2.  Go to the **"Connected Peers"** tab.
3.  Find the person you want to send a file to.
4.  Click the **"Send File"** (📁) button next to their name.
    *   *Alternatively, go to the **Files** tab and drag & drop a file.*
5.  Select the file from your computer.
6.  The transfer status will appear in the **Files** tab as "Pending" until accepted.

## 📥 Receiving a File

When someone sends you a file:

1.  You will receive a notification (if enabled).
2.  Go to the **"Files"** tab.
3.  You will see a **Pending Download** request.
4.  Click **Accept** to start the download or **Reject** to decline.
5.  Select where to save the file.
6.  The transfer will begin immediately.

## 📊 Monitoring Transfers

The **Files** tab shows a dashboard of your activity:

*   **Active**: Currently transferring files.
*   **Up**: Total bytes uploaded.
*   **Down**: Total bytes downloaded.
*   **Completed**: Total finished transfers.

## 🕵️ Troubleshooting

*   **Transfer Stuck**: Ensure both peers remain online and connected to the GoConnect network.
*   **Speed Issues**: Direct P2P connections are fastest. If traffic is relaying through a server (DERP), speed may be lower.
*   **"Peer Offline"**: You cannot send files to offline peers since GoConnect doesn't store files in the cloud.
