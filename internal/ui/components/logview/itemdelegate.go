// Custom itemDelegate: https://github.com/charmbracelet/bubbletea/blob/master/examples/list-simple/main.go#L29-L50
package logview

import (
	"fmt"
	"io"
	"strings"

	"github.com/aorith/varnishlog-tui/internal/tx"
	"github.com/aorith/varnishlog-tui/internal/ui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/truncate"
)

const (
	ellipsis = "…"
)

// itemDelegate is a standard delegate designed to work in lists.
//
// The description line can be hidden by setting Description to false, which
// renders the list as single-line-items. The spacing between items can be set
// with the SetSpacing method.
//
// Settings ShortHelpFunc and FullHelpFunc is optional. They can be set to
// include items in the list's default short and full help menus.
type itemDelegate struct {
	height  int
	spacing int
}

// newDelegate creates a new itemDelegate
func newDelegate() itemDelegate {
	return itemDelegate{
		height:  3,
		spacing: 1,
	}
}

// Height returns the delegate's preferred height.
func (d itemDelegate) Height() int {
	return d.height
}

// Spacing returns the delegate's spacing.
func (d itemDelegate) Spacing() int {
	return d.spacing
}

// Update
func (d itemDelegate) Update(msg tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render prints an item.
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, subtitle1, subtitle2 string
		parts                       []string
		matchedRunes                []int
	)

	i, ok := listItem.(tx.Tx)
	if !ok {
		return
	}

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Conditions
	var (
		isSelected  = index == m.Index()
		emptyFilter = m.SettingFilter() && m.FilterValue() == ""
		isFiltered  = m.SettingFilter() || m.IsFiltered()
	)

	if isFiltered && index < len(m.VisibleItems()) {
		// Get indices of matched characters
		matchedRunes = m.MatchesForItem(index)
	}

	list.NewDefaultDelegate()
	parts = strings.Split(i.AsItem(matchedRunes, true, isFiltered && !emptyFilter), "\n")
	title = parts[0]
	subtitle1 = parts[1]
	subtitle2 = parts[2]

	// Prevent text from exceeding list width
	textwidth := uint(m.Width() - styles.NormalItemStyle.GetPaddingLeft() - styles.NormalItemStyle.GetPaddingRight())
	title = truncate.StringWithTail(title, textwidth, ellipsis)
	subtitle1 = truncate.StringWithTail(subtitle1, textwidth, ellipsis)
	subtitle2 = truncate.StringWithTail(subtitle2, textwidth, ellipsis)

	if emptyFilter {
		title = styles.DimmedItemStyle.Width(m.Width()).Render(title)
		subtitle1 = styles.DimmedItemStyle.Width(m.Width()).Render(subtitle1)
		subtitle2 = styles.DimmedItemStyle.Width(m.Width()).Render(subtitle2)
	} else if isSelected && m.FilterState() != list.Filtering {
		title = styles.SelectedItemStyle.Width(m.Width()).Render(title)
		subtitle1 = styles.SelectedItemStyle.Width(m.Width()).Render(subtitle1)
		subtitle2 = styles.SelectedItemStyle.Width(m.Width()).Render(subtitle2)
	} else {
		title = styles.NormalItemStyle.Width(m.Width()).Render(title)
		subtitle1 = styles.NormalItemStyle.Width(m.Width()).Render(subtitle1)
		subtitle2 = styles.NormalItemStyle.Width(m.Width()).Render(subtitle2)
	}

	fmt.Fprintf(w, "%s\n%s\n%s", title, subtitle1, subtitle2)
}
