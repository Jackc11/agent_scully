package tui

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
)

var (
	baseBgColorDark  = lipgloss.Color("#2d3139")
	baseBgColorLight = lipgloss.Color("#4c566a")
	baseFgColor      = lipgloss.Color("#e5e9f0")

	lineBarColorDark  = lipgloss.Color("#00bcd4")
	lineBarColorLight = lipgloss.Color("#88c0d0")

	senderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5"))

	textareaTextStyle = lipgloss.NewStyle().
				Background(baseBgColorDark).
				Foreground(lipgloss.Color("#ffffff"))
)

func applyTextareaStyles(ta *textarea.Model) {
	s := ta.Styles()

	s.Focused.Text = textareaTextStyle
	s.Focused.CursorLine = textareaTextStyle
	s.Focused.Prompt = textareaTextStyle.Foreground(lipgloss.Color("#00bcd4"))
	s.Focused.Placeholder = textareaTextStyle.Foreground(lipgloss.Color("#626875"))
	s.Focused.EndOfBuffer = lipgloss.NewStyle().
		Background(baseBgColorDark).
		Foreground(baseBgColorDark)

	ta.SetStyles(s)
}
