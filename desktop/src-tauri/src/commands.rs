// Tauri Commands - Bridge between frontend and daemon gRPC client

use crate::daemon::{
    ChatMessage, DaemonClient, DaemonStatus, NetworkInfo, PeerInfo, Settings, 
    TransferInfo, TransferStats, VersionInfo
};
use tauri::State;
use tokio::sync::Mutex;

/// Managed state holding the daemon client connection
pub struct DaemonState(pub Mutex<Option<DaemonClient>>);

impl Default for DaemonState {
    fn default() -> Self {
        Self(Mutex::new(None))
    }
}

/// Ensure daemon client is connected
async fn get_client(state: &State<'_, DaemonState>) -> Result<DaemonClient, String> {
    let mut guard = state.0.lock().await;

    // Use existing connection if available
    if let Some(client) = guard.as_ref() {
        return Ok(client.clone());
    }

    // Otherwise create new connection
    let client = DaemonClient::connect().await.map_err(|e| e.to_string())?;
    *guard = Some(client.clone());
    
    Ok(client)
}

// =============================================================================
// DAEMON COMMANDS
// =============================================================================

#[tauri::command]
pub async fn daemon_get_status(state: State<'_, DaemonState>) -> Result<DaemonStatus, String> {
    let client = get_client(&state).await?;
    client.get_status().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_get_version(state: State<'_, DaemonState>) -> Result<VersionInfo, String> {
    let client = get_client(&state).await?;
    client.get_version().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_is_running(_state: State<'_, DaemonState>) -> Result<bool, String> {
    match DaemonClient::connect().await {
        Ok(client) => {
            match client.get_status().await {
                Ok(_) => Ok(true),
                Err(_) => Ok(false),
            }
        }
        Err(_) => Ok(false),
    }
}

// =============================================================================
// NETWORK COMMANDS
// =============================================================================

#[tauri::command]
pub async fn daemon_create_network(
    state: State<'_, DaemonState>,
    name: String,
) -> Result<NetworkInfo, String> {
    let client = get_client(&state).await?;
    client.create_network(&name).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_join_network(
    state: State<'_, DaemonState>,
    invite_code: String,
) -> Result<NetworkInfo, String> {
    let client = get_client(&state).await?;
    client.join_network(&invite_code).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_list_networks(state: State<'_, DaemonState>) -> Result<Vec<NetworkInfo>, String> {
    let client = get_client(&state).await?;
    client.list_networks().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_leave_network(
    state: State<'_, DaemonState>,
    network_id: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.leave_network(&network_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_generate_invite(
    state: State<'_, DaemonState>,
    network_id: String,
) -> Result<String, String> {
    let client = get_client(&state).await?;
    client.generate_invite(&network_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_delete_network(
    state: State<'_, DaemonState>,
    network_id: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.delete_network(&network_id).await.map_err(|e| e.to_string())
}

// =============================================================================
// PEER COMMANDS
// =============================================================================

#[tauri::command]
pub async fn daemon_get_peers(state: State<'_, DaemonState>) -> Result<Vec<PeerInfo>, String> {
    let client = get_client(&state).await?;
    client.get_peers().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_kick_peer(
    state: State<'_, DaemonState>,
    network_id: String,
    peer_id: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.kick_peer(&network_id, &peer_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_ban_peer(
    state: State<'_, DaemonState>,
    network_id: String,
    peer_id: String,
    reason: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.ban_peer(&network_id, &peer_id, &reason).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_unban_peer(
    state: State<'_, DaemonState>,
    network_id: String,
    peer_id: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.unban_peer(&network_id, &peer_id).await.map_err(|e| e.to_string())
}

// =============================================================================
// SETTINGS COMMANDS
// =============================================================================

#[tauri::command]
pub async fn daemon_get_settings(state: State<'_, DaemonState>) -> Result<Settings, String> {
    let client = get_client(&state).await?;
    client.get_settings().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_update_settings(
    state: State<'_, DaemonState>,
    settings: Settings,
) -> Result<Settings, String> {
    let client = get_client(&state).await?;
    client.update_settings(&settings).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_reset_settings(state: State<'_, DaemonState>) -> Result<Settings, String> {
    let client = get_client(&state).await?;
    client.reset_settings().await.map_err(|e| e.to_string())
}

// =============================================================================
// CHAT COMMANDS
// =============================================================================

#[tauri::command]
pub async fn daemon_get_messages(
    state: State<'_, DaemonState>,
    network_id: String,
    limit: Option<i32>,
    before: Option<String>,
) -> Result<Vec<ChatMessage>, String> {
    let client = get_client(&state).await?;
    client.get_messages(&network_id, limit.unwrap_or(50), before.as_deref())
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_send_message(
    state: State<'_, DaemonState>,
    network_id: String,
    content: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.send_message(&network_id, &content).await.map_err(|e| e.to_string())
}

// =============================================================================
// TRANSFER COMMANDS
// =============================================================================

#[tauri::command]
pub async fn daemon_list_transfers(
    state: State<'_, DaemonState>,
    status: Option<String>,
    peer_id: Option<String>,
) -> Result<Vec<TransferInfo>, String> {
    let client = get_client(&state).await?;
    client.list_transfers(status.as_deref(), peer_id.as_deref())
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_get_transfer_stats(state: State<'_, DaemonState>) -> Result<TransferStats, String> {
    let client = get_client(&state).await?;
    client.get_transfer_stats().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_cancel_transfer(
    state: State<'_, DaemonState>,
    transfer_id: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.cancel_transfer(&transfer_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_reject_transfer(
    state: State<'_, DaemonState>,
    transfer_id: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.reject_transfer(&transfer_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_send_file(
    state: State<'_, DaemonState>,
    peer_id: String,
    file_path: String,
) -> Result<String, String> {
    let client = get_client(&state).await?;
    client.send_file(&peer_id, &file_path).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn daemon_accept_transfer(
    state: State<'_, DaemonState>,
    transfer_id: String,
    save_path: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.accept_transfer(&transfer_id, &save_path).await.map_err(|e| e.to_string())
}
