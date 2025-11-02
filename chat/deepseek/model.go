package deepseek

import (
	"context"

	ds "github.com/cohesion-org/deepseek-go"
)

type Model struct {
	client       *ds.Client
	chatModel    string
	conversation []ds.ChatCompletionMessage
}

func NewModel(client *ds.Client, chatModel string, instruction string) *Model {
	return &Model{
		client:    client,
		chatModel: chatModel,
		conversation: []ds.ChatCompletionMessage{
			{Role: ds.ChatMessageRoleSystem, Content: instruction},
		},
	}
}

func (m *Model) SendMessage(ctx context.Context, message string) (string, error) {
	m.conversation = append(m.conversation, ds.ChatCompletionMessage{
		Role:    ds.ChatMessageRoleUser,
		Content: message,
	})

	request := &ds.ChatCompletionRequest{
		Model:    m.chatModel,
		Messages: m.conversation,
	}

	resp, err := m.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	msg := resp.Choices[0].Message
	m.conversation = append(m.conversation, ds.ChatCompletionMessage{
		Role:    msg.Role,
		Content: msg.Content,
	})

	return msg.Content, nil
}
