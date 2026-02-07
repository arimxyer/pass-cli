package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arimxyer/pass-cli/internal/config"
)

// T017: Integration test for loading terminal config from YAML
func TestLoadTerminalConfigFromYAML(t *testing.T) {
	// Create temporary directory for test fixtures
	tmpDir := t.TempDir()

	tests := []struct {
		name             string
		yamlContent      string
		expectValid      bool
		expectWidth      int
		expectHeight     int
		expectEnabled    bool
		expectErrorCount int
	}{
		{
			name: "valid terminal config",
			yamlContent: `terminal:
  warning_enabled: true
  min_width: 80
  min_height: 40
`,
			expectValid:   true,
			expectWidth:   80,
			expectHeight:  40,
			expectEnabled: true,
		},
		{
			name: "disabled warning",
			yamlContent: `terminal:
  warning_enabled: false
  min_width: 100
  min_height: 50
`,
			expectValid:   true,
			expectWidth:   100,
			expectHeight:  50,
			expectEnabled: false,
		},
		{
			name: "invalid negative width",
			yamlContent: `terminal:
  warning_enabled: true
  min_width: -10
  min_height: 30
`,
			expectValid:      false,
			expectErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test config file
			configPath := filepath.Join(tmpDir, "config_"+tt.name+".yml")
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test config file: %v", err)
			}

			// Load config from test path
			cfg, result := config.LoadFromPath(configPath)
			if cfg == nil {
				t.Fatal("LoadFromPath() returned nil config")
			}

			// Check validation result
			if result.Valid != tt.expectValid {
				t.Errorf("expected Valid=%v, got %v (errors: %v)", tt.expectValid, result.Valid, result.Errors)
			}

			if tt.expectErrorCount > 0 && len(result.Errors) != tt.expectErrorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.expectErrorCount, len(result.Errors), result.Errors)
			}

			// For valid configs, check values
			if tt.expectValid {
				if cfg.Terminal.MinWidth != tt.expectWidth {
					t.Errorf("expected MinWidth=%d, got %d", tt.expectWidth, cfg.Terminal.MinWidth)
				}
				if cfg.Terminal.MinHeight != tt.expectHeight {
					t.Errorf("expected MinHeight=%d, got %d", tt.expectHeight, cfg.Terminal.MinHeight)
				}
				if cfg.Terminal.WarningEnabled != tt.expectEnabled {
					t.Errorf("expected WarningEnabled=%v, got %v", tt.expectEnabled, cfg.Terminal.WarningEnabled)
				}
			}
		})
	}
}
