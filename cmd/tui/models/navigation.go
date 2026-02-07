package models

import (
	"github.com/arimxyer/pass-cli/cmd/tui/styles"
	"github.com/rivo/tview"
)

// FocusableComponent represents components that can receive focus.
type FocusableComponent int

const (
	FocusSidebar FocusableComponent = iota
	FocusTable
	FocusDetail
)

// NavigationState manages focus navigation between components.
// Provides Tab cycling and focus management with proper order tracking.
type NavigationState struct {
	app      *tview.Application
	appState *AppState

	focusOrder   []tview.Primitive
	currentIndex int

	onFocusChanged func(FocusableComponent)
}

// NewNavigationState creates a new navigation state manager.
// Builds initial focus order from AppState components.
func NewNavigationState(app *tview.Application, appState *AppState) *NavigationState {
	ns := &NavigationState{
		app:          app,
		appState:     appState,
		focusOrder:   make([]tview.Primitive, 0, 3), // Pre-allocate for 3 components
		currentIndex: 0,
	}

	// Build focus order from appState components
	// Order: Sidebar -> Table -> Detail
	if sidebar := appState.GetSidebar(); sidebar != nil {
		ns.focusOrder = append(ns.focusOrder, sidebar)
	}
	if table := appState.GetTable(); table != nil {
		ns.focusOrder = append(ns.focusOrder, table)
	}
	if detail := appState.GetDetailView(); detail != nil {
		ns.focusOrder = append(ns.focusOrder, detail)
	}

	return ns
}

// SetFocusOrder sets the components that can receive focus in tab order.
// Order: Sidebar -> Table -> Detail -> (back to Sidebar)
func (ns *NavigationState) SetFocusOrder(order []tview.Primitive) {
	ns.focusOrder = order
	ns.currentIndex = 0
}

// CycleFocus moves focus to the next component in the focus order.
// Used for Tab key navigation.
func (ns *NavigationState) CycleFocus() {
	if len(ns.focusOrder) == 0 {
		return
	}

	ns.currentIndex = (ns.currentIndex + 1) % len(ns.focusOrder)
	ns.setFocus(ns.currentIndex)
}

// CycleFocusReverse moves focus to the previous component in the focus order.
// Used for Shift+Tab navigation.
func (ns *NavigationState) CycleFocusReverse() {
	if len(ns.focusOrder) == 0 {
		return
	}

	// Calculate previous index (wrap around to end if at beginning)
	ns.currentIndex = (ns.currentIndex - 1 + len(ns.focusOrder)) % len(ns.focusOrder)

	// Use setFocus to update focus AND trigger callbacks
	ns.setFocus(ns.currentIndex)
}

// SetFocus directly sets focus to a specific component.
func (ns *NavigationState) SetFocus(target FocusableComponent) {
	if int(target) < len(ns.focusOrder) {
		ns.currentIndex = int(target)
		ns.setFocus(ns.currentIndex)
	}
}

// GetCurrentFocus returns the currently focused component.
func (ns *NavigationState) GetCurrentFocus() FocusableComponent {
	if ns.currentIndex < 0 || ns.currentIndex >= len(ns.focusOrder) {
		return FocusSidebar
	}
	return FocusableComponent(ns.currentIndex)
}

// SetOnFocusChanged registers a callback to be invoked when focus changes.
func (ns *NavigationState) SetOnFocusChanged(callback func(FocusableComponent)) {
	ns.onFocusChanged = callback
}

// setFocus is an internal helper that updates focus and triggers callbacks.
// CRITICAL: Also updates border colors for visual feedback.
func (ns *NavigationState) setFocus(index int) {
	if index < 0 || index >= len(ns.focusOrder) {
		return
	}

	primitive := ns.focusOrder[index]
	ns.app.SetFocus(primitive)

	// Update border colors for all components
	// Active component gets active border color, others get inactive
	ns.updateBorderColors(index)

	if ns.onFocusChanged != nil {
		ns.onFocusChanged(FocusableComponent(index))
	}
}

// updateBorderColors applies active/inactive border styling to components.
// Active component gets highlighted border, others get dimmed border.
func (ns *NavigationState) updateBorderColors(activeIndex int) {
	// Update sidebar
	if sidebar := ns.appState.GetSidebar(); sidebar != nil {
		sidebarIndex := 0 // Sidebar is first in focus order
		styles.ApplyBorderedStyle(sidebar, "Categories", activeIndex == sidebarIndex)
	}

	// Update table
	if table := ns.appState.GetTable(); table != nil {
		tableIndex := 1 // Table is second in focus order
		styles.ApplyBorderedStyle(table, "Credentials", activeIndex == tableIndex)
	}

	// Update detail view
	if detail := ns.appState.GetDetailView(); detail != nil {
		detailIndex := 2 // Detail is third in focus order
		styles.ApplyBorderedStyle(detail, "Details", activeIndex == detailIndex)
	}
}
