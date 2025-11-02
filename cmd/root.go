package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var (
	debugFlag bool
	rootCmd   = &cobra.Command{
		Use: "moodify",
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug mode")
	rootCmd.AddCommand(roastCmd, praiseCmd)
}

func Execute(ctx context.Context) error {
	if err := fang.Execute(ctx, rootCmd, fang.WithoutCompletions(), fang.WithNotifySignal(os.Interrupt, os.Kill)); err != nil {
		return err
	}
	return nil
}
