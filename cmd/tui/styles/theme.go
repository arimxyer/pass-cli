package styles

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ColorScheme defines the color palette for the TUI.
type ColorScheme struct {
	// Background colors
	Background      tcell.Color // Main background (dark)
	BackgroundLight tcell.Color // Slightly lighter background
	BackgroundDark  tcell.Color // Darker background (for contrast)

	// Border colors
	BorderColor    tcell.Color // Default border color (cyan)
	BorderInactive tcell.Color // Inactive/unfocused borders (gray)

	// Text colors
	TextPrimary   tcell.Color // Main text (white)
	TextSecondary tcell.Color // Secondary text (gray)
	TextAccent    tcell.Color // Accent text (cyan/yellow)

	// Status colors
	Success tcell.Color // Success messages (green)
	Error   tcell.Color // Error messages (red)
	Warning tcell.Color // Warning messages (yellow)
	Info    tcell.Color // Info messages (cyan)

	// Component-specific
	TableHeader            tcell.Color // Table header text
	TableSelected          tcell.Color // Selected row highlight
	TableSelectedHighlight tcell.Color // Brighter selected row background
	SidebarSelected        tcell.Color // Selected tree node
	StatusBarBg            tcell.Color // Status bar background
	ButtonBackground       tcell.Color // Button background
	ButtonText             tcell.Color // Button text
}

// DraculaTheme provides a Dracula-inspired color scheme (default).
var DraculaTheme = ColorScheme{
	// Backgrounds
	Background:      tcell.NewRGBColor(40, 42, 54), // #282a36
	BackgroundLight: tcell.NewRGBColor(68, 71, 90), // #44475a
	BackgroundDark:  tcell.NewRGBColor(30, 32, 44), // #1e202c

	// Borders
	BorderColor:    tcell.NewRGBColor(139, 233, 253), // #8be9fd (cyan)
	BorderInactive: tcell.NewRGBColor(98, 114, 164),  // #6272a4 (gray)

	// Text
	TextPrimary:   tcell.NewRGBColor(248, 248, 242), // #f8f8f2 (white)
	TextSecondary: tcell.NewRGBColor(98, 114, 164),  // #6272a4 (gray)
	TextAccent:    tcell.NewRGBColor(241, 250, 140), // #f1fa8c (yellow)

	// Status
	Success: tcell.NewRGBColor(80, 250, 123),  // #50fa7b (green)
	Error:   tcell.NewRGBColor(255, 85, 85),   // #ff5555 (red)
	Warning: tcell.NewRGBColor(241, 250, 140), // #f1fa8c (yellow)
	Info:    tcell.NewRGBColor(139, 233, 253), // #8be9fd (cyan)

	// Components
	TableHeader:            tcell.NewRGBColor(189, 147, 249), // #bd93f9 (purple)
	TableSelected:          tcell.NewRGBColor(68, 71, 90),    // #44475a (lighter bg)
	TableSelectedHighlight: tcell.NewRGBColor(98, 114, 164),  // #6272a4 (brighter purple-gray)
	SidebarSelected:        tcell.NewRGBColor(255, 121, 198), // #ff79c6 (pink)
	StatusBarBg:            tcell.NewRGBColor(30, 32, 44),    // #1e202c (dark)
	ButtonBackground:       tcell.NewRGBColor(68, 71, 90),    // #44475a
	ButtonText:             tcell.NewRGBColor(248, 248, 242), // #f8f8f2
}

// NordTheme provides a Nord-inspired color scheme (cool blues and grays).
var NordTheme = ColorScheme{
	// Backgrounds
	Background:      tcell.NewRGBColor(46, 52, 64),   // #2e3440
	BackgroundLight: tcell.NewRGBColor(59, 66, 82),   // #3b4252
	BackgroundDark:  tcell.NewRGBColor(36, 42, 54),   // #242a36

	// Borders
	BorderColor:    tcell.NewRGBColor(136, 192, 208), // #88c0d0 (frost cyan)
	BorderInactive: tcell.NewRGBColor(76, 86, 106),   // #4c566a (polar night)

	// Text
	TextPrimary:   tcell.NewRGBColor(236, 239, 244), // #eceff4 (snow storm)
	TextSecondary: tcell.NewRGBColor(76, 86, 106),   // #4c566a (polar night)
	TextAccent:    tcell.NewRGBColor(235, 203, 139), // #ebcb8b (aurora yellow)

	// Status
	Success: tcell.NewRGBColor(163, 190, 140), // #a3be8c (aurora green)
	Error:   tcell.NewRGBColor(191, 97, 106),  // #bf616a (aurora red)
	Warning: tcell.NewRGBColor(235, 203, 139), // #ebcb8b (aurora yellow)
	Info:    tcell.NewRGBColor(136, 192, 208), // #88c0d0 (frost cyan)

	// Components
	TableHeader:            tcell.NewRGBColor(129, 161, 193), // #81a1c1 (frost blue)
	TableSelected:          tcell.NewRGBColor(59, 66, 82),    // #3b4252
	TableSelectedHighlight: tcell.NewRGBColor(76, 86, 106),   // #4c566a
	SidebarSelected:        tcell.NewRGBColor(180, 142, 173), // #b48ead (aurora purple)
	StatusBarBg:            tcell.NewRGBColor(36, 42, 54),    // #242a36
	ButtonBackground:       tcell.NewRGBColor(59, 66, 82),    // #3b4252
	ButtonText:             tcell.NewRGBColor(236, 239, 244), // #eceff4
}

