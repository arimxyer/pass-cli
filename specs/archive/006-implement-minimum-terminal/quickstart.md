# Quickstart: Minimum Terminal Size Enforcement

**Feature**: 006-implement-minimum-terminal
**Developer Quick Reference**

## TL;DR

```go
// Location: cmd/tui/layout/manager.go
const (
    MinTerminalWidth  = 60  // Columns
    MinTerminalHeight = 30  // Rows
)

// HandleResize checks size and shows/hides warning
func (lm *LayoutManager) HandleResize(width, height int) {
    if width < MinTerminalWidth || height < MinTerminalHeight {
        lm.pageManager.ShowSizeWarning(width, height, MinTerminalWidth, MinTerminalHeight)
    } else {
        lm.pageManager.HideSizeWarning()
    }
    // ... rest of resize logic
}
```

```go
// Location: cmd/tui/layout/pages.go
func (pm *PageManager) ShowSizeWarning(currentWidth, currentHeight, minWidth, minHeight int) {
    modal := tview.NewModal().SetText(fmt.Sprintf(...)).SetBackgroundColor(tcell.ColorDarkRed)
    pm.AddPage("size-warning", modal, true, true)
    pm.sizeWarningActive = true
    pm.app.Draw()
}

func (pm *PageManager) HideSizeWarning() {
    if !pm.sizeWarningActive { return }
    pm.RemovePage("size-warning")
    pm.sizeWarningActive = false
    pm.app.Draw()
}
```

---

## Testing Locally

### 1. Test Warning Display

**Resize below minimum** (should show warning):
```bash
# On Unix/Mac - resize terminal to 50x20
resize -s 20 50
go run main.go tui

# On Windows PowerShell
$Host.UI.RawUI.WindowSize = New-Object Management.Automation.Host.Size(50, 20)
go run main.go tui
```

**Expected**: Dark red warning overlay with current (50×20) vs minimum (60×30) dimensions

### 2. Test Boundary Conditions

**Exactly at minimum** (60×30 - should NOT warn):
```bash
resize -s 30 60  # Unix/Mac
go run main.go tui
```

**Expected**: Normal interface, no warning

**One dimension failing** (70×25 - should warn):
```bash
resize -s 25 70  # Width OK, height too small
go run main.go tui
```

**Expected**: Warning shows (height < 30)

### 3. Test Automatic Recovery

1. Start in normal size (80×40)
2. Resize below minimum (50×20) → warning appears
3. Resize back to normal (80×40) → warning disappears

**Expected**: Smooth transition, no lag, interface functional after recovery

### 4. Test Modal Preservation

1. Open credential form (press 'a' to add)
2. Resize below minimum
3. Warning should overlay form
4. Resize back to normal
5. Form should still be visible and functional

**Expected**: Form data preserved, focus restored to form after warning dismissed

---

## Running Tests

### Unit Tests
```bash
# Test PageManager size warning methods
go test -v ./cmd/tui/layout -run TestShowSizeWarning
go test -v ./cmd/tui/layout -run TestHideSizeWarning
go test -v ./cmd/tui/layout -run TestSizeWarningStateTracking

# Test LayoutManager resize handling
go test -v ./cmd/tui/layout -run TestHandleResize_BelowMinimum
go test -v ./cmd/tui/layout -run TestHandleResize_ExactlyAtMinimum
go test -v ./cmd/tui/layout -run TestHandleResize_PartialFailure
go test -v ./cmd/tui/layout -run TestHandleResize_RapidOscillation
```

### Integration Tests
```bash
# End-to-end resize scenarios
go test -v ./tests/integration -run TestFullResizeFlow
go test -v ./tests/integration -run TestModalPreservation
go test -v ./tests/integration -run TestStartupWithSmallTerminal
```

### All Layout Tests
```bash
go test -v ./cmd/tui/layout
```

---

## Common Pitfalls

### ❌ Pitfall 1: Early Return in HandleResize

**Wrong**:
```go
func (lm *LayoutManager) HandleResize(width, height int) {
    if width < MinTerminalWidth || height < MinTerminalHeight {
        lm.pageManager.ShowSizeWarning(...)
        return  // ❌ DON'T RETURN HERE
    }
    // Layout update code never reached during recovery
}
```

**Right**:
```go
func (lm *LayoutManager) HandleResize(width, height int) {
    if width < MinTerminalWidth || height < MinTerminalHeight {
        lm.pageManager.ShowSizeWarning(...)
        // ✅ Continue to layout update
    } else {
        lm.pageManager.HideSizeWarning()
    }
    // Layout mode determination continues...
}
```

**Why**: Layout must continue updating even when warning is shown. Otherwise, recovery from small size is broken (layout mode never changes).

---

### ❌ Pitfall 2: Forgetting app.Draw() After Page Changes

