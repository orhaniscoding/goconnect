package websocket

import "time"

// NotificationPriority defines the urgency level of a notification.
type NotificationPriority string

const (
	// PriorityUrgent for time-sensitive notifications (e.g., incoming call).
	PriorityUrgent NotificationPriority = "urgent"
	// PriorityNormal for standard notifications (e.g., new message).
	PriorityNormal NotificationPriority = "normal"
	// PriorityLow for non-urgent notifications (e.g., status updates).
	PriorityLow NotificationPriority = "low"
)

// NotificationType categorizes the notification source.
type NotificationType string

const (
	// NotificationTypeChat for chat-related notifications.
	NotificationTypeChat NotificationType = "chat"
	// NotificationTypeNetwork for network events (join, leave, status).
	NotificationTypeNetwork NotificationType = "network"
	// NotificationTypePeer for peer-related events (online, offline).
	NotificationTypePeer NotificationType = "peer"
	// NotificationTypeTransfer for file transfer events.
	NotificationTypeTransfer NotificationType = "transfer"
	// NotificationTypeSystem for system-level notifications.
	NotificationTypeSystem NotificationType = "system"
)

// NotificationPayload represents a push notification message.
type NotificationPayload struct {
	// ID is a unique identifier for the notification.
	ID string `json:"id"`

	// Type categorizes the notification (chat, network, peer, transfer, system).
	Type NotificationType `json:"type"`

	// Priority determines how urgently to display the notification.
	Priority NotificationPriority `json:"priority"`

	// Title is the notification headline.
	Title string `json:"title"`

	// Body is the notification content.
	Body string `json:"body"`

	// Icon is an optional icon identifier or URL.
	Icon string `json:"icon,omitempty"`

	// Action is an optional deep link or action identifier.
	Action string `json:"action,omitempty"`

	// Data contains additional context-specific information.
	Data map[string]any `json:"data,omitempty"`

	// Timestamp when the notification was created.
	Timestamp time.Time `json:"timestamp"`

	// NetworkID is the related network, if applicable.
	NetworkID string `json:"network_id,omitempty"`

	// SenderID is the user who triggered the notification, if applicable.
	SenderID string `json:"sender_id,omitempty"`

	// Silent indicates whether to suppress sound/vibration.
	Silent bool `json:"silent,omitempty"`
}

// NewNotification creates a new notification with defaults.
func NewNotification(ntype NotificationType, title, body string) *NotificationPayload {
	return &NotificationPayload{
		ID:        generateNotificationID(),
		Type:      ntype,
		Priority:  PriorityNormal,
		Title:     title,
		Body:      body,
		Timestamp: time.Now(),
	}
}

// WithPriority sets the notification priority.
func (n *NotificationPayload) WithPriority(p NotificationPriority) *NotificationPayload {
	n.Priority = p
	return n
}

// WithAction sets a deep link or action.
func (n *NotificationPayload) WithAction(action string) *NotificationPayload {
	n.Action = action
	return n
}

// WithData adds additional context data.
func (n *NotificationPayload) WithData(data map[string]any) *NotificationPayload {
	n.Data = data
	return n
}

// WithNetwork sets the related network.
func (n *NotificationPayload) WithNetwork(networkID string) *NotificationPayload {
	n.NetworkID = networkID
	return n
}

// WithSender sets the triggering user.
func (n *NotificationPayload) WithSender(senderID string) *NotificationPayload {
	n.SenderID = senderID
	return n
}

// AsSilent marks the notification as silent.
func (n *NotificationPayload) AsSilent() *NotificationPayload {
	n.Silent = true
	return n
}

// generateNotificationID creates a unique notification ID.
func generateNotificationID() string {
	return "notif-" + time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random alphanumeric string.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
