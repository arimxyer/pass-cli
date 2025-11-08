package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/howeyc/gopass"
	"pass-cli/cmd/tui/components"
	"pass-cli/cmd/tui/events"
	"pass-cli/cmd/tui/layout"
	"pass-cli/cmd/tui/models"
	"pass-cli/cmd/tui/styles"
	"pass-cli/internal/config"
	"pass-cli/internal/vault"
)

const maxPasswordAttempts = 3

// Run starts the TUI application (exported for main.go to call)
// If vaultPath is empty, uses the default vault location
func Run(vaultPath string) error {
	// 1. Get vault path (use provided path or default)
	if vaultPath == "" {
		vaultPath = getDefaultVaultPath()
	}

	// 1a. Check if vault exists - if not, trigger guided initialization
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		fmt.Println("\n👋 Welcome to Pass-CLI!")
		fmt.Println("\nIt looks like this is your first time using pass-cli.")

		// Run guided initialization
		if err := vault.RunGuidedInit(vaultPath, true); err != nil {
			// User declined or error
			return fmt.Errorf("vault initialization required: %w", err)
		}
	}

	// 2. Initialize vault service
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to initialize vault service: %w", err)
	}

	// 3. Check metadata to see if keychain is enabled (T019)
	metadata, err := vaultService.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load vault metadata: %w", err)
	}

	// 4. Try keychain unlock if enabled (T019 - FR-024)
	if metadata.KeychainEnabled {
		err = vaultService.UnlockWithKeychain()
		if err != nil {
			// T019: Display clear error message (FR-026)
			fmt.Fprintf(os.Stderr, "Keychain unlock failed: %v\n", err)
			fmt.Println("Falling back to password prompt...")
		}
	}

	// 5. Prompt for password if not unlocked via keychain (T019 - FR-025)
	if !vaultService.IsUnlocked() {
		unlocked := false
		for attempt := 1; attempt <= maxPasswordAttempts; attempt++ {
			password, err := promptForPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			// Try to unlock with provided password
			err = vaultService.Unlock(password)
			if err == nil {
				unlocked = true
				break
			}

			// Show error for failed attempt
			if attempt < maxPasswordAttempts {
				fmt.Fprintf(os.Stderr, "Unlock failed: %v. Please try again (%d/%d attempts).\n",
					err, attempt, maxPasswordAttempts)
			}
		}

		// Check if unlock was successful
		if !unlocked {
			return fmt.Errorf("failed to unlock vault after %d attempts", maxPasswordAttempts)
		}
	}

	// 6. Launch TUI
	if err := launchTUI(vaultService); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

// getDefaultVaultPath returns the default vault file path
func getDefaultVaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home not available
		return ".pass-cli/vault.enc"
	}

	return filepath.Join(home, ".pass-cli", "vault.enc")
}

// promptForPassword securely prompts user for master password
func promptForPassword() ([]byte, error) {
	fmt.Print("Enter master password: ")

	// Use gopass for masked input
	passwordBytes, err := gopass.GetPasswdMasked()
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}

	return passwordBytes, nil
}

// LaunchTUI initializes and runs the TUI application
// This function is exported to be called from cmd/root.go
func LaunchTUI(vaultService *vault.VaultService) error {
	// Panic recovery to restore terminal
	defer RestoreTerminal()

	// Set rounded borders globally
	styles.SetRoundedBorders()

	// 1. Load user configuration (T022)
	cfg, validationResult := config.Load()

	// 1. Create tview.Application
	app := NewApp()

	// 2. Initialize AppState with vault service
	appState := models.NewAppState(vaultService)

	// 3. Load credentials
	if err := appState.LoadCredentials(); err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// 4. Create UI components (pass config to statusBar for keybinding hints)
	sidebar := components.NewSidebar(appState)
	table := components.NewCredentialTable(appState)
	detailView := components.NewDetailView(appState)
	statusBar := components.NewStatusBar(app, appState, cfg)

	// 5. Store components in AppState
	appState.SetSidebar(sidebar.TreeView)
	appState.SetTable(table.Table)
	appState.SetDetailView(detailView.TextView)
	appState.SetStatusBar(statusBar.TextView)

	// 6. Register callbacks
	appState.SetOnCredentialsChanged(func() {
		// Refresh all components that depend on credentials
		sidebar.Refresh()
		table.Refresh()
		detailView.Refresh()
	})

	appState.SetOnSelectionChanged(func() {
		// Refresh table to apply category filter and detail view for selection
		table.Refresh()
		detailView.Refresh()
	})

	appState.SetOnFilterChanged(func() {
		// Refresh table only (not detail view) during search filtering
		table.Refresh()
	})

	appState.SetOnError(func(err error) {
		// Display error in status bar
		statusBar.ShowError(err)
	})

	// 7. Create NavigationState
	nav := models.NewNavigationState(app, appState)

	// Register focus change callback to update statusbar
	nav.SetOnFocusChanged(func(focus models.FocusableComponent) {
		events.OnFocusChanged(focus, statusBar)
	})

	// 8. Create LayoutManager (T022: pass config)
	layoutMgr := layout.NewLayoutManager(app, appState, cfg)

	// 9. Create PageManager
	pageManager := layout.NewPageManager(app)

	// 9a. Wire PageManager to LayoutManager BEFORE creating layout
	// This ensures size warning handler is ready when resize events fire
	layoutMgr.SetPageManager(pageManager)

	// 9b. Build layout (this will trigger initial resize check)
	mainLayout := layoutMgr.CreateMainLayout()

	// 9c. T024: Display config validation errors if any
	if !validationResult.Valid && len(validationResult.Errors) > 0 {
		errorMessages := make([]string, len(validationResult.Errors))
		for i, err := range validationResult.Errors {
			if err.Line > 0 {
				errorMessages[i] = fmt.Sprintf("%s (line %d): %s", err.Field, err.Line, err.Message)
			} else {
				errorMessages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
			}
		}
		pageManager.ShowConfigValidationError(errorMessages)
	}

	// 10. Create EventHandler and setup shortcuts (pass config for keybindings)
	eventHandler := events.NewEventHandler(app, appState, nav, pageManager, statusBar, detailView, layoutMgr, cfg)
	eventHandler.SetupGlobalShortcuts()

	// 11. Set up proactive resize handling to prevent crashes
	// This is called before every screen draw.
	var lastWidth, lastHeight int
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, height := screen.Size()
		if lastWidth != width || lastHeight != height {
			// Pass resize event to layout manager to handle responsive changes
			// and show/hide size warnings.
			layoutMgr.HandleResize(width, height)
			lastWidth = width
			lastHeight = height

			// Apply pending warning changes in a goroutine to avoid deadlock
			go pageManager.ApplyPendingWarnings()
		}

		// Returning false allows the draw to continue. We don't want to suppress drawing,
		// just ensure the correct page (main or warning) is visible.
		return false
	})

	// 13. Set root primitive (use pages for modal support over main layout)
	pageManager.ShowPage("main", mainLayout)
	app.SetRoot(pageManager.Pages, true)

	// 14. Run application (blocking)
	return app.Run()
}

// launchTUI is kept as a private wrapper for backward compatibility if needed
func launchTUI(vaultService *vault.VaultService) error {
	return LaunchTUI(vaultService)
}
