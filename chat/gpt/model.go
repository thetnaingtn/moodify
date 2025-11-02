package gpt

import (
	"context"
	"log"

	"github.com/openai/openai-go/v3"
)

type Model struct {
	client       openai.Client
	chatModel    openai.ChatModel
	conversation []openai.ChatCompletionMessageParamUnion
	instruction  string
}

func NewModel(client openai.Client, chatModel openai.ChatModel, instruction string) *Model {
	systemMessage := openai.SystemMessage(instruction)

	return &Model{
		client:       client,
		chatModel:    chatModel,
		conversation: []openai.ChatCompletionMessageParamUnion{systemMessage},
		instruction:  instruction,
	}
}

func (m *Model) SendMessage(ctx context.Context, message string) (string, error) {
	m.conversation = append(m.conversation, openai.UserMessage(message))

	resp, err := m.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    m.chatModel,
		Messages: m.conversation,
	})

	if err != nil {
		log.Println("Error sending message: ", err)
		return "", err
	}

	msg := resp.Choices[0].Message
	m.conversation = append(m.conversation, msg.ToParam())

	return msg.Content, nil
}
