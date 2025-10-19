package main

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/openai/openai-go/v3"
)

type model struct {
	conversation string
	textarea     textarea.Model
	client       openai.Client
	chatModel    openai.ChatModel
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if cmd := m.updateTextArea(teaMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	ctx := context.Background()

	switch msg := teaMsg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			input := m.textarea.Value()

			resp, _ := m.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
				Model: m.chatModel,
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(input),
				},
			})

			m.conversation = resp.Choices[0].Message.Content
		}

	}
	return m, tea.Batch(cmds...)
}

func (m *model) updateTextArea(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	return cmd
}

func (m *model) View() string {
	return fmt.Sprintf("%s\n\n%s", m.textarea.View(), m.conversation)
}

func newModel() *model {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.Placeholder = "Send a message..."
	ta.SetHeight(3)

	client := openai.NewClient()

	return &model{
		textarea:  ta,
		client:    client,
		chatModel: openai.ChatModelGPT3_5Turbo,
	}
}

func main() {
	model := newModel()

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
