package state

import tea "github.com/charmbracelet/bubbletea"

// ModelState tracks which model is focused
type ModelState uint

const (
	QueryEditorView ModelState = iota
	QueryLoaderView
	LogView
)

type ChangeModelStateMsg struct {
	State ModelState
	Data  interface{}
}

func ChangeModelState(modelState ModelState, data interface{}) tea.Msg {
	return ChangeModelStateMsg{
		State: modelState,
		Data:  data,
	}
}

// NewVarnishlogScriptMsg sets the is script content to be executed.
type NewVarnishlogScriptMsg string

// NewQueryEditorScriptMsg sets the script content in the query editor.
type NewQueryEditorScriptMsg string
