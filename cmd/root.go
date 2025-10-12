package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "moodify",
}

func init() {
	rootCmd.AddCommand(roastCmd, praiseCmd)
}

func Execute(ctx context.Context) error {
	if err := fang.Execute(ctx, rootCmd, fang.WithoutCompletions(), fang.WithNotifySignal(os.Interrupt, os.Kill)); err != nil {
		return err
	}
	return nil
}
