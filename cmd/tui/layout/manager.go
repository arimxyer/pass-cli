// Package layout provides responsive layout management for the TUI.
package layout

import (
	"github.com/rivo/tview"
	"pass-cli/cmd/tui/models"
	"pass-cli/internal/config"
)

// MinTerminalWidth is the minimum terminal width (columns) required for usable interface.
// Below this width, a warning overlay is displayed prompting the user to resize.
const MinTerminalWidth = 60

// MinTerminalHeight is the minimum terminal height (rows) required for usable interface.
// Below this height, a warning overlay is displayed prompting the user to resize.
const MinTerminalHeight = 30

// LayoutMode represents the current layout configuration based on terminal width.
type LayoutMode int

const (
	// LayoutSmall is for terminals < 80 columns (table only, no sidebar)
	LayoutSmall LayoutMode = iota
	// LayoutMedium is for terminals 80-120 columns (sidebar + table)
	LayoutMedium
	// LayoutLarge is for terminals > 120 columns (sidebar + table + detail)
	LayoutLarge
)

// PageManagerInterface defines the methods needed by LayoutManager for size warnings.
// This interface allows testing with mocks while maintaining type safety.
type PageManagerInterface interface {
	ShowSizeWarning(currentWidth, currentHeight, minWidth, minHeight int)
	HideSizeWarning()
	IsSizeWarningActive() bool
}

// LayoutManager manages responsive layout composition and terminal resize handling.
// It adapts the UI based on terminal width using breakpoints.
type LayoutManager struct {
	app      *tview.Application
	appState *models.AppState

	// Current terminal dimensions
	width  int
	height int

	// Current layout mode
	currentMode LayoutMode

	// Manual visibility overrides (nil = auto/responsive, true = force show, false = force hide)
	detailPanelOverride *bool
	sidebarOverride     *bool

	// Layout primitives
	mainLayout *tview.Flex // Root layout (vertical: contentRow + statusBar)
	contentRow *tview.Flex // Main content area (horizontal: sidebar + table + detail)

	// Component references (retrieved from AppState, not created new)
	sidebar    *tview.TreeView
	table      *tview.Table
	detailView *tview.TextView
	statusBar  *tview.TextView

	// Breakpoints (configurable)
	mediumBreakpoint int // Default: 80
	largeBreakpoint  int // Default: 120

	// PageManager for terminal size warnings
	pageManager PageManagerInterface

	// User configuration (T022)
	config *config.Config

	// Detail panel positioning
	detailPosition      string // "auto", "right", "bottom"
	detailAutoThreshold int    // Width threshold for auto positioning
}

// NewLayoutManager creates a new layout manager with default breakpoints.
// Components are not retrieved until CreateMainLayout is called.
// T022: Now accepts config parameter for terminal size configuration.
func NewLayoutManager(app *tview.Application, appState *models.AppState, cfg *config.Config) *LayoutManager {
	// Use defaults if config is nil (for testing)
	detailPosition := "auto"
	detailAutoThreshold := 140
	if cfg != nil {
		detailPosition = cfg.Terminal.DetailPosition
		detailAutoThreshold = cfg.Terminal.DetailAutoThreshold
	}

	return &LayoutManager{
		app:                 app,
		appState:            appState,
		mediumBreakpoint:    80,
		largeBreakpoint:     120,
		currentMode:         LayoutSmall, // Start with small, will adjust on first draw
		config:              cfg,
		detailPosition:      detailPosition,
		detailAutoThreshold: detailAutoThreshold,
	}
}

// determineDetailPosition decides where to position the detail panel based on config and window width.
// Returns "right" for horizontal layout or "bottom" for vertical layout.
func (lm *LayoutManager) determineDetailPosition() string {
	// If manual override (not "auto"), use it directly
	if lm.detailPosition != "auto" {
		return lm.detailPosition
	}

	// Auto logic: use bottom if width < threshold, otherwise right
	if lm.width < lm.detailAutoThreshold {
		return "bottom"
	}
	return "right"
}

// CreateMainLayout builds the complete layout structure.
// Structure:
//
//	┌───────────────────────────────┐
//	│   contentRow (FlexColumn)     │  ← Main content area (flex)
//	│   ┌──────┬───────┬──────────┐ │
//	│   │ Side │ Table │  Detail  │ │
//	│   │ bar  │       │          │ │
//	│   └──────┴───────┴──────────┘ │
//	├───────────────────────────────┤
//	│       statusBar (1 row)       │  ← Fixed height
//	└───────────────────────────────┘
func (lm *LayoutManager) CreateMainLayout() *tview.Flex {
	// Get component references from appState (don't create new ones)
	lm.sidebar = lm.appState.GetSidebar()
	lm.table = lm.appState.GetTable()
	lm.detailView = lm.appState.GetDetailView()
	lm.statusBar = lm.appState.GetStatusBar()

	// Create content row (horizontal layout for main panels)
	lm.contentRow = tview.NewFlex().SetDirection(tview.FlexColumn)

	// Create main layout (vertical: content + status bar)
	lm.mainLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(lm.contentRow, 0, 1, true). // Content area (flex height)
		AddItem(lm.statusBar, 1, 0, false)  // Status bar (fixed 1 row)

	return lm.mainLayout
}

