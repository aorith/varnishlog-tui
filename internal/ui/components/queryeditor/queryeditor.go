package queryeditor

import (
	"fmt"
	"strings"

	"github.com/aorith/varnishlog-tui/internal/ui/components/queryloader"
	"github.com/aorith/varnishlog-tui/internal/ui/state"
	"github.com/aorith/varnishlog-tui/internal/ui/styles"
	"github.com/aorith/varnishlog-tui/internal/util"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	script string
	help   help.Model
	width  int
	height int
}

func New() Model {
	m := Model{
		help:   help.New(),
		script: "",
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keys.Confirm) {
			return m, submitVarnishlogQuery(state.NewVarnishlogScriptMsg(util.ParseVarnishlogArgs(m.script)))
		}
		if key.Matches(msg, keys.QueryLoader) {
			return m, func() tea.Msg {
				return state.ChangeModelState(state.QueryLoaderView, nil)
			}
		}
		if key.Matches(msg, keys.Edit) {
			return m, util.OpenEditor(strings.Split(m.script, "\n"), true, "sh")
		}
		if key.Matches(msg, keys.LogView) {
			return m, func() tea.Msg {
				return state.ChangeModelState(state.LogView, nil)
			}
		}
		if key.Matches(msg, keys.QueryToToml) {
			return m, util.OpenEditor(
				queryloader.QueryToYamlLines(queryloader.Query{
					Name:   "New Query",
					Script: m.script,
				}),
				false,
				"yaml",
			)
		}
	case util.EditorFinishedMsg:
		if msg.Err != nil {
			m.script = styles.ErrorStyle.Render(msg.Err.Error())
		} else if msg.Content != "" {
			m.script = msg.Content
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width - 4
		m.help.Width = m.width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	for _, line := range strings.Split(m.script, "\n") {
		if strings.HasPrefix(line, "#") {
			s.WriteString(styles.LabelStyle.Width(m.width-6).Render(line) + "\n")
		} else {
			s.WriteString(line + "\n")
		}
	}

	availableHeight := m.height
	head := fmt.Sprintf("%s\n",
		styles.TitleStyle.Render("Varnishlog Query Editor"),
	)
	availableHeight -= lipgloss.Height(head)

	tail := fmt.Sprintf("\n%s",
		m.help.FullHelpView(keys.FullHelp()),
	)
	availableHeight -= lipgloss.Height(tail)

	return styles.QueryEditorMarginStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			head,
			styles.QueryEditorScriptStyle.Width(m.width-4).Height(availableHeight-4).MaxHeight(availableHeight-2).Render(s.String()),
			tail,
		))
}

func submitVarnishlogQuery(e state.NewVarnishlogScriptMsg) tea.Cmd {
	return func() tea.Msg {
		return state.ChangeModelState(state.LogView, e)
	}
}

func (m *Model) SetScript(script string) {
	m.script = script
}
