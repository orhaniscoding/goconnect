package tui

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/orhaniscoding/goconnect/cli/internal/uierrors"
)

// SessionState represents the current view of the application
type SessionState int

const (
	StateDashboard SessionState = iota
	StateCreateNetwork
	StateJoinNetwork
	StatePeers
	StateLoading
)

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
	Create key.Binding
	Join   key.Binding
	Copy   key.Binding
}

var Keys = KeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit:   key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q/ctrl+c", "quit")),
	Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Create: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create network")),
	Join:   key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "join network")),
	Copy:   key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy invite")),
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Create, k.Join, k.Back, k.Quit},
	}
}

// Model is the main application model
type Model struct {
	client *UnifiedClient
	state  SessionState
	width  int
	height int

	// Components
	list    list.Model
	input   textinput.Model
	spinner spinner.Model
	help    help.Model
	keys    KeyMap

	// Data
	status     *Status
	networks   []Network
	peers      []Peer
	err        *uierrors.UserError
	loadingMsg string
	nextState  SessionState // State to go to after loading
	usingGRPC  bool         // Shows if gRPC is being used
}

// NewModel creates a new TUI model
func NewModel() Model {
	return NewModelWithState(StateDashboard)
}

// NewModelWithState creates a new TUI model with a specific initial state
func NewModelWithState(initialState SessionState) Model {
	c := NewUnifiedClient()

	// Initialize List (Menu)
	items := []list.Item{
		item{title: "Create Network", desc: "Host a new private LAN"},
		item{title: "Join Network", desc: "Connect to an existing network"},
		item{title: "List Networks", desc: "View your networks"},
		item{title: "View Peers", desc: "See connected peers"},
		item{title: "Settings", desc: "Configure daemon settings"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "GoConnect Menu"
	l.SetShowHelp(false) // We use our own help

	// Initialize Input
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 156
	ti.Width = 30

	// Initialize Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorSecondary)

	model := Model{
		client:    c,
		state:     initialState,
		list:      l,
		input:     ti,
		spinner:   s,
		help:      help.New(),
		keys:      Keys,
		usingGRPC: c.IsUsingGRPC(),
	}

	// If starting in Create or Join state, prepare the input
	if initialState == StateCreateNetwork {
		model.input.Placeholder = "Network Name"
		model.input.Focus()
	} else if initialState == StateJoinNetwork {
		model.input.Placeholder = "Invite Code / Link"
		model.input.Focus()
	}

	return model
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.EnableMouseCellMotion,
		spinner.Tick,
		m.fetchStatusCmd(),
		m.fetchNetworksCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4) // Adjust for header/footer
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Copy):
			if m.status != nil && m.status.Connected {
				_ = clipboard.WriteAll(m.status.InviteCode)
				m.loadingMsg = "Copied to clipboard!"
				return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return tickMsg{} })
			}
		}

	case statusMsg:
		m.status = msg.status
		m.err = nil
		// Re-fetch status periodically
		cmds = append(cmds, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
			return tickMsg{}
		}))

	case networksMsg:
		m.networks = msg.networks

	case peersMsg:
		m.peers = msg.peers
		m.state = StatePeers

	case networkActionMsg:
		m.state = StateDashboard
		m.loadingMsg = ""
		// Refresh data
		cmds = append(cmds, m.fetchStatusCmd(), m.fetchNetworksCmd())

	case errMsg:
		m.err = uierrors.Map(msg.err)
		m.state = StateDashboard // Go back to dashboard on error?

	case tickMsg:
		cmds = append(cmds, m.fetchStatusCmd())
	}

	// Handle State-specific updates
	switch m.state {
	case StateDashboard:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

		if msg, ok := msg.(tea.KeyMsg); ok {
			if key.Matches(msg, m.keys.Enter) {
				selectedItem := m.list.SelectedItem().(item)
				switch selectedItem.title {
				case "Create Network":
					m.state = StateCreateNetwork
					m.input.Placeholder = "Network Name"
					m.input.SetValue("")
					m.input.Focus()
					return m, textinput.Blink
				case "Join Network":
					m.state = StateJoinNetwork
					m.input.Placeholder = "Invite Code / Link"
					m.input.SetValue("")
					m.input.Focus()
					return m, textinput.Blink
				case "View Peers":
					m.state = StateLoading
					m.loadingMsg = "Fetching peers..."
					return m, m.fetchPeersCmd()
				}
			}
		}

	case StateCreateNetwork, StateJoinNetwork:
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)

		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, m.keys.Enter):
				val := m.input.Value()
				if m.state == StateCreateNetwork {
					return m, m.createNetworkCmd(val)
				} else {
					return m, m.joinNetworkCmd(val)
				}
			case key.Matches(msg, m.keys.Back):
				m.state = StateDashboard
			}
		}

	case StateLoading:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case StatePeers:
		if msg, ok := msg.(tea.KeyMsg); ok {
			if key.Matches(msg, m.keys.Back) {
				m.state = StateDashboard
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.state {
	case StateDashboard:
		content = m.viewDashboard()
	case StateCreateNetwork:
		content = m.viewForm("Create Network", "Enter a name for your new network:")
	case StateJoinNetwork:
		content = m.viewForm("Join Network", "Enter the invite code or link:")
	case StatePeers:
		content = m.viewPeers()
	case StateLoading:
		content = fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.loadingMsg)
	}

	// Footer (Help)
	helpView := m.help.View(m.keys)

	// Layout
	return lipgloss.JoinVertical(lipgloss.Left,
		RenderHeader("GoConnect CLI"),
		content,
		"\n",
		helpView,
	)
}

// --- Commands ---

type statusMsg struct{ status *Status }
type networksMsg struct{ networks []Network }
type peersMsg struct{ peers []Peer }
type networkActionMsg struct{}
type errMsg struct{ err error }
type tickMsg struct{}

func (m Model) fetchStatusCmd() tea.Cmd {
	return func() tea.Msg {
		status, err := m.client.GetStatus()
		if err != nil {
			return errMsg{err}
		}
		return statusMsg{status}
	}
}

func (m Model) fetchNetworksCmd() tea.Cmd {
	return func() tea.Msg {
		networks, err := m.client.GetNetworks()
		if err != nil {
			return errMsg{err}
		}
		return networksMsg{networks}
	}
}

func (m Model) fetchPeersCmd() tea.Cmd {
	return func() tea.Msg {
		peers, err := m.client.GetPeers()
		if err != nil {
			return errMsg{err}
		}
		return peersMsg{peers}
	}
}

func (m Model) createNetworkCmd(name string) tea.Cmd {
	m.state = StateLoading
	m.loadingMsg = "Creating network..."
	return func() tea.Msg {
		_, err := m.client.CreateNetwork(name, "")
		if err != nil {
			return errMsg{err}
		}
		return networkActionMsg{}
	}
}

func (m Model) joinNetworkCmd(code string) tea.Cmd {
	m.state = StateLoading
	m.loadingMsg = "Joining network..."
	return func() tea.Msg {
		_, err := m.client.JoinNetwork(code)
		if err != nil {
			return errMsg{err}
		}
		return networkActionMsg{}
	}
}

// --- List Item Helper ---
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
