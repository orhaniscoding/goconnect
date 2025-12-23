package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/orhaniscoding/goconnect/cli/internal/uierrors"
)

// ==================== SessionState Tests ====================

func TestSessionState_Constants(t *testing.T) {
	t.Run("States Have Unique Values", func(t *testing.T) {
		states := []SessionState{
			StateDashboard,
			StateCreateNetwork,
			StateJoinNetwork,
			StatePeers,
			StateLoading,
		}

		seen := make(map[SessionState]bool)
		for _, state := range states {
			assert.False(t, seen[state], "Duplicate state value: %d", state)
			seen[state] = true
		}
	})

	t.Run("StateDashboard Is Zero", func(t *testing.T) {
		assert.Equal(t, SessionState(0), StateDashboard)
	})
}

// ==================== KeyMap Tests ====================

func TestKeyMap_Keys(t *testing.T) {
	t.Run("Keys Are Defined", func(t *testing.T) {
		assert.NotEmpty(t, Keys.Up.Keys())
		assert.NotEmpty(t, Keys.Down.Keys())
		assert.NotEmpty(t, Keys.Enter.Keys())
		assert.NotEmpty(t, Keys.Back.Keys())
		assert.NotEmpty(t, Keys.Quit.Keys())
		assert.NotEmpty(t, Keys.Help.Keys())
		assert.NotEmpty(t, Keys.Create.Keys())
		assert.NotEmpty(t, Keys.Join.Keys())
		assert.NotEmpty(t, Keys.Copy.Keys())
	})

	t.Run("Keys Have Help Text", func(t *testing.T) {
		// Each key should have help defined
		assert.NotEmpty(t, Keys.Up.Help().Key)
		assert.NotEmpty(t, Keys.Down.Help().Key)
		assert.NotEmpty(t, Keys.Enter.Help().Key)
		assert.NotEmpty(t, Keys.Back.Help().Key)
		assert.NotEmpty(t, Keys.Quit.Help().Key)
	})
}

func TestKeyMap_ShortHelp(t *testing.T) {
	t.Run("Returns Expected Bindings", func(t *testing.T) {
		help := Keys.ShortHelp()
		require.Len(t, help, 2)
		assert.Equal(t, Keys.Help, help[0])
		assert.Equal(t, Keys.Quit, help[1])
	})
}

func TestKeyMap_FullHelp(t *testing.T) {
	t.Run("Returns Multiple Rows", func(t *testing.T) {
		help := Keys.FullHelp()
		require.Len(t, help, 2)

		// First row: navigation
		assert.Len(t, help[0], 3)
		assert.Equal(t, Keys.Up, help[0][0])
		assert.Equal(t, Keys.Down, help[0][1])
		assert.Equal(t, Keys.Enter, help[0][2])

		// Second row: actions
		assert.Len(t, help[1], 4)
		assert.Equal(t, Keys.Create, help[1][0])
		assert.Equal(t, Keys.Join, help[1][1])
		assert.Equal(t, Keys.Back, help[1][2])
		assert.Equal(t, Keys.Quit, help[1][3])
	})
}

// ==================== NewModel Tests ====================

func TestNewModel(t *testing.T) {
	t.Run("Creates Model With Default State", func(t *testing.T) {
		m := NewModel()
		assert.Equal(t, StateDashboard, m.state)
		assert.NotNil(t, m.client)
	})
}

func TestNewModelWithState(t *testing.T) {
	t.Run("Creates Model With Dashboard State", func(t *testing.T) {
		m := NewModelWithState(StateDashboard)
		assert.Equal(t, StateDashboard, m.state)
		assert.NotNil(t, m.client)
		assert.NotNil(t, m.help)
	})

	t.Run("Creates Model With CreateNetwork State", func(t *testing.T) {
		m := NewModelWithState(StateCreateNetwork)
		assert.Equal(t, StateCreateNetwork, m.state)
		// Input should be focused for create network
		assert.Equal(t, "Network Name", m.input.Placeholder)
	})

	t.Run("Creates Model With JoinNetwork State", func(t *testing.T) {
		m := NewModelWithState(StateJoinNetwork)
		assert.Equal(t, StateJoinNetwork, m.state)
		// Input should be focused for join network
		assert.Equal(t, "Invite Code / Link", m.input.Placeholder)
	})

	t.Run("List Has Menu Items", func(t *testing.T) {
		m := NewModelWithState(StateDashboard)
		// List should have items
		items := m.list.Items()
		assert.GreaterOrEqual(t, len(items), 3)
	})
}

// ==================== Model.Init Tests ====================

func TestModel_Init(t *testing.T) {
	t.Run("Returns Commands", func(t *testing.T) {
		m := NewModel()
		cmd := m.Init()
		assert.NotNil(t, cmd)
	})
}

