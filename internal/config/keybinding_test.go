package config

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// T025: Unit tests for keybinding string parsing
// Following TDD - these tests will fail until ParseKeybinding() is implemented

func TestParseKeybinding(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantKey     tcell.Key
		wantRune    rune
		wantMod     tcell.ModMask
		wantErr     bool
		errContains string
	}{
		// Simple keys
		{
			name:     "simple letter 'a'",
			input:    "a",
			wantKey:  tcell.KeyRune,
			wantRune: 'a',
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "simple letter 'q'",
			input:    "q",
			wantKey:  tcell.KeyRune,
			wantRune: 'q',
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "simple letter with uppercase input",
			input:    "Q",
			wantKey:  tcell.KeyRune,
			wantRune: 'q', // Normalized to lowercase
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "question mark",
			input:    "?",
			wantKey:  tcell.KeyRune,
			wantRune: '?',
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "forward slash",
			input:    "/",
			wantKey:  tcell.KeyRune,
			wantRune: '/',
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},

		// Special keys
		{
			name:     "enter key",
			input:    "enter",
			wantKey:  tcell.KeyEnter,
			wantRune: 0,
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "escape key",
			input:    "esc",
			wantKey:  tcell.KeyEscape,
			wantRune: 0,
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "escape key (full name)",
			input:    "escape",
			wantKey:  tcell.KeyEscape,
			wantRune: 0,
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "tab key",
			input:    "tab",
			wantKey:  tcell.KeyTab,
			wantRune: 0,
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "space key",
			input:    "space",
			wantKey:  tcell.KeyRune,
			wantRune: ' ',
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},

		// Function keys
		{
			name:     "function key F1",
			input:    "f1",
			wantKey:  tcell.KeyF1,
			wantRune: 0,
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},
		{
			name:     "function key F12",
			input:    "f12",
			wantKey:  tcell.KeyF12,
			wantRune: 0,
			wantMod:  tcell.ModNone,
			wantErr:  false,
		},

		// Ctrl combinations
		{
			name:     "ctrl+c",
			input:    "ctrl+c",
			wantKey:  tcell.KeyCtrlC,
			wantRune: 0,
			wantMod:  tcell.ModCtrl,
			wantErr:  false,
		},
		{
			name:     "ctrl+q",
			input:    "ctrl+q",
			wantKey:  tcell.KeyCtrlQ,
			wantRune: 0,
			wantMod:  tcell.ModCtrl,
			wantErr:  false,
		},

		// Alt combinations
		{
			name:     "alt+f",
			input:    "alt+f",
			wantKey:  tcell.KeyRune,
			wantRune: 'f',
			wantMod:  tcell.ModAlt,
			wantErr:  false,
		},
		{
			name:     "alt+enter",
			input:    "alt+enter",
			wantKey:  tcell.KeyEnter,
			wantRune: 0,
			wantMod:  tcell.ModAlt,
			wantErr:  false,
		},

		// Shift combinations
		{
			name:     "shift+tab",
			input:    "shift+tab",
			wantKey:  tcell.KeyBacktab,
			wantRune: 0,
			wantMod:  tcell.ModShift,
			wantErr:  false,
		},

		// Multiple modifiers
		{
			name:     "ctrl+alt+s",
			input:    "ctrl+alt+s",
			wantKey:  tcell.KeyRune,
			wantRune: 's',
			wantMod:  tcell.ModCtrl | tcell.ModAlt,
			wantErr:  false,
		},
		{
			name:     "ctrl+shift+f",
			input:    "ctrl+shift+f",
			wantKey:  tcell.KeyRune,
			wantRune: 'f',
			wantMod:  tcell.ModCtrl | tcell.ModShift,
			wantErr:  false,
		},

		// Error cases
		{
			name:        "empty string",
			input:       "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "only whitespace",
			input:       "   ",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "unknown modifier",
			input:       "super+a",
			wantErr:     true,
			errContains: "unknown modifier",
		},
		{
			name:        "modifier without key",
			input:       "ctrl+",
			wantErr:     true,
			errContains: "empty key",
		},
		{
			name:        "unknown key name",
			input:       "unknownkey",
			wantErr:     true,
			errContains: "unknown key",
		},
		{
			name:        "invalid multi-char key",
			input:       "abc",
			wantErr:     true,
			errContains: "unknown key",
		},
		{
			name:        "duplicate modifier",
			input:       "ctrl+ctrl+a",
			wantErr:     true,
			errContains: "duplicate modifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, r, mod, err := ParseKeybinding(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseKeybinding(%q) expected error containing %q, got nil",
						tt.input, tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseKeybinding(%q) error = %q, want error containing %q",
						tt.input, err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseKeybinding(%q) unexpected error: %v", tt.input, err)
				return
			}

			if key != tt.wantKey {
				t.Errorf("ParseKeybinding(%q) key = %v, want %v", tt.input, key, tt.wantKey)
			}
			if r != tt.wantRune {
				t.Errorf("ParseKeybinding(%q) rune = %q, want %q", tt.input, r, tt.wantRune)
			}
			if mod != tt.wantMod {
				t.Errorf("ParseKeybinding(%q) modifiers = %v, want %v", tt.input, mod, tt.wantMod)
			}
		})
	}
}

