package tui

import (
	"agent_scully/client"
	"agent_scully/tools"
	"agent_scully/utils"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func runToolCmd(cmdName, target string, width int, c *client.LLMClient) tea.Cmd {
	return func() tea.Msg {
		if target == "" {
			res, _ := utils.RenderMarkdown(fmt.Sprintf("**Usage:** `%s <target>`", cmdName), width)
			return LLMResponseMsg{text: res}
		}

		var out string
		var err error

		switch cmdName {
		case "/curl":
			out = tools.CurlCommand(target)
		case "/ping":
			out = tools.PingCommand(target)
		case "/nslookup":
			out = tools.NslookupCommand(target)
		case "/search":
			out, err = tools.SearchWeb(target)
			if err != nil {
				out = fmt.Sprintf("Search error: %v", err)
			}
		case "/read":
			out, err = tools.ReadWebpage(target)
			if err != nil {
				out = fmt.Sprintf("Read error: %v", err)
			}
		}

		text, err := c.RunPrompt("Summarize this output. Include the numbered source markers and URLs from the output, and add a Sources section at the end:\n\n" + out)
		rendered, _ := utils.RenderMarkdown(text, width)
		return LLMResponseMsg{text: rendered, err: err}
	}
}

func (m *Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) {
	vpWidth := msg.Width - 1
	if vpWidth < 1 {
		vpWidth = 1
	}

	taWidth := msg.Width - 2
	if taWidth < 1 {
		taWidth = 1
	}

	m.viewport.SetWidth(vpWidth)
	m.textarea.SetWidth(taWidth)

	vpHeight := msg.Height - (m.textarea.Height() + 2)
	if vpHeight < 1 {
		vpHeight = 1
	}
	m.viewport.SetHeight(vpHeight)

	if len(m.messages) > 0 {
		m.updateViewportContent()
	}
	m.viewport.GotoBottom()
}

func (m *Model) handleMouseReleaseMsg(msg tea.MouseReleaseMsg) (tea.Model, tea.Cmd) {
	if m.draggingScrollbar {
		m.draggingScrollbar = false
		return m, nil
	}

	if m.selecting {
		m.selecting = false
		selected := m.selectedText()
		if selected != "" {
			m.copied = selected
			return m, tea.SetClipboard(selected)
		}
	}

	return m, nil
}

func (m Model) handleKeyPressMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+q":
		return m, tea.Quit
	case "pgup", "ctrl+up":
		m.viewport.ScrollUp(3)
		return m, nil
	case "pgdown", "ctrl+down":
		m.viewport.ScrollDown(3)
		return m, nil
	case "end":
		m.viewport.GotoBottom()
		return m, nil
	case "enter":
		v := strings.TrimSpace(m.textarea.Value())
		if v != "" {
			m.loading = true
			m.thinkingDots = 0

			m.messages = append(m.messages, m.senderStyle.Render("You: ")+v)
			m.messages = append(m.messages, m.agentStyle.Render("BlackZetsu: ")+"Thinking...")
			m.updateViewportContent()
			m.textarea.Reset()
			m.viewport.GotoBottom()

			if cmdName, arg, isCmd := utils.ParseSlashCommand(v); isCmd {
				switch cmdName {
				case "/curl", "/ping", "/nslookup", "/search", "/read":
					return m, tea.Batch(
						runToolCmd(cmdName, arg, m.viewport.Width(), m.client),
						thinkingTick(),
					)
				default:
					m.loading = false
					m.messages[len(m.messages)-1] = m.agentStyle.Render("BlackZetsu: ") +
						fmt.Sprintf("Unknown command `%s`. Available: `/curl <domain>`, `/ping <domain>`, `/nslookup <domain>`, `/search QUERY`, `/read <url>`", cmdName)
					m.updateViewportContent()
					m.viewport.GotoBottom()
					return m, nil
				}
			}

			return m, tea.Batch(
				askAgentCmd(m.client, v, m.viewport.Width()),
				thinkingTick(),
			)
		}
		return m, nil
	}

	var tiCmd tea.Cmd
	m.textarea, tiCmd = m.textarea.Update(msg)
	return m, tiCmd
}

func (m *Model) handleMouseMotionMsg(msg tea.MouseMotionMsg, vpCmd *tea.Cmd) (tea.Model, tea.Cmd) {
	if m.draggingScrollbar {
		if msg.Button == tea.MouseLeft {
			m.scrollToMouseY(msg.Y)
		} else {
			m.draggingScrollbar = false
		}
		return m, nil
	}

	if m.selecting {
		if msg.Button == tea.MouseLeft {
			m.lastMouseX = msg.X
			m.lastMouseY = msg.Y

			if msg.Y <= 0 {
				m.viewport.ScrollUp(1)
			} else if msg.Y >= m.viewport.Height()-1 {
				m.viewport.ScrollDown(1)
			}

			y := max(0, min(msg.Y, m.viewport.Height()-1))
			x := max(0, min(msg.X, m.viewport.Width()-1))
			m.end = pos{
				line: m.viewport.YOffset() + y,
				col:  x,
			}
		} else {
			m.selecting = false
		}
		return m, nil
	}

	m.viewport, *vpCmd = m.viewport.Update(msg)
	return m, *vpCmd
}

func (m *Model) handleMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if len(m.messages) > 0 && msg.Button == tea.MouseLeft && msg.Y < m.viewport.Height() {
		if msg.X == m.viewport.Width() {
			m.selecting = false
			switch {
			case msg.Y == 0:
				m.viewport.ScrollUp(3)
			case msg.Y == m.viewport.Height()-1:
				m.viewport.ScrollDown(3)
			default:
				m.draggingScrollbar = true
				m.scrollToMouseY(msg.Y)
			}
			return m, nil
		}

		if msg.X < m.viewport.Width() {
			p := pos{
				line: m.viewport.YOffset() + msg.Y,
				col:  msg.X,
			}
			m.selecting = true
			m.start = p
			m.end = p
			m.lastMouseX = msg.X
			m.lastMouseY = msg.Y
			m.copied = ""
			return m, selScrollTick()
		}
	}

	return m, nil
}
