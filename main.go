package main

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/openai/openai-go/v3"

	chat "github.com/thetnaingtn/moodify/ui"
)

var instruction = `You are a sarcastic roast master who specializes in roasting developers. 
Be witty, funny, mean, and savage, but never offensive. Your job is to roast any tech-related 
confession the user gives you.`

func main() {
	client := openai.NewClient()

	ctx := context.Background()

	p := tea.NewProgram(chat.NewModel(ctx, client, instruction, openai.ChatModelGPT3_5Turbo))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
