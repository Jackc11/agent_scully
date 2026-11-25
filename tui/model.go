package tui

import (
	"agent_scully/client"
	"agent_scully/utils"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type LLMResponseMsg struct {
	text string
	err  error
}

func askAgentCmd(c *client.LLMClient, prompt string, width int) tea.Cmd {
	return func() tea.Msg {
		text, err := c.RunPrompt(prompt)
		if err != nil {
			return LLMResponseMsg{err: err}
		}

		res, err := utils.RenderMarkdown(text, width)
		return LLMResponseMsg{text: res, err: err}
	}
}

type pos struct {
	line int
	col  int
}

type Model struct {
	viewport          viewport.Model
	messages          []string
	textarea          textarea.Model
	senderStyle       lipgloss.Style
	agentStyle        lipgloss.Style
	err               error
	client            *client.LLMClient
	loading           bool
	draggingScrollbar bool
	lines             []string
	selecting         bool
	start             pos
	end               pos
	lastMouseX        int
	lastMouseY        int
	copied            string
	thinkingDots      int
}

func InitializeModel(conf *utils.AgentConfig) Model {
	ta := textarea.New()
	ta.Placeholder = "Ask a question..."
	ta.SetVirtualCursor(false)
	ta.Focus()

	ta.Prompt = ""
	ta.CharLimit = 0
	ta.SetWidth(30)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	applyTextareaStyles(&ta)

	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(5))
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)
	vp.MouseWheelEnabled = true

	return Model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: senderStyle,
		agentStyle:  lipgloss.NewStyle().Foreground(lineBarColorLight).Bold(true),
		err:         nil,
		client:      client.CreateLLMClient(conf),
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd, vpCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSizeMsg(msg)
	case tea.MouseClickMsg:
		return m.handleMouseClick(msg)
	case tea.MouseMotionMsg:
		return m.handleMouseMotionMsg(msg, &vpCmd)
	case setScrollTickMsg:
		return m.handleScrollTickMsg()
	case tea.MouseReleaseMsg:
		return m.handleMouseReleaseMsg(msg)
	case tea.KeyPressMsg:
		return m.handleKeyPressMsg(msg)
	case thinkingTickMsg:
		return m.handleThinkingTickMsg()
	case LLMResponseMsg:
		m.loading = false
		lastIdx := len(m.messages) - 1
		if msg.err != nil {
			m.messages[lastIdx] = m.agentStyle.Render("Agent Scully: ") + "Error: " + msg.err.Error()
		} else {
			m.messages[lastIdx] = m.agentStyle.Render("Agent Scully: ") + msg.text
		}
		m.updateViewportContent()
		m.viewport.GotoBottom()
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}

	switch msg.(type) {
	case tea.MouseWheelMsg, tea.MouseMotionMsg:
		m.viewport, vpCmd = m.viewport.Update(msg)
	default:
		m.textarea, tiCmd = m.textarea.Update(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() tea.View {
	var viewportView string
	if len(m.messages) == 0 {
		viewportView = renderWelcomeArea(m.textarea.Width(), m.viewport.Height())
	} else {
		viewportView = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.renderHighlightedViewport(),
			m.renderScrollbar(),
		)
	}

	textareaBoxStyle := lipgloss.NewStyle().
		Padding(1, 1).
		Background(baseBgColorDark).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(lineBarColorDark).
		BorderBackground(baseBgColorDark).
		Padding(1, 1).
		Width(m.textarea.Width() + 2).
		MaxWidth(m.textarea.Width() + 3)

	styledTextarea := textareaBoxStyle.Render(m.textarea.View())

	v := tea.NewView(viewportView + "\n" + styledTextarea)

	c := m.textarea.Cursor()
	if c != nil {
		c.X += 2
		c.Y += lipgloss.Height(viewportView) + 1
	}
	v.Cursor = c
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	return v
}

func (m *Model) updateViewportContent() {
	content := lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "\n\n"))
	m.lines = strings.Split(content, "\n")
	m.viewport.SetContent(content)
}
