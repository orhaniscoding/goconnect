package websocket

import (
	"testing"
	"time"
)

func TestNewNotification(t *testing.T) {
	tests := []struct {
		name      string
		ntype     NotificationType
		title     string
		body      string
		wantType  NotificationType
		wantTitle string
		wantBody  string
	}{
		{
			name:      "chat notification",
			ntype:     NotificationTypeChat,
			title:     "New Message",
			body:      "John: Hello!",
			wantType:  NotificationTypeChat,
			wantTitle: "New Message",
			wantBody:  "John: Hello!",
		},
		{
			name:      "network notification",
			ntype:     NotificationTypeNetwork,
			title:     "Network Update",
			body:      "A new member joined",
			wantType:  NotificationTypeNetwork,
			wantTitle: "Network Update",
			wantBody:  "A new member joined",
		},
		{
			name:      "system notification",
			ntype:     NotificationTypeSystem,
			title:     "Update Available",
			body:      "Version 3.1.0 is ready",
			wantType:  NotificationTypeSystem,
			wantTitle: "Update Available",
			wantBody:  "Version 3.1.0 is ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNotification(tt.ntype, tt.title, tt.body)

			if n.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", n.Type, tt.wantType)
			}
			if n.Title != tt.wantTitle {
				t.Errorf("Title = %v, want %v", n.Title, tt.wantTitle)
			}
			if n.Body != tt.wantBody {
				t.Errorf("Body = %v, want %v", n.Body, tt.wantBody)
			}
			if n.Priority != PriorityNormal {
				t.Errorf("Priority = %v, want %v", n.Priority, PriorityNormal)
			}
			if n.ID == "" {
				t.Error("ID should not be empty")
			}
			if n.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
		})
	}
}

func TestNotificationBuilder(t *testing.T) {
	n := NewNotification(NotificationTypeChat, "Test", "Body").
		WithPriority(PriorityUrgent).
		WithAction("goconnect://network/abc123").
		WithNetwork("net-123").
		WithSender("user-456").
		WithData(map[string]any{"message_id": "msg-789"})

	if n.Priority != PriorityUrgent {
		t.Errorf("Priority = %v, want %v", n.Priority, PriorityUrgent)
	}
	if n.Action != "goconnect://network/abc123" {
		t.Errorf("Action = %v, want goconnect://network/abc123", n.Action)
	}
	if n.NetworkID != "net-123" {
		t.Errorf("NetworkID = %v, want net-123", n.NetworkID)
	}
	if n.SenderID != "user-456" {
		t.Errorf("SenderID = %v, want user-456", n.SenderID)
	}
	if n.Data["message_id"] != "msg-789" {
		t.Errorf("Data[message_id] = %v, want msg-789", n.Data["message_id"])
	}
}

func TestNotificationSilent(t *testing.T) {
	n := NewNotification(NotificationTypePeer, "Peer Online", "John is online").AsSilent()

	if !n.Silent {
		t.Error("Expected Silent to be true")
	}
}

func TestNotificationPriorityConstants(t *testing.T) {
	if PriorityUrgent != "urgent" {
		t.Errorf("PriorityUrgent = %v, want urgent", PriorityUrgent)
	}
	if PriorityNormal != "normal" {
		t.Errorf("PriorityNormal = %v, want normal", PriorityNormal)
	}
	if PriorityLow != "low" {
		t.Errorf("PriorityLow = %v, want low", PriorityLow)
	}
}

func TestNotificationTypeConstants(t *testing.T) {
	types := []struct {
		got  NotificationType
		want string
	}{
		{NotificationTypeChat, "chat"},
		{NotificationTypeNetwork, "network"},
		{NotificationTypePeer, "peer"},
		{NotificationTypeTransfer, "transfer"},
		{NotificationTypeSystem, "system"},
	}

	for _, tt := range types {
		if string(tt.got) != tt.want {
			t.Errorf("NotificationType = %v, want %v", tt.got, tt.want)
		}
	}
}

func TestGenerateNotificationID(t *testing.T) {
	id1 := generateNotificationID()
	time.Sleep(time.Millisecond)
	id2 := generateNotificationID()

	if id1 == "" {
		t.Error("ID should not be empty")
	}
	if id1 == id2 {
		t.Error("IDs should be unique")
	}
	if len(id1) < 20 {
		t.Error("ID should be at least 20 characters")
	}
}
