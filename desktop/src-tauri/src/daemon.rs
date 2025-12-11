// GoConnect Daemon gRPC Client
// Communicates with the local daemon via gRPC with IPC token authentication

use std::path::PathBuf;
use tonic::transport::Channel;
use tonic::metadata::MetadataValue;
use tonic::{Request, Status};

// Include generated protobuf code
pub mod proto {
    tonic::include_proto!("daemon");
}

use proto::daemon_service_client::DaemonServiceClient;
use proto::network_service_client::NetworkServiceClient;
use proto::peer_service_client::PeerServiceClient;
use proto::settings_service_client::SettingsServiceClient;
use proto::chat_service_client::ChatServiceClient;
use proto::transfer_service_client::TransferServiceClient;

const IPC_TOKEN_HEADER: &str = "x-goconnect-ipc-token";

/// DaemonClient wraps gRPC connections to the local GoConnect daemon
pub struct DaemonClient {
    channel: Channel,
    token: String,
}

impl DaemonClient {
    /// Connect to the daemon with IPC token authentication
    pub async fn connect() -> Result<Self, DaemonError> {
        let token = Self::load_ipc_token().await?;
        let endpoint = Self::get_daemon_endpoint();
        
        let channel = Channel::from_static(endpoint)
            .connect()
            .await
            .map_err(|e| DaemonError::Connection(e.to_string()))?;

        Ok(Self { channel, token })
    }

