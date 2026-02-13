package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type updateCache struct {
	LatestVersion string `json:"latest_version"`
	CheckedAt     string `json:"checked_at"`
}

func cachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".switcher-update.json"), nil
}

// CheckForUpdate reads the cache and returns a notification message if a newer
// version exists. It fires a background goroutine to refresh the cache when
// stale (>24h) or missing. Returns "" immediately for dev builds.
func CheckForUpdate(currentVersion string) string {
	if currentVersion == "dev" {
		return ""
	}

	path, err := cachePath()
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Cache missing — refresh in background
		go refreshCache(path)
		return ""
	}

	var cache updateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		go refreshCache(path)
		return ""
	}

	checkedAt, err := time.Parse(time.RFC3339, cache.CheckedAt)
	if err != nil || time.Since(checkedAt) > 24*time.Hour {
		go refreshCache(path)
	}

	if cache.LatestVersion != "" && isNewer(cache.LatestVersion, currentVersion) {
		return fmt.Sprintf(
			"A new version of switcher is available: %s (current: %s)\nRun \"switcher update\" to update.",
			cache.LatestVersion, currentVersion,
		)
	}

	return ""
}

// isNewer reports whether latest is a newer semver than current.
// Both may optionally have a "v" prefix.
func isNewer(latest, current string) bool {
	latestParts := parseSemver(latest)
	currentParts := parseSemver(current)
	if latestParts == nil || currentParts == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		// Strip any pre-release suffix (e.g. "0-rc1")
		if idx := strings.IndexByte(p, '-'); idx >= 0 {
			p = p[:idx]
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func fetchLatestRelease(timeout time.Duration) (*githubRelease, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get("https://api.github.com/repos/lucas-stellet/switcher/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func refreshCache(path string) {
	release, err := fetchLatestRelease(10 * time.Second)
	if err != nil {
		return
	}

	cache := updateCache{
		LatestVersion: release.TagName,
		CheckedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	// Atomic write: temp file + rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return
	}
	os.Rename(tmp, path)
}

// WriteCache updates the cache file with the given version. Used after a
// successful self-update so the notification disappears.
func WriteCache(version string) {
	path, err := cachePath()
	if err != nil {
		return
	}

	cache := updateCache{
		LatestVersion: version,
		CheckedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return
	}
	os.Rename(tmp, path)
}
