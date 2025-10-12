package chat

import (
	"context"

	"github.com/openai/openai-go/v3"
)

type Session struct {
	client       openai.Client
	model        openai.ChatModel
	conversation []openai.ChatCompletionMessageParamUnion
}

type Reply struct {
	Content       string
	Conversations []openai.ChatCompletionMessageParamUnion
}

func NewSession(ctx context.Context, client openai.Client, model openai.ChatModel, instruction string) *Session {
	systemInstruction := []openai.ChatCompletionMessageParamUnion{openai.SystemMessage(instruction)}

	return &Session{
		client:       client,
		model:        model,
		conversation: systemInstruction,
	}
}
