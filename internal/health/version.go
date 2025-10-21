package health

import "context"

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
	// Placeholder - will be implemented in Phase 3
	return CheckResult{
		Name:    v.Name(),
		Status:  CheckPass,
		Message: "Version check not yet implemented",
		Details: VersionCheckDetails{
			Current:  v.currentVersion,
			UpToDate: true,
		},
	}
}
