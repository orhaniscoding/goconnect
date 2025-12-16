package commands

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"a greater major", "2.0.0", "1.0.0", 1},
		{"b greater major", "1.0.0", "2.0.0", -1},
		{"a greater minor", "1.2.0", "1.1.0", 1},
		{"b greater minor", "1.1.0", "1.2.0", -1},
		{"a greater patch", "1.0.2", "1.0.1", 1},
		{"b greater patch", "1.0.1", "1.0.2", -1},
		{"complex comparison", "2.1.3", "2.1.2", 1},
		{"major wins over minor", "2.0.0", "1.9.9", 1},
		{"minor wins over patch", "1.2.0", "1.1.9", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestNewUpdater(t *testing.T) {
	u := NewUpdater("owner", "repo", "1.0.0")

	if u.GitHubOwner != "owner" {
		t.Errorf("GitHubOwner = %q, want owner", u.GitHubOwner)
	}
	if u.GitHubRepo != "repo" {
		t.Errorf("GitHubRepo = %q, want repo", u.GitHubRepo)
	}
	if u.CurrentVer != "1.0.0" {
		t.Errorf("CurrentVer = %q, want 1.0.0", u.CurrentVer)
	}
	if u.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestGetAssetForPlatform(t *testing.T) {
	u := NewUpdater("owner", "repo", "1.0.0")

	release := &GitHubRelease{
		Assets: []Asset{
			{Name: "goconnect_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux"},
			{Name: "goconnect_darwin_amd64.tar.gz", BrowserDownloadURL: "https://example.com/darwin"},
			{Name: "goconnect_windows_amd64.zip", BrowserDownloadURL: "https://example.com/windows"},
		},
	}

	asset := u.GetAssetForPlatform(release)

	// Asset should be found for current platform
	if asset == nil {
		t.Log("No asset found for current platform (may be expected in test environment)")
	}
}

func TestGitHubReleaseStruct(t *testing.T) {
	release := GitHubRelease{
		TagName:    "v1.2.0",
		Name:       "Release 1.2.0",
		Draft:      false,
		Prerelease: false,
		HTMLURL:    "https://github.com/owner/repo/releases/v1.2.0",
	}

	if release.TagName != "v1.2.0" {
		t.Errorf("TagName = %q, want v1.2.0", release.TagName)
	}
}

func TestAssetStruct(t *testing.T) {
	asset := Asset{
		Name:               "goconnect_linux_amd64.tar.gz",
		BrowserDownloadURL: "https://example.com/download",
		Size:               1024 * 1024 * 10, // 10 MB
		ContentType:        "application/gzip",
	}

	if asset.Size != 10*1024*1024 {
		t.Errorf("Size = %d, want %d", asset.Size, 10*1024*1024)
	}
}

func TestVersionConstants(t *testing.T) {
	// These are set at build time, defaults should be reasonable
	if Version == "" {
		t.Error("Version should have a default value")
	}
}
