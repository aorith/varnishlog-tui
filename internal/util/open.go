package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type EditorFinishedMsg struct {
	Err     error
	Content string
}

// OpenEditor opens the editor found in $EDITOR or a fallback editor
// with the contents of 'lines' in a temp file.
func OpenEditor(lines []string, returnBody bool, extension string) tea.Cmd {
	// Determine which editor to use
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
		if _, err := exec.LookPath("vim"); err != nil {
			editor = "nano"
			if _, err := exec.LookPath("nano"); err != nil {
				return func() tea.Msg {
					return EditorFinishedMsg{Err: fmt.Errorf("no suitable editor found")}
				}
			}
		}
	} else {
		if _, err := exec.LookPath(editor); err != nil {
			return func() tea.Msg {
				return EditorFinishedMsg{Err: fmt.Errorf(`exec: "%s": $EDITOR executable file not found in $PATH`, editor)}
			}
		}
	}

	return execTeaProcess(lines, returnBody, extension, editor)
}

// OpenInBrowserWithFallbackToEditor opens the browser using $BROWSER, xdg-open or open
// with the contents of 'lines' in a temp file. If xdg-open or open are not
// executable OpenEditor is executed instead.
func OpenInBrowserWithFallbackToEditor(lines []string) tea.Cmd {
	var open string = os.Getenv("BROWSER")
	if open == "" {
		open = "xdg-open"
	}

	if _, err := exec.LookPath(open); err != nil {
		open = "xdg-open"
		if _, err := exec.LookPath(open); err != nil {
			open = "open"
			if _, err := exec.LookPath(open); err != nil {
				// Neither $BROWSER, xdg-open or open found
				return OpenEditor(lines, false, "html")
			}
		}
	}

	return execTeaProcess(lines, false, "html", open)
}

// execTeaProcess is a helper function to save lines in a temporary file and open
// with a command using tea.execTeaProcess
func execTeaProcess(lines []string, returnBody bool, extension, command string) tea.Cmd {
	tempFile, err := os.CreateTemp("", "varnishlog-*."+extension)
	if err != nil {
		return func() tea.Msg {
			return EditorFinishedMsg{Err: fmt.Errorf("could not create temp file: %w", err)}
		}
	}
	defer tempFile.Close()

	content := strings.Join(lines, "\n")
	if _, err := tempFile.WriteString(content); err != nil {
		return func() tea.Msg {
			return EditorFinishedMsg{Err: fmt.Errorf("could not write to temp file: %w", err)}
		}
	}

	c := exec.Command(command, tempFile.Name()) //nolint:gosec
	c.Stderr = io.Discard
	fileName := tempFile.Name()
	return tea.ExecProcess(c, func(err error) tea.Msg {
		go func() {
			time.Sleep(time.Second * 10) // Give some time until the file is read
			os.Remove(fileName)
		}()

		// Outer error
		if err != nil {
			return EditorFinishedMsg{Err: err}
		}

		if returnBody {
			file, err := os.Open(fileName)
			if err != nil {
				return EditorFinishedMsg{Err: err}
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				return EditorFinishedMsg{Err: err}
			}

			return EditorFinishedMsg{Err: nil, Content: string(content)}
		}

		return EditorFinishedMsg{Err: nil}
	})
}
