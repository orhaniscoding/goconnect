// GoConnect Desktop Client
// Tauri 2.x application with gRPC daemon communication

mod daemon;
mod commands;

use commands::DaemonState;
use tauri::{
    menu::{Menu, MenuItem, PredefinedMenuItem},
    tray::TrayIconBuilder,
    Manager,
};

#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

use tauri_plugin_updater::UpdaterExt;
use tauri_plugin_notification::NotificationExt;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_os::init())
        .manage(DaemonState::default())
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                window.hide().unwrap();
                api.prevent_close();
            }
        })
        .setup(|app| {
            let status_i = MenuItem::with_id(app, "status", "Status: Checking...", false, None::<&str>)?;
            let check_update_i = MenuItem::with_id(app, "check_update", "Check for Updates", true, None::<&str>)?;
            let quit_i = MenuItem::with_id(app, "quit", "Quit", true, None::<&str>)?;
            let show_i = MenuItem::with_id(app, "show", "Show", true, None::<&str>)?;
            let sep1 = PredefinedMenuItem::separator(app)?;
            let sep2 = PredefinedMenuItem::separator(app)?;
            
            let menu = Menu::with_items(app, &[
                &status_i, 
                &sep1, 
                &check_update_i, 
                &sep2, 
                &show_i, 
                &quit_i
            ])?;

            let _tray = TrayIconBuilder::with_id("tray")
                .icon(app.default_window_icon().unwrap().clone())
                .menu(&menu)
                .show_menu_on_left_click(true)
                .on_menu_event(|app, event| match event.id.as_ref() {
                    "quit" => {
                        app.exit(0);
                    }
                    "show" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                    "check_update" => {
                        let handle = app.handle().clone();
                        tauri::async_runtime::spawn(async move {
                            match handle.updater().check().await {
                                Ok(Some(update)) => {
                                    handle.notification()
                                        .builder()
                                        .title("GoConnect Update")
                                        .body(format!("Update available: v{}", update.version).as_str())
                                        .show()
                                        .unwrap();
                                        
                                    // Optionally trigger download/install logic here or via dialog
                                    // For now, just notify.
                                }
                                Ok(None) => {
                                    handle.notification()
                                        .builder()
                                        .title("GoConnect")
                                        .body("You are on the latest version.")
                                        .show()
                                        .unwrap();
                                }
                                Err(e) => {
                                    handle.notification()
                                        .builder()
                                        .title("Update Check Failed")
                                        .body(format!("Error: {}", e).as_str())
                                        .show()
                                        .unwrap();
                                }
                            }
                        });
                    }
                    _ => {}
                })
                .build(app)?;
            
            // Spawn background task to update status
            let status_handle = status_i.clone();
            tauri::async_runtime::spawn(async move {
                loop {
                    let status_text = match crate::daemon::DaemonClient::connect().await {
                        Ok(client) => match client.get_status().await {
                            Ok(status) => {
                                if status.connected {
                                    format!("Status: Connected ({})", status.network_name)
                                } else {
                                    "Status: Disconnected".to_string()
                                }
                            }
                            Err(_) => "Status: Daemon Error".to_string(),
                        },
                        Err(_) => "Status: Daemon Stopped".to_string(),
                    };

                    let _ = status_handle.set_text(status_text);
                    tokio::time::sleep(std::time::Duration::from_secs(5)).await;
                }
            });

            Ok(())
        })
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .invoke_handler(tauri::generate_handler![
            greet,
            // Daemon commands
            commands::daemon_get_status,
            commands::daemon_get_version,
            commands::daemon_is_running,
            // Network commands
            commands::daemon_create_network,
            commands::daemon_join_network,
            commands::daemon_list_networks,
            commands::daemon_leave_network,
            commands::daemon_generate_invite,
            // Peer commands
            commands::daemon_get_peers,
            commands::daemon_kick_peer,
            commands::daemon_ban_peer,
            commands::daemon_unban_peer,
            // Settings commands
            commands::daemon_get_settings,
            commands::daemon_update_settings,
            commands::daemon_reset_settings,
            // Chat commands
            commands::daemon_get_messages,
            commands::daemon_send_message,
            // Transfer commands
            commands::daemon_list_transfers,
            commands::daemon_get_transfer_stats,
            commands::daemon_cancel_transfer,
            commands::daemon_reject_transfer,
            commands::daemon_send_file,
            commands::daemon_accept_transfer,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
