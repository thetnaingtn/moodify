package chat

import "context"

type Session interface {
	SendMessage(ctx context.Context, message string) (string, error)
}
