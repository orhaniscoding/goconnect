package voice

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)
	assert.NotNil(t, m.stopChan)
	assert.NotNil(t, m.subscribers)
}

func TestManager_StartStop(t *testing.T) {
	m := NewManager()
	// Bind to port 0 to let OS choose
	err := m.Start("127.0.0.1", 0)
	require.NoError(t, err)

	assert.NotNil(t, m.listener)
	
	// Check we can get the port
	_, port, err := net.SplitHostPort(m.listener.Addr().String())
	require.NoError(t, err)
	t.Logf("Listening on port %s", port)

	m.Stop()
	
	// Listener should be closed
	// Try to dial should fail or refuse
	_, err = net.Dial("tcp", m.listener.Addr().String())
	assert.Error(t, err)
}

func TestManager_Subscribe(t *testing.T) {
	m := NewManager()
	ch := m.Subscribe()
	assert.NotNil(t, ch)
	
	// Check subscriber map
	m.subscribersMu.RLock()
	_, exists := m.subscribers[ch]
	m.subscribersMu.RUnlock()
	assert.True(t, exists)

	m.Unsubscribe(ch)
	
	m.subscribersMu.RLock()
	_, exists = m.subscribers[ch]
	m.subscribersMu.RUnlock()
	assert.False(t, exists)
	
	// Channel should be closed
	_, open := <-ch
	assert.False(t, open)
}

func TestManager_Communication(t *testing.T) {
	receiver := NewManager()
	err := receiver.Start("127.0.0.1", 0)
	require.NoError(t, err)
	defer receiver.Stop()

	// Get receiver port
	_, portStr, _ := net.SplitHostPort(receiver.listener.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	// Setup receiver callback
	receivedSigChan := make(chan Signal, 1)
	receiver.OnSignal(func(sig Signal) {
		receivedSigChan <- sig
	})
	
	// Setup receiver subscription
	subCh := receiver.Subscribe()
	defer receiver.Unsubscribe(subCh)

	// Sender
	sender := NewManager() // Doesn't need to start listening to send
	
	sig := Signal{
		Type:      "offer",
		SDP:       "v=0...",
		SenderID:  "sender1",
		TargetID:  "target1",
		NetworkID: "net1",
	}

	// Send signal
	err = sender.SendSignal("127.0.0.1", port, sig)
	require.NoError(t, err)

	// Verify callback received it
	select {
	case received := <-receivedSigChan:
		assert.Equal(t, sig.Type, received.Type)
		assert.Equal(t, sig.SenderID, received.SenderID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for signal via OnSignal")
	}

	// Verify subscriber received it
	select {
	case received := <-subCh:
		assert.Equal(t, sig.Type, received.Type)
		assert.Equal(t, sig.SenderID, received.SenderID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for signal via Subscribe")
	}
}
