// Package components provides TUI form components for credential management.
// All forms support the complete credential model: service, username, password, category, URL, and notes.
package components

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"pass-cli/cmd/tui/models"
	"pass-cli/cmd/tui/styles"
	"pass-cli/internal/security"
	"pass-cli/internal/vault"
)

// normalizeCategory converts the "Uncategorized" UI label to empty string for storage.
// Prevents the UI label from leaking into credential data.
func normalizeCategory(c string) string {
	if c == "Uncategorized" {
		return ""
	}
	return c
}

// AddForm provides a modal form for adding new credentials.
// Embeds tview.Flex (which contains Form + hints footer) and manages validation and submission.
type AddForm struct {
	*tview.Flex
	form *tview.Form

	appState *models.AppState

	passwordVisible bool // Track password visibility state for toggle

	onSubmit        func()
	onCancel        func()
	onCancelConfirm func(message string, onYes func(), onNo func()) // Callback to show confirmation dialog
}

// EditForm provides a modal form for editing existing credentials.
// Embeds tview.Flex (which contains Form + hints footer) and pre-populates fields from credential.
type EditForm struct {
	*tview.Flex
	form *tview.Form

	appState   *models.AppState
	credential *vault.CredentialMetadata

	originalPassword string // Track original password to detect changes
	passwordFetched  bool   // Track if password has been fetched (lazy loading)
	passwordVisible  bool   // Track password visibility state for toggle
	clearTOTP        bool   // Track if user wants to clear TOTP

	onSubmit        func()
	onCancel        func()
	onCancelConfirm func(message string, onYes func(), onNo func()) // Callback to show confirmation dialog
}

// NewAddForm creates a new form for adding credentials.
// Creates input fields for Service, Username, Password, Category, URL, Notes.
func NewAddForm(appState *models.AppState) *AddForm {
	form := tview.NewForm()

	af := &AddForm{
		form:     form,
		appState: appState,
	}

	af.buildFormFields()
	af.applyStyles()
	af.setupKeyboardShortcuts()
	af.wrapInFrame()

	return af
}

// buildFormFields constructs all input fields for the add form.
func (af *AddForm) buildFormFields() {
	categories := af.getCategories()

	// Ensure "Uncategorized" is always present for autocomplete
	hasUncategorized := false
	for _, cat := range categories {
		if cat == "Uncategorized" {
			hasUncategorized = true
			break
		}
	}

	// If "Uncategorized" is not present, prepend it to the list
	if !hasUncategorized {
		categories = append([]string{"Uncategorized"}, categories...)
	}

	// Core credential fields
	// Use 0 width to make fields fill available space (prevents black rectangles)
	af.form.AddInputField("Service (UID)", "", 0, nil, nil)
	af.form.AddInputField("Username", "", 0, nil, nil)

	// T048, T049: Password field with real-time strength indicator
	passwordField := tview.NewInputField().
		SetLabel("Password").
		SetFieldWidth(0).
		SetMaskCharacter('*')

	// T049: Update password label with strength indicator inline
	passwordField.SetChangedFunc(func(text string) {
		af.updatePasswordLabel(passwordField, []byte(text))
	})

	af.form.AddFormItem(passwordField)

	// Optional metadata fields - default to "Uncategorized"

	// Original dropdown approach (commented out for autocomplete field)
	// af.form.AddDropDown("Category", categories, uncategorizedIndex, nil)

	// New autocomplete input field for Category
	categoryField := tview.NewInputField().
		SetLabel("Category").
		SetFieldWidth(0)

	categoryField.SetAutocompleteFunc(func(currentText string) []string {
		if currentText == "" {
			return categories
		}
		var matches []string
		lowerText := strings.ToLower(currentText)
		for _, cat := range categories {
			if strings.HasPrefix(strings.ToLower(cat), lowerText) {
				matches = append(matches, cat)
			}
		}
		return matches
	})

	af.form.AddFormItem(categoryField)
	af.form.AddInputField("URL", "", 0, nil, nil)
	af.form.AddTextArea("Notes", "", 0, 5, 0, nil)

	// TOTP field (optional) - accepts base32 secret or otpauth:// URI
	af.form.AddInputField("TOTP Secret/URI", "", 0, nil, nil)

	// Action buttons
	af.form.AddButton("Generate Password", af.onGeneratePassword)
	af.form.AddButton("Add", af.onAddPressed)
	af.form.AddButton("Cancel", af.onCancelPressed)
}

