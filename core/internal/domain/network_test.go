package domain

import (
	"testing"
)

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		wantErr bool
	}{
		{
			name:    "valid CIDR",
			cidr:    "10.0.0.0/24",
			wantErr: false,
		},
		{
			name:    "valid smaller subnet",
			cidr:    "192.168.1.0/28",
			wantErr: false,
		},
		{
			name:    "invalid format",
			cidr:    "10.0.0.0",
			wantErr: true,
		},
		{
			name:    "host address instead of network",
			cidr:    "10.0.0.1/24",
			wantErr: true,
		},
		{
			name:    "invalid IP",
			cidr:    "256.0.0.0/24",
			wantErr: true,
		},
		{
			name:    "invalid subnet mask",
			cidr:    "10.0.0.0/33",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCIDR(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCIDR() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckCIDROverlap(t *testing.T) {
	tests := []struct {
		name        string
		cidr1       string
		cidr2       string
		wantOverlap bool
		wantErr     bool
	}{
		{
			name:        "no overlap - different networks",
			cidr1:       "10.0.0.0/24",
			cidr2:       "192.168.1.0/24",
			wantOverlap: false,
			wantErr:     false,
		},
		{
			name:        "overlap - same network",
			cidr1:       "10.0.0.0/24",
			cidr2:       "10.0.0.0/24",
			wantOverlap: true,
			wantErr:     false,
		},
		{
			name:        "overlap - subnet within larger network",
			cidr1:       "10.0.0.0/16",
			cidr2:       "10.0.1.0/24",
			wantOverlap: true,
			wantErr:     false,
		},
		{
			name:        "overlap - larger network contains subnet",
			cidr1:       "10.0.1.0/24",
			cidr2:       "10.0.0.0/16",
			wantOverlap: true,
			wantErr:     false,
		},
		{
			name:        "no overlap - adjacent networks",
			cidr1:       "10.0.0.0/24",
			cidr2:       "10.0.1.0/24",
			wantOverlap: false,
			wantErr:     false,
		},
		{
			name:        "error - invalid first CIDR",
			cidr1:       "invalid",
			cidr2:       "10.0.0.0/24",
			wantOverlap: false,
			wantErr:     true,
		},
		{
			name:        "error - invalid second CIDR",
			cidr1:       "10.0.0.0/24",
			cidr2:       "invalid",
			wantOverlap: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOverlap, err := CheckCIDROverlap(tt.cidr1, tt.cidr2)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckCIDROverlap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOverlap != tt.wantOverlap {
				t.Errorf("CheckCIDROverlap() = %v, want %v", gotOverlap, tt.wantOverlap)
			}
		})
	}
}

func TestGenerateNetworkID(t *testing.T) {
	id1 := GenerateNetworkID()
	id2 := GenerateNetworkID()

	if id1 == "" {
		t.Error("GenerateNetworkID() returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateNetworkID() returned duplicate IDs")
	}

	// Check that ID starts with expected prefix
	if len(id1) < 5 || id1[:4] != "net_" {
		t.Errorf("GenerateNetworkID() = %v, expected to start with 'net_'", id1)
	}
}

func TestHashRequestBody(t *testing.T) {
	req1 := CreateNetworkRequest{
		Name:       "test",
		Visibility: NetworkVisibilityPublic,
		JoinPolicy: JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
	}

	req2 := CreateNetworkRequest{
		Name:       "test",
		Visibility: NetworkVisibilityPublic,
		JoinPolicy: JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
	}

	req3 := CreateNetworkRequest{
		Name:       "different",
		Visibility: NetworkVisibilityPublic,
		JoinPolicy: JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
	}

	hash1, err := HashRequestBody(req1)
	if err != nil {
		t.Fatalf("HashRequestBody() error = %v", err)
	}

	hash2, err := HashRequestBody(req2)
	if err != nil {
		t.Fatalf("HashRequestBody() error = %v", err)
	}

	hash3, err := HashRequestBody(req3)
	if err != nil {
		t.Fatalf("HashRequestBody() error = %v", err)
	}

	if hash1 != hash2 {
		t.Error("HashRequestBody() should return same hash for identical requests")
	}

	if hash1 == hash3 {
		t.Error("HashRequestBody() should return different hash for different requests")
	}
}

// unMarshalable is a type that causes json.Marshal to fail
type unMarshalable struct {
	Ch chan int
}

func TestHashRequestBody_Error(t *testing.T) {
	// json.Marshal fails for channels
	_, err := HashRequestBody(unMarshalable{Ch: make(chan int)})
	if err == nil {
		t.Error("HashRequestBody() should return error for un-marshalable type")
	}
}

func TestValidateNetworkName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid simple name",
			input:    "My Network",
			wantName: "My Network",
			wantErr:  false,
		},
		{
			name:     "valid name with hyphens",
			input:    "Gaming-LAN-Server",
			wantName: "Gaming-LAN-Server",
			wantErr:  false,
		},
		{
			name:     "valid name with underscores",
			input:    "Dev_Team_Network",
			wantName: "Dev_Team_Network",
			wantErr:  false,
		},
		{
			name:     "valid minimum length",
			input:    "ABC",
			wantName: "ABC",
			wantErr:  false,
		},
		{
			name:     "valid 50 char name",
			input:    "12345678901234567890123456789012345678901234567890",
			wantName: "12345678901234567890123456789012345678901234567890",
			wantErr:  false,
		},
		{
			name:     "trims whitespace",
			input:    "  My Network  ",
			wantName: "My Network",
			wantErr:  false,
		},
		{
			name:     "collapses multiple spaces",
			input:    "My    Network",
			wantName: "My Network",
			wantErr:  false,
		},
		{
			name:    "too short",
			input:   "AB",
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name:    "too short after trim",
			input:   "  A  ",
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name:    "too long",
			input:   "123456789012345678901234567890123456789012345678901",
			wantErr: true,
			errMsg:  "at most 50 characters",
		},
		{
			name:    "leading hyphen",
			input:   "-Network",
			wantErr: true,
			errMsg:  "cannot start or end with hyphens",
		},
		{
			name:    "trailing hyphen",
			input:   "Network-",
			wantErr: true,
			errMsg:  "cannot start or end with hyphens",
		},
		{
			name:    "leading underscore",
			input:   "_Network",
			wantErr: true,
			errMsg:  "cannot start or end with hyphens",
		},
		{
			name:    "trailing underscore",
			input:   "Network_",
			wantErr: true,
			errMsg:  "cannot start or end with hyphens",
		},
		{
			name:    "invalid special chars",
			input:   "My@Network!",
			wantErr: true,
			errMsg:  "can only contain letters, numbers",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := ValidateNetworkName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNetworkName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotName != tt.wantName {
				t.Errorf("ValidateNetworkName() = %q, want %q", gotName, tt.wantName)
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateNetworkName() error = %v, should contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

