package layout

import (
	"testing"

	"github.com/arimxyer/pass-cli/cmd/tui/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestDetermineLayoutMode verifies breakpoint logic at critical boundaries.
func TestDetermineLayoutMode(t *testing.T) {
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
	}

	tests := []struct {
		name     string
		width    int
		expected LayoutMode
	}{
		// Small mode (< 80)
		{"Very narrow terminal", 40, LayoutSmall},
		{"Just below medium breakpoint", 79, LayoutSmall},

		// Medium mode (80-119)
		{"Exactly at medium breakpoint", 80, LayoutMedium},
		{"Middle of medium range", 100, LayoutMedium},
		{"Just below large breakpoint", 119, LayoutMedium},

		// Large mode (>= 120)
		{"Exactly at large breakpoint", 120, LayoutLarge},
		{"Wide terminal", 150, LayoutLarge},
		{"Very wide terminal", 200, LayoutLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lm.determineLayoutMode(tt.width)
			if result != tt.expected {
				t.Errorf("determineLayoutMode(%d) = %v, want %v", tt.width, result, tt.expected)
			}
		})
	}
}

// TestSetBreakpoints verifies custom breakpoint configuration.
func TestSetBreakpoints(t *testing.T) {
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
	}

	// Set custom breakpoints
	lm.SetBreakpoints(90, 140)

	if lm.mediumBreakpoint != 90 {
		t.Errorf("mediumBreakpoint = %d, want 90", lm.mediumBreakpoint)
	}
	if lm.largeBreakpoint != 140 {
		t.Errorf("largeBreakpoint = %d, want 140", lm.largeBreakpoint)
	}

	// Verify layout mode calculation uses new breakpoints
	if mode := lm.determineLayoutMode(85); mode != LayoutSmall {
		t.Errorf("With breakpoint 90, width 85 should be Small, got %v", mode)
	}
	if mode := lm.determineLayoutMode(95); mode != LayoutMedium {
		t.Errorf("With breakpoint 90-140, width 95 should be Medium, got %v", mode)
	}
	if mode := lm.determineLayoutMode(145); mode != LayoutLarge {
		t.Errorf("With breakpoint 140, width 145 should be Large, got %v", mode)
	}
}

// TestGetCurrentMode verifies mode tracking.
func TestGetCurrentMode(t *testing.T) {
	lm := &LayoutManager{
		currentMode: LayoutMedium,
	}

	if mode := lm.GetCurrentMode(); mode != LayoutMedium {
		t.Errorf("GetCurrentMode() = %v, want LayoutMedium", mode)
	}

	lm.currentMode = LayoutLarge
	if mode := lm.GetCurrentMode(); mode != LayoutLarge {
		t.Errorf("GetCurrentMode() = %v, want LayoutLarge", mode)
	}
}

// TestHandleResize verifies resize detection and mode changes.
func TestHandleResize(t *testing.T) {
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
	}

	// Note: contentRow is nil, but rebuildLayout() guards against this
	// In real usage, CreateMainLayout() initializes all components

	// Resize to medium mode
	lm.HandleResize(100, 40)

	if lm.width != 100 {
		t.Errorf("width = %d, want 100", lm.width)
	}
	if lm.height != 40 {
		t.Errorf("height = %d, want 40", lm.height)
	}
	if lm.currentMode != LayoutMedium {
		t.Errorf("currentMode = %v, want LayoutMedium", lm.currentMode)
	}

	// Resize to large mode
	lm.HandleResize(150, 50)

	if lm.width != 150 {
		t.Errorf("width = %d, want 150", lm.width)
	}
	if lm.height != 50 {
		t.Errorf("height = %d, want 50", lm.height)
	}
	if lm.currentMode != LayoutLarge {
		t.Errorf("currentMode = %v, want LayoutLarge", lm.currentMode)
	}

	// Resize within same mode (should not rebuild)
	previousMode := lm.currentMode
	lm.HandleResize(155, 50)

	if lm.currentMode != previousMode {
		t.Errorf("Mode should not change for resize within same range")
	}
}

// TestLayoutModeConstants verifies enum values are distinct.
func TestLayoutModeConstants(t *testing.T) {
	if LayoutSmall == LayoutMedium {
		t.Error("LayoutSmall and LayoutMedium should be distinct")
	}
	if LayoutMedium == LayoutLarge {
		t.Error("LayoutMedium and LayoutLarge should be distinct")
	}
	if LayoutSmall == LayoutLarge {
		t.Error("LayoutSmall and LayoutLarge should be distinct")
	}
}

// =============================================================================
// User Story 1 Tests: Terminal Size Warning Display
// =============================================================================

