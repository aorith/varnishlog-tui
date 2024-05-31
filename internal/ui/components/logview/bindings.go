package logview

import "github.com/charmbracelet/bubbles/key"

func additionalFullHelpKeys() []key.Binding {
	var keys []key.Binding
	keys = append(keys,
		key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "query editor"),
		),
		key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "query loader"),
		),
		key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "clear transactions"),
		),
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "run"),
		),
		key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stop"),
		),
		key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "open current tx in $EDITOR"),
		),
		key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "open related txs in $EDITOR"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "open all visible txs in $EDITOR"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open HTML report in $BROWSER or $EDITOR"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	)
	return keys
}

func additionalShortHelpKeys() []key.Binding {
	var keys []key.Binding
	keys = append(keys,
		key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stop"),
		),
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "run"),
		),
		key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "query editor"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	)
	return keys
}
