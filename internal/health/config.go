package health

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

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
	details := ConfigCheckDetails{
		Path:        c.configPath,
		Errors:      []ConfigError{},
		UnknownKeys: []string{},
	}

	// Check if config file exists
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		// Config file doesn't exist - this is OK (use defaults)
		details.Exists = false
		details.Valid = true
		return CheckResult{
			Name:    c.Name(),
			Status:  CheckPass,
			Message: "No config file (using defaults)",
			Details: details,
		}
	}

	details.Exists = true

	// Parse config file using Viper
	v := viper.New()
	v.SetConfigFile(c.configPath)

	if err := v.ReadInConfig(); err != nil {
		details.Valid = false
		return CheckResult{
			Name:           c.Name(),
			Status:         CheckError,
			Message:        fmt.Sprintf("Failed to parse config: %v", err),
			Recommendation: "Check YAML syntax in config file",
			Details:        details,
		}
	}

	details.Valid = true

	// Known valid config keys
	knownKeys := map[string]bool{
		"vault_path":         true,
		"clipboard_timeout":  true,
		"audit_enabled":      true,
		"keychain_enabled":   true,
		"default_vault_path": true,
	}

	// Check for unknown keys
	allKeys := v.AllKeys()
	for _, key := range allKeys {
		if !knownKeys[key] {
			details.UnknownKeys = append(details.UnknownKeys, key)
		}
	}

	// Validate clipboard_timeout range (5-300 seconds)
	if v.IsSet("clipboard_timeout") {
		timeout := v.GetInt("clipboard_timeout")
		if timeout < 5 || timeout > 300 {
			details.Errors = append(details.Errors, ConfigError{
				Key:           "clipboard_timeout",
				Problem:       "value out of range",
				CurrentValue:  strconv.Itoa(timeout),
				ExpectedValue: "5-300",
			})
		}
	}

	// Determine status
	if len(details.Errors) > 0 || len(details.UnknownKeys) > 0 {
		message := "Config validation issues"
		if len(details.Errors) > 0 {
			message = fmt.Sprintf("%d validation error(s)", len(details.Errors))
		}
		if len(details.UnknownKeys) > 0 {
			message = fmt.Sprintf("%s, %d unknown key(s)", message, len(details.UnknownKeys))
		}

		recommendation := "Review config file"
		if len(details.UnknownKeys) > 0 {
			recommendation = fmt.Sprintf("Check for typos in: %v", details.UnknownKeys)
		}

		return CheckResult{
			Name:           c.Name(),
			Status:         CheckWarning,
			Message:        message,
			Recommendation: recommendation,
			Details:        details,
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  CheckPass,
		Message: "Config valid",
		Details: details,
	}
}
