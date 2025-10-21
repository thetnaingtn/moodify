package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/openai/openai-go/v3"
)

type model struct {
	conversation []openai.ChatCompletionMessageParamUnion
	textarea     textarea.Model
	client       openai.Client
	chatModel    openai.ChatModel
	spinner      spinner.Model
	reply        string
	loading      bool
	offset       int
	messages     []chatEntry
	viewport     viewport.Model
}

type apiCallStartedMsg struct{}

type apiResponseMsg struct {
	reply string
	param openai.ChatCompletionMessageParamUnion
}

type chatEntry struct {
	content string
	sender  string
	loading bool
}

type apiErrorMsg struct {
	err error
}

func (e *apiErrorMsg) Error() string {
	return e.err.Error()
}

func (m *model) Init() tea.Cmd {
	return m.textarea.Cursor.BlinkCmd()
}

func (m *model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if cmd := m.updateTextArea(teaMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.updateSpinner(teaMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.updateViewport(teaMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	ctx := context.Background()

	switch msg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height("\n\n")
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, chatEntry{
				sender:  "You",
				content: m.textarea.Value(),
			})
			cmds = append(cmds, m.sendMessage(ctx))
			m.refreshViewport()
		}
	case apiCallStartedMsg:
		m.loading = true
		m.offset = len(m.messages)
		m.messages = append(m.messages, chatEntry{
			loading: true,
			sender:  "Roast Master",
			content: "Thinking",
		})
		cmds = append(cmds, m.spinner.Tick)
		m.refreshViewport()
	case apiResponseMsg:
		m.loading = false
		if m.offset >= 0 && m.offset < len(m.messages) {
			m.messages[m.offset] = chatEntry{
				loading: false,
				sender:  "Roast Master",
				content: msg.reply,
			}
		}

		m.refreshViewport()
	}
	return m, tea.Batch(cmds...)
}

func (m *model) updateTextArea(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	return cmd
}

func (m *model) updateSpinner(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)

	if _, ok := msg.(spinner.TickMsg); ok && m.messages[m.offset].loading {
		m.refreshViewport()
	}

	return cmd
}

func (m *model) updateViewport(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return cmd
}

func (m *model) refreshViewport() {
	lines := make([]string, 0, len(m.messages))

	for _, entry := range m.messages {
		content := entry.content
		if entry.loading {
			content = fmt.Sprintf("%s %s", entry.content, m.spinner.View())
		}

		lines = append(lines, fmt.Sprintf("%s: %s", entry.sender, content))
	}

	if len(lines) < 1 {
		return
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
}

func (m *model) View() string {
	return fmt.Sprintf("%s\n\n%s", m.viewport.View(), m.textarea.View())
}

func (m *model) sendMessage(ctx context.Context) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return apiCallStartedMsg{}
		},
		func() tea.Msg {
			input := m.textarea.Value()
			input = strings.TrimSpace(input)

			m.conversation = append(m.conversation, openai.UserMessage(input))
			resp, err := m.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
				Model:    m.chatModel,
				Messages: m.conversation,
			})

			if err != nil {
				return &apiErrorMsg{err: err}
			}

			message := resp.Choices[0].Message

			m.conversation = append(m.conversation, message.ToParam())
			m.reply = message.Content
			return apiResponseMsg{
				reply: message.Content,
				param: message.ToParam(),
			}
		},
	)
}

func newModel() *model {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.Placeholder = "Send a message..."
	ta.SetHeight(3)

	sp := spinner.New()
	sp.Spinner = spinner.Points

	vp := viewport.New(30, 10)
	vp.SetContent("Welcome to the Roast Master! Type your tech-related confession below and hit Enter to receive a savage roast.")

	client := openai.NewClient()
	return &model{
		textarea:  ta,
		client:    client,
		chatModel: openai.ChatModelGPT3_5Turbo,
		spinner:   sp,
		viewport:  vp,
		offset:    -1,
		conversation: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a sarcastic roast master who specializes in roasting developers. Be witty, funny, mean, and savage, but never offensive. Your job is to roast any tech-related confession the user gives you. Be concise and to the point."),
		},
	}
}

func main() {
	model := newModel()

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
