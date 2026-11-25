package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type thinkingTickMsg struct{}

func thinkingTick() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(time.Time) tea.Msg {
		return thinkingTickMsg{}
	})
}

func (m Model) handleThinkingTickMsg() (tea.Model, tea.Cmd) {
	if !m.loading {
		return m, nil // This is needed for stoping
	}
	m.thinkingDots = (m.thinkingDots + 1) % 4
	dots := strings.Repeat(".", m.thinkingDots)

	if len(m.messages) > 0 {
		m.messages[len(m.messages)-1] = m.agentStyle.Render("BlackZetsu: ") + "Thinking" + dots
		m.updateViewportContent()
	}
	return m, thinkingTick()
}