// GruvboxTheme provides a Gruvbox dark color scheme (warm retro vibes).
var GruvboxTheme = ColorScheme{
	// Backgrounds
	Background:      tcell.NewRGBColor(40, 40, 40),  // #282828
	BackgroundLight: tcell.NewRGBColor(60, 56, 54),  // #3c3836
	BackgroundDark:  tcell.NewRGBColor(29, 32, 33),  // #1d2021

	// Borders
	BorderColor:    tcell.NewRGBColor(142, 192, 124), // #8ec07c (aqua)
	BorderInactive: tcell.NewRGBColor(146, 131, 116), // #928374 (gray)

	// Text
	TextPrimary:   tcell.NewRGBColor(235, 219, 178), // #ebdbb2 (fg)
	TextSecondary: tcell.NewRGBColor(146, 131, 116), // #928374 (gray)
	TextAccent:    tcell.NewRGBColor(250, 189, 47),  // #fabd2f (yellow)

	// Status
	Success: tcell.NewRGBColor(184, 187, 38),  // #b8bb26 (green)
	Error:   tcell.NewRGBColor(251, 73, 52),   // #fb4934 (red)
	Warning: tcell.NewRGBColor(250, 189, 47),  // #fabd2f (yellow)
	Info:    tcell.NewRGBColor(131, 165, 152), // #83a598 (blue)

	// Components
	TableHeader:            tcell.NewRGBColor(211, 134, 155), // #d3869b (purple)
	TableSelected:          tcell.NewRGBColor(60, 56, 54),    // #3c3836
	TableSelectedHighlight: tcell.NewRGBColor(80, 73, 69),    // #504945
	SidebarSelected:        tcell.NewRGBColor(254, 128, 25),  // #fe8019 (orange)
	StatusBarBg:            tcell.NewRGBColor(29, 32, 33),    // #1d2021
	ButtonBackground:       tcell.NewRGBColor(60, 56, 54),    // #3c3836
	ButtonText:             tcell.NewRGBColor(235, 219, 178), // #ebdbb2
}

// MonokaiTheme provides a Monokai-inspired color scheme (vibrant and colorful).
var MonokaiTheme = ColorScheme{
	// Backgrounds
	Background:      tcell.NewRGBColor(39, 40, 34),  // #272822
	BackgroundLight: tcell.NewRGBColor(49, 51, 45),  // #31332d
	BackgroundDark:  tcell.NewRGBColor(29, 30, 24),  // #1d1e18

	// Borders
	BorderColor:    tcell.NewRGBColor(102, 217, 239), // #66d9ef (cyan)
	BorderInactive: tcell.NewRGBColor(117, 113, 94),  // #75715e (gray)

	// Text
	TextPrimary:   tcell.NewRGBColor(248, 248, 242), // #f8f8f2 (white)
	TextSecondary: tcell.NewRGBColor(117, 113, 94),  // #75715e (gray)
	TextAccent:    tcell.NewRGBColor(230, 219, 116), // #e6db74 (yellow)

	// Status
	Success: tcell.NewRGBColor(166, 226, 46),  // #a6e22e (green)
	Error:   tcell.NewRGBColor(249, 38, 114),  // #f92672 (pink/red)
	Warning: tcell.NewRGBColor(230, 219, 116), // #e6db74 (yellow)
	Info:    tcell.NewRGBColor(102, 217, 239), // #66d9ef (cyan)

	// Components
	TableHeader:            tcell.NewRGBColor(174, 129, 255), // #ae81ff (purple)
	TableSelected:          tcell.NewRGBColor(49, 51, 45),    // #31332d
	TableSelectedHighlight: tcell.NewRGBColor(73, 72, 62),    // #49483e
	SidebarSelected:        tcell.NewRGBColor(253, 151, 31),  // #fd971f (orange)
	StatusBarBg:            tcell.NewRGBColor(29, 30, 24),    // #1d1e18
	ButtonBackground:       tcell.NewRGBColor(49, 51, 45),    // #31332d
	ButtonText:             tcell.NewRGBColor(248, 248, 242), // #f8f8f2
}

// currentTheme holds the active theme (defaults to Dracula)
var currentTheme = DraculaTheme

// GetCurrentTheme returns the currently active color scheme.
func GetCurrentTheme() ColorScheme {
	return currentTheme
}