// onAddPressed handles the Add button submission.
// Validates inputs, calls AppState.AddCredential(), invokes onSubmit callback.
func (af *AddForm) onAddPressed() {
	// Validate inputs before submission
	if err := af.validate(); err != nil {
		// Validation failed - error will be shown via status bar or modal
		// Form stays open for correction
		return
	}

	// Extract field values (using form item index)
	service := af.form.GetFormItem(0).(*tview.InputField).GetText()
	username := af.form.GetFormItem(1).(*tview.InputField).GetText()
	password := af.form.GetFormItem(2).(*tview.InputField).GetText()

	// Extract category from input field
	category := af.form.GetFormItem(3).(*tview.InputField).GetText()
	category = normalizeCategory(category) // Convert "Uncategorized" to empty string

	url := af.form.GetFormItem(4).(*tview.InputField).GetText()
	notes := af.form.GetFormItem(5).(*tview.TextArea).GetText()
	totpInput := af.form.GetFormItem(6).(*tview.InputField).GetText()

	// Call AppState to add credential with all 6 fields
	err := af.appState.AddCredential(service, username, password, category, url, notes)
	if err != nil {
		// Error already handled by AppState onError callback
		// Form stays open for correction
		return
	}

	// If TOTP was provided, update credential with TOTP fields
	if totpInput != "" {
		totpConfig, err := vault.ParseTOTPURI(strings.TrimSpace(totpInput))
		if err != nil {
			// TOTP parsing failed - credential was added but without TOTP
			// Could show warning but don't fail the whole operation
			// Form will close, user can edit later to fix TOTP
		} else {
			// Update credential with TOTP fields
			opts := models.UpdateCredentialOpts{
				TOTPSecret:    &totpConfig.Secret,
				TOTPAlgorithm: &totpConfig.Algorithm,
				TOTPDigits:    &totpConfig.Digits,
				TOTPPeriod:    &totpConfig.Period,
			}
			if totpConfig.Issuer != "" {
				opts.TOTPIssuer = &totpConfig.Issuer
			}
			// Ignore error - credential was added, TOTP is optional
			_ = af.appState.UpdateCredential(service, opts)
		}
	}

	// Success - invoke callback to close modal
	if af.onSubmit != nil {
		af.onSubmit()
	}
}

// onCancelPressed handles the Cancel button.
// Shows confirmation dialog if there's unsaved data, otherwise closes immediately.
func (af *AddForm) onCancelPressed() {
	// Check if any fields have data
	if af.hasUnsavedData() && af.onCancelConfirm != nil {
		af.onCancelConfirm(
			"Discard unsaved credential?\nAll entered data will be lost.",
			func() {
				// Yes - discard and close
				if af.onCancel != nil {
					af.onCancel()
				}
			},
			func() {
				// No - return to form (do nothing)
			},
		)
	} else {
		// No data or no confirmation callback - close immediately
		if af.onCancel != nil {
			af.onCancel()
		}
	}
}

// onGeneratePassword generates a secure password and fills the password field.
func (af *AddForm) onGeneratePassword() {
	// Generate a 20-character password
	password, err := generateSecurePassword(20)
	if err != nil {
		// Error - could show in status bar but form doesn't have direct access
		// Just silently fail for now
		return
	}

	// Set the generated password in the password field
	passwordField := af.form.GetFormItem(2).(*tview.InputField)
	passwordField.SetText(password)

	// Update the password strength indicator
	af.updatePasswordLabel(passwordField, []byte(password))

	// Copy to clipboard
	_ = clipboard.WriteAll(password) // Ignore errors silently
}