**Wrong**:
```go
func (pm *PageManager) ShowSizeWarning(...) {
    pm.AddPage("size-warning", modal, true, true)
    pm.sizeWarningActive = true
    // ❌ No explicit Draw() - relies on next render cycle
}
```

**Right**:
```go
func (pm *PageManager) ShowSizeWarning(...) {
    pm.AddPage("size-warning", modal, true, true)
    pm.sizeWarningActive = true
    pm.app.Draw()  // ✅ Force immediate redraw
}
```

**Why**: Without explicit `Draw()`, warning may not appear until next automatic render cycle (could be delayed). Explicit `Draw()` ensures <100ms response time.

---

### ❌ Pitfall 3: Using AND Logic Instead of OR

**Wrong**:
```go
if width < MinTerminalWidth && height < MinTerminalHeight {
    // ❌ Warning only shows when BOTH dimensions fail
}
```

**Right**:
```go
if width < MinTerminalWidth || height < MinTerminalHeight {
    // ✅ Warning shows if EITHER dimension fails
}
```

**Why**: Per spec clarification, warning should trigger if width < 60 OR height < 30. Terminal with 70×25 should warn (height too small).

---

### ❌ Pitfall 4: Inclusive Boundary Check

**Wrong**:
```go
if width <= MinTerminalWidth || height <= MinTerminalHeight {
    // ❌ Warning shows at exactly 60×30
}
```

**Right**:
```go
if width < MinTerminalWidth || height < MinTerminalHeight {
    // ✅ Warning shows only when strictly less than minimum
}
```

**Why**: Per spec clarification, 60×30 is acceptable (inclusive boundary). Warning only for < 60 or < 30.

---

## Debugging Tips

### Enable Verbose Resize Logging

Add temporary debug logging in `HandleResize`:

```go
func (lm *LayoutManager) HandleResize(width, height int) {
    log.Printf("[DEBUG] Resize event: %dx%d (min: %dx%d)",
        width, height, MinTerminalWidth, MinTerminalHeight)

    // ... rest of implementation
}
```

### Check Warning State

Query PageManager state in debugger:

```go
// In test or interactive debugging:
if pm.IsSizeWarningActive() {
    t.Logf("Warning is currently active")
} else {
    t.Logf("Warning is hidden")
}
```

### Verify Modal Z-Order

Check tview.Pages internal state:

```go
// In debugger, inspect pm.Pages to see page stack
// Last page in slice = top of z-order
```

### Simulate Rapid Resize in Tests

```go
func TestHandleResize_RapidOscillation(t *testing.T) {
    for i := 0; i < 100; i++ {
        if i%2 == 0 {
            lm.HandleResize(50, 20)  // Below minimum
        } else {
            lm.HandleResize(80, 40)  // Normal size
        }
    }
    // Should not crash or leak memory
}
```

---

## Performance Benchmarks

### Measure Resize Latency

```go
func BenchmarkHandleResize(b *testing.B) {
    lm := setupLayoutManager(t)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        lm.HandleResize(50, 20)
        lm.HandleResize(80, 40)
    }
}
```

**Expected**: <1ms per HandleResize call

### Measure Warning Display Latency

```go
func BenchmarkShowSizeWarning(b *testing.B) {
    pm := setupPageManager(t)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        pm.ShowSizeWarning(50, 20, 60, 30)
        pm.HideSizeWarning()
    }
}
```

**Expected**: <10ms per show/hide cycle (well below 100ms requirement)

---

## Adding Size-Dependent UI Elements

### Pattern: Check Size Before Layout

```go
func (lm *LayoutManager) RebuildLayout() {
    // Existing pattern: check width for sidebar/detail visibility
    if lm.width >= lm.mediumBreakpoint {
        // Show sidebar
    }

    // NEW: Can add height-dependent elements
    if lm.height >= 35 {
        // Show hints footer (requires extra 5 rows)
    }
}
```

### Pattern: Dynamic Content Based on Size

```go
func (pm *PageManager) ShowForm(form *tview.Form) {
    // Calculate max form height from terminal size
    _, termHeight := pm.app.GetScreen().Size()
    maxFormHeight := termHeight - 4  // Leave room for borders

    pm.ShowModal("form", form, FormModalWidth, maxFormHeight)
}
```

---

## File Locations Reference

| File | Purpose | Changes |
|------|---------|---------|
| `cmd/tui/layout/manager.go` | Layout management | Add constants, modify HandleResize |
| `cmd/tui/layout/pages.go` | Modal/page management | Add ShowSizeWarning/HideSizeWarning |
| `cmd/tui/layout/manager_test.go` | Layout tests | Add resize boundary tests |
| `cmd/tui/layout/pages_test.go` | Page tests | Add size warning tests |
| `tests/integration/tui_resize_test.go` | E2E tests | Add full resize flow tests |

---

## Questions?

- Check [research.md](./research.md) for technical decisions and rationale
- Check [plan.md](./plan.md) for architecture and design details
- Check [spec.md](./spec.md) for requirements and acceptance criteria
