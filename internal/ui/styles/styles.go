package styles

import (
	"github.com/charmbracelet/lipgloss"
)

/* Colors */
var (
	BrightRedFGColor = lipgloss.AdaptiveColor{Light: "#730a05", Dark: "#ff7e8f"}
	PaleRedFGColor   = lipgloss.AdaptiveColor{Light: "#b74244", Dark: "#f38ba8"}
	YellowFGColor    = lipgloss.AdaptiveColor{Light: "#b58850", Dark: "#f9e2af"}
	BrownFGColor     = lipgloss.AdaptiveColor{Light: "#bf6f46", Dark: "#fab387"}
	BlueFGColor      = lipgloss.AdaptiveColor{Light: "#5c7abe", Dark: "#89b4fa"}
	OrangeFGColor    = lipgloss.AdaptiveColor{Light: "#761801", Dark: "#f78e2f"}
	GrayFGColor      = lipgloss.AdaptiveColor{Light: "#434964", Dark: "#6c7086"}
	DarkGrayFGColor  = lipgloss.AdaptiveColor{Light: "#1a1b1e", Dark: "#b4abac"}

	LightGrayBGColor = lipgloss.AdaptiveColor{Light: "#cacbce", Dark: "#4a4b4e"}
)

/* Styles */
var (
	NoStyle    = lipgloss.NewStyle()
	ErrorStyle = lipgloss.NewStyle().Foreground(BrightRedFGColor)
	TitleStyle = lipgloss.NewStyle().Foreground(BlueFGColor).Bold(true)
	LabelStyle = lipgloss.NewStyle().Foreground(DarkGrayFGColor)

	MainMarginStyle = lipgloss.NewStyle().Margin(1, 1)
	PagerStyle      = lipgloss.NewStyle().Foreground(GrayFGColor).Italic(true)

	RecordColorStyle        = lipgloss.NewStyle().Foreground(PaleRedFGColor)
	ReasonColorStyle        = lipgloss.NewStyle().Foreground(YellowFGColor)
	TxidColorStyle          = lipgloss.NewStyle().Foreground(BrownFGColor)
	TimestampsColorStyle    = lipgloss.NewStyle().Foreground(GrayFGColor)
	HostMethodURLColorStyle = lipgloss.NewStyle().Foreground(DarkGrayFGColor)

	RecordTypeStyle = lipgloss.NewStyle().Inline(true).Inherit(RecordColorStyle)
	ReasonStyle     = lipgloss.NewStyle().Inline(true).Inherit(ReasonColorStyle)
	TxidStyle       = lipgloss.NewStyle().Inline(true).Inherit(TxidColorStyle)
	HostStyle       = lipgloss.NewStyle().Inline(true).Inherit(HostMethodURLColorStyle)
	MethodStyle     = lipgloss.NewStyle().Inline(true).Inherit(HostMethodURLColorStyle)
	UrlStyle        = lipgloss.NewStyle().Inline(true).Inherit(HostMethodURLColorStyle)
	TsFlowStyle     = lipgloss.NewStyle().Inline(true).Inherit(TimestampsColorStyle)

	QueryEditorMarginStyle = lipgloss.NewStyle().Margin(1, 3)
	QueryEditorScriptStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).
				Padding(1).BorderForeground(LightGrayBGColor)

	NormalItemStyle   = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	SelectedItemStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder(), false, false, false, true).
				BorderForeground(OrangeFGColor).
				Padding(0, 0, 0, 1)
	DimmedItemStyle  = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	MatchedItemStyle = lipgloss.NewStyle().Background(LightGrayBGColor)
)