// hasUnsavedData checks if any form fields contain data.
func (af *AddForm) hasUnsavedData() bool {
	service := af.form.GetFormItem(0).(*tview.InputField).GetText()
	username := af.form.GetFormItem(1).(*tview.InputField).GetText()
	password := af.form.GetFormItem(2).(*tview.InputField).GetText()
	category := af.form.GetFormItem(3).(*tview.InputField).GetText()
	url := af.form.GetFormItem(4).(*tview.InputField).GetText()
	notes := af.form.GetFormItem(5).(*tview.TextArea).GetText()
	totp := af.form.GetFormItem(6).(*tview.InputField).GetText()

	// Consider form "dirty" if any field has non-empty value
	// Ignore "Uncategorized" since it's the default
	return service != "" || username != "" || password != "" ||
		(category != "" && category != "Uncategorized") ||
		url != "" || notes != "" || totp != ""
}

// validate checks that required fields are filled.
// Returns error describing first validation failure, or nil if valid.
func (af *AddForm) validate() error {
	// Service is required (cannot be empty)
	service := af.form.GetFormItem(0).(*tview.InputField).GetText()
	if service == "" {
		return fmt.Errorf("service is required")
	}

	// Username is required (minimum validation)
	username := af.form.GetFormItem(1).(*tview.InputField).GetText()
	if username == "" {
		return fmt.Errorf("username is required")
	}

	// Password validation (basic check)
	password := af.form.GetFormItem(2).(*tview.InputField).GetText()
	if password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}

// T048, T049: updatePasswordLabel updates the password field label with strength indicator
func (af *AddForm) updatePasswordLabel(field *tview.InputField, password []byte) {
	policy := security.DefaultPasswordPolicy
	strength := policy.Strength(password)

	var label string
	if len(password) == 0 {
		label = "Password"
	} else {
		switch strength {
		case security.PasswordStrengthWeak:
			label = "Password [yellow](Weak)[-]"
		case security.PasswordStrengthMedium:
			label = "Password [orange](Medium)[-]"
		case security.PasswordStrengthStrong:
			label = "Password [green](Strong)[-]"
		}
	}

	field.SetLabel(label)
}

// getCategories retrieves available categories from AppState.
// Returns default "Uncategorized" if no categories exist.
func (af *AddForm) getCategories() []string {
	categories := af.appState.GetCategories()
	if len(categories) == 0 {
		return []string{"Uncategorized"}
	}
	return categories
}

// wrapInFrame wraps the form in a Flex with a TextView footer for keyboard hints.
func (af *AddForm) wrapInFrame() {
	theme := styles.GetCurrentTheme()

	// Create hints footer as a TextView with wrapping enabled
	// Match statusbar style: [yellow] for keys, [white] for separators
	hintsText := "[yellow]Tab[white]/[yellow]Shift+Tab[-]:Navigate  [yellow]Ctrl+S[-]:Add  [yellow]Ctrl+G[-]:Generate password  [yellow]Ctrl+P[-]:Toggle password  [yellow]Esc[-]:Cancel"
	hints := tview.NewTextView().
		SetText(hintsText).
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true)
	hints.SetBackgroundColor(theme.Background)

	// Create Flex layout: form on top, hints at bottom
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(af.form, 0, 1, true). // Form takes all available space
		AddItem(hints, 2, 0, false)   // Hints fixed at 2 rows (enough for wrapped text)

	// Apply border and title to the flex container
	flex.SetBorder(true).
		SetTitle(" Add Credential ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderColor)

	af.Flex = flex
}

// setupKeyboardShortcuts configures form-level keyboard shortcuts.
// Adds Ctrl+S for quick-save and ensures Tab/Shift+Tab stay within form.
func (af *AddForm) setupKeyboardShortcuts() {
	af.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			// Ctrl+S for quick-save
			af.onAddPressed()
			return nil

		case tcell.KeyCtrlG:
			// Ctrl+G for password generation
			af.onGeneratePassword()
			return nil

		case tcell.KeyCtrlP:
			// Ctrl+P for password visibility toggle
			af.togglePasswordVisibility()
			return nil

		case tcell.KeyEscape:
			// Handle Escape to close form
			af.onCancelPressed()
			return nil

		case tcell.KeyTab, tcell.KeyBacktab:
			// Let form handle Tab/Shift+Tab internally
			// tview's Form will cycle through focusable items and skip disabled ones
			// Return event to allow built-in navigation
			return event
		}
		return event
	})
}