// mockPageManager is a test double for PageManager to verify ShowSizeWarning calls.
type mockPageManager struct {
	showSizeWarningCalled bool
	showSizeWarningArgs   struct {
		currentWidth  int
		currentHeight int
		minWidth      int
		minHeight     int
	}
	hideSizeWarningCalled bool
	sizeWarningActive     bool
}

func (m *mockPageManager) ShowSizeWarning(currentWidth, currentHeight, minWidth, minHeight int) {
	m.showSizeWarningCalled = true
	m.showSizeWarningArgs.currentWidth = currentWidth
	m.showSizeWarningArgs.currentHeight = currentHeight
	m.showSizeWarningArgs.minWidth = minWidth
	m.showSizeWarningArgs.minHeight = minHeight
	m.sizeWarningActive = true
}

func (m *mockPageManager) HideSizeWarning() {
	m.hideSizeWarningCalled = true
	m.sizeWarningActive = false
}

func (m *mockPageManager) IsSizeWarningActive() bool {
	return m.sizeWarningActive
}

// TestHandleResize_BelowMinimum verifies HandleResize calls ShowSizeWarning
// when width < 60 OR height < 30.
func TestHandleResize_BelowMinimum(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		height        int
		shouldTrigger bool
	}{
		{"Both dimensions below minimum", 50, 20, true},
		{"Width below minimum", 50, 40, true},
		{"Height below minimum", 80, 20, true},
		{"Both dimensions at minimum", 60, 30, false},
		{"Both dimensions above minimum", 80, 40, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := &mockPageManager{}
			lm := &LayoutManager{
				mediumBreakpoint: 80,
				largeBreakpoint:  120,
				currentMode:      LayoutSmall,
				pageManager:      mockPM,
			}

			lm.HandleResize(tt.width, tt.height)

			if tt.shouldTrigger && !mockPM.showSizeWarningCalled {
				t.Errorf("Expected ShowSizeWarning to be called for %dx%d", tt.width, tt.height)
			}
			if !tt.shouldTrigger && mockPM.showSizeWarningCalled {
				t.Errorf("Expected ShowSizeWarning NOT to be called for %dx%d", tt.width, tt.height)
			}

			// Verify correct arguments were passed
			if tt.shouldTrigger {
				if mockPM.showSizeWarningArgs.currentWidth != tt.width {
					t.Errorf("currentWidth = %d, want %d", mockPM.showSizeWarningArgs.currentWidth, tt.width)
				}
				if mockPM.showSizeWarningArgs.currentHeight != tt.height {
					t.Errorf("currentHeight = %d, want %d", mockPM.showSizeWarningArgs.currentHeight, tt.height)
				}
				if mockPM.showSizeWarningArgs.minWidth != MinTerminalWidth {
					t.Errorf("minWidth = %d, want %d", mockPM.showSizeWarningArgs.minWidth, MinTerminalWidth)
				}
				if mockPM.showSizeWarningArgs.minHeight != MinTerminalHeight {
					t.Errorf("minHeight = %d, want %d", mockPM.showSizeWarningArgs.minHeight, MinTerminalHeight)
				}
			}
		})
	}
}

// TestHandleResize_StartupCheck verifies startup size check triggers warning
// if terminal is already too small.
func TestHandleResize_StartupCheck(t *testing.T) {
	mockPM := &mockPageManager{}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
		pageManager:      mockPM,
	}

	// Simulate startup with small terminal
	lm.HandleResize(50, 20)

	// Verify warning was shown
	if !mockPM.showSizeWarningCalled {
		t.Error("Expected ShowSizeWarning to be called on startup with small terminal")
	}

	// Verify correct dimensions passed
	if mockPM.showSizeWarningArgs.currentWidth != 50 {
		t.Errorf("currentWidth = %d, want 50", mockPM.showSizeWarningArgs.currentWidth)
	}
	if mockPM.showSizeWarningArgs.currentHeight != 20 {
		t.Errorf("currentHeight = %d, want 20", mockPM.showSizeWarningArgs.currentHeight)
	}
}

// =============================================================================
// User Story 2 Tests: Automatic Recovery
// =============================================================================

// TestHideSizeWarning verifies HideSizeWarning removes the warning page
// and clears the state flag.
func TestHideSizeWarning(t *testing.T) {
	mockPM := &mockPageManager{sizeWarningActive: true}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
		pageManager:      mockPM,
	}

	// Resize to adequate size
	lm.HandleResize(80, 40)

	// Verify HideSizeWarning was called
	if !mockPM.hideSizeWarningCalled {
		t.Error("Expected HideSizeWarning to be called when resizing to adequate size")
	}

	// Verify state was cleared
	if mockPM.sizeWarningActive {
		t.Error("Expected sizeWarningActive to be false after HideSizeWarning")
	}
}

