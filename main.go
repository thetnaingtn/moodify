package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
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
}

type apiResponseMsg struct {
	reply string
	param openai.ChatCompletionMessageParamUnion
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

	ctx := context.Background()

	switch msg := teaMsg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.reply = fmt.Sprintf("%s %s", m.spinner.View(), "Thinking...")
			m.loading = true
			cmds = append(cmds, m.spinner.Tick, m.sendMessage(ctx))
		}
	case apiResponseMsg:
		m.loading = false
		m.reply = fmt.Sprintf("%s %s", "Response: ", msg.reply)
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

	if _, ok := msg.(spinner.TickMsg); ok && m.loading {
		m.reply = fmt.Sprintf("%s %s", m.spinner.View(), "Thinking...") // update reply with new spinner frame
	}

	return cmd
}

func (m *model) View() string {
	return fmt.Sprintf("%s\n\n%s", m.textarea.View(), m.reply)
}

func (m *model) sendMessage(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
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
	}
}

func newModel() *model {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.Placeholder = "Send a message..."
	ta.SetHeight(3)

	client := openai.NewClient()

	sp := spinner.New()
	sp.Spinner = spinner.Points

	return &model{
		textarea:  ta,
		client:    client,
		chatModel: openai.ChatModelGPT3_5Turbo,
		spinner:   sp,
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