// applyStyles applies theme colors and border styling to the form.
func (af *AddForm) applyStyles() {
	theme := styles.GetCurrentTheme()

	// Apply form-level styling
	styles.ApplyFormStyle(af.form)

	// Style individual input fields (7 fields: Service, Username, Password, Category, URL, Notes, TOTP)
	// Use BackgroundLight for input fields - lighter than form Background for contrast
	for i := 0; i < 7; i++ {
		item := af.form.GetFormItem(i)
		switch field := item.(type) {
		case *tview.InputField:
			field.SetFieldBackgroundColor(theme.BackgroundLight).
				SetFieldTextColor(theme.TextPrimary)
		case *tview.TextArea:
			field.SetTextStyle(tcell.StyleDefault.
				Background(theme.BackgroundLight).
				Foreground(theme.TextPrimary))
		case *tview.DropDown:
			field.SetFieldBackgroundColor(theme.BackgroundLight).
				SetFieldTextColor(theme.TextPrimary)
		}
	}

	// Button alignment
	af.form.SetButtonsAlign(tview.AlignRight)
}

// togglePasswordVisibility switches between masked and plaintext password display.
// Updates both the mask character and the field label to indicate current state.
func (af *AddForm) togglePasswordVisibility() {
	af.passwordVisible = !af.passwordVisible

	passwordField := af.form.GetFormItem(2).(*tview.InputField)

	if af.passwordVisible {
		passwordField.SetMaskCharacter(0) // 0 = plaintext (tview convention)
		passwordField.SetLabel("Password [VISIBLE]")
	} else {
		passwordField.SetMaskCharacter('*')
		// Restore label with current strength if password exists
		text := passwordField.GetText()
		if text != "" {
			af.updatePasswordLabel(passwordField, []byte(text))
		} else {
			passwordField.SetLabel("Password")
		}
	}
}

// SetOnSubmit registers a callback to be invoked after successful add.
func (af *AddForm) SetOnSubmit(callback func()) {
	af.onSubmit = callback
}

// SetOnCancel registers a callback to be invoked when cancel is pressed.
func (af *AddForm) SetOnCancel(callback func()) {
	af.onCancel = callback
}

// SetOnCancelConfirm registers a callback to show confirmation dialogs.
func (af *AddForm) SetOnCancelConfirm(callback func(message string, onYes func(), onNo func())) {
	af.onCancelConfirm = callback
}

// GetFormItem delegates to the internal form for test access.
func (af *AddForm) GetFormItem(index int) tview.FormItem {
	return af.form.GetFormItem(index)
}

// GetInputCapture delegates to the internal form for test access.
func (af *AddForm) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return af.form.GetInputCapture()
}

// NewEditForm creates a new form for editing an existing credential.
// Pre-populates all fields with current credential values.
func NewEditForm(appState *models.AppState, credential *vault.CredentialMetadata) *EditForm {
	form := tview.NewForm()

	ef := &EditForm{
		form:       form,
		appState:   appState,
		credential: credential,
	}

	ef.buildFormFieldsWithValues()
	ef.applyStyles()
	ef.setupKeyboardShortcuts()
	ef.wrapInFrame()

	return ef
}