// TestHideSizeWarning_WhenNotActive verifies safe no-op when warning not showing.
func TestHideSizeWarning_WhenNotActive(t *testing.T) {
	mockPM := &mockPageManager{sizeWarningActive: false}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
		pageManager:      mockPM,
	}

	// Resize to adequate size when warning already hidden
	lm.HandleResize(80, 40)

	// Should still call HideSizeWarning (method handles idempotency)
	if !mockPM.hideSizeWarningCalled {
		t.Error("Expected HideSizeWarning to be called even when not active")
	}

	// State should remain false
	if mockPM.sizeWarningActive {
		t.Error("Expected sizeWarningActive to remain false")
	}
}

// TestHandleResize_ExactlyAtMinimum verifies 60×30 does NOT trigger warning
// (inclusive boundary).
func TestHandleResize_ExactlyAtMinimum(t *testing.T) {
	mockPM := &mockPageManager{}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
		pageManager:      mockPM,
	}

	// Resize to exactly minimum dimensions
	lm.HandleResize(60, 30)

	// Should NOT trigger warning
	if mockPM.showSizeWarningCalled {
		t.Error("Expected NO warning at exactly 60×30 (inclusive boundary)")
	}

	// Should call HideSizeWarning (to clear any existing warning)
	if !mockPM.hideSizeWarningCalled {
		t.Error("Expected HideSizeWarning to be called at adequate size")
	}
}

// TestHandleResize_PartialFailure verifies 70×25 triggers warning
// (height < 30, OR logic).
func TestHandleResize_PartialFailure(t *testing.T) {
	mockPM := &mockPageManager{}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
		pageManager:      mockPM,
	}

	// Width OK (70 >= 60), but height too small (25 < 30)
	lm.HandleResize(70, 25)

	// Should trigger warning (OR logic: width OK but height fails)
	if !mockPM.showSizeWarningCalled {
		t.Error("Expected warning to be shown for 70×25 (height < 30)")
	}

	// Verify correct dimensions passed
	if mockPM.showSizeWarningArgs.currentWidth != 70 {
		t.Errorf("currentWidth = %d, want 70", mockPM.showSizeWarningArgs.currentWidth)
	}
	if mockPM.showSizeWarningArgs.currentHeight != 25 {
		t.Errorf("currentHeight = %d, want 25", mockPM.showSizeWarningArgs.currentHeight)
	}
}

// TestFullResizeFlow_ShowAndHide verifies end-to-end resize flow:
// start 50×20 (warning shows), resize 80×40 (warning hides), interface functional.
func TestFullResizeFlow_ShowAndHide(t *testing.T) {
	mockPM := &mockPageManager{}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
		pageManager:      mockPM,
	}

	// Step 1: Startup with small terminal
	lm.HandleResize(50, 20)

	// Verify warning shown
	if !mockPM.showSizeWarningCalled {
		t.Error("Expected warning to be shown at 50×20")
	}
	if !mockPM.sizeWarningActive {
		t.Error("Expected sizeWarningActive=true after showing warning")
	}

	// Step 2: Resize to adequate size
	mockPM.hideSizeWarningCalled = false // Reset flag
	lm.HandleResize(80, 40)

	// Verify warning hidden
	if !mockPM.hideSizeWarningCalled {
		t.Error("Expected HideSizeWarning to be called at 80×40")
	}
	if mockPM.sizeWarningActive {
		t.Error("Expected sizeWarningActive=false after hiding warning")
	}

	// Step 3: Verify layout mode updated correctly (recovery functional)
	if lm.currentMode != LayoutMedium {
		t.Errorf("Expected LayoutMedium at width 80, got %v", lm.currentMode)
	}
	if lm.width != 80 {
		t.Errorf("Expected width=80, got %d", lm.width)
	}
	if lm.height != 40 {
		t.Errorf("Expected height=40, got %d", lm.height)
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

// TestHandleResize_RapidOscillation verifies no crashes or memory leaks
// during rapid resize oscillation across the minimum boundary.
func TestHandleResize_RapidOscillation(t *testing.T) {
	mockPM := &mockPageManager{}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutSmall,
		pageManager:      mockPM,
	}

	// Rapidly oscillate 100 times between small and adequate sizes
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			lm.HandleResize(50, 20) // Below minimum
		} else {
			lm.HandleResize(80, 40) // Adequate size
		}
	}

	// Verify no crashes occurred (test completes without panic)
	// Verify final state is correct (last iteration was 80×40)
	if lm.width != 80 {
		t.Errorf("Expected final width=80, got %d", lm.width)
	}
	if lm.height != 40 {
		t.Errorf("Expected final height=40, got %d", lm.height)
	}

	// Verify HideSizeWarning was called on final iteration
	if !mockPM.hideSizeWarningCalled {
		t.Error("Expected HideSizeWarning to be called after oscillation")
	}
}