    /// Get the platform-specific daemon endpoint
    fn get_daemon_endpoint() -> &'static str {
        #[cfg(target_os = "windows")]
        {
            "http://127.0.0.1:34101"
        }
        #[cfg(not(target_os = "windows"))]
        {
            "http://[::1]:34101" // Unix socket would be better but tonic needs extra setup
        }
    }

    /// Load IPC auth token from the token file
    async fn load_ipc_token() -> Result<String, DaemonError> {
        let token_path = Self::get_token_path()?;
        
        let token = tokio::fs::read_to_string(&token_path)
            .await
            .map_err(|e| DaemonError::TokenNotFound(format!(
                "Failed to read token from {:?}: {}", token_path, e
            )))?;
        
        Ok(token.trim().to_string())
    }

    /// Get platform-specific token path
    fn get_token_path() -> Result<PathBuf, DaemonError> {
        #[cfg(target_os = "windows")]
        {
            let local_app_data = dirs::data_local_dir()
                .ok_or_else(|| DaemonError::TokenNotFound("Cannot find LOCALAPPDATA".into()))?;
            Ok(local_app_data.join("GoConnect").join("ipc.token"))
        }
        #[cfg(target_os = "macos")]
        {
            let home = dirs::home_dir()
                .ok_or_else(|| DaemonError::TokenNotFound("Cannot find home directory".into()))?;
            Ok(home.join("Library/Application Support/GoConnect/ipc.token"))
        }
        #[cfg(target_os = "linux")]
        {
            let home = dirs::home_dir()
                .ok_or_else(|| DaemonError::TokenNotFound("Cannot find home directory".into()))?;
            Ok(home.join(".local/share/goconnect/ipc.token"))
        }
    }

    /// Add auth token to a gRPC request
    fn add_auth<T>(&self, mut request: Request<T>) -> Request<T> {
        if let Ok(token) = self.token.parse::<MetadataValue<_>>() {
            request.metadata_mut().insert(IPC_TOKEN_HEADER, token);
        }
        request
    }

    // =========================================================================
    // DAEMON SERVICE
    // =========================================================================

    /// Get daemon status
    pub async fn get_status(&self) -> Result<DaemonStatus, DaemonError> {
        let mut client = DaemonServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::GetStatusRequest {}));
        
        let response = client.get_status(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let status = response.into_inner();
        Ok(DaemonStatus {
            connected: status.status == proto::ConnectionStatus::Connected as i32,
            virtual_ip: status.virtual_ip,
            active_peers: status.active_peers as u32,
            network_name: status.current_network_name,
        })
    }

    /// Get daemon version info
    pub async fn get_version(&self) -> Result<VersionInfo, DaemonError> {
        let mut client = DaemonServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(()));
        
        let response = client.get_version(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let v = response.into_inner();
        Ok(VersionInfo {
            version: v.version,
            build_date: v.build_date,
            commit: v.commit,
            go_version: v.go_version,
            os: v.os,
            arch: v.arch,
        })
    }

    // =========================================================================
    // NETWORK SERVICE
    // =========================================================================

    /// Create a new network
    pub async fn create_network(&self, name: &str) -> Result<NetworkInfo, DaemonError> {
        let mut client = NetworkServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::CreateNetworkRequest {
            name: name.to_string(),
            description: String::new(),
        }));
        
        let response = client.create_network(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let resp = response.into_inner();
        let network = resp.network.ok_or_else(|| DaemonError::InvalidResponse("missing network".into()))?;
        
        Ok(NetworkInfo {
            id: network.id,
            name: network.name,
            invite_code: resp.invite_code,
        })
    }

    /// Join a network via invite code
    pub async fn join_network(&self, invite_code: &str) -> Result<NetworkInfo, DaemonError> {
        let mut client = NetworkServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::JoinNetworkRequest {
            invite_code: invite_code.to_string(),
        }));
        
        let response = client.join_network(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let resp = response.into_inner();
        let network = resp.network.ok_or_else(|| DaemonError::InvalidResponse("missing network".into()))?;
        
        Ok(NetworkInfo {
            id: network.id,
            name: network.name,
            invite_code: String::new(),
        })
    }

    /// List all networks
    pub async fn list_networks(&self) -> Result<Vec<NetworkInfo>, DaemonError> {
        let mut client = NetworkServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(()));
        
        let response = client.list_networks(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let networks = response.into_inner().networks
            .into_iter()
            .map(|n| NetworkInfo {
                id: n.id,
                name: n.name,
                invite_code: n.invite_code,
            })
            .collect();
        
        Ok(networks)
    }

    /// Leave a network
    pub async fn leave_network(&self, network_id: &str) -> Result<(), DaemonError> {
        let mut client = NetworkServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::LeaveNetworkRequest {
            network_id: network_id.to_string(),
        }));
        
        client.leave_network(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(())
    }

    /// Generate an invite code for a network
    pub async fn generate_invite(&self, network_id: &str) -> Result<String, DaemonError> {
        let mut client = NetworkServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::GenerateInviteRequest {
            network_id: network_id.to_string(),
            max_uses: 0, // Unlimited
            expires_hours: 0, // No expiry
        }));
        
        let response = client.generate_invite(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(response.into_inner().invite_code)
    }

    // =========================================================================
    // PEER SERVICE
    // =========================================================================

    /// Get list of peers
    pub async fn get_peers(&self) -> Result<Vec<PeerInfo>, DaemonError> {
        let mut client = PeerServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::GetPeersRequest {
            network_id: String::new(), // Empty = current network
        }));
        
        let response = client.get_peers(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let peers = response.into_inner().peers
            .into_iter()
            .map(|p| PeerInfo {
                id: p.id,
                name: p.name,
                display_name: p.display_name,
                virtual_ip: p.virtual_ip,
                connected: p.status == proto::ConnectionStatus::Connected as i32,
                is_relay: p.connection_type == proto::ConnectionType::Relay as i32,
                latency_ms: p.latency_ms,
            })
            .collect();
        
        Ok(peers)
    }

    /// Kick a peer from a network
    pub async fn kick_peer(&self, network_id: &str, peer_id: &str) -> Result<(), DaemonError> {
        let mut client = PeerServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::KickPeerRequest {
            network_id: network_id.to_string(),
            peer_id: peer_id.to_string(),
            reason: String::new(),
        }));
        
        client.kick_peer(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(())
    }

    /// Ban a peer from a network
    pub async fn ban_peer(&self, network_id: &str, peer_id: &str, reason: &str) -> Result<(), DaemonError> {
        let mut client = PeerServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::BanPeerRequest {
            network_id: network_id.to_string(),
            peer_id: peer_id.to_string(),
            reason: reason.to_string(),
        }));
        
        client.ban_peer(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(())
    }

    /// Unban a peer from a network
    pub async fn unban_peer(&self, network_id: &str, peer_id: &str) -> Result<(), DaemonError> {
        let mut client = PeerServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::UnbanPeerRequest {
            network_id: network_id.to_string(),
            peer_id: peer_id.to_string(),
        }));
        
        client.unban_peer(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(())
    }

    // =========================================================================
    // SETTINGS SERVICE
    // =========================================================================

    /// Get daemon settings
    pub async fn get_settings(&self) -> Result<Settings, DaemonError> {
        let mut client = SettingsServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(()));
        
        let response = client.get_settings(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let s = response.into_inner();
        Ok(Settings {
            auto_connect: s.auto_connect,
            start_minimized: s.start_minimized,
            notifications_enabled: s.notifications_enabled,
            log_level: String::new(), // Not in proto, use default
        })
    }

    /// Update daemon settings
    pub async fn update_settings(&self, settings: &Settings) -> Result<Settings, DaemonError> {
        let mut client = SettingsServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::UpdateSettingsRequest {
            settings: Some(proto::Settings {
                auto_connect: settings.auto_connect,
                start_minimized: settings.start_minimized,
                notifications_enabled: settings.notifications_enabled,
                auto_accept_files: false,
                download_path: String::new(),
                max_upload_speed_kbps: 0,
                max_download_speed_kbps: 0,
                theme: String::new(),
                language: String::new(),
            }),
        }));
        
        let response = client.update_settings(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let s = response.into_inner();
        Ok(Settings {
            auto_connect: s.auto_connect,
            start_minimized: s.start_minimized,
            notifications_enabled: s.notifications_enabled,
            log_level: String::new(),
        })
    }

    /// Reset settings to defaults
    pub async fn reset_settings(&self) -> Result<Settings, DaemonError> {
        let mut client = SettingsServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(()));
        
        let response = client.reset_settings(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let s = response.into_inner();
        Ok(Settings {
            auto_connect: s.auto_connect,
            start_minimized: s.start_minimized,
            notifications_enabled: s.notifications_enabled,
            log_level: String::new(),
        })
    }

    // =========================================================================
    // CHAT SERVICE
    // =========================================================================

    /// Get chat messages
    pub async fn get_messages(&self, network_id: &str, limit: i32, before: Option<&str>) -> Result<Vec<ChatMessage>, DaemonError> {
        let mut client = ChatServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::GetMessagesRequest {
            network_id: network_id.to_string(),
            limit,
            before_id: before.unwrap_or_default().to_string(),
        }));
        
        let response = client.get_messages(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let messages = response.into_inner().messages
            .into_iter()
            .map(|m| ChatMessage {
                id: m.id,
                peer_id: m.sender_id.clone(),
                content: m.content,
                timestamp: m.sent_at.map(|t| t.seconds.to_string()).unwrap_or_default(),
                is_self: false, // Determine from sender_id comparison if needed
            })
            .collect();
        
        Ok(messages)
    }

    /// Send a chat message
    pub async fn send_message(&self, network_id: &str, content: &str) -> Result<(), DaemonError> {
        let mut client = ChatServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::SendMessageRequest {
            network_id: network_id.to_string(),
            content: content.to_string(),
            recipient_id: String::new(), // Empty = broadcast to network
        }));
        
        client.send_message(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(())
    }

    // =========================================================================
    // TRANSFER SERVICE
    // =========================================================================

    /// List transfers
    pub async fn list_transfers(&self, _status: Option<&str>, _peer_id: Option<&str>) -> Result<Vec<TransferInfo>, DaemonError> {
        let mut client = TransferServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(()));
        
        let response = client.list_transfers(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        let transfers = response.into_inner().transfers
            .into_iter()
            .map(|t| TransferInfo {
                id: t.id,
                peer_id: t.peer_id,
                file_name: t.filename,
                file_size: t.size_bytes as u64,
                transferred: t.transferred_bytes as u64,
                status: match t.status {
                    0 => "pending".to_string(),
                    1 => "pending".to_string(),
                    2 => "active".to_string(),
                    3 => "completed".to_string(),
                    4 => "failed".to_string(),
                    5 => "cancelled".to_string(),
                    _ => "unknown".to_string(),
                },
                direction: if t.is_incoming { "download".to_string() } else { "upload".to_string() },
                error: if t.error_message.is_empty() { None } else { Some(t.error_message) },
            })
            .collect();
        
        Ok(transfers)
    }

    /// Get transfer statistics
    pub async fn get_transfer_stats(&self) -> Result<TransferStats, DaemonError> {
        // Note: This would require a new gRPC method. For now, aggregate from list_transfers
        let transfers = self.list_transfers(None, None).await?;
        
        let mut stats = TransferStats {
            total_uploads: 0,
            total_downloads: 0,
            active_transfers: 0,
            completed_transfers: 0,
            failed_transfers: 0,
            total_bytes_sent: 0,
            total_bytes_received: 0,
        };
        
        for t in &transfers {
            if t.direction == "upload" {
                stats.total_uploads += 1;
                stats.total_bytes_sent += t.transferred;
            } else {
                stats.total_downloads += 1;
                stats.total_bytes_received += t.transferred;
            }
            
            match t.status.as_str() {
                "active" => stats.active_transfers += 1,
                "completed" => stats.completed_transfers += 1,
                "failed" | "cancelled" | "rejected" => stats.failed_transfers += 1,
                _ => {}
            }
        }
        
        Ok(stats)
    }

    /// Cancel an active transfer
    pub async fn cancel_transfer(&self, transfer_id: &str) -> Result<(), DaemonError> {
        let mut client = TransferServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::CancelTransferRequest {
            transfer_id: transfer_id.to_string(),
        }));
        
        client.cancel_transfer(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(())
    }

    /// Reject an incoming transfer
    pub async fn reject_transfer(&self, transfer_id: &str) -> Result<(), DaemonError> {
        let mut client = TransferServiceClient::new(self.channel.clone());
        let request = self.add_auth(Request::new(proto::RejectTransferRequest {
            transfer_id: transfer_id.to_string(),
        }));
        
        client.reject_transfer(request)
            .await
            .map_err(|e| DaemonError::Rpc(e))?;
        
        Ok(())
    }
}