// HandleResize responds to terminal size changes.
// It determines if the layout mode needs to change and triggers a rebuild if necessary.
// Also checks terminal size against minimum requirements and shows/hides warning overlay.
// T022+T023: Now uses config values for terminal size thresholds and respects WarningEnabled flag.
func (lm *LayoutManager) HandleResize(width, height int) {
	lm.width = width
	lm.height = height

	// T022+T023: Get terminal config values (use defaults if config not set)
	minWidth := MinTerminalWidth
	minHeight := MinTerminalHeight
	warningEnabled := true

	if lm.config != nil {
		minWidth = lm.config.Terminal.MinWidth
		minHeight = lm.config.Terminal.MinHeight
		warningEnabled = lm.config.Terminal.WarningEnabled
	}

	// Check minimum terminal size and show/hide warning (T023: respect WarningEnabled flag)
	if lm.pageManager != nil && warningEnabled {
		if width < minWidth || height < minHeight {
			lm.pageManager.ShowSizeWarning(width, height, minWidth, minHeight)
		} else {
			lm.pageManager.HideSizeWarning()
		}
	} else if lm.pageManager != nil && !warningEnabled {
		// T023: Warning disabled, always hide
		lm.pageManager.HideSizeWarning()
	}

	// Determine the new layout mode based on width
	newMode := lm.determineLayoutMode(width)

	// Only rebuild layout if the mode changed
	if newMode != lm.currentMode {
		lm.currentMode = newMode
		lm.RebuildLayout()
	}
}

// RebuildLayout reconstructs the layout based on the current mode.
// It clears the content row and adds components according to breakpoint rules:
//   - Small: Table only (full width)
//   - Medium: Sidebar (20 cols) + Table (flex)
//   - Large: Sidebar + Table + Detail (position determined by determineDetailPosition)
//     - Horizontal (detail on right): Sidebar (20 cols) + Table (flex) + Detail (40 cols)
//     - Vertical (detail on bottom): Sidebar (20 cols) + (Table above Detail stacked vertically)
//
// Manual overrides (detailPanelOverride) take precedence over responsive breakpoints.
func (lm *LayoutManager) RebuildLayout() {
	// Skip rebuild if layout hasn't been initialized yet
	if lm.contentRow == nil {
		return
	}

	// Clear existing content
	lm.contentRow.Clear()

	// Determine effective mode (considering manual overrides)
	effectiveMode := lm.currentMode

	// Apply detail panel override if set
	if lm.detailPanelOverride != nil {
		if *lm.detailPanelOverride {
			// Force detail panel on (upgrade to Large mode)
			effectiveMode = LayoutLarge
		} else if lm.currentMode == LayoutLarge {
			// Force detail panel off (downgrade to Medium)
			effectiveMode = LayoutMedium
		}
	}

	// Determine sidebar visibility
	showSidebar := lm.shouldShowSidebar()

	// Get table area (may include search input if active)
	tableArea := lm.getTableArea()

	// Build layout based on effective mode and sidebar visibility
	switch effectiveMode {
	case LayoutSmall:
		if showSidebar {
			// Sidebar + Table (forced by override in small mode)
			lm.contentRow.
				AddItem(lm.sidebar, 20, 0, false).
				AddItem(tableArea, 0, 1, true)
		} else {
			// Table only (full width)
			lm.contentRow.AddItem(tableArea, 0, 1, true)
		}

	case LayoutMedium:
		if showSidebar {
			// Sidebar + Table
			lm.contentRow.
				AddItem(lm.sidebar, 20, 0, false).
				AddItem(tableArea, 0, 1, true)
		} else {
			// Table only (sidebar hidden by override)
			lm.contentRow.AddItem(tableArea, 0, 1, true)
		}

	case LayoutLarge:
		// Determine detail panel position
		detailPos := lm.determineDetailPosition()

		if detailPos == "bottom" {
			// Vertical layout: detail on bottom
			// Create right area with table and detail stacked vertically
			rightArea := tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(tableArea, 0, 2, true).      // Table gets 2/3 height
				AddItem(lm.detailView, 0, 1, false)  // Detail gets 1/3 height

			if showSidebar {
				// Sidebar + (Table above Detail)
				lm.contentRow.
					AddItem(lm.sidebar, 20, 0, false).
					AddItem(rightArea, 0, 1, true)
			} else {
				// (Table above Detail) - sidebar hidden by override
				lm.contentRow.AddItem(rightArea, 0, 1, true)
			}
		} else {
			// Horizontal layout: detail on right (default/original behavior)
			if showSidebar {
				// Sidebar + Table + Detail
				lm.contentRow.
					AddItem(lm.sidebar, 20, 0, false).
					AddItem(tableArea, 0, 1, true).
					AddItem(lm.detailView, 40, 0, false)
			} else {
				// Table + Detail (sidebar hidden by override)
				lm.contentRow.
					AddItem(tableArea, 0, 1, true).
					AddItem(lm.detailView, 40, 0, false)
			}
		}
	}
}