// TestModalPreservation_DuringWarning verifies that when a modal is open
// and terminal resizes below minimum, the warning overlays on top while
// preserving the modal state.
func TestModalPreservation_DuringWarning(t *testing.T) {
	// Note: This is a structural test since we use mocks.
	// Real integration would require tview app running.

	mockPM := &mockPageManager{}
	lm := &LayoutManager{
		mediumBreakpoint: 80,
		largeBreakpoint:  120,
		currentMode:      LayoutMedium,
		pageManager:      mockPM,
	}

	// Simulate: user opens form modal (PageManager tracks this separately)
	// Then resize below minimum
	lm.HandleResize(50, 20)

	// Verify warning was shown (will overlay on top of modal)
	if !mockPM.showSizeWarningCalled {
		t.Error("Expected warning to show even when modal is open")
	}

	// The modal state is preserved because ShowSizeWarning uses AddPage
	// which adds to the page stack without removing existing pages.
	// This is verified by the implementation in pages.go:238 using AddPage.

	// Resize back to adequate size
	mockPM.hideSizeWarningCalled = false
	lm.HandleResize(80, 40)

	// Verify warning was hidden (modal should now be visible again)
	if !mockPM.hideSizeWarningCalled {
		t.Error("Expected warning to hide after resize to adequate size")
	}
}

// TestHandleResize_BoundaryEdgeCases verifies exact boundary conditions
// around the minimum size threshold.
func TestHandleResize_BoundaryEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		shouldWarn  bool
		description string
	}{
		{
			name:        "59×30 - width just below minimum",
			width:       59,
			height:      30,
			shouldWarn:  true,
			description: "Width 59 < 60, should warn",
		},
		{
			name:        "60×29 - height just below minimum",
			width:       60,
			height:      29,
			shouldWarn:  true,
			description: "Height 29 < 30, should warn",
		},
		{
			name:        "61×31 - both above minimum",
			width:       61,
			height:      31,
			shouldWarn:  false,
			description: "Both dimensions above minimum, no warning",
		},
		{
			name:        "59×29 - both below minimum",
			width:       59,
			height:      29,
			shouldWarn:  true,
			description: "Both dimensions below minimum, should warn",
		},
		{
			name:        "60×30 - exactly at minimum",
			width:       60,
			height:      30,
			shouldWarn:  false,
			description: "Exactly at minimum (inclusive), no warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := &mockPageManager{}
			lm := &LayoutManager{
				mediumBreakpoint: 80,
				largeBreakpoint:  120,
				currentMode:      LayoutSmall,
				pageManager:      mockPM,
			}

			lm.HandleResize(tt.width, tt.height)

			if tt.shouldWarn && !mockPM.showSizeWarningCalled {
				t.Errorf("%s: Expected warning to show but it didn't. %s", tt.name, tt.description)
			}
			if !tt.shouldWarn && mockPM.showSizeWarningCalled {
				t.Errorf("%s: Expected NO warning but it showed. %s", tt.name, tt.description)
			}
		})
	}
}

// =============================================================================
// Sidebar Toggle Tests (migrated from test/tui/layout_test.go)
// =============================================================================

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// simulateApp creates a minimal tview.Application for testing
func simulateApp(t *testing.T) *tview.Application {
	app := tview.NewApplication()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("Failed to initialize simulation screen: %v", err)
	}
	app.SetScreen(screen)
	return app
}

// TestSidebarToggleCycles verifies sidebar toggle cycles through nil → false → true → nil
func TestSidebarToggleCycles(t *testing.T) {
	app := simulateApp(t)
	appState := models.NewAppState(nil)        // nil vault for testing
	lm := NewLayoutManager(app, appState, nil) // nil config for testing

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
	app := simulateApp(t)
	appState := models.NewAppState(nil)
	lm := NewLayoutManager(app, appState, nil)

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
			app := simulateApp(t)
			appState := models.NewAppState(nil)
			lm := NewLayoutManager(app, appState, nil)
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
	app := simulateApp(t)
	appState := models.NewAppState(nil)
	lm := NewLayoutManager(app, appState, nil)

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
	app := simulateApp(t)
	appState := models.NewAppState(nil)
	lm := NewLayoutManager(app, appState, nil)

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
