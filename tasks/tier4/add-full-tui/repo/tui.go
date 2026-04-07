package tui

// KeyType identifies the type of a key event.
type KeyType int

const (
	KeyRune KeyType = iota
	KeyUp
	KeyDown
	KeyEnter
	KeyTab
	KeyBackspace
	KeyEscape
)

// KeyEvent represents a keyboard event.
type KeyEvent struct {
	Type KeyType
	Rune rune // only valid when Type == KeyRune
}

// Component is a renderable, interactive UI element.
type Component interface {
	Render() string
	HandleKey(key KeyEvent) bool
	Focused() bool
	SetFocused(focused bool)
}

// App is the top-level TUI application.
// TODO: Implement component management, focus cycling, rendering.
type App struct {
	components []Component
	focusIndex int
}

// NewApp creates a new TUI application with the given components.
func NewApp(components ...Component) *App {
	a := &App{
		components: components,
	}
	if len(components) > 0 {
		components[0].SetFocused(true)
	}
	return a
}

// Render returns the full ANSI-escaped string for the current UI state.
// TODO: Implement — compose all components with ANSI escape sequences.
func (a *App) Render() string {
	return ""
}

// HandleKey processes a key event.
// Tab cycles focus between components.
// Other keys go to the focused component.
// Returns true if the key was handled.
// TODO: Implement.
func (a *App) HandleKey(key KeyEvent) bool {
	return false
}

// FocusedComponent returns the currently focused component.
func (a *App) FocusedComponent() Component {
	if len(a.components) == 0 {
		return nil
	}
	return a.components[a.focusIndex]
}

// ANSI escape helpers — these should be used by components.
// TODO: Implement.

func ClearScreen() string   { return "" }
func MoveCursor(row, col int) string { return "" }
func Bold(s string) string  { return s }
func Reverse(s string) string { return s }
func Color(s string, fg int) string { return s }
func Reset() string         { return "" }