// buildFormFieldsWithValues constructs form fields pre-populated with credential data.
func (ef *EditForm) buildFormFieldsWithValues() {
	categories := ef.getCategories()

	// Pre-populate fields with existing credential data
	// Service field is read-only (cannot be changed in edit mode)
	// Use 0 width to make fields fill available space (prevents black rectangles)
	ef.form.AddInputField("Service (UID)", ef.credential.Service, 0, nil, nil)
	serviceField := ef.form.GetFormItem(0).(*tview.InputField)
	serviceField.SetDisabled(true) // Make read-only to prevent confusion

	ef.form.AddInputField("Username", ef.credential.Username, 0, nil, nil)

	// T048, T049: Password field with real-time strength indicator
	// Defer fetching until user focuses field (lazy loading)
	// This prevents blocking UI and avoids incrementing usage stats on form open
	passwordField := tview.NewInputField().
		SetLabel("Password").
		SetFieldWidth(0).
		SetMaskCharacter('*')

	// Attach focus handler to fetch password lazily
	passwordField.SetFocusFunc(func() {
		ef.fetchPasswordIfNeeded(passwordField)
	})

	// T049: Update password label with strength indicator inline
	passwordField.SetChangedFunc(func(text string) {
		ef.updatePasswordLabel(passwordField, []byte(text))
	})

	ef.form.AddFormItem(passwordField)

	// Optional metadata fields - pre-populated from credential
	// ef.AddDropDown("Category", categories, categoryIndex, nil)

	// Replace DropDown with InputField + autocomplete
	// Pre-populate with existing category (or "Uncategorized" if empty)
	initialCategory := ef.credential.Category
	if initialCategory == "" {
		initialCategory = "Uncategorized"
	}

	categoryField := tview.NewInputField().
		SetLabel("Category").
		SetFieldWidth(0).
		SetText(initialCategory)

	categoryField.SetAutocompleteFunc(func(currentText string) []string {
		if currentText == "" {
			return categories
		}
		var matches []string
		lowerText := strings.ToLower(currentText)
		for _, cat := range categories {
			if strings.HasPrefix(strings.ToLower(cat), lowerText) {
				matches = append(matches, cat)
			}
		}
		return matches
	})

	ef.form.AddFormItem(categoryField)

	ef.form.AddInputField("URL", ef.credential.URL, 0, nil, nil)
	ef.form.AddTextArea("Notes", ef.credential.Notes, 0, 5, 0, nil)

	// TOTP field - show current status in label, allow adding/updating
	totpLabel := "TOTP Secret/URI"
	if ef.credential.HasTOTP {
		if ef.credential.TOTPIssuer != "" {
			totpLabel = fmt.Sprintf("TOTP [%s] (leave empty to keep)", ef.credential.TOTPIssuer)
		} else {
			totpLabel = "TOTP [configured] (leave empty to keep)"
		}
	}
	ef.form.AddInputField(totpLabel, "", 0, nil, nil)

	// Clear TOTP checkbox - only meaningful if credential has TOTP
	if ef.credential.HasTOTP {
		ef.form.AddCheckbox("Clear TOTP", false, func(checked bool) {
			ef.clearTOTP = checked
		})
	}

	// Action buttons
	ef.form.AddButton("Generate Password", ef.onGeneratePassword)
	ef.form.AddButton("Save", ef.onSavePressed)
	ef.form.AddButton("Cancel", ef.onCancelPressed)
}

// fetchPasswordIfNeeded lazily fetches the password when the field is focused.
// Uses track=false to avoid incrementing usage statistics on form pre-population.
// Caches result to avoid redundant fetches on repeated focus events.
func (ef *EditForm) fetchPasswordIfNeeded(passwordField *tview.InputField) {
	// Only fetch once
	if ef.passwordFetched {
		return
	}

	// Fetch credential without tracking (track=false)
	cred, err := ef.appState.GetFullCredentialWithTracking(ef.credential.Service, false)
	if err != nil {
		// Surface error via AppState error handler without blocking UI
		// Leave password field empty on error
		ef.passwordFetched = true // Mark as attempted to avoid retry loops
		return
	}

	// Set password field and cache original value
	// T020d: Convert []byte password to string for display
	if cred != nil {
		ef.originalPassword = string(cred.Password)
		passwordField.SetText(string(cred.Password))
	}

	ef.passwordFetched = true
}

// onSavePressed handles the Save button submission.
// Shows confirmation if data changed, validates, and saves credential.
func (ef *EditForm) onSavePressed() {
	// Check if any changes were made
	if ef.hasUnsavedChanges() && ef.onCancelConfirm != nil {
		ef.onCancelConfirm(
			"Save changes to credential?",
			func() {
				// Yes - proceed with save
				ef.performSave()
			},
			func() {
				// No - return to form (do nothing)
			},
		)
	} else {
		// No changes or no confirmation callback - save immediately
		ef.performSave()
	}
}

