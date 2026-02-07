package events

import (
	"fmt"

	"github.com/arimxyer/pass-cli/cmd/tui/components"
	"github.com/arimxyer/pass-cli/cmd/tui/layout"
	"github.com/arimxyer/pass-cli/cmd/tui/models"
)

// IsComponentVisible checks if a component is visible in the current layout mode.
// Components visibility rules:
// - Small (<80 cols): Only table visible
// - Medium (80-120 cols): Sidebar and table visible
// - Large (>120 cols): All components visible
func IsComponentVisible(component models.FocusableComponent, mode layout.LayoutMode) bool {
	switch component {
	case models.FocusSidebar:
		// Sidebar hidden in small layout
		return mode != layout.LayoutSmall

	case models.FocusDetail:
		// Detail only visible in large layout
		return mode == layout.LayoutLarge

	case models.FocusTable:
		// Table always visible
		return true

	default:
		return false
	}
}

// GetNextVisibleComponent finds the next visible component in tab order.
// Skips hidden components based on the current layout mode.
func GetNextVisibleComponent(current models.FocusableComponent, mode layout.LayoutMode) models.FocusableComponent {
	// Focus order: Sidebar -> Table -> Detail
	order := []models.FocusableComponent{
		models.FocusSidebar,
		models.FocusTable,
		models.FocusDetail,
	}

	// Find current position in order
	currentIndex := -1
	for i, comp := range order {
		if comp == current {
			currentIndex = i
			break
		}
	}

	// If current not found, start from beginning
	if currentIndex == -1 {
		currentIndex = 0
	}

	// Cycle through order, skipping hidden components
	// Try up to len(order) times to find a visible component
	for i := 1; i <= len(order); i++ {
		nextIndex := (currentIndex + i) % len(order)
		next := order[nextIndex]

		if IsComponentVisible(next, mode) {
			return next
		}
	}

	// Fallback: return current if no visible components found
	return current
}

// GetPreviousVisibleComponent finds the previous visible component in tab order.
// Skips hidden components based on the current layout mode.
func GetPreviousVisibleComponent(current models.FocusableComponent, mode layout.LayoutMode) models.FocusableComponent {
	// Focus order: Sidebar -> Table -> Detail
	order := []models.FocusableComponent{
		models.FocusSidebar,
		models.FocusTable,
		models.FocusDetail,
	}

	// Find current position in order
	currentIndex := -1
	for i, comp := range order {
		if comp == current {
			currentIndex = i
			break
		}
	}

	// If current not found, start from end
	if currentIndex == -1 {
		currentIndex = len(order) - 1
	}

	// Cycle through order in reverse, skipping hidden components
	for i := 1; i <= len(order); i++ {
		prevIndex := (currentIndex - i + len(order)) % len(order)
		prev := order[prevIndex]

		if IsComponentVisible(prev, mode) {
			return prev
		}
	}

	// Fallback: return current if no visible components found
	return current
}

// OnFocusChanged is a standard callback for focus changes.
// Updates the status bar to show context-appropriate shortcuts.
func OnFocusChanged(focus models.FocusableComponent, statusBar *components.StatusBar) {
	// Convert FocusableComponent to FocusContext
	var context components.FocusContext
	switch focus {
	case models.FocusSidebar:
		context = components.FocusSidebar
	case models.FocusTable:
		context = components.FocusTable
	case models.FocusDetail:
		context = components.FocusDetail
	default:
		context = components.FocusTable // Default to table
	}

	statusBar.UpdateForContext(context)
}

// SetFocusToComponent safely sets focus to a component with validation.
// Returns error if component is not visible in current layout.
func SetFocusToComponent(nav *models.NavigationState, target models.FocusableComponent, mode layout.LayoutMode) error {
	// Validate component is visible
	if !IsComponentVisible(target, mode) {
		return fmt.Errorf("component %d is not visible in current layout mode", target)
	}

	// Set focus
	nav.SetFocus(target)
	return nil
}

// FocusOnFirstCredential focuses on the first credential in the table.
// Use case: After loading vault, start at first item.
func FocusOnFirstCredential(nav *models.NavigationState) {
	nav.SetFocus(models.FocusTable)
}

// RestoreFocusAfterModal restores focus to the table after closing a modal.
// Default behavior: focus table (most common case).
func RestoreFocusAfterModal(nav *models.NavigationState) {
	nav.SetFocus(models.FocusTable)
}

// CycleFocusWithLayoutAwareness cycles focus to the next visible component.
// Automatically skips hidden components based on layout mode.
func CycleFocusWithLayoutAwareness(nav *models.NavigationState, layoutManager *layout.LayoutManager) {
	current := nav.GetCurrentFocus()
	mode := layoutManager.GetCurrentMode()

	next := GetNextVisibleComponent(current, mode)
	nav.SetFocus(next)
}

// CycleFocusReverseWithLayoutAwareness cycles focus to the previous visible component.
// Automatically skips hidden components based on layout mode.
func CycleFocusReverseWithLayoutAwareness(nav *models.NavigationState, layoutManager *layout.LayoutManager) {
	current := nav.GetCurrentFocus()
	mode := layoutManager.GetCurrentMode()

	prev := GetPreviousVisibleComponent(current, mode)
	nav.SetFocus(prev)
}