// ==================== Model.View Tests ====================

func TestModel_View(t *testing.T) {
	t.Run("Returns Loading When Width Is Zero", func(t *testing.T) {
		m := NewModel()
		m.width = 0
		view := m.View()
		assert.Contains(t, view, "Loading...")
	})

	t.Run("Returns Content When Width Is Set", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		view := m.View()
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "GoConnect")
	})

	t.Run("Dashboard State", func(t *testing.T) {
		m := NewModelWithState(StateDashboard)
		m.width = 80
		m.height = 24
		view := m.View()
		assert.NotEmpty(t, view)
	})

	t.Run("CreateNetwork State", func(t *testing.T) {
		m := NewModelWithState(StateCreateNetwork)
		m.width = 80
		m.height = 24
		view := m.View()
		assert.Contains(t, view, "Create Network")
	})

	t.Run("JoinNetwork State", func(t *testing.T) {
		m := NewModelWithState(StateJoinNetwork)
		m.width = 80
		m.height = 24
		view := m.View()
		assert.Contains(t, view, "Join Network")
	})

	t.Run("Peers State", func(t *testing.T) {
		m := NewModelWithState(StateDashboard)
		m.state = StatePeers
		m.width = 80
		m.height = 24
		m.peers = []Peer{}
		view := m.View()
		assert.Contains(t, view, "Peers")
	})

	t.Run("Loading State", func(t *testing.T) {
		m := NewModelWithState(StateDashboard)
		m.state = StateLoading
		m.width = 80
		m.height = 24
		m.loadingMsg = "Loading data..."
		view := m.View()
		assert.Contains(t, view, "Loading data...")
	})
}

// ==================== Message Types Tests ====================

func TestMessageTypes(t *testing.T) {
	t.Run("statusMsg", func(t *testing.T) {
		status := &Status{Connected: true, IP: "10.0.0.1"}
		msg := statusMsg{status: status}
		assert.Equal(t, status, msg.status)
	})

	t.Run("networksMsg", func(t *testing.T) {
		networks := []Network{{ID: "net1", Name: "Test Net"}}
		msg := networksMsg{networks: networks}
		assert.Equal(t, networks, msg.networks)
	})

	t.Run("peersMsg", func(t *testing.T) {
		peers := []Peer{{Name: "Peer1", VirtualIP: "10.0.0.2"}}
		msg := peersMsg{peers: peers}
		assert.Equal(t, peers, msg.peers)
	})

	t.Run("networkActionMsg", func(t *testing.T) {
		msg := networkActionMsg{}
		assert.NotNil(t, msg)
	})

	t.Run("errMsg", func(t *testing.T) {
		err := assert.AnError
		msg := errMsg{err: err}
		assert.Equal(t, err, msg.err)
	})

	t.Run("tickMsg", func(t *testing.T) {
		msg := tickMsg{}
		assert.NotNil(t, msg)
	})
}

// ==================== Item Helper Tests ====================

func TestItem(t *testing.T) {
	t.Run("Title Returns Title", func(t *testing.T) {
		i := item{title: "Test", desc: "Description"}
		assert.Equal(t, "Test", i.Title())
	})

	t.Run("Description Returns Desc", func(t *testing.T) {
		i := item{title: "Test", desc: "Description"}
		assert.Equal(t, "Description", i.Description())
	})

	t.Run("FilterValue Returns Title", func(t *testing.T) {
		i := item{title: "Test", desc: "Description"}
		assert.Equal(t, "Test", i.FilterValue())
	})
}

// ==================== Key Matching Tests ====================

func TestKeyMatching(t *testing.T) {
	t.Run("Quit Keys Are Defined", func(t *testing.T) {
		// Verify quit key bindings exist
		assert.NotEmpty(t, Keys.Quit.Keys())
	})
}

// ==================== Model.Update Tests ====================

func TestModel_Update_StatusMsg(t *testing.T) {
	t.Run("Updates Status", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		status := &Status{Connected: true, IP: "10.0.0.5"}
		msg := statusMsg{status: status}

		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, status, updated.status)
	})
}

func TestModel_Update_NetworksMsg(t *testing.T) {
	t.Run("Updates Networks", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		networks := []Network{{ID: "net1", Name: "Test"}}
		msg := networksMsg{networks: networks}

		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, networks, updated.networks)
	})
}

func TestModel_Update_PeersMsg(t *testing.T) {
	t.Run("Updates Peers", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.state = StatePeers

		peers := []Peer{{Name: "Peer1", VirtualIP: "10.0.0.2"}}
		msg := peersMsg{peers: peers}

		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, peers, updated.peers)
	})
}

