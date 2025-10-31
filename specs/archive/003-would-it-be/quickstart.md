# Quickstart: Password Visibility Toggle Implementation

**Feature**: Password Visibility Toggle
**Target Audience**: Developers implementing this feature
**Estimated Time**: 2-3 hours (including tests)

## Overview

Add password visibility toggle to TUI add/edit forms, allowing users to verify password entries before saving via Ctrl+H keyboard shortcut and visual state indicator.

## Prerequisites

- Go 1.25.1 installed
- Familiarity with tview v0.42.0 and tcell v2.9.0 APIs
- Codebase cloned and buildable (`go build` succeeds)
- Familiarity with existing forms.go structure (AddForm, EditForm)

## Key Files

```
cmd/tui/components/forms.go    # MODIFY: Add toggle logic to AddForm and EditForm
tests/unit/tui_forms_test.go       # CREATE: Unit tests for toggle functionality
tests/integration/tui_password_toggle_test.go  # CREATE: Integration tests
```

## Implementation Steps

### Step 1: Understand Existing Code (15 min)

Read these sections in `cmd/tui/components/forms.go`:

1. **AddForm struct** (lines 27-34): Note the `appState`, `onSubmit`, `onCancel` fields
2. **AddForm.buildFormFields()** (lines 69-127): See how password field created at line 91 with `AddPasswordField(..., '*', ...)`
3. **AddForm.setupKeyboardShortcuts()** (lines 230-249): Pattern for Ctrl+S shortcut
4. **EditForm struct** (lines 38-49): Similar structure to AddForm
5. **EditForm password field** (lines 329-339): Password field with lazy loading via SetFocusFunc

**Key Insight**: Password field is added as item index 2 in both forms. Access via `GetFormItem(2).(*tview.InputField)`.

### Step 2: Add State Field to Structs (5 min)

**File**: `cmd/tui/components/forms.go`

**AddForm**:
```go
// Line ~34 (after existing fields)
type AddForm struct {
	*tview.Form

	appState *models.AppState
	passwordVisible bool  // ADD THIS LINE - tracks current visibility state

	onSubmit func()
	onCancel func()
}
```

**EditForm**:
```go
// Line ~49 (after existing fields)
type EditForm struct {
	*tview.Form

	appState   *models.AppState
	credential *vault.CredentialMetadata
	passwordVisible bool  // ADD THIS LINE - tracks current visibility state

	originalPassword string
	passwordFetched  bool

	onSubmit func()
	onCancel func()
}
```

### Step 3: Add Toggle Method to AddForm (10 min)

**File**: `cmd/tui/components/forms.go`

**Location**: After `applyStyles()` method, before `SetOnSubmit()` (around line 284)

```go
// togglePasswordVisibility switches between masked and plaintext password display.
// Updates both the mask character and the field label to indicate current state.
func (af *AddForm) togglePasswordVisibility() {
	af.passwordVisible = !af.passwordVisible

	passwordField := af.GetFormItem(2).(*tview.InputField)

	if af.passwordVisible {
		passwordField.SetMaskCharacter(0) // 0 = plaintext (tview convention)
		passwordField.SetLabel("Password [VISIBLE]")
	} else {
		passwordField.SetMaskCharacter('*')
		passwordField.SetLabel("Password")
	}
}
```

### Step 4: Add Ctrl+H Shortcut to AddForm (5 min)

**File**: `cmd/tui/components/forms.go`

**Modify**: `AddForm.setupKeyboardShortcuts()` method (line ~231)

```go
func (af *AddForm) setupKeyboardShortcuts() {
	af.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			// Ctrl+S for quick-save
			af.onAddPressed()
			return nil

		case tcell.KeyCtrlH:  // ADD THIS CASE
			// Ctrl+H for password visibility toggle
			af.togglePasswordVisibility()
			return nil

		case tcell.KeyTab:
			// Let form handle Tab internally
			return event

		case tcell.KeyBacktab:
			// Let form handle Shift+Tab internally
			return event
		}
		return event
	})
}
```

### Step 5: Update Keyboard Hints for AddForm (5 min)

**File**: `cmd/tui/components/forms.go`

**Modify**: `AddForm.addKeyboardHints()` method (line ~217)

```go
func (af *AddForm) addKeyboardHints() {
	theme := styles.GetCurrentTheme()

	// UPDATE THIS LINE to include Ctrl+H hint
	hintsText := "  Tab: Next field  •  Shift+Tab: Previous  •  Ctrl+S: Add  •  Ctrl+H: Toggle password  •  Esc: Cancel"

	hints := tview.NewTextView()
	hints.SetText(hintsText)
	hints.SetTextAlign(tview.AlignCenter)
	hints.SetTextColor(theme.TextSecondary)
	hints.SetBackgroundColor(theme.Background)

	af.AddFormItem(hints)
}
```

### Step 6: Repeat for EditForm (20 min)

**File**: `cmd/tui/components/forms.go`

Apply identical changes to EditForm:

1. **Add toggle method** (after `applyStyles()`, around line 602):
   ```go
   func (ef *EditForm) togglePasswordVisibility() {
       ef.passwordVisible = !ef.passwordVisible

       passwordField := ef.GetFormItem(2).(*tview.InputField)

       if ef.passwordVisible {
           passwordField.SetMaskCharacter(0)
           passwordField.SetLabel("Password [VISIBLE]")
       } else {
           passwordField.SetMaskCharacter('*')
           passwordField.SetLabel("Password")
       }
   }
   ```