// performSave validates and saves the credential changes.
func (ef *EditForm) performSave() {
	// Validate inputs before submission
	if err := ef.validate(); err != nil {
		// Validation failed - form stays open for correction
		return
	}

	// Extract field values
	// Note: Use original credential.Service as identifier, not form field
	// This prevents ErrCredentialNotFound if user tries to edit service name
	// For service renaming, a dedicated rename flow should be implemented
	service := ef.credential.Service
	username := ef.form.GetFormItem(1).(*tview.InputField).GetText()
	password := ef.form.GetFormItem(2).(*tview.InputField).GetText()

	// Extract category from input field
	category := ef.form.GetFormItem(3).(*tview.InputField).GetText()
	category = normalizeCategory(category) // Convert "Uncategorized" to empty string

	url := ef.form.GetFormItem(4).(*tview.InputField).GetText()
	notes := ef.form.GetFormItem(5).(*tview.TextArea).GetText()
	totpInput := ef.form.GetFormItem(6).(*tview.InputField).GetText()

	// Build UpdateCredentialOpts with only non-empty fields
	opts := models.UpdateCredentialOpts{}

	if username != "" {
		opts.Username = &username
	}

	// Only update password if user changed it (not empty AND different from original)
	// This prevents unnecessary updates when user just views the form
	// T020d: Convert string password to []byte for UpdateOpts
	if password != "" && password != ef.originalPassword {
		passwordBytes := []byte(password)
		opts.Password = &passwordBytes
	}

	// Always set category (even if empty, to allow clearing)
	opts.Category = &category

	// Always set URL (even if empty, to allow clearing)
	opts.URL = &url

	// Always set notes (even if empty, to allow clearing)
	opts.Notes = &notes

	// Handle TOTP: clear takes precedence, then update if provided
	if ef.clearTOTP {
		opts.ClearTOTP = true
	} else if totpInput != "" {
		// Parse and validate TOTP input
		totpConfig, err := vault.ParseTOTPURI(strings.TrimSpace(totpInput))
		if err != nil {
			// TOTP parsing failed - don't fail the whole save, just skip TOTP update
			// User can try again
		} else {
			opts.TOTPSecret = &totpConfig.Secret
			opts.TOTPAlgorithm = &totpConfig.Algorithm
			opts.TOTPDigits = &totpConfig.Digits
			opts.TOTPPeriod = &totpConfig.Period
			if totpConfig.Issuer != "" {
				opts.TOTPIssuer = &totpConfig.Issuer
			}
		}
	}

	// Call AppState to update credential with options struct
	err := ef.appState.UpdateCredential(service, opts)
	if err != nil {
		// Error already handled by AppState onError callback
		// Form stays open for correction
		return
	}

	// Success - invoke callback to close modal
	if ef.onSubmit != nil {
		ef.onSubmit()
	}
}

// onCancelPressed handles the Cancel button.
// Shows confirmation dialog if there are unsaved changes, otherwise closes immediately.
func (ef *EditForm) onCancelPressed() {
	// Check if any changes were made
	if ef.hasUnsavedChanges() && ef.onCancelConfirm != nil {
		ef.onCancelConfirm(
			"Discard unsaved changes?\nAll modifications will be lost.",
			func() {
				// Yes - discard and close
				if ef.onCancel != nil {
					ef.onCancel()
				}
			},
			func() {
				// No - return to form (do nothing)
			},
		)
	} else {
		// No changes or no confirmation callback - close immediately
		if ef.onCancel != nil {
			ef.onCancel()
		}
	}
}

// onGeneratePassword generates a secure password and fills the password field.
func (ef *EditForm) onGeneratePassword() {
	// Generate a 20-character password
	password, err := generateSecurePassword(20)
	if err != nil {
		// Error - could show in status bar but form doesn't have direct access
		// Just silently fail for now
		return
	}

	// Set the generated password in the password field
	passwordField := ef.form.GetFormItem(2).(*tview.InputField)
	passwordField.SetText(password)

	// Update the password strength indicator
	ef.updatePasswordLabel(passwordField, []byte(password))

	// Copy to clipboard
	_ = clipboard.WriteAll(password) // Ignore errors silently
}

// hasUnsavedChanges checks if any form fields have been modified from original values.
func (ef *EditForm) hasUnsavedChanges() bool {
	username := ef.form.GetFormItem(1).(*tview.InputField).GetText()
	password := ef.form.GetFormItem(2).(*tview.InputField).GetText()
	category := ef.form.GetFormItem(3).(*tview.InputField).GetText()
	url := ef.form.GetFormItem(4).(*tview.InputField).GetText()
	notes := ef.form.GetFormItem(5).(*tview.TextArea).GetText()
	totpInput := ef.form.GetFormItem(6).(*tview.InputField).GetText()

	// Normalize current category for comparison
	normalizedCategory := normalizeCategory(category)

	// Compare with original values
	return username != ef.credential.Username ||
		password != ef.originalPassword ||
		normalizedCategory != ef.credential.Category ||
		url != ef.credential.URL ||
		notes != ef.credential.Notes ||
		totpInput != "" || // Any TOTP input means changes
		ef.clearTOTP // Clear TOTP checkbox is checked
}

