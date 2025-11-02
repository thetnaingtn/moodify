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
	praiseCmd = &cobra.Command{
		Use:   "praise",
		Short: "Start a praising session",
		Long:  `Start a praising session where the AI will praise your tech-related confessions.`,
		RunE:  praiseRunE,
	}

	praiseInstruction = "You are an enthusiastic cheerleader who specializes in praising developers. Be uplifting, funny, and encouraging. Your job is to praise any tech-related confession the user gives you."
)

func praiseRunE(cmd *cobra.Command, args []string) error {
	client := openai.NewClient()

	ctx := context.Background()

	model := gpt.NewModel(client, openai.ChatModelGPT3_5Turbo, praiseInstruction)

	p := tea.NewProgram(ui.NewModel(ctx, model))
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
