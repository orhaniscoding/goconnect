// GoConnect Desktop Client
// Tauri 2.x application with gRPC daemon communication

mod daemon;
mod commands;

use commands::DaemonState;
use tauri::{
    menu::{Menu, MenuItem},
    tray::TrayIconBuilder,
    Manager,
};

#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_os::init())
        .manage(DaemonState::default())
        .setup(|app| {
            let quit_i = MenuItem::with_id(app, "quit", "Quit", true, None::<&str>)?;
            let show_i = MenuItem::with_id(app, "show", "Show", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show_i, &quit_i])?;

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
                    _ => {}
                })
                .build(app)?;

            Ok(())
        })
        .plugin(tauri_plugin_opener::init())
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
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
