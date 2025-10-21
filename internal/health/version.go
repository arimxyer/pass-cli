package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VersionChecker checks if the binary version is up to date
type VersionChecker struct {
	currentVersion string
	githubRepo     string
	apiBaseURL     string // For testing - defaults to GitHub API
}

// NewVersionChecker creates a new version checker
func NewVersionChecker(currentVersion string, githubRepo string) HealthChecker {
	return &VersionChecker{
		currentVersion: currentVersion,
		githubRepo:     githubRepo,
		apiBaseURL:     "https://api.github.com", // Default to real GitHub API
	}
}

// Name returns the check name
func (v *VersionChecker) Name() string {
	return "version"
}

// Run executes the version check
func (v *VersionChecker) Run(ctx context.Context) CheckResult {
	details := VersionCheckDetails{
		Current:  v.currentVersion,
		UpToDate: true,
	}

	// Query GitHub API for latest release
	latest, updateURL, err := v.fetchLatestRelease(ctx)
	if err != nil {
		// Network error - graceful offline fallback
		details.CheckError = err.Error()
		return CheckResult{
			Name:    v.Name(),
			Status:  CheckPass,
			Message: fmt.Sprintf("Current version: %s (unable to check for updates: offline)", v.currentVersion),
			Details: details,
		}
	}

	details.Latest = latest
	details.UpdateURL = updateURL

	// Compare versions
	if v.isNewer(latest, v.currentVersion) {
		details.UpToDate = false
		return CheckResult{
			Name:           v.Name(),
			Status:         CheckWarning,
			Message:        fmt.Sprintf("Update available: %s → %s", v.currentVersion, latest),
			Recommendation: fmt.Sprintf("Update to latest version: %s", updateURL),
			Details:        details,
		}
	}

	return CheckResult{
		Name:    v.Name(),
		Status:  CheckPass,
		Message: fmt.Sprintf("%s (up to date)", v.currentVersion),
		Details: details,
	}
}

// fetchLatestRelease queries GitHub API for the latest release
func (v *VersionChecker) fetchLatestRelease(ctx context.Context) (string, string, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", v.apiBaseURL, v.githubRepo)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", err
	}

	return release.TagName, release.HTMLURL, nil
}

// isNewer checks if version a is newer than version b
// Simple comparison: v1.2.4 > v1.2.3
func (v *VersionChecker) isNewer(a, b string) bool {
	// Strip 'v' prefix if present
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	// Simple string comparison for now
	// In production, use semver library for proper comparison
	return a > b
}
