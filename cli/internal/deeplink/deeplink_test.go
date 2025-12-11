package deeplink

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *DeepLink
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid join URL",
			url:  "goconnect://join/ABC123",
			want: &DeepLink{
				Action: ActionJoin,
				Target: "ABC123",
				Params: map[string]string{},
				RawURL: "goconnect://join/ABC123",
			},
		},
		{
			name: "valid join URL with long invite code",
			url:  "goconnect://join/inv_a1b2c3d4e5f6g7h8",
			want: &DeepLink{
				Action: ActionJoin,
				Target: "inv_a1b2c3d4e5f6g7h8",
				Params: map[string]string{},
				RawURL: "goconnect://join/inv_a1b2c3d4e5f6g7h8",
			},
		},
		{
			name: "valid network URL",
			url:  "goconnect://network/net-123-456",
			want: &DeepLink{
				Action: ActionNetwork,
				Target: "net-123-456",
				Params: map[string]string{},
				RawURL: "goconnect://network/net-123-456",
			},
		},
		{
			name: "valid connect URL",
			url:  "goconnect://connect/peer-abc-xyz",
			want: &DeepLink{
				Action: ActionConnect,
				Target: "peer-abc-xyz",
				Params: map[string]string{},
				RawURL: "goconnect://connect/peer-abc-xyz",
			},
		},
		{
			name: "valid login URL with params",
			url:  "goconnect://login?token=jwt-token-here&server=https://api.example.com",
			want: &DeepLink{
				Action: ActionLogin,
				Target: "",
				Params: map[string]string{
					"token":  "jwt-token-here",
					"server": "https://api.example.com",
				},
				RawURL: "goconnect://login?token=jwt-token-here&server=https://api.example.com",
			},
		},
		{
			name: "unknown action",
			url:  "goconnect://unknown-action",
			want: &DeepLink{
				Action: ActionUnknown,
				Target: "unknown-action",
				Params: map[string]string{},
				RawURL: "goconnect://unknown-action",
			},
		},
		{
			name: "case insensitive action",
			url:  "goconnect://JOIN/CODE123",
			want: &DeepLink{
				Action: ActionJoin,
				Target: "CODE123",
				Params: map[string]string{},
				RawURL: "goconnect://JOIN/CODE123",
			},
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "empty URL",
		},
		{
			name:    "wrong scheme",
			url:     "https://example.com/join/ABC123",
			wantErr: true,
			errMsg:  "invalid scheme",
		},
		{
			name:    "join without invite code",
			url:     "goconnect://join",
			wantErr: true,
			errMsg:  "join action requires invite code",
		},
		{
			name:    "join with empty path",
			url:     "goconnect://join/",
			wantErr: true,
			errMsg:  "join action requires invite code",
		},
		{
			name:    "network without ID",
			url:     "goconnect://network",
			wantErr: true,
			errMsg:  "network action requires network ID",
		},
		{
			name:    "connect without peer ID",
			url:     "goconnect://connect",
			wantErr: true,
			errMsg:  "connect action requires peer ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.url)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want.Action, got.Action)
			assert.Equal(t, tt.want.Target, got.Target)
			assert.Equal(t, tt.want.RawURL, got.RawURL)
			assert.Equal(t, tt.want.Params, got.Params)
		})
	}
}

func TestDeepLink_String(t *testing.T) {
	tests := []struct {
		name     string
		deeplink *DeepLink
		want     string
	}{
		{
			name: "login action",
			deeplink: &DeepLink{
				Action: ActionLogin,
			},
			want: "Login request",
		},
		{
			name: "join action",
			deeplink: &DeepLink{
				Action: ActionJoin,
				Target: "ABC123",
			},
			want: "Join network with invite code: ABC123",
		},
		{
			name: "network action",
			deeplink: &DeepLink{
				Action: ActionNetwork,
				Target: "net-123",
			},
			want: "View network: net-123",
		},
		{
			name: "connect action",
			deeplink: &DeepLink{
				Action: ActionConnect,
				Target: "peer-456",
			},
			want: "Connect to peer: peer-456",
		},
		{
			name: "unknown action",
			deeplink: &DeepLink{
				Action: ActionUnknown,
				Target: "something",
			},
			want: "Unknown action: something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.deeplink.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name   string
		action Action
		target string
		params map[string]string
		want   string
	}{
		{
			name:   "join URL",
			action: ActionJoin,
			target: "ABC123",
			params: nil,
			want:   "goconnect://join/ABC123",
		},
		{
			name:   "network URL",
			action: ActionNetwork,
			target: "net-123",
			params: nil,
			want:   "goconnect://network/net-123",
		},
		{
			name:   "login URL with params",
			action: ActionLogin,
			target: "",
			params: map[string]string{
				"token":  "mytoken",
				"server": "https://api.example.com",
			},
			// Note: Query params order may vary
		},
		{
			name:   "connect URL",
			action: ActionConnect,
			target: "peer-xyz",
			params: nil,
			want:   "goconnect://connect/peer-xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildURL(tt.action, tt.target, tt.params)

			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			} else {
				// For URLs with params, verify it can be parsed back
				dl, err := Parse(got)
				require.NoError(t, err)
				assert.Equal(t, tt.action, dl.Action)
				assert.Equal(t, tt.target, dl.Target)
				for k, v := range tt.params {
					assert.Equal(t, v, dl.Params[k])
				}
			}
		})
	}
}

func TestActionConstants(t *testing.T) {
	// Verify action constants have expected string values
	assert.Equal(t, Action("unknown"), ActionUnknown)
	assert.Equal(t, Action("login"), ActionLogin)
	assert.Equal(t, Action("join"), ActionJoin)
	assert.Equal(t, Action("network"), ActionNetwork)
	assert.Equal(t, Action("connect"), ActionConnect)
}

func TestParse_QueryParams(t *testing.T) {
	// Test that query params are properly parsed
	url := "goconnect://join/ABC123?ref=friend&campaign=summer"
	dl, err := Parse(url)
	require.NoError(t, err)

	assert.Equal(t, ActionJoin, dl.Action)
	assert.Equal(t, "ABC123", dl.Target)
	assert.Equal(t, "friend", dl.Params["ref"])
	assert.Equal(t, "summer", dl.Params["campaign"])
}

func TestParse_MultipleSlashesInPath(t *testing.T) {
	// Test with multiple path segments
	url := "goconnect://join/path/with/slashes"
	dl, err := Parse(url)
	require.NoError(t, err)

	assert.Equal(t, ActionJoin, dl.Action)
	// The full path (minus leading slash) should be the target
	assert.Equal(t, "path/with/slashes", dl.Target)
}
