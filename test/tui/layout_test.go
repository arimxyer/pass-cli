package tui_test

import (
	"testing"

	"pass-cli/cmd/tui/layout"
	"pass-cli/cmd/tui/models"
)

// TestSidebarToggleCycles verifies sidebar toggle cycles through nil → false → true → nil
func TestSidebarToggleCycles(t *testing.T) {
	app := SimulateApp(t)
	appState := models.NewAppState(nil)               // nil vault for testing
	lm := layout.NewLayoutManager(app, appState, nil) // nil config for testing

	// Initial state should be Auto (nil)
	if lm.GetSidebarOverride() != nil {
		t.Errorf("Expected initial sidebarOverride to be nil (Auto), got %v", lm.GetSidebarOverride())
	}

	// First toggle: Auto → Hide
	msg := lm.ToggleSidebar()
	override := lm.GetSidebarOverride()
	if override == nil {
		t.Error("Expected sidebarOverride to be non-nil after first toggle")
	} else if *override != false {
		t.Errorf("Expected sidebarOverride to be false (Hide) after first toggle, got %v", *override)
	}
	if msg != "Sidebar: Hidden" {
		t.Errorf("Expected message 'Sidebar: Hidden', got %q", msg)
	}

	// Second toggle: Hide → Show
	msg = lm.ToggleSidebar()
	override = lm.GetSidebarOverride()
	if override == nil {
		t.Error("Expected sidebarOverride to be non-nil after second toggle")
	} else if *override != true {
		t.Errorf("Expected sidebarOverride to be true (Show) after second toggle, got %v", *override)
	}
	if msg != "Sidebar: Visible" {
		t.Errorf("Expected message 'Sidebar: Visible', got %q", msg)
	}

	// Third toggle: Show → Auto
	msg = lm.ToggleSidebar()
	override = lm.GetSidebarOverride()
	if override != nil {
		t.Errorf("Expected sidebarOverride to be nil (Auto) after third toggle, got %v", override)
	}
	if msg != "Sidebar: Auto (responsive)" {
		t.Errorf("Expected message 'Sidebar: Auto (responsive)', got %q", msg)
	}
}

// TestSidebarManualOverridePersistsAcrossResize verifies manual override persists during terminal resize
func TestSidebarManualOverridePersistsAcrossResize(t *testing.T) {
	app := SimulateApp(t)
	appState := models.NewAppState(nil)
	lm := layout.NewLayoutManager(app, appState, nil)

	// Create layout to initialize components
	lm.CreateMainLayout()

	// Set manual override to Hide
	lm.ToggleSidebar()
	if lm.GetSidebarOverride() == nil || *lm.GetSidebarOverride() != false {
		t.Fatal("Failed to set override to Hide")
	}

	// Simulate resize to large terminal (should normally show sidebar)
	lm.HandleResize(150, 40)

	// Manual override should persist
	override := lm.GetSidebarOverride()
	if override == nil {
		t.Error("Expected manual override to persist after resize, got nil")
	} else if *override != false {
		t.Errorf("Expected manual override to remain false after resize, got %v", *override)
	}
}

// TestShouldShowSidebarLogic verifies shouldShowSidebar() respects override precedence
func TestShouldShowSidebarLogic(t *testing.T) {
	tests := []struct {
		name             string
		width            int
		mediumBreakpoint int
		override         *bool
		expectedShow     bool
	}{
		{
			name:             "Auto mode - narrow terminal",
			width:            60,
			mediumBreakpoint: 80,
			override:         nil,
			expectedShow:     false,
		},
		{
			name:             "Auto mode - wide terminal",
			width:            100,
			mediumBreakpoint: 80,
			override:         nil,
			expectedShow:     true,
		},
		{
			name:             "Force hide - wide terminal",
			width:            150,
			mediumBreakpoint: 80,
			override:         boolPtr(false),
			expectedShow:     false,
		},
		{
			name:             "Force show - narrow terminal",
			width:            60,
			mediumBreakpoint: 80,
			override:         boolPtr(true),
			expectedShow:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := SimulateApp(t)
			appState := models.NewAppState(nil)
			lm := layout.NewLayoutManager(app, appState, nil)
			lm.SetBreakpoints(tt.mediumBreakpoint, 120)

			// Create layout
			lm.CreateMainLayout()

			// Set override if specified
			if tt.override != nil {
				lm.SetSidebarOverride(tt.override)
			}

			// Trigger resize to set width
			lm.HandleResize(tt.width, 40)

			// Check result
			actualShow := lm.ShouldShowSidebar()
			if actualShow != tt.expectedShow {
				t.Errorf("Expected shouldShowSidebar()=%v, got %v", tt.expectedShow, actualShow)
			}
		})
	}
}

// TestSidebarToggleStatusBarMessages verifies status bar messages for each toggle state
func TestSidebarToggleStatusBarMessages(t *testing.T) {
	app := SimulateApp(t)
	appState := models.NewAppState(nil)
	lm := layout.NewLayoutManager(app, appState, nil)

	tests := []struct {
		name            string
		expectedMessage string
	}{
		{"Auto → Hide", "Sidebar: Hidden"},
		{"Hide → Show", "Sidebar: Visible"},
		{"Show → Auto", "Sidebar: Auto (responsive)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := lm.ToggleSidebar()
			if msg != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, msg)
			}
		})
	}
}

// TestSidebarToggleWithActiveSelection verifies toggle works when table focused and sidebar has selection
func TestSidebarToggleWithActiveSelection(t *testing.T) {
	app := SimulateApp(t)
	appState := models.NewAppState(nil)
	lm := layout.NewLayoutManager(app, appState, nil)

	// Create layout
	lm.CreateMainLayout()

	// Simulate table focus and sidebar selection (just verify toggle doesn't panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ToggleSidebar panicked with table focus: %v", r)
		}
	}()

	// Toggle sidebar while table is conceptually focused
	msg := lm.ToggleSidebar()
	if msg == "" {
		t.Error("Expected non-empty message from ToggleSidebar")
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
