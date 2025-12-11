package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewDashboard() string {
	// 1. Status Bar
	var statusText string
	var statusColor lipgloss.Style

	if m.status != nil && m.status.Connected {
		statusText = "CONNECTED"
		statusColor = StatusValStyle.Foreground(ColorSuccess)
	} else {
		statusText = "DISCONNECTED"
		statusColor = StatusValStyle.Foreground(ColorError)
	}

	// Add IPC mode indicator
	ipcMode := "HTTP"
	if m.usingGRPC {
		ipcMode = "gRPC"
	}

	statusBar := RenderStatusBar(
		StatusKeyStyle.Render(" STATUS: "),
		statusColor.Render(statusText),
		StatusKeyStyle.Render(" IP: "),
		StatusValStyle.Render(func() string {
			if m.status != nil && m.status.IP != "" {
				return m.status.IP
			}
			return "N/A"
		}()),
		StatusKeyStyle.Render(" NETWORK: "),
		StatusValStyle.Render(func() string {
			if m.status != nil && m.status.NetworkName != "" {
				return m.status.NetworkName
			}
			return "None"
		}()),
		StatusKeyStyle.Render(" IPC: "),
		StatusValStyle.Foreground(ColorSecondary).Render(ipcMode),
	)

	// 2. Main Content (Split View)
	// Left: Menu
	menuView := m.list.View()

	// Right: Details / Stats
	var detailsView string
	if m.status != nil {
		detailsView = fmt.Sprintf(
			"%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s",
			CardTitleStyle.Render("Network Statistics"),
			StatusKeyStyle.Render("Online Members:"),
			StatusValStyle.Render(fmt.Sprintf("%d", m.status.OnlineMembers)),
			StatusKeyStyle.Render("Your Role:"),
			StatusValStyle.Render(m.status.Role),
			StatusKeyStyle.Render("Networks Joined:"),
			StatusValStyle.Render(fmt.Sprintf("%d", len(m.status.Networks))),
		)
	} else {
		detailsView = "Loading stats..."
	}

	rightPanel := CardStyle.Width(40).Height(m.height - 8).Render(detailsView)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().MarginRight(2).Render(menuView),
		rightPanel,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		mainContent,
		"\n",
		statusBar,
	)
}

func (m Model) viewForm(title, label string) string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		CardStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				CardTitleStyle.Render(title),
				lipgloss.NewStyle().MarginBottom(1).Render(label),
				m.input.View(),
				"\n",
				StatusValStyle.Render("(Press Enter to confirm, Esc to cancel)"),
			),
		),
	)
}

func (m Model) viewPeers() string {
	var content strings.Builder

	content.WriteString(CardTitleStyle.Render("Connected Peers"))
	content.WriteString("\n\n")

	if len(m.peers) == 0 {
		content.WriteString(StatusValStyle.Foreground(ColorMuted).Render("No peers connected"))
	} else {
		for _, peer := range m.peers {
			statusIcon := "●"
			statusStyle := StatusValStyle.Foreground(ColorMuted)

			switch peer.Status {
			case "connected":
				statusStyle = StatusValStyle.Foreground(ColorSuccess)
			case "connecting":
				statusStyle = StatusValStyle.Foreground(ColorWarning)
			case "failed":
				statusStyle = StatusValStyle.Foreground(ColorError)
			}

			connType := ""
			if peer.ConnectionType == "direct" {
				connType = " (P2P)"
			} else if peer.ConnectionType == "relay" {
				connType = " (Relay)"
			}

			latency := ""
			if peer.LatencyMs > 0 {
				latency = fmt.Sprintf(" %dms", peer.LatencyMs)
			}

			content.WriteString(fmt.Sprintf(
				"%s %s %s%s%s\n",
				statusStyle.Render(statusIcon),
				StatusKeyStyle.Render(peer.Name),
				StatusValStyle.Foreground(ColorMuted).Render(peer.VirtualIP),
				StatusValStyle.Foreground(ColorSecondary).Render(connType),
				StatusValStyle.Foreground(ColorMuted).Render(latency),
			))
		}
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		CardStyle.Width(60).Render(content.String()),
	)
}