// validate checks that required fields are filled.
// Returns error describing first validation failure, or nil if valid.
func (ef *EditForm) validate() error {
	// Service is required (cannot be empty)
	service := ef.form.GetFormItem(0).(*tview.InputField).GetText()
	if service == "" {
		return fmt.Errorf("service is required")
	}

	// Username is required (minimum validation)
	username := ef.form.GetFormItem(1).(*tview.InputField).GetText()
	if username == "" {
		return fmt.Errorf("username is required")
	}

	// Password not required in edit form (can keep existing)

	return nil
}

// getCategories retrieves available categories from AppState.
// Returns default "Uncategorized" if no categories exist.
func (ef *EditForm) getCategories() []string {
	categories := ef.appState.GetCategories()
	if len(categories) == 0 {
		return []string{"Uncategorized"}
	}
	return categories
}

// wrapInFrame wraps the form in a Flex with a TextView footer for keyboard hints.
func (ef *EditForm) wrapInFrame() {
	theme := styles.GetCurrentTheme()

	// Create hints footer as a TextView with wrapping enabled
	// Match statusbar style: [yellow] for keys, [white] for separators
	hintsText := "[yellow]Tab[white]/[yellow]Shift+Tab[-]:Navigate  [yellow]Ctrl+S[-]:Save  [yellow]Ctrl+G[-]:Generate password  [yellow]Ctrl+P[-]:Toggle password  [yellow]Esc[-]:Cancel"
	hints := tview.NewTextView().
		SetText(hintsText).
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true)
	hints.SetBackgroundColor(theme.Background)

	// Create Flex layout: form on top, hints at bottom
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ef.form, 0, 1, true). // Form takes all available space
		AddItem(hints, 2, 0, false)   // Hints fixed at 2 rows (enough for wrapped text)

	// Apply border and title to the flex container
	flex.SetBorder(true).
		SetTitle(" Edit Credential ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderColor)

	ef.Flex = flex
}

// setupKeyboardShortcuts configures form-level keyboard shortcuts.
// Adds Ctrl+S for quick-save and ensures Tab/Shift+Tab stay within form.
func (ef *EditForm) setupKeyboardShortcuts() {
	ef.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			// Ctrl+S for quick-save
			ef.onSavePressed()
			return nil

		case tcell.KeyCtrlG:
			// Ctrl+G for password generation
			ef.onGeneratePassword()
			return nil

		case tcell.KeyCtrlP:
			// Ctrl+P for password visibility toggle
			ef.togglePasswordVisibility()
			return nil

		case tcell.KeyEscape:
			// Handle Escape to close form
			ef.onCancelPressed()
			return nil

		case tcell.KeyTab, tcell.KeyBacktab:
			// Let form handle Tab/Shift+Tab internally
			// tview's Form will cycle through focusable items and skip disabled ones
			// Return event to allow built-in navigation
			return event
		}
		return event
	})
}

// T048, T049: updatePasswordLabel updates the password field label with strength indicator
func (ef *EditForm) updatePasswordLabel(field *tview.InputField, password []byte) {
	policy := security.DefaultPasswordPolicy
	strength := policy.Strength(password)

	var label string
	if len(password) == 0 {
		label = "Password"
	} else {
		switch strength {
		case security.PasswordStrengthWeak:
			label = "Password [yellow](Weak)[-]"
		case security.PasswordStrengthMedium:
			label = "Password [orange](Medium)[-]"
		case security.PasswordStrengthStrong:
			label = "Password [green](Strong)[-]"
		}
	}

	field.SetLabel(label)
}

