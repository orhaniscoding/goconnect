package daemon

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"log/slog"
)

// NotificationPriority defines the urgency level of a notification.
type NotificationPriority string

const (
	// PriorityUrgent for time-sensitive notifications.
	PriorityUrgent NotificationPriority = "urgent"
	// PriorityNormal for standard notifications.
	PriorityNormal NotificationPriority = "normal"
	// PriorityLow for non-urgent notifications.
	PriorityLow NotificationPriority = "low"
)

// Notification represents a push notification.
type Notification struct {
	ID        string               `json:"id"`
	Type      string               `json:"type"` // chat, network, peer, transfer, system
	Priority  NotificationPriority `json:"priority"`
	Title     string               `json:"title"`
	Body      string               `json:"body"`
	Icon      string               `json:"icon,omitempty"`
	Action    string               `json:"action,omitempty"`
	NetworkID string               `json:"network_id,omitempty"`
	SenderID  string               `json:"sender_id,omitempty"`
	Silent    bool                 `json:"silent,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
}

// NotificationSettings holds user preferences for notifications.
type NotificationSettings struct {
	Enabled           bool     `json:"enabled"`
	DoNotDisturb      bool     `json:"do_not_disturb"`
	DoNotDisturbStart string   `json:"dnd_start,omitempty"` // HH:MM format
	DoNotDisturbEnd   string   `json:"dnd_end,omitempty"`   // HH:MM format
	MutedNetworks     []string `json:"muted_networks,omitempty"`
	SoundEnabled      bool     `json:"sound_enabled"`
	ShowPreview       bool     `json:"show_preview"`
}

// DefaultNotificationSettings returns default settings.
func DefaultNotificationSettings() NotificationSettings {
	return NotificationSettings{
		Enabled:      true,
		SoundEnabled: true,
		ShowPreview:  true,
	}
}

// NotificationService handles cross-platform notifications.
type NotificationService struct {
	settings NotificationSettings
	mu       sync.RWMutex
	history  []Notification
	maxHist  int

	// Callbacks
	onNotification func(Notification)
}

// NewNotificationService creates a new notification service.
func NewNotificationService(settings NotificationSettings) *NotificationService {
	return &NotificationService{
		settings: settings,
		history:  make([]Notification, 0, 100),
		maxHist:  100,
	}
}

// SetOnNotification sets a callback for incoming notifications.
func (ns *NotificationService) SetOnNotification(fn func(Notification)) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.onNotification = fn
}

// UpdateSettings updates notification preferences.
func (ns *NotificationService) UpdateSettings(settings NotificationSettings) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.settings = settings
}

// GetSettings returns current notification settings.
func (ns *NotificationService) GetSettings() NotificationSettings {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.settings
}

// Show displays a notification on the system.
func (ns *NotificationService) Show(ctx context.Context, n Notification) error {
	ns.mu.RLock()
	settings := ns.settings
	callback := ns.onNotification
	ns.mu.RUnlock()

	// Check if notifications are enabled
	if !settings.Enabled {
		slog.Debug("Notifications disabled, skipping", "title", n.Title)
		return nil
	}

	// Check Do Not Disturb
	if settings.DoNotDisturb {
		slog.Debug("Do Not Disturb active, skipping", "title", n.Title)
		return nil
	}

	// Check if network is muted
	if n.NetworkID != "" {
		for _, mutedNet := range settings.MutedNetworks {
			if mutedNet == n.NetworkID {
				slog.Debug("Network muted, skipping", "network", n.NetworkID)
				return nil
			}
		}
	}

	// Add to history
	ns.addToHistory(n)

	// Trigger callback if set
	if callback != nil {
		go callback(n)
	}

	// Display based on platform
	return ns.showPlatformNotification(ctx, n, settings)
}

// showPlatformNotification displays notification using platform-specific method.
func (ns *NotificationService) showPlatformNotification(ctx context.Context, n Notification, settings NotificationSettings) error {
	switch runtime.GOOS {
	case "linux":
		return ns.showLinuxNotification(ctx, n, settings)
	case "darwin":
		return ns.showMacNotification(ctx, n, settings)
	case "windows":
		return ns.showWindowsNotification(ctx, n, settings)
	default:
		slog.Warn("Unsupported platform for notifications", "os", runtime.GOOS)
		return nil
	}
}

// showLinuxNotification uses notify-send on Linux.
func (ns *NotificationService) showLinuxNotification(ctx context.Context, n Notification, settings NotificationSettings) error {
	args := []string{
		n.Title,
		n.Body,
	}

	// Set urgency based on priority
	switch n.Priority {
	case PriorityUrgent:
		args = append([]string{"-u", "critical"}, args...)
	case PriorityLow:
		args = append([]string{"-u", "low"}, args...)
	default:
		args = append([]string{"-u", "normal"}, args...)
	}

	// Add icon if available
	if n.Icon != "" {
		args = append([]string{"-i", n.Icon}, args...)
	}

	// Add app name
	args = append([]string{"-a", "GoConnect"}, args...)

	cmd := exec.CommandContext(ctx, "notify-send", args...)
	if err := cmd.Run(); err != nil {
		slog.Warn("Failed to show Linux notification", "error", err)
		// Not a critical error, don't return it
		return nil
	}

	slog.Debug("Linux notification shown", "title", n.Title)
	return nil
}

// showMacNotification uses osascript on macOS.
func (ns *NotificationService) showMacNotification(ctx context.Context, n Notification, settings NotificationSettings) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, n.Body, n.Title)

	if !n.Silent && settings.SoundEnabled {
		script = fmt.Sprintf(`display notification "%s" with title "%s" sound name "default"`, n.Body, n.Title)
	}

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		slog.Warn("Failed to show macOS notification", "error", err)
		return nil
	}

	slog.Debug("macOS notification shown", "title", n.Title)
	return nil
}

// showWindowsNotification uses PowerShell on Windows.
func (ns *NotificationService) showWindowsNotification(ctx context.Context, n Notification, settings NotificationSettings) error {
	// PowerShell script to show toast notification
	script := fmt.Sprintf(`
		[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

		$template = @"
		<toast>
			<visual>
				<binding template="ToastGeneric">
					<text>%s</text>
					<text>%s</text>
				</binding>
			</visual>
		</toast>
"@

		$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
		$xml.LoadXml($template)
		$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
		$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("GoConnect")
		$notifier.Show($toast)
	`, n.Title, n.Body)

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", script)
	if err := cmd.Run(); err != nil {
		slog.Warn("Failed to show Windows notification", "error", err)
		return nil
	}

	slog.Debug("Windows notification shown", "title", n.Title)
	return nil
}

// addToHistory adds a notification to the history.
func (ns *NotificationService) addToHistory(n Notification) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	// Trim if exceeding max
	if len(ns.history) >= ns.maxHist {
		ns.history = ns.history[1:]
	}

	ns.history = append(ns.history, n)
}

// GetHistory returns notification history.
func (ns *NotificationService) GetHistory(limit int) []Notification {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	if limit <= 0 || limit > len(ns.history) {
		limit = len(ns.history)
	}

	// Return most recent first
	result := make([]Notification, limit)
	for i := 0; i < limit; i++ {
		result[i] = ns.history[len(ns.history)-1-i]
	}

	return result
}

// ClearHistory clears the notification history.
func (ns *NotificationService) ClearHistory() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.history = make([]Notification, 0, ns.maxHist)
}