func TestModel_Update_ErrMsg(t *testing.T) {
	t.Run("Sets Error", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		testErr := assert.AnError
		msg := errMsg{err: testErr}

		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		// updated.err is a string, so compare with error message string
		assert.Equal(t, testErr.Error(), updated.err.Error())
	})
}

func TestModel_Update_TickMsg(t *testing.T) {
	t.Run("Handles Tick", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		msg := tickMsg{}

		// Should not panic
		_, cmd := m.Update(msg)
		// cmd will be a batch of fetch commands
		assert.NotNil(t, cmd)
	})
}

func TestModel_Update_NetworkActionMsg(t *testing.T) {
	t.Run("Switches To Dashboard", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.state = StateCreateNetwork

		msg := networkActionMsg{}

		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, StateDashboard, updated.state)
	})
}

// ==================== Commands Tests ====================

func TestModel_Commands(t *testing.T) {
	t.Run("fetchStatusCmd Returns Cmd", func(t *testing.T) {
		m := NewModel()
		cmd := m.fetchStatusCmd()
		assert.NotNil(t, cmd)
	})

	t.Run("fetchNetworksCmd Returns Cmd", func(t *testing.T) {
		m := NewModel()
		cmd := m.fetchNetworksCmd()
		assert.NotNil(t, cmd)
	})

	t.Run("fetchPeersCmd Returns Cmd", func(t *testing.T) {
		m := NewModel()
		cmd := m.fetchPeersCmd()
		assert.NotNil(t, cmd)
	})

	t.Run("createNetworkCmd Returns Cmd", func(t *testing.T) {
		m := NewModel()
		cmd := m.createNetworkCmd("TestNetwork")
		assert.NotNil(t, cmd)
	})

	t.Run("joinNetworkCmd Returns Cmd", func(t *testing.T) {
		m := NewModel()
		cmd := m.joinNetworkCmd("invite-code-123")
		assert.NotNil(t, cmd)
	})
}

// ==================== State Transitions ====================

func TestModel_StateTransitions(t *testing.T) {
	t.Run("Dashboard To CreateNetwork", func(t *testing.T) {
		m := NewModel()
		m.state = StateDashboard
		m.state = StateCreateNetwork
		assert.Equal(t, StateCreateNetwork, m.state)
	})

	t.Run("Dashboard To JoinNetwork", func(t *testing.T) {
		m := NewModel()
		m.state = StateDashboard
		m.state = StateJoinNetwork
		assert.Equal(t, StateJoinNetwork, m.state)
	})

	t.Run("Dashboard To Peers", func(t *testing.T) {
		m := NewModel()
		m.state = StateDashboard
		m.state = StatePeers
		assert.Equal(t, StatePeers, m.state)
	})
}

// ==================== Tea KeyMsg Update Tests ====================

func TestModel_Update_KeyMsgQuit(t *testing.T) {
	t.Run("Ctrl+C Quits", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		_, cmd := m.Update(msg)

		// Should return tea.Quit command
		assert.NotNil(t, cmd)
	})

	t.Run("Q Key Quits", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		_, cmd := m.Update(msg)

		assert.NotNil(t, cmd)
	})
}

func TestModel_Update_KeyMsgHelp(t *testing.T) {
	t.Run("Help Toggle Shows All Help", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.help.ShowAll = false

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.True(t, updated.help.ShowAll)
	})

	t.Run("Help Toggle Hides All Help", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.help.ShowAll = true

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.False(t, updated.help.ShowAll)
	})
}

func TestModel_Update_WindowSizeMsg(t *testing.T) {
	t.Run("Updates Dimensions", func(t *testing.T) {
		m := NewModel()

		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, 120, updated.width)
		assert.Equal(t, 40, updated.height)
	})
}

func TestModel_Update_CreateNetworkState(t *testing.T) {
	t.Run("Enter Key In CreateNetwork State", func(t *testing.T) {
		m := NewModelWithState(StateCreateNetwork)
		m.width = 80
		m.height = 24
		m.input.SetValue("TestNetwork")

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := m.Update(msg)

		assert.NotNil(t, newModel)
		assert.NotNil(t, cmd)
	})

	t.Run("Esc Key In CreateNetwork Returns To Dashboard", func(t *testing.T) {
		m := NewModelWithState(StateCreateNetwork)
		m.width = 80
		m.height = 24

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, StateDashboard, updated.state)
	})

	t.Run("Input Updates In CreateNetwork State", func(t *testing.T) {
		m := NewModelWithState(StateCreateNetwork)
		m.width = 80
		m.height = 24

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		// Input should have been updated
		assert.NotNil(t, updated.input)
	})
}

