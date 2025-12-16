package daemon

import (
	"context"
	"testing"
	"time"
)

func TestNewNotificationService(t *testing.T) {
	settings := DefaultNotificationSettings()
	ns := NewNotificationService(settings)

	if ns == nil {
		t.Fatal("Expected non-nil NotificationService")
	}
	if len(ns.history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(ns.history))
	}
}

func TestDefaultNotificationSettings(t *testing.T) {
	settings := DefaultNotificationSettings()

	if !settings.Enabled {
		t.Error("Expected Enabled to be true by default")
	}
	if !settings.SoundEnabled {
		t.Error("Expected SoundEnabled to be true by default")
	}
	if !settings.ShowPreview {
		t.Error("Expected ShowPreview to be true by default")
	}
	if settings.DoNotDisturb {
		t.Error("Expected DoNotDisturb to be false by default")
	}
}

func TestUpdateSettings(t *testing.T) {
	ns := NewNotificationService(DefaultNotificationSettings())

	newSettings := NotificationSettings{
		Enabled:      false,
		DoNotDisturb: true,
		MutedNetworks: []string{"net-1", "net-2"},
	}

	ns.UpdateSettings(newSettings)
	got := ns.GetSettings()

	if got.Enabled {
		t.Error("Expected Enabled to be false")
	}
	if !got.DoNotDisturb {
		t.Error("Expected DoNotDisturb to be true")
	}
	if len(got.MutedNetworks) != 2 {
		t.Errorf("Expected 2 muted networks, got %d", len(got.MutedNetworks))
	}
}

func TestShow_DisabledNotifications(t *testing.T) {
	settings := DefaultNotificationSettings()
	settings.Enabled = false
	ns := NewNotificationService(settings)

	notification := Notification{
		ID:        "test-1",
		Title:     "Test",
		Body:      "Test body",
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
	}

	err := ns.Show(context.Background(), notification)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should not be added to history when disabled
	history := ns.GetHistory(10)
	if len(history) != 0 {
		t.Error("Expected empty history when notifications disabled")
	}
}

func TestShow_DoNotDisturb(t *testing.T) {
	settings := DefaultNotificationSettings()
	settings.DoNotDisturb = true
	ns := NewNotificationService(settings)

	notification := Notification{
		ID:        "test-2",
		Title:     "Test DND",
		Body:      "Should be skipped",
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
	}

	err := ns.Show(context.Background(), notification)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	history := ns.GetHistory(10)
	if len(history) != 0 {
		t.Error("Expected empty history when DND active")
	}
}

func TestShow_MutedNetwork(t *testing.T) {
	settings := DefaultNotificationSettings()
	settings.MutedNetworks = []string{"muted-net-id"}
	ns := NewNotificationService(settings)

	notification := Notification{
		ID:        "test-3",
		Title:     "Test Muted",
		Body:      "Should be skipped",
		NetworkID: "muted-net-id",
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
	}

	err := ns.Show(context.Background(), notification)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	history := ns.GetHistory(10)
	if len(history) != 0 {
		t.Error("Expected empty history for muted network")
	}
}

func TestOnNotificationCallback(t *testing.T) {
	ns := NewNotificationService(DefaultNotificationSettings())

	var received *Notification
	ns.SetOnNotification(func(n Notification) {
		received = &n
	})

	notification := Notification{
		ID:        "test-4",
		Title:     "Callback Test",
		Body:      "Testing callback",
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
	}

	_ = ns.Show(context.Background(), notification)
	time.Sleep(50 * time.Millisecond) // Wait for goroutine

	if received == nil {
		t.Fatal("Expected callback to be triggered")
	}
	if received.Title != "Callback Test" {
		t.Errorf("Expected title 'Callback Test', got '%s'", received.Title)
	}
}

func TestGetHistory(t *testing.T) {
	ns := NewNotificationService(DefaultNotificationSettings())

	// Add notifications
	for i := 0; i < 5; i++ {
		ns.addToHistory(Notification{
			ID:        string(rune('1' + i)),
			Title:     "Test",
			Timestamp: time.Now(),
		})
	}

	// Get limited history
	history := ns.GetHistory(3)
	if len(history) != 3 {
		t.Errorf("Expected 3 items, got %d", len(history))
	}

	// Most recent first
	if history[0].ID != "5" {
		t.Errorf("Expected most recent first, got ID '%s'", history[0].ID)
	}
}

func TestClearHistory(t *testing.T) {
	ns := NewNotificationService(DefaultNotificationSettings())

	ns.addToHistory(Notification{ID: "1", Title: "Test", Timestamp: time.Now()})
	ns.addToHistory(Notification{ID: "2", Title: "Test", Timestamp: time.Now()})

	ns.ClearHistory()

	history := ns.GetHistory(10)
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(history))
	}
}

func TestHistoryLimit(t *testing.T) {
	settings := DefaultNotificationSettings()
	ns := NewNotificationService(settings)
	ns.maxHist = 5 // Override for testing

	// Add more than limit
	for i := 0; i < 10; i++ {
		ns.addToHistory(Notification{
			ID:        string(rune('0' + i)),
			Title:     "Test",
			Timestamp: time.Now(),
		})
	}

	if len(ns.history) != 5 {
		t.Errorf("Expected history to be capped at 5, got %d", len(ns.history))
	}
}

func TestPriorityConstants(t *testing.T) {
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