// determineLayoutMode calculates the appropriate layout mode based on terminal width.
// Breakpoint rules:
//   - width < mediumBreakpoint (default 80): LayoutSmall
//   - width >= mediumBreakpoint && width < largeBreakpoint (default 80-119): LayoutMedium
//   - width >= largeBreakpoint (default 120+): LayoutLarge
func (lm *LayoutManager) determineLayoutMode(width int) LayoutMode {
	if width < lm.mediumBreakpoint {
		return LayoutSmall
	}
	if width < lm.largeBreakpoint {
		return LayoutMedium
	}
	return LayoutLarge
}

// GetCurrentMode returns the current layout mode.
func (lm *LayoutManager) GetCurrentMode() LayoutMode {
	return lm.currentMode
}

// SetBreakpoints allows customizing the layout breakpoints.
// Default values are medium=80, large=120.
func (lm *LayoutManager) SetBreakpoints(medium, large int) {
	lm.mediumBreakpoint = medium
	lm.largeBreakpoint = large
}

// SetPageManager injects the PageManager for terminal size warning functionality.
// This must be called before HandleResize can display size warnings.
func (lm *LayoutManager) SetPageManager(pm PageManagerInterface) {
	lm.pageManager = pm
}

// ToggleDetailPanel manually shows/hides the detail panel, cycling through three states:
//   - Auto (nil): Detail panel follows responsive breakpoints (default)
//   - ForceHide (false): Detail panel hidden even at large terminal widths
//   - ForceShow (true): Detail panel visible regardless of terminal width
//
// Returns a message describing the new state for status bar display.
func (lm *LayoutManager) ToggleDetailPanel() string {
	var message string

	if lm.detailPanelOverride == nil {
		// Auto -> ForceHide
		forceHide := false
		lm.detailPanelOverride = &forceHide
		message = "Detail panel: Hidden"
	} else if !*lm.detailPanelOverride {
		// ForceHide -> ForceShow
		forceShow := true
		lm.detailPanelOverride = &forceShow
		message = "Detail panel: Visible"
	} else {
		// ForceShow -> Auto
		lm.detailPanelOverride = nil
		message = "Detail panel: Auto (responsive)"
	}

	// Rebuild layout with new override
	lm.RebuildLayout()

	return message
}

// shouldShowSidebar determines if the sidebar should be visible based on override and responsive logic.
// Manual override takes precedence over responsive breakpoints.
func (lm *LayoutManager) shouldShowSidebar() bool {
	if lm.sidebarOverride != nil {
		return *lm.sidebarOverride // Manual override takes precedence
	}
	// Fallback to responsive logic: show sidebar if width >= medium breakpoint
	return lm.width >= lm.mediumBreakpoint
}

// getTableArea returns the table area, optionally wrapped with search input.
// When search is active, returns a vertical Flex with InputField + Table.
// When search is inactive, returns just the table.
func (lm *LayoutManager) getTableArea() tview.Primitive {
	searchState := lm.appState.GetSearchState()

	// If search is not active, return table directly
	if searchState == nil || !searchState.Active {
		return lm.table
	}

	// Search is active - create vertical Flex with InputField + Table
	tableArea := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(searchState.InputField, 1, 0, true). // Search input (1 row, focusable)
		AddItem(lm.table, 0, 1, false)               // Table (flex height, not directly focusable)

	return tableArea
}

// ToggleSidebar manually shows/hides the sidebar, cycling through three states:
//   - Auto (nil): Sidebar follows responsive breakpoints (default)
//   - ForceHide (false): Sidebar hidden regardless of terminal width
//   - ForceShow (true): Sidebar visible regardless of terminal width
//
// Returns a message describing the new state for status bar display.
func (lm *LayoutManager) ToggleSidebar() string {
	var message string

	if lm.sidebarOverride == nil {
		// Auto -> ForceHide
		forceHide := false
		lm.sidebarOverride = &forceHide
		message = "Sidebar: Hidden"
	} else if !*lm.sidebarOverride {
		// ForceHide -> ForceShow
		forceShow := true
		lm.sidebarOverride = &forceShow
		message = "Sidebar: Visible"
	} else {
		// ForceShow -> Auto
		lm.sidebarOverride = nil
		message = "Sidebar: Auto (responsive)"
	}

	// Rebuild layout with new override
	lm.RebuildLayout()

	return message
}

// GetSidebarOverride returns the current sidebar override state for testing.
func (lm *LayoutManager) GetSidebarOverride() *bool {
	return lm.sidebarOverride
}

// SetSidebarOverride sets the sidebar override state for testing.
func (lm *LayoutManager) SetSidebarOverride(override *bool) {
	lm.sidebarOverride = override
}

// ShouldShowSidebar exposes the sidebar visibility logic for testing.
func (lm *LayoutManager) ShouldShowSidebar() bool {
	return lm.shouldShowSidebar()
}