// =============================================================================
// DATA TYPES (Rust-friendly versions of proto messages)
// =============================================================================

#[derive(Debug, Clone, serde::Serialize)]
pub struct DaemonStatus {
    pub connected: bool,
    pub virtual_ip: String,
    pub active_peers: u32,
    pub network_name: String,
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct VersionInfo {
    pub version: String,
    pub build_date: String,
    pub commit: String,
    pub go_version: String,
    pub os: String,
    pub arch: String,
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct NetworkInfo {
    pub id: String,
    pub name: String,
    pub invite_code: String,
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct PeerInfo {
    pub id: String,
    pub name: String,
    pub display_name: String,
    pub virtual_ip: String,
    pub connected: bool,
    pub is_relay: bool,
    pub latency_ms: i64,
}

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct Settings {
    pub auto_connect: bool,
    pub start_minimized: bool,
    pub notifications_enabled: bool,
    pub log_level: String,
}

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ChatMessage {
    pub id: String,
    pub peer_id: String,
    pub content: String,
    pub timestamp: String,
    pub is_self: bool,
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct TransferInfo {
    pub id: String,
    pub peer_id: String,
    pub file_name: String,
    pub file_size: u64,
    pub transferred: u64,
    pub status: String,
    pub direction: String,
    pub error: Option<String>,
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct TransferStats {
    pub total_uploads: u32,
    pub total_downloads: u32,
    pub active_transfers: u32,
    pub completed_transfers: u32,
    pub failed_transfers: u32,
    pub total_bytes_sent: u64,
    pub total_bytes_received: u64,
}

// =============================================================================
// ERROR TYPES
// =============================================================================

#[derive(Debug, thiserror::Error)]
pub enum DaemonError {
    #[error("Daemon not running or token file missing: {0}")]
    TokenNotFound(String),

    #[error("Failed to connect to daemon: {0}")]
    Connection(String),

    #[error("gRPC error: {0}")]
    Rpc(#[from] Status),

    #[error("Invalid response: {0}")]
    InvalidResponse(String),
}

impl serde::Serialize for DaemonError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.to_string())
    }
}
