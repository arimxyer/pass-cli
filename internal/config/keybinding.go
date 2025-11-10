package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Keybinding represents a parsed keybinding for runtime matching
type Keybinding struct {
	Action    string // Action name (e.g., "add_credential")
	KeyString string // Original string from config (e.g., "ctrl+a")

	// Parsed tcell representation
	Key       tcell.Key
	Rune      rune
	Modifiers tcell.ModMask
}

// validActions is the list of recognized action names
var validActions = []string{
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

// GetValidActions returns the list of valid action names
func GetValidActions() []string {
	result := make([]string, len(validActions))
	copy(result, validActions)
	return result
}

// ParseKeybinding parses a keybinding string (e.g., "ctrl+a", "enter", "f1")
// into tcell key components. Returns (key, rune, modifiers, error).
//
// Format: [modifier+]*key
// Valid modifiers: ctrl, alt, shift
// Valid keys: single letter, enter, esc/escape, tab, space, f1-f12
func ParseKeybinding(keyStr string) (tcell.Key, rune, tcell.ModMask, error) {
	// Normalize: trim and lowercase
	keyStr = strings.ToLower(strings.TrimSpace(keyStr))

	if keyStr == "" {
		return 0, 0, 0, fmt.Errorf("empty keybinding string")
	}

	// Split on '+' to separate modifiers from key
	parts := strings.Split(keyStr, "+")
	if len(parts) == 0 {
		return 0, 0, 0, fmt.Errorf("empty keybinding string")
	}

	var mods tcell.ModMask
	seenModifiers := make(map[string]bool)

	// Parse modifiers (all parts except the last)
	for i := 0; i < len(parts)-1; i++ {
		modifier := parts[i]

		// Check for duplicate modifiers
		if seenModifiers[modifier] {
			return 0, 0, 0, fmt.Errorf("duplicate modifier: %s", modifier)
		}
		seenModifiers[modifier] = true

		switch modifier {
		case "ctrl":
			mods |= tcell.ModCtrl
		case "alt":
			mods |= tcell.ModAlt
		case "shift":
			mods |= tcell.ModShift
		default:
			return 0, 0, 0, fmt.Errorf("unknown modifier: %s", modifier)
		}
	}

	// Parse the final key part
	keyPart := parts[len(parts)-1]
	if keyPart == "" {
		return 0, 0, 0, fmt.Errorf("empty key after modifiers")
	}

	// Handle special keys
	switch keyPart {
	case "enter":
		return tcell.KeyEnter, 0, mods, nil
	case "esc", "escape":
		return tcell.KeyEscape, 0, mods, nil
	case "tab":
		if mods&tcell.ModShift != 0 {
			return tcell.KeyBacktab, 0, mods, nil
		}
		return tcell.KeyTab, 0, mods, nil
	case "space":
		return tcell.KeyRune, ' ', mods, nil
	case "backspace":
		return tcell.KeyBackspace, 0, mods, nil
	case "delete", "del":
		return tcell.KeyDelete, 0, mods, nil
	case "insert", "ins":
		return tcell.KeyInsert, 0, mods, nil
	case "home":
		return tcell.KeyHome, 0, mods, nil
	case "end":
		return tcell.KeyEnd, 0, mods, nil
	case "pageup", "pgup":
		return tcell.KeyPgUp, 0, mods, nil
	case "pagedown", "pgdn":
		return tcell.KeyPgDn, 0, mods, nil
	case "up":
		return tcell.KeyUp, 0, mods, nil
	case "down":
		return tcell.KeyDown, 0, mods, nil
	case "left":
		return tcell.KeyLeft, 0, mods, nil
	case "right":
		return tcell.KeyRight, 0, mods, nil
	}

	// Handle function keys (f1-f12)
	if len(keyPart) >= 2 && keyPart[0] == 'f' {
		if num, err := strconv.Atoi(keyPart[1:]); err == nil {
			if num >= 1 && num <= 12 {
				// tcell.KeyF1 through tcell.KeyF12 are sequential
				// #nosec G115 -- num is bounded [1,12] by validation, safe conversion to tcell.Key
				return tcell.KeyF1 + tcell.Key(num-1), 0, mods, nil
			}
		}
	}

	// Handle Ctrl+letter combinations (special tcell keys)
	// ONLY when Ctrl is the ONLY modifier (not combined with Alt or Shift)
	if mods == tcell.ModCtrl && len(keyPart) == 1 && keyPart[0] >= 'a' && keyPart[0] <= 'z' {
		// Map to tcell.KeyCtrlA through tcell.KeyCtrlZ
		offset := keyPart[0] - 'a'
		return tcell.KeyCtrlA + tcell.Key(offset), 0, mods, nil
	}

	// Handle single character keys (including those with multiple modifiers)
	if len(keyPart) == 1 {
		return tcell.KeyRune, rune(keyPart[0]), mods, nil
	}

	// Unknown key
	return 0, 0, 0, fmt.Errorf("unknown key: %s", keyPart)
}

// ValidateActions checks if all action names in the bindings map are valid.
// Returns a slice of error messages for invalid actions.
func ValidateActions(bindings map[string]string) []string {
	var errors []string

	validSet := make(map[string]bool)
	for _, action := range validActions {
		validSet[action] = true
	}

	for action := range bindings {
		if !validSet[action] {
			errors = append(errors, fmt.Sprintf("keybindings.%s: unknown action name (valid actions: %s)",
				action, strings.Join(validActions, ", ")))
		}
	}

	return errors
}

// DetectKeybindingConflicts checks for duplicate key assignments.
// Returns a slice of error messages describing conflicts.
func DetectKeybindingConflicts(bindings map[string]string) []string {
	var conflicts []string

	// Map from normalized key string to list of actions using that key
	keyToActions := make(map[string][]string)

	for action, keyStr := range bindings {
		// Normalize the key string for comparison
		normalized := strings.ToLower(strings.TrimSpace(keyStr))
		keyToActions[normalized] = append(keyToActions[normalized], action)
	}

	// Check for conflicts (multiple actions with same key)
	for keyStr, actions := range keyToActions {
		if len(actions) > 1 {
			// Report conflict
			actionList := strings.Join(actions, ", ")
			conflicts = append(conflicts, fmt.Sprintf("keybindings conflict: multiple actions (%s) mapped to same key '%s'",
				actionList, keyStr))
		}
	}

	return conflicts
}

// GetKeybindingForAction returns the keybinding for a specific action from the config.
// Returns the key string (e.g., "ctrl+q", "n") or empty string if not found.
func (c *Config) GetKeybindingForAction(action string) string {
	if keyStr, ok := c.Keybindings[action]; ok {
		return keyStr
	}
	return ""
}

// MatchesKeybinding checks if a tcell.EventKey matches the configured binding for an action.
// Returns true if the event matches the configured keybinding for the action.
func (c *Config) MatchesKeybinding(event *tcell.EventKey, action string) bool {
	// Find parsed keybinding for this action
	for _, kb := range c.ParsedKeybindings {
		if kb.Action != action {
			continue
		}

		// Check if event matches the parsed keybinding
		if event.Key() == tcell.KeyRune {
			// For rune keys, match rune and check meaningful modifiers (Ctrl, Alt)
			// Ignore Shift modifier since it's implicit in the rune itself
			// (e.g., '?' is inherently Shift+/, 'A' is inherently Shift+a)
			if event.Rune() != kb.Rune {
				continue
			}

			// Compare modifiers, but mask out Shift since it's implicit
			eventMods := event.Modifiers() &^ tcell.ModShift
			kbMods := kb.Modifiers &^ tcell.ModShift

			return eventMods == kbMods
		} else {
			// For special keys, match key and modifiers exactly
			return event.Key() == kb.Key && event.Modifiers() == kb.Modifiers
		}
	}

	return false
}

// GetDisplayString returns a human-readable display string for a keybinding.
// Formats the key string for UI display (e.g., "Ctrl+Q", "N", "Enter").
func GetDisplayString(keyStr string) string {
	if keyStr == "" {
		return ""
	}

	// Normalize the key string
	normalized := strings.ToLower(strings.TrimSpace(keyStr))
	parts := strings.Split(normalized, "+")

	// Capitalize modifiers and key for display
	displayParts := make([]string, len(parts))
	for i, part := range parts {
		switch part {
		case "ctrl":
			displayParts[i] = "Ctrl"
		case "alt":
			displayParts[i] = "Alt"
		case "shift":
			displayParts[i] = "Shift"
		case "enter":
			displayParts[i] = "Enter"
		case "esc", "escape":
			displayParts[i] = "Esc"
		case "tab":
			displayParts[i] = "Tab"
		case "space":
			displayParts[i] = "Space"
		case "backspace":
			displayParts[i] = "Backspace"
		case "delete", "del":
			displayParts[i] = "Del"
		default:
			// Keep lowercase for single-letter keys to match terminal behavior
			// (typing 'a' not 'A', 'q' not 'Q', etc.)
			// Capitalize function keys (F1, F2, etc.)
			if len(part) > 0 {
				if len(part) >= 2 && part[0] == 'f' {
					// Function key - capitalize (f1 -> F1)
					displayParts[i] = strings.ToUpper(part[0:1]) + part[1:]
				} else {
					// Single letter or other - keep as-is (lowercase)
					displayParts[i] = part
				}
			}
		}
	}

	return strings.Join(displayParts, "+")
}
