package internal

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "wordly",
	Short:         "Wordly is a terminal based wordle",
	Long:          `Wordly is a terminal based wordle game. It is a clone of the popular wordle game, but it is played in the terminal.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

func Execute() error {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	return rootCmd.Execute()
}
