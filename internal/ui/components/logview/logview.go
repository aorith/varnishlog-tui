package logview

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/aorith/varnishlog-tui/internal/tx"
	"github.com/aorith/varnishlog-tui/internal/ui/state"
	"github.com/aorith/varnishlog-tui/internal/ui/styles"
	"github.com/aorith/varnishlog-tui/internal/util"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

var frameHoriz, frameVert = styles.MainMarginStyle.GetFrameSize()

type initFetchTxsMsg struct{}

type cancelTxsFetchMsg struct {
	clear bool
}

type Model struct {
	list         list.Model
	execSettings state.NewVarnishlogScriptMsg
	txs          map[string]*tx.Tx
	fetching     bool
	cancelChan   chan struct{}
	txChan       chan tx.Tx
	err          error
}

func New() Model {
	s := spinner.Spinner{
		Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		FPS:    time.Second / 10,
	}

	p := paginator.New()
	p.Type = paginator.Arabic
	p.ArabicFormat = styles.PagerStyle.Render("page %d of %d")

	l := list.New([]list.Item{}, newDelegate(), 80, 60)
	l.FilterInput.Cursor.Style = styles.NoStyle
	l.Paginator = p
	l.SetSpinner(s)
	l.SetShowTitle(true)
	l.Title = "Transactions"
	l.Styles.Title = styles.TitleStyle
	l.SetStatusBarItemName("tx", "txs")
	l.SetShowStatusBar(true)
	l.StatusMessageLifetime = time.Second * 10
	l.DisableQuitKeybindings()
	l.AdditionalFullHelpKeys = additionalFullHelpKeys
	l.AdditionalShortHelpKeys = additionalShortHelpKeys

	return Model{
		list:     l,
		fetching: false,
		txs:      make(map[string]*tx.Tx),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Dismiss errors on any keypress
		if m.err != nil {
			m.err = nil
		}

		// Don't match any of the keys below if we're filtering
		if m.list.SettingFilter() {
			break
		}

		key := msg.String()
		switch key {
		case "q":
			return m, tea.Sequence(m.CancelTxsFetchCmd(false), switchToQueryEditorView())
		case "d":
			return m, tea.Sequence(m.CancelTxsFetchCmd(false), switchToQueryLoaderView())
		case "s":
			return m, m.CancelTxsFetchCmd(false)
		case "x":
			return m, tea.Sequence(m.CancelTxsFetchCmd(true), m.list.SetItems([]list.Item{}))
		case "r":
			return m, m.FetchTxsCmd()
		case "e":
			return m, m.openEditorForCurrentTxCmd()
		case "E":
			return m, m.openEditorForCurrentAndRelatedTxsCmd()
		case "ctrl+e":
			return m, util.OpenEditor(m.getAllRawTx(), false, "txt")
		case "enter":
			currTx := m.getCurrentTx()
			if currTx != nil {
				report, err := currTx.GenerateHtmlReport()
				if err != nil {
					return m, func() tea.Msg { return util.EditorFinishedMsg{Err: err} }
				} else {
					return m, util.OpenInBrowserWithFallbackToEditor(report)
				}
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width-frameHoriz, msg.Height-frameVert)
	case tx.NewTxMsg:
		if m.fetching {
			newTx := tx.Tx(msg)
			return m, tea.Batch(m.addNewTxCmd(newTx), tx.ListenForTxsCmd(m.txChan))
		}
	case tx.FetchEndMsg:
		m.err = msg.Err
		if m.err != nil {
			log.Debug(m.err.Error())
		}
		return m, m.CancelTxsFetchCmd(false)
	case cancelTxsFetchMsg:
		if m.cancelChan != nil {
			close(m.cancelChan)
		}
		m.fetching = false
		m.list.StopSpinner()
		if msg.clear {
			m.txs = make(map[string]*tx.Tx)
		}
		m.cancelChan = make(chan struct{}) // reset the cancel channel to avoid errors on repeated 'c' press
	case initFetchTxsMsg:
		if !m.fetching {
			m.fetching = true
			m.cancelChan = make(chan struct{})
			m.txChan = make(chan tx.Tx)
			return m, tea.Batch(
				m.list.StartSpinner(),
				tx.ListenForTxsCmd(m.txChan),
				tx.ExecVarnishlogAndFetchTxs(m.execSettings, m.cancelChan, m.txChan),
			)
		}
	case util.EditorFinishedMsg:
		m.err = msg.Err
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return styles.MainMarginStyle.
			Width(m.list.Width()-2).
			Render(
				"Error executing the command:\n\n"+
					styles.ErrorStyle.Width(m.list.Width()-4).Render(m.err.Error()),
				"\nPress any key to continue.\n",
			)
	}

	return styles.MainMarginStyle.Render(m.list.View())
}

func switchToQueryEditorView() tea.Cmd {
	return func() tea.Msg {
		return state.ChangeModelState(state.QueryEditorView, "")
	}
}

func switchToQueryLoaderView() tea.Cmd {
	return func() tea.Msg {
		return state.ChangeModelState(state.QueryLoaderView, "")
	}
}

func (m *Model) addNewTxCmd(newTx tx.Tx) tea.Cmd {
	m.txs[newTx.Txid] = &newTx

	// Update parent and children relationships
	for _, currTx := range m.txs {
		for childId := range currTx.Children {
			child, childExists := m.txs[childId]
			if childExists {
				child.Parent = currTx
				currTx.Children[childId] = child
			}
		}
	}

	// Extract and sort the keys
	keys := make([]string, 0, len(m.txs))
	for k := range m.txs {
		keys = append(keys, m.txs[k].Txid)
	}
	slices.Sort(keys)

	// Create a sorted slice of list.Item
	items := make([]list.Item, 0, len(m.txs))
	for _, k := range keys {
		items = append(items, *m.txs[k])
	}

	return m.list.SetItems(items)
}

func (m *Model) getCurrentTx() *tx.Tx {
	currTx, ok := m.list.SelectedItem().(tx.Tx)
	if !ok {
		return nil
	}
	return &currTx
}

// getVisibleTx gets all the visible txs
func (m *Model) getVisibleTx() []tx.Tx {
	var (
		currTx       tx.Tx
		ok           bool
		visibleItems = m.list.VisibleItems()
	)

	visibleTxs := make([]tx.Tx, 0, len(visibleItems))
	for _, i := range visibleItems {
		currTx, ok = i.(tx.Tx)
		if ok {
			visibleTxs = append(visibleTxs, currTx)
		}
	}
	return visibleTxs
}

// getAllRawTx gets the rawTxs of all the visible txs
func (m *Model) getAllRawTx() (rawTxs []string) {
	visibleTxs := m.getVisibleTx()

	keys := make([]string, 0, len(visibleTxs))
	for _, v := range visibleTxs {
		keys = append(keys, v.Txid)
	}
	slices.Sort(keys)

	for _, k := range keys {
		rawTxs = append(rawTxs, m.txs[k].RawTx...)
		rawTxs = append(rawTxs, "") // New line
	}
	return rawTxs
}

func (m *Model) openEditorForCurrentTxCmd() tea.Cmd {
	currTx := m.getCurrentTx()
	if currTx == nil {
		m.err = fmt.Errorf("could not open the current tx")
		return nil
	}

	var finalText []string
	finalText = append(finalText, "Tx Duration", "===========")
	finalText = append(finalText, strings.Split(currTx.GenerateTimestampHistogram(), "\n")...)

	finalText = append(finalText, "", "Raw log", "=======", "")
	finalText = append(finalText, currTx.RawTx...)
	finalText = append(finalText, "")

	return util.OpenEditor(finalText, false, "txt")
}

func (m *Model) openEditorForCurrentAndRelatedTxsCmd() tea.Cmd {
	currTx := m.getCurrentTx()
	if currTx == nil {
		m.err = fmt.Errorf("could not open the current tx")
		return nil
	}

	parent := currTx.FindRootParent()
	txs := []*tx.Tx{parent}
	txs = append(txs, parent.GetSortedChildren()...)

	tree := parent.PrintTree("", currTx.Txid, false)
	histogram := currTx.GenerateAllTxsHistogram(txs)

	var finalText []string
	finalText = append(finalText, "Txs Tree", "========", "")
	finalText = append(finalText, strings.Split(tree, "\n")...)
	finalText = append(finalText, "", "Txs Duration", "============")
	finalText = append(finalText, strings.Split(histogram, "\n")...)

	acctReceived := currTx.GenerateAccountingHistogram(txs, false)
	acctTransmitted := currTx.GenerateAccountingHistogram(txs, true)
	finalText = append(finalText, "", "Txs Received Accounting", "=======================")
	finalText = append(finalText, strings.Split(acctReceived, "\n")...)
	finalText = append(finalText, "", "Txs Transmitted Accounting", "==========================")
	finalText = append(finalText, strings.Split(acctTransmitted, "\n")...)

	finalText = append(finalText, "", "Raw log", "=======", "")
	for _, tx := range txs {
		finalText = append(finalText, tx.RawTx...)
		finalText = append(finalText, "")
	}

	return util.OpenEditor(finalText, false, "txt")
}

func (m *Model) SetVarnishlogExecSettings(execSettings state.NewVarnishlogScriptMsg) {
	m.execSettings = execSettings
}

func (m *Model) FetchTxsCmd() tea.Cmd {
	return func() tea.Msg {
		return initFetchTxsMsg{}
	}
}

func (m *Model) CancelTxsFetchCmd(clear bool) tea.Cmd {
	return func() tea.Msg {
		return cancelTxsFetchMsg{clear: clear}
	}
}
