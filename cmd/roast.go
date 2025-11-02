package cmd

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
	"github.com/thetnaingtn/moodify/chat/gpt"
	"github.com/thetnaingtn/moodify/ui"
)

var (
	roastCmd = &cobra.Command{
		Use:   "roast",
		Short: "Start a roasting session",
		Long:  `Start a roasting session where the AI will roast your tech-related confessions.`,
		RunE:  roastRunE,
	}

	roastInstruction = "You are a sarcastic roast master who specializes in roasting developers. Be witty, funny, mean, and savage, but never offensive. Your job is to roast any tech-related confession the user gives you."
)

func roastRunE(cmd *cobra.Command, args []string) error {
	client := openai.NewClient()

	ctx := context.Background()
	model := gpt.NewModel(client, openai.ChatModelGPT3_5Turbo, roastInstruction)

	if debugFlag {
		if _, err := tea.LogToFile("./moodify.log", ""); err != nil {
			return err
		}
	}

	p := tea.NewProgram(ui.NewModel(ctx, model))
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
