package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/openai/openai-go/v3"
)

const (
	gap            = "\n\n"
	userLabel      = "You"
	assistantLabel = "Assistant"
	thinkingText   = "Thinking…"
)

type chatEntry struct {
	sender  string
	content string
	loading bool
}

type apiCallStartedMsg struct{}

type apiResponseMsg struct {
	content string
	param   openai.ChatCompletionMessageParamUnion
}

type apiErrorMsg struct {
	err error
}

type errMsg struct {
	err error
}

func (m errMsg) Error() string {
	if m.err == nil {
		return ""
	}
	return m.err.Error()
}

// Model implements the Bubble Tea chat UI backed by OpenAI responses.
type Model struct {
	ctx            context.Context
	client         openai.Client
	model          openai.ChatModel
	conversation   []openai.ChatCompletionMessageParamUnion
	viewport       viewport.Model
	textarea       textarea.Model
	spinner        spinner.Model
	messages       []chatEntry
	pendingIndex   int
	loading        bool
	senderStyle    lipgloss.Style
	assistantStyle lipgloss.Style
	errorStyle     lipgloss.Style
}

// NewModel sets up the chat UI and primes the OpenAI conversation with the optional
// systemInstruction. The provided context must remain valid for the life of the model.
func NewModel(ctx context.Context, client openai.Client, systemInstruction string, model openai.ChatModel) *Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 1024
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 10)
	vp.SetContent(`Welcome to the OpenAI chat.
Type a message and press Enter to send.`)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	var conversation []openai.ChatCompletionMessageParamUnion
	if systemInstruction != "" {
		conversation = append(conversation, openai.SystemMessage(systemInstruction))
	}

	return &Model{
		ctx:            ctx,
		client:         client,
		model:          model,
		conversation:   conversation,
		viewport:       vp,
		textarea:       ta,
		spinner:        sp,
		messages:       []chatEntry{},
		pendingIndex:   -1,
		senderStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		assistantStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		errorStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
	}
}

func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if textareaCmd := m.updateTextarea(msg); textareaCmd != nil {
		cmds = append(cmds, textareaCmd)
	}

	if viewportCmd := m.updateViewportModel(msg); viewportCmd != nil {
		cmds = append(cmds, viewportCmd)
	}

	if spinnerCmd := m.updateSpinner(msg); spinnerCmd != nil {
		cmds = append(cmds, spinnerCmd)
	}

	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = typed.Width
		m.textarea.SetWidth(typed.Width)
		m.viewport.Height = typed.Height - m.textarea.Height() - lipgloss.Height(gap)
		m.refreshViewport()
	case tea.KeyMsg:
		switch typed.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.loading {
				cmd := m.handleUserSubmit()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	case apiCallStartedMsg:
		m.loading = true
		m.pendingIndex = len(m.messages)
		m.messages = append(m.messages, chatEntry{
			sender:  assistantLabel,
			content: thinkingText,
			loading: true,
		})
		cmds = append(cmds, m.spinner.Tick)
		m.refreshViewport()
	case apiResponseMsg:
		m.loading = false
		if m.pendingIndex >= 0 && m.pendingIndex < len(m.messages) {
			m.messages[m.pendingIndex].content = typed.content
			m.messages[m.pendingIndex].loading = false
		}
		m.pendingIndex = -1
		m.conversation = append(m.conversation, typed.param)
		m.refreshViewport()
	case apiErrorMsg:
		m.loading = false
		errorText := "Something went wrong."
		if typed.err != nil {
			errorText = typed.err.Error()
		}
		if m.pendingIndex >= 0 && m.pendingIndex < len(m.messages) {
			m.messages[m.pendingIndex] = chatEntry{
				sender:  assistantLabel,
				content: errorText,
				loading: false,
			}
		} else {
			m.messages = append(m.messages, chatEntry{
				sender:  assistantLabel,
				content: errorText,
				loading: false,
			})
		}
		m.pendingIndex = -1
		m.refreshViewport()
	case errMsg:
		cmds = append(cmds, tea.Println(typed.Error()))
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

func (m *Model) updateTextarea(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return cmd
}

func (m *Model) updateViewportModel(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return cmd
}

func (m *Model) updateSpinner(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	if _, ok := msg.(spinner.TickMsg); ok && m.loading {
		m.refreshViewport()
	}
	return cmd
}

func (m *Model) handleUserSubmit() tea.Cmd {
	input := strings.TrimSpace(m.textarea.Value())
	if input == "" {
		return nil
	}

	m.messages = append(m.messages, chatEntry{
		sender:  userLabel,
		content: input,
	})
	m.conversation = append(m.conversation, openai.UserMessage(input))
	m.textarea.Reset()
	m.refreshViewport()
	m.viewport.GotoBottom()

	return m.callOpenAI()
}

func (m *Model) refreshViewport() {
	if m.viewport.Width <= 0 {
		return
	}

	wrapper := lipgloss.NewStyle().Width(m.viewport.Width)
	lines := make([]string, 0, len(m.messages))

	for _, entry := range m.messages {
		content := entry.content
		if entry.loading {
			content = fmt.Sprintf("%s %s", m.spinner.View(), content)
		}

		line := fmt.Sprintf("%s: %s", entry.sender, content)
		switch entry.sender {
		case userLabel:
			line = m.senderStyle.Render(line)
		case assistantLabel:
			if entry.loading {
				line = m.assistantStyle.Render(line)
			} else {
				line = m.assistantStyle.Render(line)
			}
		default:
			line = wrapper.Render(line)
		}

		if entry.sender == assistantLabel && !entry.loading && strings.Contains(strings.ToLower(content), "error") {
			line = m.errorStyle.Render(line)
		}

		lines = append(lines, wrapper.Render(line))
	}

	if len(lines) == 0 {
		return
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.GotoBottom()
}

func (m *Model) callOpenAI() tea.Cmd {
	messagesCopy := append([]openai.ChatCompletionMessageParamUnion(nil), m.conversation...)

	return tea.Batch(
		func() tea.Msg { return apiCallStartedMsg{} },
		func() tea.Msg {
			resp, err := m.client.Chat.Completions.New(m.ctx, openai.ChatCompletionNewParams{
				Messages: messagesCopy,
				Model:    m.model,
			})
			if err != nil {
				return apiErrorMsg{err: err}
			}
			if len(resp.Choices) == 0 {
				return apiErrorMsg{err: errors.New("no response from OpenAI")}
			}
			message := resp.Choices[0].Message
			return apiResponseMsg{
				content: message.Content,
				param:   message.ToParam(),
			}
		},
	)
}
