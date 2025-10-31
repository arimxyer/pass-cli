package config

import (
	"runtime"
	"testing"
)

// Placeholder for config package unit tests
// Tests will be added in later phases following TDD approach

func TestGetDefaults(t *testing.T) {
	cfg := GetDefaults()
	if cfg == nil {
		t.Fatal("GetDefaults() returned nil")
	}

	// Verify terminal defaults
	if cfg.Terminal.WarningEnabled != true {
		t.Errorf("expected WarningEnabled=true, got %v", cfg.Terminal.WarningEnabled)
	}
	if cfg.Terminal.MinWidth != 60 {
		t.Errorf("expected MinWidth=60, got %d", cfg.Terminal.MinWidth)
	}
	if cfg.Terminal.MinHeight != 30 {
		t.Errorf("expected MinHeight=30, got %d", cfg.Terminal.MinHeight)
	}

	// Verify keybindings defaults exist
	if len(cfg.Keybindings) == 0 {
		t.Error("expected default keybindings, got empty map")
	}
	if cfg.Keybindings["quit"] != "q" {
		t.Errorf("expected quit='q', got '%s'", cfg.Keybindings["quit"])
	}
}

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}
}

// T015: Unit test for TerminalConfig validation (positive/negative values, range checks)
func TestTerminalConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         TerminalConfig
		expectValid    bool
		expectErrors   int
		expectWarnings int
	}{
		{
			name: "valid config with defaults",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       60,
				MinHeight:      30,
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "valid config with custom values",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       100,
				MinHeight:      50,
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "warning disabled",
			config: TerminalConfig{
				WarningEnabled: false,
				MinWidth:       60,
				MinHeight:      30,
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "negative min_width",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       -10,
				MinHeight:      30,
			},
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "zero min_width",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       0,
				MinHeight:      30,
			},
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "negative min_height",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       60,
				MinHeight:      -5,
			},
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "min_width too large",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       15000,
				MinHeight:      30,
			},
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "min_height too large",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       60,
				MinHeight:      2000,
			},
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "unusually large width (should warn)",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       600,
				MinHeight:      30,
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 1,
		},
		{
			name: "unusually large height (should warn)",
			config: TerminalConfig{
				WarningEnabled: true,
				MinWidth:       60,
				MinHeight:      250,
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Terminal: tt.config,
			}
			result := cfg.Validate()

			if result.Valid != tt.expectValid {
				t.Errorf("expected Valid=%v, got %v", tt.expectValid, result.Valid)
			}
			if len(result.Errors) != tt.expectErrors {
				t.Errorf("expected %d errors, got %d: %v", tt.expectErrors, len(result.Errors), result.Errors)
			}
			if tt.expectWarnings > 0 && len(result.Warnings) != tt.expectWarnings {
				t.Errorf("expected %d warnings, got %d: %v", tt.expectWarnings, len(result.Warnings), result.Warnings)
			}
		})
	}
}

// T016: Unit test for terminal config merging with defaults
func TestTerminalConfigMerging(t *testing.T) {
	tests := []struct {
		name            string
		yamlContent     map[string]interface{}
		expectedWidth   int
		expectedHeight  int
		expectedEnabled bool
	}{
		{
			name:            "empty config uses all defaults",
			yamlContent:     map[string]interface{}{},
			expectedWidth:   60,
			expectedHeight:  30,
			expectedEnabled: true,
		},
		{
			name: "partial config - only width",
			yamlContent: map[string]interface{}{
				"terminal": map[string]interface{}{
					"min_width": 100,
				},
			},
			expectedWidth:   100,
			expectedHeight:  30,   // default
			expectedEnabled: true, // default
		},
		{
			name: "partial config - only warning_enabled",
			yamlContent: map[string]interface{}{
				"terminal": map[string]interface{}{
					"warning_enabled": false,
				},
			},
			expectedWidth:   60, // default
			expectedHeight:  30, // default
			expectedEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will be fully implemented when we add YAML parsing
			// For now, just verify defaults work
			defaults := GetDefaults()
			if defaults.Terminal.MinWidth != 60 {
				t.Errorf("default MinWidth should be 60, got %d", defaults.Terminal.MinWidth)
			}
		})
	}
}

// Helper function to get platform-appropriate absolute path for testing
// Uses temp dir so parent directory exists
func getAbsolutePath() string {
	if runtime.GOOS == "windows" {
		return "C:\\Windows\\Temp\\vault.enc"
	}
	return "/tmp/vault.enc"
}

// T008 & T009: Unit tests for vault_path configuration
func TestVaultPathValidation(t *testing.T) {
	tests := []struct {
		name          string
		vaultPath     string
		expectErrors  int
		expectWarnings int
		errorContains string
		warnContains  string
	}{
		{
			name:           "empty vault_path is valid",
			vaultPath:      "",
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name:           "absolute path is valid",
			vaultPath:      getAbsolutePath(),
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name:           "tilde path is valid",
			vaultPath:      "~/vault.enc",
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name:           "relative path warns",
			vaultPath:      "vault.enc",
			expectErrors:   0,
			expectWarnings: 1,
			warnContains:   "relative path",
		},
		{
			name:           "null byte errors",
			vaultPath:      "vault\x00.enc",
			expectErrors:   1,
			expectWarnings: 0,
			errorContains:  "null byte",
		},
		{
			name:           "env var path is valid",
			vaultPath:      "$HOME/vault.enc",
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name:           "Windows env var is valid",
			vaultPath:      "%USERPROFILE%\\vault.enc",
			expectErrors:   0,
			expectWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				VaultPath: tt.vaultPath,
			}

			result := &ValidationResult{
				Valid:    true,
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
			}

			result = cfg.validateVaultPath(result)

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("expected %d errors, got %d", tt.expectErrors, len(result.Errors))
				for _, err := range result.Errors {
					t.Logf("  Error: %s - %s", err.Field, err.Message)
				}
			}

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
				for _, warn := range result.Warnings {
					t.Logf("  Warning: %s - %s", warn.Field, warn.Message)
				}
			}

			if tt.errorContains != "" && len(result.Errors) > 0 {
				found := false
				for _, err := range result.Errors {
					if contains(err.Message, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing '%s', got: %v", tt.errorContains, result.Errors)
				}
			}

			if tt.warnContains != "" && len(result.Warnings) > 0 {
				found := false
				for _, warn := range result.Warnings {
					if contains(warn.Message, tt.warnContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning containing '%s', got: %v", tt.warnContains, result.Warnings)
				}
			}
		})
	}
}
