package config_test

import (
	"testing"

	"pass-cli/internal/config"
)

// T028: Integration test for custom keybindings in TUI event handling
// Following TDD - these tests will fail until keybinding loading is implemented

func TestKeybindingIntegration_LoadAndValidate(t *testing.T) {
	tests := []struct {
		name       string
		configData map[string]string
		wantValid  bool
		wantErrors int
	}{
		{
			name: "valid custom keybindings",
			configData: map[string]string{
				"quit":              "q",
				"add_credential":    "n", // Remapped from 'a' to 'n'
				"edit_credential":   "e",
				"delete_credential": "d",
			},
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "custom keybindings with modifiers",
			configData: map[string]string{
				"quit":           "ctrl+q",
				"add_credential": "ctrl+n",
				"help":           "f1",
			},
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "conflicting keybindings should error",
			configData: map[string]string{
				"add_credential":    "d",
				"delete_credential": "d", // Conflict!
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "unknown action should error",
			configData: map[string]string{
				"quit":           "q",
				"invalid_action": "x",
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "invalid key format should error",
			configData: map[string]string{
				"quit": "unknownkey",
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "multiple errors detected",
			configData: map[string]string{
				"quit":              "unknownkey", // Invalid key
				"invalid_action":    "x",          // Unknown action
				"add_credential":    "d",
				"delete_credential": "d", // Conflict
			},
			wantValid:  false,
			wantErrors: 3, // All 3 errors should be detected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Config with custom keybindings
			cfg := &config.Config{
				Terminal: config.TerminalConfig{
					WarningEnabled: true,
					MinWidth:       60,
					MinHeight:      30,
				},
				Keybindings: tt.configData,
			}

			// Validate config (this should parse and validate keybindings)
			result := cfg.Validate()

			if tt.wantValid && !result.Valid {
				t.Errorf("Config.Validate() expected valid, got invalid with errors: %v", result.Errors)
			}
			if !tt.wantValid && result.Valid {
				t.Errorf("Config.Validate() expected invalid, got valid")
			}

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Config.Validate() expected %d errors, got %d: %v",
					tt.wantErrors, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestKeybindingIntegration_ParsedKeybindingsAccessible(t *testing.T) {
	cfg := &config.Config{
		Terminal: config.TerminalConfig{
			WarningEnabled: true,
			MinWidth:       60,
			MinHeight:      30,
		},
		Keybindings: map[string]string{
			"quit":           "q",
			"add_credential": "n",
			"help":           "ctrl+h",
		},
	}

	// Validate to trigger parsing
	result := cfg.Validate()
	if !result.Valid {
		t.Fatalf("Config.Validate() failed: %v", result.Errors)
	}

	// Get parsed keybindings (this function needs to exist)
	parsedBindings := cfg.GetParsedKeybindings()
	if parsedBindings == nil {
		t.Fatal("GetParsedKeybindings() returned nil")
	}

	// Check that all actions were parsed
	expectedActions := []string{"quit", "add_credential", "help"}
	for _, action := range expectedActions {
		binding, exists := parsedBindings[action]
		if !exists {
			t.Errorf("GetParsedKeybindings() missing action: %s", action)
			continue
		}

		if binding.Action != action {
			t.Errorf("Keybinding for %s has wrong action: got %s", action, binding.Action)
		}
		if binding.KeyString == "" {
			t.Errorf("Keybinding for %s has empty KeyString", action)
		}
	}
}

func TestKeybindingIntegration_DefaultsUsedWhenEmpty(t *testing.T) {
	cfg := &config.Config{
		Terminal: config.TerminalConfig{
			WarningEnabled: true,
			MinWidth:       60,
			MinHeight:      30,
		},
		Keybindings: map[string]string{}, // Empty - should use defaults
	}

	result := cfg.Validate()
	if !result.Valid {
		t.Fatalf("Config.Validate() with empty keybindings failed: %v", result.Errors)
	}

	// Should merge with defaults
	defaults := config.GetDefaults()
	parsedBindings := cfg.GetParsedKeybindings()

	if len(parsedBindings) == 0 {
		t.Error("GetParsedKeybindings() returned empty map, expected defaults to be used")
	}

	// Check that default actions exist
	for action := range defaults.Keybindings {
		if _, exists := parsedBindings[action]; !exists {
			t.Errorf("Default action %s not found in parsed keybindings", action)
		}
	}
}