2. **Modify setupKeyboardShortcuts()** (line ~549): Add `case tcell.KeyCtrlH` handler

3. **Modify addKeyboardHints()** (line ~535): Update to:
   ```go
   hintsText := "  Tab: Next field  •  Shift+Tab: Previous  •  Ctrl+S: Save  •  Ctrl+H: Toggle password  •  Esc: Cancel"
   ```

### Step 7: Manual Testing (15 min)

Build and run the TUI:

```bash
go build -o pass-cli.exe
./pass-cli.exe tui
```

**Test AddForm**:
1. Press `a` to open add form
2. Focus password field, type `test123`
3. Press Ctrl+H → Verify password shows as `test123` and label shows `[VISIBLE]`
4. Press Ctrl+H again → Verify password shows as `*******` and label shows `Password`
5. Press Esc to cancel → Reopen form → Verify password field defaults to masked

**Test EditForm**:
1. Select an existing credential (arrow keys + Enter)
2. Press `e` to open Edit form
3. Focus password field (Tab to it)
4. Press Ctrl+H → Verify password becomes visible with `[VISIBLE]` label
5. Press Ctrl+H → Verify password returns to masked
6. Press Esc to cancel → Reopen form → Verify defaults to masked

### Step 8: Write Unit Tests (30 min)

**Create File**: `tests/unit/tui_forms_test.go`

```go
package unit

import (
	"testing"

	"pass-cli/cmd/tui/components"
	"pass-cli/cmd/tui/models"
)

// TestAddFormPasswordToggle verifies Ctrl+H toggles mask character
func TestAddFormPasswordToggle(t *testing.T) {
	// Setup
	appState := &models.AppState{} // Mock AppState
	form := components.NewAddForm(appState)

	passwordField := form.GetFormItem(2).(*tview.InputField)

	// Initial state: masked
	if passwordField.GetMaskCharacter() != '*' {
		t.Errorf("Expected initial mask character '*', got %c", passwordField.GetMaskCharacter())
	}

	// Toggle to visible
	form.TogglePasswordVisibility() // Make method public for testing
	if passwordField.GetMaskCharacter() != 0 {
		t.Errorf("Expected visible mask character 0, got %c", passwordField.GetMaskCharacter())
	}

	// Toggle back to masked
	form.TogglePasswordVisibility()
	if passwordField.GetMaskCharacter() != '*' {
		t.Errorf("Expected masked character '*', got %c", passwordField.GetMaskCharacter())
	}
}

// Additional tests: TestEditFormPasswordToggle, TestLabelUpdatesOnToggle, etc.
```

**Note**: To test toggle behavior without violating Library-First Architecture (Principle II), use one of these approaches:
- **Integration tests**: Simulate keyboard events (`tcell.EventKey` with `KeyCtrlH`) and verify password field state via `GetFormItem(2).GetMaskCharacter()`
- **Public API testing**: Test observable behavior through public methods - verify mask character changes from `'*'` to `0` after simulated Ctrl+H event
- **NEVER** export private methods solely for testing - this pollutes the public API and violates architectural boundaries

### Step 9: Write Integration Tests (30 min)

**Create File**: `tests/integration/tui_password_toggle_test.go`

Focus on:
- Keyboard event simulation (Ctrl+H)
- Cursor position preservation
- Unicode character support
- Form reset behavior on close

### Step 10: Run Tests and Build (10 min)

```bash
# Run all tests
go test ./... -v

# Check coverage
go test ./cmd/tui/components -cover

# Build final binary
go build -o pass-cli.exe
```

## Verification Checklist

Before marking complete:

- [ ] Ctrl+H toggles password visibility in AddForm
- [ ] Ctrl+H toggles password visibility in EditForm
- [ ] Label shows `[VISIBLE]` when password visible
- [ ] Label shows `Password` when password masked
- [ ] Password defaults to masked on form open
- [ ] Keyboard hints mention Ctrl+H
- [ ] Cursor position preserved after toggle
- [ ] Unicode characters display correctly when visible
- [ ] Unit tests pass with >80% coverage
- [ ] Integration tests pass
- [ ] Manual testing completed in TUI

## Common Issues

**Issue**: Password field not at index 2
- **Solution**: Verify buildFormFields() order - Service(0), Username(1), Password(2)

**Issue**: Ctrl+H not firing
- **Solution**: Check SetInputCapture returns `nil` for KeyCtrlH case

**Issue**: Label not updating
- **Solution**: Ensure SetLabel() called after SetMaskCharacter()

**Issue**: Tests can't call toggle method
- **Solution**: Either export method (capitalize) or use integration tests with keyboard events

**Issue**: Unicode/emoji passwords display inconsistently across terminals
- **Solution**: Terminal rendering of wide characters (CJK, emoji) varies - tview masks each rune as single '*', but visible display depends on terminal's Unicode support. This is expected behavior and outside our control.

## Next Steps

After implementation:
1. Commit changes: `git commit -m "feat: Add password visibility toggle to TUI forms"`
2. Update tasks.md to mark tasks complete
3. Consider future enhancement: Mouse click activation (see research.md decision on deferred mouse support)

## References

- [tview Documentation](https://pkg.go.dev/github.com/rivo/tview)
- [tcell Key Constants](https://pkg.go.dev/github.com/gdamore/tcell/v2#pkg-constants)
- [Existing Implementation](../../../cmd/tui/components/forms.go)
- [Research Findings](./research.md)
