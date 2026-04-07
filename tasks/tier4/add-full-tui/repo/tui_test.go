package tui

import "testing"

func TestNewApp(t *testing.T) {
	list := NewListSelector([]string{"a", "b", "c"})
	app := NewApp(list)
	if app.FocusedComponent() != list {
		t.Fatal("first component should be focused")
	}
}