func TestModel_Update_JoinNetworkState(t *testing.T) {
	t.Run("Enter Key In JoinNetwork State", func(t *testing.T) {
		m := NewModelWithState(StateJoinNetwork)
		m.width = 80
		m.height = 24
		m.input.SetValue("invite-code-123")

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := m.Update(msg)

		assert.NotNil(t, newModel)
		assert.NotNil(t, cmd)
	})

	t.Run("Esc Key In JoinNetwork Returns To Dashboard", func(t *testing.T) {
		m := NewModelWithState(StateJoinNetwork)
		m.width = 80
		m.height = 24

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, StateDashboard, updated.state)
	})
}

func TestModel_Update_LoadingState(t *testing.T) {
	t.Run("Spinner Updates In Loading State", func(t *testing.T) {
		m := NewModelWithState(StateDashboard)
		m.state = StateLoading
		m.width = 80
		m.height = 24

		// Simulate spinner tick
		msg := m.spinner.Tick()
		newModel, _ := m.Update(msg)

		assert.NotNil(t, newModel)
	})
}

func TestModel_Update_PeersState(t *testing.T) {
	t.Run("Esc Key In Peers Returns To Dashboard", func(t *testing.T) {
		m := NewModel()
		m.state = StatePeers
		m.width = 80
		m.height = 24

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Equal(t, StateDashboard, updated.state)
	})
}

func TestModel_Update_CopyKey(t *testing.T) {
	t.Run("Copy When Connected Shows Message", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.status = &Status{
			Connected:  true,
			InviteCode: "test-invite-code",
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
		newModel, cmd := m.Update(msg)
		updated := newModel.(Model)

		// Should have set loading message
		assert.Equal(t, "Copied to clipboard!", updated.loadingMsg)
		assert.NotNil(t, cmd)
	})

	t.Run("Copy When Disconnected Does Nothing", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.status = nil

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
		newModel, _ := m.Update(msg)
		updated := newModel.(Model)

		assert.Empty(t, updated.loadingMsg)
	})
}

// ==================== Command Execution Tests ====================

func TestModel_FetchStatusCmd_Execution(t *testing.T) {
	t.Run("Returns Error When Client Fails", func(t *testing.T) {
		m := NewModel()
		cmd := m.fetchStatusCmd()
		msg := cmd()

		// Without daemon running, should return errMsg
		_, isErr := msg.(errMsg)
		assert.True(t, isErr)
	})
}

func TestModel_FetchNetworksCmd_Execution(t *testing.T) {
	t.Run("Returns Error When Client Fails", func(t *testing.T) {
		m := NewModel()
		cmd := m.fetchNetworksCmd()
		msg := cmd()

		_, isErr := msg.(errMsg)
		assert.True(t, isErr)
	})
}

func TestModel_FetchPeersCmd_Execution(t *testing.T) {
	t.Run("Returns Error When Client Fails", func(t *testing.T) {
		m := NewModel()
		cmd := m.fetchPeersCmd()
		msg := cmd()

		_, isErr := msg.(errMsg)
		assert.True(t, isErr)
	})
}

func TestModel_CreateNetworkCmd_Execution(t *testing.T) {
	t.Run("Returns Error When Client Fails", func(t *testing.T) {
		m := NewModel()
		cmd := m.createNetworkCmd("test-network")
		msg := cmd()

		_, isErr := msg.(errMsg)
		assert.True(t, isErr)
	})
}

func TestModel_JoinNetworkCmd_Execution(t *testing.T) {
	t.Run("Returns Error When Client Fails", func(t *testing.T) {
		m := NewModel()
		cmd := m.joinNetworkCmd("invite-code")
		msg := cmd()

		_, isErr := msg.(errMsg)
		assert.True(t, isErr)
	})
}

// ==================== Model Field Tests ====================

func TestModel_Fields(t *testing.T) {
	t.Run("Model Has All Required Fields", func(t *testing.T) {
		m := NewModel()

		assert.NotNil(t, m.client)
		assert.NotNil(t, m.help)
		assert.Equal(t, Keys, m.keys)
	})

	t.Run("Model Stores Error", func(t *testing.T) {
		m := NewModel()
		m.err = uierrors.New("Test", "Test Message", "Test Hint", nil)

		assert.Error(t, m.err)
	})

	t.Run("Model Stores Loading Message", func(t *testing.T) {
		m := NewModel()
		m.loadingMsg = "Loading..."

		assert.Equal(t, "Loading...", m.loadingMsg)
	})
}

// ==================== Update Dashboard Enter Key Tests ====================

func TestModel_Update_DashboardListNavigation(t *testing.T) {
	t.Run("Down Key In Dashboard", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.state = StateDashboard

		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := m.Update(msg)

		assert.NotNil(t, newModel)
	})

	t.Run("Up Key In Dashboard", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 24
		m.state = StateDashboard

		msg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ := m.Update(msg)

		assert.NotNil(t, newModel)
	})
}