// applyStyles applies theme colors and border styling to the form.
func (ef *EditForm) applyStyles() {
	theme := styles.GetCurrentTheme()

	// Apply form-level styling
	styles.ApplyFormStyle(ef.form)

	// Style individual input fields
	// Form has 7 fields (Service, Username, Password, Category, URL, Notes, TOTP)
	// Plus optional Clear TOTP checkbox (8th item) if credential has TOTP
	// Use BackgroundLight for input fields - lighter than form Background for contrast
	numFields := 7
	if ef.credential.HasTOTP {
		numFields = 8 // Include Clear TOTP checkbox
	}
	for i := 0; i < numFields; i++ {
		item := ef.form.GetFormItem(i)
		switch field := item.(type) {
		case *tview.InputField:
			field.SetFieldBackgroundColor(theme.BackgroundLight).
				SetFieldTextColor(theme.TextPrimary)
		case *tview.TextArea:
			field.SetTextStyle(tcell.StyleDefault.
				Background(theme.BackgroundLight).
				Foreground(theme.TextPrimary))
		case *tview.DropDown:
			field.SetFieldBackgroundColor(theme.BackgroundLight).
				SetFieldTextColor(theme.TextPrimary)
		case *tview.Checkbox:
			field.SetFieldBackgroundColor(theme.BackgroundLight).
				SetFieldTextColor(theme.TextPrimary)
		}
	}

	// Button alignment
	ef.form.SetButtonsAlign(tview.AlignRight)
}

// togglePasswordVisibility switches between masked and plaintext password display.
// Updates both the mask character and the field label to indicate current state.
func (ef *EditForm) togglePasswordVisibility() {
	ef.passwordVisible = !ef.passwordVisible

	passwordField := ef.form.GetFormItem(2).(*tview.InputField)

	if ef.passwordVisible {
		passwordField.SetMaskCharacter(0) // 0 = plaintext (tview convention)
		passwordField.SetLabel("Password [VISIBLE]")
	} else {
		passwordField.SetMaskCharacter('*')
		// Restore label with current strength if password exists
		text := passwordField.GetText()
		if text != "" {
			ef.updatePasswordLabel(passwordField, []byte(text))
		} else {
			passwordField.SetLabel("Password")
		}
	}
}

// SetOnSubmit registers a callback to be invoked after successful update.
func (ef *EditForm) SetOnSubmit(callback func()) {
	ef.onSubmit = callback
}

// SetOnCancel registers a callback to be invoked when cancel is pressed.
func (ef *EditForm) SetOnCancel(callback func()) {
	ef.onCancel = callback
}

// SetOnCancelConfirm registers a callback to show confirmation dialogs.
func (ef *EditForm) SetOnCancelConfirm(callback func(message string, onYes func(), onNo func())) {
	ef.onCancelConfirm = callback
}

// GetFormItem delegates to the internal form for test access.
func (ef *EditForm) GetFormItem(index int) tview.FormItem {
	return ef.form.GetFormItem(index)
}

// GetInputCapture delegates to the internal form for test access.
func (ef *EditForm) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return ef.form.GetInputCapture()
}

// generateSecurePassword generates a cryptographically secure password.
// Reuses the same logic as the CLI generate command.
func generateSecurePassword(length int) (string, error) {
	// Validate length
	if length < 8 {
		return "", fmt.Errorf("password length must be at least 8 characters")
	}
	if length > 128 {
		return "", fmt.Errorf("password length cannot exceed 128 characters")
	}

	// Build character set (always include all types for security)
	const (
		lowerChars  = "abcdefghijklmnopqrstuvwxyz"
		upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digitChars  = "0123456789"
		symbolChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	charset := lowerChars + upperChars + digitChars + symbolChars
	password := make([]byte, length)

	// Ensure at least one character from each required set
	requiredSets := []string{lowerChars, upperChars, digitChars, symbolChars}
	for i, reqSet := range requiredSets {
		if i >= length {
			break
		}
		setLen := big.NewInt(int64(len(reqSet)))
		randomIndex, err := rand.Int(rand.Reader, setLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = reqSet[randomIndex.Int64()]
	}

	// Fill remaining positions with random chars from full charset
	charsetLen := big.NewInt(int64(len(charset)))
	for i := len(requiredSets); i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = charset[randomIndex.Int64()]
	}

	// Shuffle password to avoid predictable positions
	for i := length - 1; i > 0; i-- {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		j := randomIndex.Int64()
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}
