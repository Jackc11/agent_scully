package tui

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type setScrollTickMsg struct{}

func selScrollTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
		return setScrollTickMsg{}
	})
}

func (m *Model) scrollToMouseY(y int) {
	height := m.viewport.Height()
	if height <= 2 {
		return
	}
	// Clamp y to the track area between ▲ (0) and ▼ (height-1)
	if y < 1 {
		y = 1
	}
	if y > height-2 {
		y = height - 2
	}

	trackHeight := height - 2
	maxOffset := m.viewport.TotalLineCount() - height
	if maxOffset <= 0 {
		return
	}

	if trackHeight == 1 {
		m.viewport.SetYOffset(0)
		return
	}

	ratio := float64(y-1) / float64(trackHeight-1)
	m.viewport.SetYOffset(int(math.Round(ratio * float64(maxOffset))))
}

func (m Model) normalized() (pos, pos, bool) {
	if m.start == m.end {
		return pos{}, pos{}, false
	}
	s, e := m.start, m.end
	if s.line > e.line || (s.line == e.line && s.col > e.col) {
		s, e = e, s
	}
	return s, e, true
}

func (m Model) renderScrollbar() string {
	height := m.viewport.Height()
	if height <= 0 {
		return ""
	}

	arrowStyle := lipgloss.NewStyle().Foreground(lineBarColorLight).Bold(true)
	trackStyle := lipgloss.NewStyle().Foreground(baseBgColorLight)
	thumbStyle := lipgloss.NewStyle().Foreground(lineBarColorLight)

	if height == 1 {
		return thumbStyle.Render("█")
	}
	if height == 2 {
		return arrowStyle.Render("▲") + "\n" + arrowStyle.Render("▼")
	}

	trackHeight := height - 2
	totalLines := m.viewport.TotalLineCount()

	// Calculate thumb size and position
	thumbHeight := trackHeight
	thumbPos := 0
	if totalLines > height {
		thumbHeight = int(math.Round(float64(height) / float64(totalLines) * float64(trackHeight)))
		if thumbHeight < 1 {
			thumbHeight = 1
		}
		if thumbHeight > trackHeight {
			thumbHeight = trackHeight
		}
		scrollableTrack := trackHeight - thumbHeight
		thumbPos = int(math.Round(m.viewport.ScrollPercent() * float64(scrollableTrack)))
	}

	lines := make([]string, height)
	lines[0] = arrowStyle.Render("▲")
	for i := 0; i < trackHeight; i++ {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			lines[i+1] = thumbStyle.Render("█")
		} else {
			lines[i+1] = trackStyle.Render("│")
		}
	}
	lines[height-1] = arrowStyle.Render("▼")

	return strings.Join(lines, "\n")
}

func (m Model) handleScrollTickMsg() (tea.Model, tea.Cmd) {
	if !m.selecting {
		return m, nil
	}

	if m.lastMouseY <= 0 {
		m.viewport.ScrollUp(1)
		x := max(0, min(m.lastMouseX, m.viewport.Width()-1))
		m.end = pos{
			line: m.viewport.YOffset(),
			col:  x,
		}
	} else if m.lastMouseY >= m.viewport.Height()-1 {
		m.viewport.ScrollDown(1)
		y := m.viewport.Height() - 1
		x := max(0, min(m.lastMouseX, m.viewport.Width()-1))
		m.end = pos{
			line: m.viewport.YOffset() + y,
			col:  x,
		}
	}

	return m, selScrollTick()
}
