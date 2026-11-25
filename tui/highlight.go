package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) selectedText() string {
	s, e, ok := m.normalized()
	if !ok || s.line >= len(m.lines) {
		return ""
	}
	e.line = min(e.line, len(m.lines)-1)

	var out []string
	for i := s.line; i <= e.line; i++ {
		runes := []rune(ansi.Strip(m.lines[i]))
		cStart, cEnd := 0, len(runes)
		if i == s.line {
			cStart = max(0, min(s.col, len(runes)))
		}
		if i == e.line {
			cEnd = max(0, min(e.col+1, len(runes)))
		}
		if cStart < cEnd {
			out = append(out, string(runes[cStart:cEnd]))
		} else {
			out = append(out, "")
		}
	}
	return strings.Join(out, "\n")
}

func (m Model) renderHighlightedViewport() string {
	view := m.viewport.View()
	s, e, hasSel := m.normalized()
	if !hasSel {
		return view
	}

	hl := lipgloss.NewStyle().
		Background(lineBarColorLight).
		Foreground(baseBgColorDark)

	visibleLines := strings.Split(view, "\n")
	offset := m.viewport.YOffset()

	for i, line := range visibleLines {
		docLine := offset + i
		if docLine < s.line || docLine > e.line {
			continue
		}

		w := ansi.StringWidth(line)
		cStart, cEnd := 0, w
		if docLine == s.line {
			cStart = max(0, min(s.col, w))
		}
		if docLine == e.line {
			cEnd = max(0, min(e.col+1, w))
		}
		if cStart >= cEnd {
			continue
		}

		left := ansi.Truncate(line, cStart, "")
		mid := ansi.Strip(ansi.Cut(line, cStart, cEnd))
		right := ansi.Cut(line, cEnd, w)
		visibleLines[i] = left + hl.Render(mid) + right
	}

	return strings.Join(visibleLines, "\n")
}
