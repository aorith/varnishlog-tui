package queryeditor

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Confirm     key.Binding
	QueryLoader key.Binding
	QueryToToml key.Binding
	LogView     key.Binding
	Edit        key.Binding
	Quit        key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.QueryLoader}, // first column
		{k.Edit, k.QueryToToml},    // second column
		{k.LogView, k.Quit},        // third column
	}
}

var keys = keyMap{
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "execute"),
	),
	QueryLoader: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "query loader"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit query in $EDITOR"),
	),
	QueryToToml: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "open query as YAML in $EDITOR"),
	),
	LogView: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "txs view"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}
