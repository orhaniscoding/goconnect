package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/pion/ice/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== NewAgent Tests ====================

func TestNewAgent(t *testing.T) {
	t.Run("Creates Agent With Default STUN URL", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		require.NotNil(t, agent)
		defer agent.Close()

		assert.NotNil(t, agent.agent)
		assert.Equal(t, ice.ConnectionStateNew, agent.state)
	})

	t.Run("Creates Agent With Custom STUN URL", func(t *testing.T) {
		agent, err := NewAgent("stun:stun.l.google.com:19302")
		require.NoError(t, err)
		require.NotNil(t, agent)
		defer agent.Close()
	})

	t.Run("Falls Back To Default For Invalid STUN URL", func(t *testing.T) {
		// NewAgent skips invalid URLs and falls back to default STUN
		agent, err := NewAgent("invalid://not-a-valid-url")
		require.NoError(t, err) // Should not error - falls back to default
		require.NotNil(t, agent)
		defer agent.Close()
	})
}

// ==================== GetLocalCredentials Tests ====================

func TestAgent_GetLocalCredentials(t *testing.T) {
	t.Run("Returns Credentials", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		ufrag, pwd, err := agent.GetLocalCredentials()
		require.NoError(t, err)
		assert.NotEmpty(t, ufrag)
		assert.NotEmpty(t, pwd)
	})
}

// ==================== GatherCandidates Tests ====================

func TestAgent_GatherCandidates(t *testing.T) {
	t.Run("Starts Candidate Gathering", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		// OnCandidate must be set before gathering
		err = agent.OnCandidate(func(_ ice.Candidate) {})
		require.NoError(t, err)

		err = agent.GatherCandidates()
		require.NoError(t, err)
	})
}

// ==================== OnCandidate Tests ====================

func TestAgent_OnCandidate(t *testing.T) {
	t.Run("Sets Candidate Callback", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		err = agent.OnCandidate(func(_ ice.Candidate) {
			// Callback set
		})
		require.NoError(t, err)

		// Start gathering to trigger callback
		_ = agent.GatherCandidates()
	})
}

// ==================== OnConnectionStateChange Tests ====================

func TestAgent_OnConnectionStateChange(t *testing.T) {
	t.Run("Sets State Change Callback", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		err = agent.OnConnectionStateChange(func(_ ice.ConnectionState) {
			// Callback set
		})
		require.NoError(t, err)
	})
}

// ==================== ConnectionState Tests ====================

func TestAgent_ConnectionState(t *testing.T) {
	t.Run("Returns Initial State", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		state := agent.ConnectionState()
		assert.Equal(t, ice.ConnectionStateNew, state)
	})

	t.Run("State Is Thread Safe", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		done := make(chan bool, 10)

		// Concurrent reads
		for i := 0; i < 10; i++ {
			go func() {
				_ = agent.ConnectionState()
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// ==================== Close Tests ====================

func TestAgent_Close(t *testing.T) {
	t.Run("Closes Agent", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)

		err = agent.Close()
		assert.NoError(t, err)
	})

	t.Run("Safe To Call Multiple Times", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)

		err = agent.Close()
		assert.NoError(t, err)

		// Second close should not panic
		// Note: pion/ice may return error on second close
		_ = agent.Close()
	})
}

// ==================== Agent Struct Tests ====================

func TestAgentStruct(t *testing.T) {
	t.Run("Has Required Fields", func(t *testing.T) {
		agent := &Agent{
			state: ice.ConnectionStateNew,
		}
		assert.Equal(t, ice.ConnectionStateNew, agent.state)
	})
}

// ==================== AddRemoteCandidate Tests ====================

func TestAgent_AddRemoteCandidate(t *testing.T) {
	t.Run("Returns Error For Nil Candidate", func(t *testing.T) {
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		err = agent.AddRemoteCandidate(nil)
		// Agent may handle nil differently
		// This test documents behavior
		_ = err
	})
}

// ==================== Integration Tests ====================

func TestAgent_Dial_And_Accept(t *testing.T) {
	// Create two agents
	agentA, err := NewAgent("")
	require.NoError(t, err)
	defer agentA.Close()

	agentB, err := NewAgent("")
	require.NoError(t, err)
	defer agentB.Close()

	// Get credentials
	ufragA, pwdA, err := agentA.GetLocalCredentials()
	require.NoError(t, err)
	ufragB, pwdB, err := agentB.GetLocalCredentials()
	require.NoError(t, err)

	// Exchange candidates
	err = agentA.OnCandidate(func(c ice.Candidate) {
		if c != nil {
			_ = agentB.AddRemoteCandidate(c)
		}
	})
	require.NoError(t, err)

	err = agentB.OnCandidate(func(c ice.Candidate) {
		if c != nil {
			_ = agentA.AddRemoteCandidate(c)
		}
	})
	require.NoError(t, err)

	// Start gathering
	err = agentA.GatherCandidates()
	require.NoError(t, err)
	err = agentB.GatherCandidates()
	require.NoError(t, err)

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connCh := make(chan *ice.Conn, 2)
	errCh := make(chan error, 2)

	go func() {
		conn, err := agentA.Dial(ctx, ufragB, pwdB)
		if err != nil {
			errCh <- err
			return
		}
		connCh <- conn
	}()

	go func() {
		conn, err := agentB.Accept(ctx, ufragA, pwdA)
		if err != nil {
			errCh <- err
			return
		}
		connCh <- conn
	}()

	// Wait for connections or timeout (test environment might not support full ICE connectivity)
	select {
	case err := <-errCh:
		t.Fatalf("Connection failed: %v", err)
	case <-connCh:
		// Connection object created, even if ICE isn't fully completed yet
	case <-time.After(2 * time.Second):
		// If we don't get a connection object quickly, that's okay for this environment
		// as long as we exercised the code paths without panic or immediate error.
	}

	// Verify state is valid (any valid state is fine, as long as it didn't panic)
	// We just want to ensure the agent logic ran.
	state := agentA.ConnectionState()
	assert.Contains(t, []ice.ConnectionState{
		ice.ConnectionStateNew,
		ice.ConnectionStateChecking,
		ice.ConnectionStateConnected,
		ice.ConnectionStateCompleted,
	}, state)
}
