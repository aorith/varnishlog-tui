package ui

import (
	"github.com/aorith/varnishlog-tui/internal/ui/components/logview"
	"github.com/aorith/varnishlog-tui/internal/ui/components/queryeditor"
	"github.com/aorith/varnishlog-tui/internal/ui/components/queryloader"
	"github.com/aorith/varnishlog-tui/internal/ui/state"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type model struct {
	quitting        bool
	state           state.ModelState
	queryEditorView queryeditor.Model
	queryLoaderView queryloader.Model
	logView         logview.Model
}

func StartUI(configQueries *queryloader.QueriesConfig) {
	p := tea.NewProgram(
		NewModel(configQueries),
		tea.WithAltScreen(),
	)

	log.Debug("Starting ...")
	if _, err := p.Run(); err != nil {
		log.Fatal("Failed starting the TUI", err)
	}
}

func NewModel(configQueries *queryloader.QueriesConfig) model {
	return model{
		quitting:        false,
		state:           state.QueryEditorView,
		logView:         logview.New(),
		queryEditorView: queryeditor.New(),
		queryLoaderView: queryloader.New(configQueries),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Sequence(m.queryEditorView.Init(), m.queryLoaderView.Init(), m.logView.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.Cmd:
		panic("A tea.Cmd has been erroneously sent (tea.Cmd inside of a tea.Msg?).")
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Sequence(m.logView.CancelTxsFetchCmd(true), tea.Quit)
		}

		// Only process the keys for the current state
		switch m.state {
		case state.QueryEditorView:
			m.queryEditorView, cmd = m.queryEditorView.Update(msg)
			return m, cmd
		case state.QueryLoaderView:
			m.queryLoaderView, cmd = m.queryLoaderView.Update(msg)
			return m, cmd
		case state.LogView:
			m.logView, cmd = m.logView.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		// Update dimensions on all models
		m.logView, cmd = m.logView.Update(msg)
		cmds = append(cmds, cmd)
		m.queryEditorView, cmd = m.queryEditorView.Update(msg)
		cmds = append(cmds, cmd)
		m.queryLoaderView, cmd = m.queryLoaderView.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case state.ChangeModelStateMsg:
		if m.state == state.QueryEditorView && msg.State == state.LogView {
			m.state = msg.State
			switch data := msg.Data.(type) {
			case state.NewVarnishlogScriptMsg:
				m.logView.SetVarnishlogExecSettings(data)
			}
			return m, m.logView.FetchTxsCmd()
		}

		switch data := msg.Data.(type) {
		case state.NewQueryEditorScriptMsg:
			m.queryEditorView.SetScript(string(data))
		}

		m.state = msg.State
	default:
		// Everything else can go through the model update even if it's not the active one
		m.logView, cmd = m.logView.Update(msg)
		cmds = append(cmds, cmd)
		m.queryEditorView, cmd = m.queryEditorView.Update(msg)
		cmds = append(cmds, cmd)
		m.queryLoaderView, cmd = m.queryLoaderView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return "Bye!"
	}

	switch m.state {
	case state.LogView:
		return m.logView.View()
	case state.QueryEditorView:
		return m.queryEditorView.View()
	case state.QueryLoaderView:
		return m.queryLoaderView.View()
	default:
		return "..."
	}
}