// SetTheme sets the active theme by name.
// Valid names: "dracula", "nord", "gruvbox", "monokai"
// Returns error if theme name is invalid.
func SetTheme(name string) error {
	switch name {
	case "dracula":
		currentTheme = DraculaTheme
	case "nord":
		currentTheme = NordTheme
	case "gruvbox":
		currentTheme = GruvboxTheme
	case "monokai":
		currentTheme = MonokaiTheme
	default:
		return fmt.Errorf("unknown theme: %s (valid themes: dracula, nord, gruvbox, monokai)", name)
	}
	return nil
}

// GetAvailableThemes returns a list of all available theme names.
func GetAvailableThemes() []string {
	return []string{"dracula", "nord", "gruvbox", "monokai"}
}

// SetRoundedBorders configures tview to use rounded border characters.
func SetRoundedBorders() {
	tview.Borders.Horizontal = '─'
	tview.Borders.Vertical = '│'
	tview.Borders.TopLeft = '╭'
	tview.Borders.TopRight = '╮'
	tview.Borders.BottomLeft = '╰'
	tview.Borders.BottomRight = '╯'
}

// ApplyBorderedStyle applies consistent border styling to a component.
// Uses type switch to handle all tview primitive types.
func ApplyBorderedStyle(p tview.Primitive, title string, active bool) {
	theme := GetCurrentTheme()
	borderColor := theme.BorderInactive
	if active {
		borderColor = theme.BorderColor
	}

	switch v := p.(type) {
	case *tview.Box:
		v.SetBorder(true).
			SetTitle(" " + title + " ").
			SetTitleAlign(tview.AlignLeft).
			SetBorderColor(borderColor).
			SetBackgroundColor(theme.Background)

	case *tview.Table:
		v.SetBorder(true).
			SetTitle(" " + title + " ").
			SetTitleAlign(tview.AlignLeft).
			SetBorderColor(borderColor).
			SetBackgroundColor(theme.Background)

	case *tview.TreeView:
		v.SetBorder(true).
			SetTitle(" " + title + " ").
			SetTitleAlign(tview.AlignLeft).
			SetBorderColor(borderColor).
			SetBackgroundColor(theme.Background)

	case *tview.TextView:
		v.SetBorder(true).
			SetTitle(" " + title + " ").
			SetTitleAlign(tview.AlignLeft).
			SetBorderColor(borderColor).
			SetBackgroundColor(theme.Background)

	case *tview.Form:
		v.SetBorder(true).
			SetTitle(" " + title + " ").
			SetTitleAlign(tview.AlignLeft).
			SetBorderColor(borderColor).
			SetBackgroundColor(theme.Background)

	case *tview.Modal:
		v.SetBorder(true).
			SetTitle(" " + title + " ").
			SetTitleAlign(tview.AlignLeft).
			SetBorderColor(borderColor).
			SetBackgroundColor(theme.Background)

	case *tview.List:
		v.SetBorder(true).
			SetTitle(" " + title + " ").
			SetTitleAlign(tview.AlignLeft).
			SetBorderColor(borderColor).
			SetBackgroundColor(theme.Background)
	}
}

// ApplyTableStyle applies consistent styling to table components.
func ApplyTableStyle(table *tview.Table) {
	theme := GetCurrentTheme()

	table.SetBackgroundColor(theme.Background)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.TableSelectedHighlight).
		Foreground(theme.TextPrimary).
		Bold(true).
		Underline(true))
}

// ApplyFormStyle applies consistent styling to form components.
func ApplyFormStyle(form *tview.Form) {
	theme := GetCurrentTheme()

	// Use Background (darker) for form container, BackgroundLight for fields
	// This creates visible input boxes that stand out from the form background
	form.SetBackgroundColor(theme.Background)
	form.SetButtonBackgroundColor(theme.ButtonBackground)
	form.SetButtonTextColor(theme.ButtonText)
	form.SetLabelColor(theme.TextPrimary) // Use primary text for labels (better contrast)

	// Set field style using tcell.Style for complete background coverage
	form.SetFieldStyle(tcell.StyleDefault.
		Background(theme.BackgroundLight).
		Foreground(theme.TextPrimary))

	// Also set individual colors for compatibility
	form.SetFieldBackgroundColor(theme.BackgroundLight)
	form.SetFieldTextColor(theme.TextPrimary)
}

// Lighten makes a color lighter by the given percentage.
func Lighten(color tcell.Color, amount float64) tcell.Color {
	r, g, b := color.RGB()
	r = clampUint8(float64(r) * (1.0 + amount))
	g = clampUint8(float64(g) * (1.0 + amount))
	b = clampUint8(float64(b) * (1.0 + amount))
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// Darken makes a color darker by the given percentage.
func Darken(color tcell.Color, amount float64) tcell.Color {
	r, g, b := color.RGB()
	r = clampUint8(float64(r) * (1.0 - amount))
	g = clampUint8(float64(g) * (1.0 - amount))
	b = clampUint8(float64(b) * (1.0 - amount))
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// clampUint8 clamps a float64 value to the uint8 range [0, 255].
func clampUint8(v float64) int32 {
	if v > 255 {
		return 255
	}
	if v < 0 {
		return 0
	}
	return int32(v)
}
