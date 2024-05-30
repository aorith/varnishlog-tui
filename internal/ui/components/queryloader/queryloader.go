package queryloader

import (
	"fmt"

	"github.com/aorith/varnishlog-tui/assets"
	"github.com/aorith/varnishlog-tui/internal/ui/state"
	"github.com/aorith/varnishlog-tui/internal/ui/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

type Model struct {
	list list.Model
}

func New(configQueries *QueriesConfig) Model {
	var builtInQueries QueriesConfig
	err := yaml.Unmarshal([]byte(assets.BuiltInQueries), &builtInQueries)
	if err != nil {
		panic(fmt.Sprintf("could not unmarshal YAML: %s", err.Error()))
	}
	if len(builtInQueries.Queries) <= 0 {
		panic("At least one built-in query is required in the file \"assets/queries/built-in.yaml\".")
	}

	var queries []list.Item
	if configQueries != nil {
		for _, q := range configQueries.Queries {
			queries = append(queries, q)
		}
	}

	for _, q := range builtInQueries.Queries {
		queries = append(queries, q)
	}

	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = styles.NormalItemStyle.Inherit(styles.TxidColorStyle)
	d.Styles.SelectedTitle = styles.SelectedItemStyle.Inherit(styles.TxidColorStyle)
	d.Styles.DimmedTitle = styles.DimmedItemStyle.Inherit(styles.TxidColorStyle)
	d.Styles.NormalDesc = styles.NormalItemStyle.Inherit(styles.HostMethodURLColorStyle)
	d.Styles.SelectedDesc = styles.SelectedItemStyle.Inherit(styles.HostMethodURLColorStyle)
	d.Styles.DimmedDesc = styles.DimmedItemStyle.Inherit(styles.HostMethodURLColorStyle)

	l := list.New(queries, d, 80, 60)
	l.FilterInput.Cursor.Style = styles.NoStyle
	l.SetShowTitle(true)
	l.Title = "Query Loader"
	l.Styles.Title = styles.TitleStyle
	l.SetStatusBarItemName("query", "queries")
	l.SetShowStatusBar(true)
	l.DisableQuitKeybindings()
	l.AdditionalFullHelpKeys = additionalFullHelpKeys
	l.AdditionalShortHelpKeys = additionalShortHelpKeys

	return Model{
		list: l,
	}
}

func (m Model) Init() tea.Cmd {
	// Initialize the model with the first query (explicitly loaded with -queries or built-in)
	firstQuery, ok := m.list.Items()[0].(Query)
	if ok {
		return func() tea.Msg {
			return state.ChangeModelState(state.QueryEditorView, firstQuery.newQueryEditorData())
		}
	}

	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't match any of the keys below if we're filtering
		if m.list.SettingFilter() {
			break
		}

		key := msg.String()
		switch key {
		case "f":
			return m, func() tea.Msg {
				return state.ChangeModelState(state.QueryEditorView, nil)
			}
		case "o":
			return m, func() tea.Msg {
				return state.ChangeModelState(state.LogView, nil)
			}
		case "enter":
			currQuery := m.getCurrentQuery()
			if currQuery == nil {
				break
			}

			return m, func() tea.Msg {
				return state.ChangeModelState(state.QueryEditorView, currQuery.newQueryEditorData())
			}
		}
	case tea.WindowSizeMsg:
		m.updateSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return styles.MainMarginStyle.Render(m.list.View())
}

func (m *Model) updateSize(width, height int) {
	h, v := styles.MainMarginStyle.GetFrameSize()
	m.list.SetSize(width-h, height-v)
}

func additionalFullHelpKeys() []key.Binding {
	var keys []key.Binding
	keys = append(keys,
		key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "query editor"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select query"),
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
			key.WithKeys("f"),
			key.WithHelp("f", "query editor"),
		),
		key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "txs view"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select query"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	)
	return keys
}

func (m *Model) getCurrentQuery() *Query {
	currQuery, ok := m.list.SelectedItem().(Query)
	if !ok {
		return nil
	}
	return &currQuery
}
