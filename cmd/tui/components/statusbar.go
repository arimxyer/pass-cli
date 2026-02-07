package components

import (
	"fmt"
	"time"

	"github.com/rivo/tview"

	"github.com/arimxyer/pass-cli/cmd/tui/models"
	"github.com/arimxyer/pass-cli/cmd/tui/styles"
	"github.com/arimxyer/pass-cli/internal/config"
)

// FocusContext represents the current focus context for determining which shortcuts to display.
type FocusContext int

const (
	// FocusSidebar indicates the sidebar is focused
	FocusSidebar FocusContext = iota
	// FocusTable indicates the credential table is focused
	FocusTable
	// FocusDetail indicates the detail view is focused
	FocusDetail
	// FocusModal indicates a modal (form or dialog) is focused
	FocusModal
)

// StatusBar displays context-aware keyboard shortcuts and temporary status messages.
type StatusBar struct {
	*tview.TextView

	app          *tview.Application // For forcing redraws
	appState     *models.AppState
	config       *config.Config // User configuration for keybindings
	currentFocus FocusContext
	messageTimer *time.Timer
}

// NewStatusBar creates and initializes a new status bar.
func NewStatusBar(app *tview.Application, appState *models.AppState, cfg *config.Config) *StatusBar {
	theme := styles.GetCurrentTheme()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Configure styling: no borders, dark background, fixed height
	textView.SetBackgroundColor(theme.StatusBarBg).
		SetBorder(false)

	sb := &StatusBar{
		TextView:     textView,
		app:          app,
		appState:     appState,
		config:       cfg,
		currentFocus: FocusSidebar, // Default focus
	}

	// Set initial shortcuts display (direct SetText, no queue - app not running yet)
	shortcuts := sb.getShortcutsForContext(FocusSidebar)
	sb.SetText(shortcuts)

	return sb
}

// UpdateForContext updates the displayed shortcuts based on the current focus context.
func (sb *StatusBar) UpdateForContext(focus FocusContext) {
	sb.currentFocus = focus
	shortcuts := sb.getShortcutsForContext(focus)

	// Direct SetText is sufficient - tview redraws automatically on next frame
	sb.SetText(shortcuts)
}

// ShowSuccess displays a temporary success message (green text, 3 seconds).
func (sb *StatusBar) ShowSuccess(message string) {
	formatted := fmt.Sprintf("[green]%s[-]", message)
	sb.showTemporaryMessage(formatted, 3*time.Second)
}

// ShowInfo displays a temporary info message (cyan text, 3 seconds).
func (sb *StatusBar) ShowInfo(message string) {
	formatted := fmt.Sprintf("[cyan]%s[-]", message)
	sb.showTemporaryMessage(formatted, 3*time.Second)
}

// ShowError displays a temporary error message (red text, 5 seconds).
func (sb *StatusBar) ShowError(err error) {
	formatted := fmt.Sprintf("[red]Error: %s[-]", err.Error())
	sb.showTemporaryMessage(formatted, 5*time.Second)
}

// showTemporaryMessage displays a message for the specified duration, then restores shortcuts.
func (sb *StatusBar) showTemporaryMessage(message string, duration time.Duration) {
	// Cancel previous message timer if it exists
	if sb.messageTimer != nil {
		sb.messageTimer.Stop()
	}

	// Display the message
	sb.SetText(message)

	// Schedule restoration of shortcuts after duration
	sb.messageTimer = time.AfterFunc(duration, func() {
		sb.UpdateForContext(sb.currentFocus)
	})
}

// getShortcutsForContext returns the appropriate shortcut text for the given focus context.
func (sb *StatusBar) getShortcutsForContext(focus FocusContext) string {
	// Check if search is active
	searchState := sb.appState.GetSearchState()
	isSearchActive := searchState != nil && searchState.Active

	// Helper to get display string for a keybinding action
	getKey := func(action string) string {
		keyStr := sb.config.GetKeybindingForAction(action)
		if keyStr == "" {
			return ""
		}
		return config.GetDisplayString(keyStr)
	}

	// Format key hint with color
	formatKey := func(action string) string {
		key := getKey(action)
		if key == "" {
			return ""
		}
		return fmt.Sprintf("[yellow]%s[-]", key)
	}

	if isSearchActive {
		// Search mode shortcuts - First Esc exits search input (keeps filter), second Esc clears filter
		helpKey := formatKey("help")
		return fmt.Sprintf("[white]Type to filter  [yellow]↑↓[-]:Navigate  [yellow]Esc[-]:Exit (Esc again clears)  [yellow]p[-]:Show  [yellow]c[-]:Copy  %s:Help", helpKey)
	}

	// Get common keys
	newKey := formatKey("add_credential")
	editKey := formatKey("edit_credential")
	delKey := formatKey("delete_credential")
	searchKey := formatKey("search")
	sidebarKey := formatKey("toggle_sidebar")
	detailKey := formatKey("toggle_detail")
	helpKey := formatKey("help")
	quitKey := formatKey("quit")

	switch focus {
	case FocusSidebar:
		return fmt.Sprintf("[yellow]Tab/Shift+Tab[-]:Switch  [yellow]↑↓[-]:Navigate  [yellow]Enter[-]:Select  %s:New  %s:Search  %s:Sidebar  %s:Details  %s:Help  %s:Quit",
			newKey, searchKey, sidebarKey, detailKey, helpKey, quitKey)

	case FocusTable:
		return fmt.Sprintf("[yellow]Tab/Shift+Tab[-]:Switch  [yellow]↑↓[-]:Navigate  %s:New  %s:Edit  %s:Delete  [yellow]p[-]:Show  [yellow]c[-]:Copy  %s:Search  %s:Sidebar  %s:Details  %s:Help  %s:Quit",
			newKey, editKey, delKey, searchKey, sidebarKey, detailKey, helpKey, quitKey)

	case FocusDetail:
		return fmt.Sprintf("[yellow]Tab/Shift+Tab[-]:Switch  %s:Edit  %s:Delete  [yellow]p[-]:Toggle  [yellow]c[-]:Copy  %s:Search  %s:Sidebar  %s:Details  %s:Help  %s:Quit",
			editKey, delKey, searchKey, sidebarKey, detailKey, helpKey, quitKey)

	case FocusModal:
		return "[yellow]Tab/Shift+Tab[-]:Next/Prev field  [yellow]Enter[-]:Submit  [yellow]Esc[-]:Cancel"

	default:
		return fmt.Sprintf("[yellow]Tab/Shift+Tab[-]:Switch  %s:Search  %s:Sidebar  %s:Details  %s:Help  %s:Quit",
			searchKey, sidebarKey, detailKey, helpKey, quitKey)
	}
}