// Helper function for error string checking
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// T026: Unit tests for keybinding conflict detection
func TestDetectKeybindingConflicts(t *testing.T) {
	tests := []struct {
		name          string
		bindings      map[string]string
		wantConflicts bool
		conflictPairs [][2]string // Pairs of actions that conflict
	}{
		{
			name: "no conflicts - all unique keys",
			bindings: map[string]string{
				"quit":              "q",
				"add_credential":    "a",
				"edit_credential":   "e",
				"delete_credential": "d",
			},
			wantConflicts: false,
		},
		{
			name: "no conflicts - modifiers make keys unique",
			bindings: map[string]string{
				"quit":       "q",
				"quick_save": "ctrl+q",
				"search":     "alt+q",
			},
			wantConflicts: false,
		},
		{
			name: "conflict - same simple key",
			bindings: map[string]string{
				"add_credential":    "d",
				"delete_credential": "d",
			},
			wantConflicts: true,
			conflictPairs: [][2]string{
				{"add_credential", "delete_credential"},
			},
		},
		{
			name: "conflict - same modifier+key",
			bindings: map[string]string{
				"save":   "ctrl+s",
				"search": "ctrl+s",
			},
			wantConflicts: true,
			conflictPairs: [][2]string{
				{"save", "search"},
			},
		},
		{
			name: "multiple conflicts",
			bindings: map[string]string{
				"action1": "a",
				"action2": "a",
				"action3": "b",
				"action4": "b",
			},
			wantConflicts: true,
			conflictPairs: [][2]string{
				{"action1", "action2"},
				{"action3", "action4"},
			},
		},
		{
			name: "three-way conflict",
			bindings: map[string]string{
				"action1": "x",
				"action2": "x",
				"action3": "x",
			},
			wantConflicts: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts := DetectKeybindingConflicts(tt.bindings)

			if tt.wantConflicts && len(conflicts) == 0 {
				t.Errorf("DetectKeybindingConflicts() expected conflicts, got none")
			}
			if !tt.wantConflicts && len(conflicts) > 0 {
				t.Errorf("DetectKeybindingConflicts() expected no conflicts, got %d", len(conflicts))
				for _, c := range conflicts {
					t.Logf("  Conflict: %s", c)
				}
			}

			// Check that specific conflict pairs are detected
			if tt.conflictPairs != nil {
				conflictMap := make(map[string]bool)
				for _, c := range conflicts {
					conflictMap[c] = true
				}

				for _, pair := range tt.conflictPairs {
					// Check both orderings since we don't know which order they appear in
					found := false
					for _, c := range conflicts {
						if contains(c, pair[0]) && contains(c, pair[1]) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected conflict between %q and %q not found", pair[0], pair[1])
					}
				}
			}
		})
	}
}

// T027: Unit tests for unknown action validation
func TestValidateActions(t *testing.T) {
	tests := []struct {
		name           string
		bindings       map[string]string
		wantValid      bool
		unknownActions []string
	}{
		{
			name: "all valid actions",
			bindings: map[string]string{
				"quit":              "q",
				"add_credential":    "a",
				"edit_credential":   "e",
				"delete_credential": "d",
				"toggle_detail":     "tab",
				"toggle_sidebar":    "s",
				"help":              "?",
				"search":            "/",
				"confirm":           "enter",
				"cancel":            "esc",
			},
			wantValid: true,
		},
		{
			name: "subset of valid actions",
			bindings: map[string]string{
				"quit":   "q",
				"help":   "?",
				"search": "/",
			},
			wantValid: true,
		},
		{
			name: "single unknown action",
			bindings: map[string]string{
				"quit":           "q",
				"invalid_action": "x",
			},
			wantValid:      false,
			unknownActions: []string{"invalid_action"},
		},
		{
			name: "multiple unknown actions",
			bindings: map[string]string{
				"quit":     "q",
				"unknown1": "a",
				"unknown2": "b",
			},
			wantValid:      false,
			unknownActions: []string{"unknown1", "unknown2"},
		},
		{
			name: "all unknown actions",
			bindings: map[string]string{
				"invalid1": "a",
				"invalid2": "b",
				"invalid3": "c",
			},
			wantValid:      false,
			unknownActions: []string{"invalid1", "invalid2", "invalid3"},
		},
		{
			name:      "empty bindings",
			bindings:  map[string]string{},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateActions(tt.bindings)

			if tt.wantValid && len(errors) > 0 {
				t.Errorf("ValidateActions() expected valid, got errors: %v", errors)
			}
			if !tt.wantValid && len(errors) == 0 {
				t.Errorf("ValidateActions() expected errors, got none")
			}

			// Check that specific unknown actions are reported
			if tt.unknownActions != nil {
				for _, unknown := range tt.unknownActions {
					found := false
					for _, err := range errors {
						if contains(err, unknown) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for unknown action %q not found in errors: %v",
							unknown, errors)
					}
				}
			}
		})
	}
}

// Test that GetValidActions returns the expected list
func TestGetValidActions(t *testing.T) {
	validActions := GetValidActions()

	expectedActions := []string{
		"quit",
		"add_credential",
		"edit_credential",
		"delete_credential",
		"toggle_detail",
		"toggle_sidebar",
		"help",
		"search",
		"confirm",
		"cancel",
	}

	if len(validActions) != len(expectedActions) {
		t.Errorf("GetValidActions() returned %d actions, expected %d",
			len(validActions), len(expectedActions))
	}

	validSet := make(map[string]bool)
	for _, action := range validActions {
		validSet[action] = true
	}

	for _, expected := range expectedActions {
		if !validSet[expected] {
			t.Errorf("GetValidActions() missing expected action: %s", expected)
		}
	}
}

func TestKeybindingStruct(t *testing.T) {
	// Basic struct instantiation test
	kb := Keybinding{
		Action:    "quit",
		KeyString: "q",
	}
	if kb.Action != "quit" {
		t.Errorf("expected Action='quit', got '%s'", kb.Action)
	}
}
