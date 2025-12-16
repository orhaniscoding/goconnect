package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Version info (set at build time via ldflags)
var (
	Version   = "dev"
	BuildDate = "unknown"
	Commit    = "unknown"
)

// GitHubRelease represents a GitHub release response.
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
	HTMLURL     string    `json:"html_url"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

// Updater handles checking and applying updates.
type Updater struct {
	GitHubOwner  string
	GitHubRepo   string
	CurrentVer   string
	HTTPClient   *http.Client
}

// NewUpdater creates a new updater instance.
func NewUpdater(owner, repo, currentVersion string) *Updater {
	return &Updater{
		GitHubOwner: owner,
		GitHubRepo:  repo,
		CurrentVer:  currentVersion,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckForUpdates checks GitHub for the latest release.
func (u *Updater) CheckForUpdates(ctx context.Context) (*GitHubRelease, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.GitHubOwner, u.GitHubRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "GoConnect-CLI/"+u.CurrentVer)

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil // No releases yet
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Skip drafts and prereleases
	if release.Draft || release.Prerelease {
		return nil, false, nil
	}

	// Compare versions
	latestVer := strings.TrimPrefix(release.TagName, "v")
	currentVer := strings.TrimPrefix(u.CurrentVer, "v")

	if currentVer == "dev" {
		return &release, false, nil // Dev builds always considered up-to-date
	}

	hasUpdate := compareVersions(latestVer, currentVer) > 0

	return &release, hasUpdate, nil
}

// GetAssetForPlatform finds the matching asset for current OS/arch.
func (u *Updater) GetAssetForPlatform(release *GitHubRelease) *Asset {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Map to common naming patterns
	patterns := []string{
		fmt.Sprintf("goconnect_%s_%s", osName, archName),
		fmt.Sprintf("goconnect-%s-%s", osName, archName),
		fmt.Sprintf("GoConnect_%s_%s", osName, archName),
	}

	for _, asset := range release.Assets {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(pattern)) {
				return &asset
			}
		}
	}

	return nil
}

// DownloadAsset downloads the release asset to a temporary file.
func (u *Updater) DownloadAsset(ctx context.Context, asset *Asset, progress func(downloaded, total int64)) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Create temp file
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "goconnect-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return "", fmt.Errorf("failed to write: %w", writeErr)
			}
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, asset.Size)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("failed to read: %w", err)
		}
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

// ApplyUpdate replaces the current binary with the downloaded one.
func (u *Updater) ApplyUpdate(downloadPath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// For .tar.gz files, extract first
	if strings.HasSuffix(downloadPath, ".tar.gz") || strings.HasSuffix(downloadPath, ".tgz") {
		extractDir := filepath.Dir(downloadPath)
		cmd := exec.Command("tar", "-xzf", downloadPath, "-C", extractDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
		// Find the binary in extracted files
		downloadPath = filepath.Join(extractDir, "goconnect")
		if runtime.GOOS == "windows" {
			downloadPath += ".exe"
		}
	}

	// Make executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(downloadPath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}

	// Backup current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary
	if err := os.Rename(downloadPath, execPath); err != nil {
		// Restore backup
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	return nil
}

// compareVersions compares semver strings. Returns 1 if a > b, -1 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < 3; i++ {
		var aVal, bVal int
		if i < len(aParts) {
			fmt.Sscanf(aParts[i], "%d", &aVal)
		}
		if i < len(bParts) {
			fmt.Sscanf(bParts[i], "%d", &bVal)
		}

		if aVal > bVal {
			return 1
		}
		if aVal < bVal {
			return -1
		}
	}
	return 0
}

// UpdateCmd creates the update command.
func UpdateCmd(version string) *cobra.Command {
	var (
		checkOnly bool
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for and apply updates",
		Long:  "Check for new GoConnect releases on GitHub and optionally apply updates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			updater := NewUpdater("orhaniscoding", "goconnect", version)

			fmt.Printf("🔍 Checking for updates...\n")
			fmt.Printf("   Current version: %s\n\n", version)

			release, hasUpdate, err := updater.CheckForUpdates(ctx)
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if release == nil {
				fmt.Println("ℹ️  No releases found.")
				return nil
			}

			fmt.Printf("📦 Latest version: %s\n", release.TagName)
			fmt.Printf("   Published: %s\n", release.PublishedAt.Format("2006-01-02"))
			fmt.Printf("   URL: %s\n\n", release.HTMLURL)

			if !hasUpdate && !force {
				fmt.Println("✅ You are running the latest version!")
				return nil
			}

			if hasUpdate {
				fmt.Println("🆕 A new version is available!")
			}

			if checkOnly {
				fmt.Println("\nRun 'goconnect update' without --check to install.")
				return nil
			}

			// Find asset
			asset := updater.GetAssetForPlatform(release)
			if asset == nil {
				fmt.Printf("⚠️  No binary found for %s/%s\n", runtime.GOOS, runtime.GOARCH)
				fmt.Printf("   Please download manually from: %s\n", release.HTMLURL)
				return nil
			}

			fmt.Printf("📥 Downloading %s (%d MB)...\n", asset.Name, asset.Size/1024/1024)

			downloadPath, err := updater.DownloadAsset(ctx, asset, func(downloaded, total int64) {
				pct := float64(downloaded) / float64(total) * 100
				fmt.Printf("\r   Progress: %.1f%%", pct)
			})
			if err != nil {
				return fmt.Errorf("failed to download: %w", err)
			}
			fmt.Println()

			fmt.Println("🔧 Applying update...")
			if err := updater.ApplyUpdate(downloadPath); err != nil {
				return fmt.Errorf("failed to apply update: %w", err)
			}

			fmt.Println("✅ Update complete! Please restart GoConnect.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force update even if already on latest")

	return cmd
}
