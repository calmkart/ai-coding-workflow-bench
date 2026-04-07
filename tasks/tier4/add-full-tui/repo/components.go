package tui

// ListSelector is a navigable list of items.
// TODO: Implement Render, HandleKey, selection.
type ListSelector struct {
	Items    []string
	Selected int
	focused  bool
	// OnSelect is called when Enter is pressed on an item.
	OnSelect func(index int, item string)
}

func NewListSelector(items []string) *ListSelector {
	return &ListSelector{Items: items}
}

func (l *ListSelector) Render() string     { return "" }
func (l *ListSelector) HandleKey(key KeyEvent) bool { return false }
func (l *ListSelector) Focused() bool      { return l.focused }
func (l *ListSelector) SetFocused(f bool)  { l.focused = f }

// TextInput is a single-line text input field.
// TODO: Implement Render, HandleKey, cursor, text editing.
type TextInput struct {
	Prompt string
	Value  string
	Cursor int
	focused bool
	// OnSubmit is called when Enter is pressed.
	OnSubmit func(value string)
}

func NewTextInput(prompt string) *TextInput {
	return &TextInput{Prompt: prompt}
}

func (t *TextInput) Render() string     { return "" }
func (t *TextInput) HandleKey(key KeyEvent) bool { return false }
func (t *TextInput) Focused() bool      { return t.focused }
func (t *TextInput) SetFocused(f bool)  { t.focused = f }

// StatusBar displays a status message at the bottom.
// TODO: Implement Render.
type StatusBar struct {
	Message string
	focused bool
}

func NewStatusBar(message string) *StatusBar {
	return &StatusBar{Message: message}
}

func (s *StatusBar) Render() string     { return "" }
func (s *StatusBar) HandleKey(key KeyEvent) bool { return false }
func (s *StatusBar) Focused() bool      { return s.focused }
func (s *StatusBar) SetFocused(f bool)  { s.focused = f }
