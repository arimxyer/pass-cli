package health

import "context"

// ConfigChecker checks config file validity
type ConfigChecker struct {
	configPath string
}

// NewConfigChecker creates a new config checker
func NewConfigChecker(configPath string) HealthChecker {
	return &ConfigChecker{
		configPath: configPath,
	}
}

// Name returns the check name
func (c *ConfigChecker) Name() string {
	return "config"
}

// Run executes the config check
func (c *ConfigChecker) Run(ctx context.Context) CheckResult {
	// Placeholder - will be implemented in Phase 3
	return CheckResult{
		Name:    c.Name(),
		Status:  CheckPass,
		Message: "Config check not yet implemented",
		Details: ConfigCheckDetails{
			Path:   c.configPath,
			Exists: true,
			Valid:  true,
		},
	}
}
