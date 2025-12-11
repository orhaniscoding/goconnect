// Package deeplink handles goconnect:// URL parsing and processing.
// It provides a unified interface for deep link actions like joining networks,
// viewing network details, and authentication flows.
package deeplink

import (
	"fmt"
	"net/url"
	"strings"
)

// Action represents the type of deep link action
type Action string

const (
	// ActionUnknown represents an unrecognized action
	ActionUnknown Action = "unknown"
	// ActionLogin handles authentication via goconnect://login?token=...&server=...
	ActionLogin Action = "login"
	// ActionJoin handles network joining via goconnect://join/<invite-code>
	ActionJoin Action = "join"
	// ActionNetwork handles network viewing via goconnect://network/<network-id>
	ActionNetwork Action = "network"
	// ActionConnect handles peer connection via goconnect://connect/<peer-id>
	ActionConnect Action = "connect"
)

// DeepLink represents a parsed goconnect:// URL
type DeepLink struct {
	// Action is the type of deep link (join, network, login, etc.)
	Action Action
	// Target is the primary identifier (invite code, network ID, etc.)
	Target string
	// Params contains any query parameters
	Params map[string]string
	// RawURL is the original URL string
	RawURL string
}

// Parse parses a goconnect:// URL into a DeepLink struct
func Parse(rawURL string) (*DeepLink, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("empty URL")
	}

	// Ensure it's a goconnect:// URL
	if !strings.HasPrefix(rawURL, "goconnect://") {
		return nil, fmt.Errorf("invalid scheme: expected goconnect://, got %s", rawURL)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	dl := &DeepLink{
		RawURL: rawURL,
		Params: make(map[string]string),
	}

	// Parse query parameters
	for key, values := range u.Query() {
		if len(values) > 0 {
			dl.Params[key] = values[0]
		}
	}

	// Determine action from host
	action := strings.ToLower(u.Host)
	path := strings.Trim(u.Path, "/")

	switch action {
	case "login":
		dl.Action = ActionLogin
		// Login doesn't have a path target, uses query params
	case "join":
		dl.Action = ActionJoin
		dl.Target = path
		if dl.Target == "" {
			return nil, fmt.Errorf("join action requires invite code: goconnect://join/<invite-code>")
		}
	case "network":
		dl.Action = ActionNetwork
		dl.Target = path
		if dl.Target == "" {
			return nil, fmt.Errorf("network action requires network ID: goconnect://network/<network-id>")
		}
	case "connect":
		dl.Action = ActionConnect
		dl.Target = path
		if dl.Target == "" {
			return nil, fmt.Errorf("connect action requires peer ID: goconnect://connect/<peer-id>")
		}
	default:
		dl.Action = ActionUnknown
		dl.Target = action
	}

	return dl, nil
}

// String returns a human-readable representation of the deep link
func (d *DeepLink) String() string {
	switch d.Action {
	case ActionLogin:
		return "Login request"
	case ActionJoin:
		return fmt.Sprintf("Join network with invite code: %s", d.Target)
	case ActionNetwork:
		return fmt.Sprintf("View network: %s", d.Target)
	case ActionConnect:
		return fmt.Sprintf("Connect to peer: %s", d.Target)
	default:
		return fmt.Sprintf("Unknown action: %s", d.Target)
	}
}

// BuildURL creates a goconnect:// URL from components
func BuildURL(action Action, target string, params map[string]string) string {
	u := url.URL{
		Scheme: "goconnect",
		Host:   string(action),
	}

	if target != "" {
		u.Path = "/" + target
	}

	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}
