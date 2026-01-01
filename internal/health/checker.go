package health

import (
	"context"
	"time"

	"pass-cli/internal/config"
)

// Exit codes for doctor command
const (
	ExitHealthy       = 0 // All checks passed
	ExitWarnings      = 1 // Non-critical issues detected
	ExitErrors        = 2 // Critical issues detected
	ExitSecurityError = 3 // Security-related issues (reserved for future use)
)

// CheckStatus represents health check outcome
type CheckStatus string

const (
	CheckPass    CheckStatus = "pass"
	CheckWarning CheckStatus = "warning"
	CheckError   CheckStatus = "error"
)

// HealthChecker defines interface for individual health checks
type HealthChecker interface {
	Name() string                        // Check name (e.g., "version", "vault")
	Run(ctx context.Context) CheckResult // Execute check and return result
}

// CheckResult represents a single health check outcome
type CheckResult struct {
	Name           string      `json:"name"`           // e.g., "version", "vault", "config"
	Status         CheckStatus `json:"status"`         // pass, warning, error
	Message        string      `json:"message"`        // Human-readable result
	Recommendation string      `json:"recommendation"` // Actionable fix (empty if passed)
	Details        interface{} `json:"details"`        // Check-specific structured data
}

// HealthSummary provides high-level statistics
type HealthSummary struct {
	Passed   int `json:"passed"`    // Number of checks that passed
	Warnings int `json:"warnings"`  // Number of warnings
	Errors   int `json:"errors"`    // Number of errors
	ExitCode int `json:"exit_code"` // 0=healthy, 1=warnings, 2=errors, 3=security
}

// HealthReport aggregates all health check results
type HealthReport struct {
	Summary   HealthSummary `json:"summary"`   // High-level statistics
	Checks    []CheckResult `json:"checks"`    // Individual check results
	Timestamp time.Time     `json:"timestamp"` // When report was generated
}

// CheckOptions contains configuration for health check execution
type CheckOptions struct {
	CurrentVersion  string            // Current binary version
	GitHubRepo      string            // GitHub repository (format: owner/repo)
	VaultPath       string            // Path to vault file
	VaultPathSource string            // Source of vault path ("config" or "default")
	VaultDir        string            // Directory containing vault
	ConfigPath      string            // Path to config file
	SyncConfig      config.SyncConfig // ARI-53: Sync configuration for health check
}

// DetermineExitCode maps health summary to exit code
func (s HealthSummary) DetermineExitCode() int {
	if s.Errors > 0 {
		return ExitErrors
	}
	if s.Warnings > 0 {
		return ExitWarnings
	}
	return ExitHealthy
}

// RunChecks executes all health checks and returns aggregated report
func RunChecks(ctx context.Context, opts CheckOptions) HealthReport {
	// Register all health checkers
	checkers := []HealthChecker{
		NewVersionChecker(opts.CurrentVersion, opts.GitHubRepo),
		NewVaultChecker(opts.VaultPath),
		NewConfigChecker(opts.ConfigPath),
		NewKeychainChecker(opts.VaultPath),
		NewBackupChecker(opts.VaultDir),
		NewSyncChecker(opts.SyncConfig), // ARI-53: Cloud sync health check
	}

	// Execute all checks
	results := make([]CheckResult, 0, len(checkers))
	for _, checker := range checkers {
		result := checker.Run(ctx)
		results = append(results, result)
	}

	// Build summary
	summary := buildSummary(results)

	return HealthReport{
		Summary:   summary,
		Checks:    results,
		Timestamp: time.Now(),
	}
}

// buildSummary aggregates check results into summary statistics
func buildSummary(results []CheckResult) HealthSummary {
	summary := HealthSummary{}

	for _, result := range results {
		switch result.Status {
		case CheckPass:
			summary.Passed++
		case CheckWarning:
			summary.Warnings++
		case CheckError:
			summary.Errors++
		}
	}

	summary.ExitCode = summary.DetermineExitCode()
	return summary
}
