package internal

import (
	"fmt"
	"os"

	"github.com/nikhil25803/wordly/internal/db"
	"github.com/nikhil25803/wordly/internal/game"
	"github.com/nikhil25803/wordly/internal/ui"
	"github.com/spf13/cobra"
)

var (
	words bool
	reset bool
	stats bool
)

var rootCmd = &cobra.Command{
	Use:           "wordly",
	Short:         "Wordly is a terminal based wordle",
	Long:          `Wordly is a terminal based wordle game. It is a clone of the popular wordle game, but it is played in the terminal.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := db.ConnectDatabase(); err != nil {
			return err
		}
		defer db.DB.Close()

		if words {
			count, err := db.GetWordCount()
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), count)
			return nil
		}

		if reset {
			user, err := game.ResetCurrentUser()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Reset stats for %s.\n", user.Username)
			return nil
		}
		if stats {
			currentStats, err := game.GetCurrentUserStats()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderStats(currentStats))
			return nil
		}

		currentGame, err := game.StartGame()
		if err != nil {
			return err
		}
		return ui.Run(currentGame, cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.Flags().BoolVar(
		&words,
		"words",
		false,
		"Show the number of available words",
	)

	rootCmd.Flags().BoolVar(
		&reset,
		"reset",
		false,
		"Reset the current user's game statistics",
	)
	rootCmd.Flags().BoolVar(
		&stats,
		"stats",
		false,
		"Show the current user's game statistics",
	)
	rootCmd.MarkFlagsMutuallyExclusive("words", "reset", "stats")
}

func Execute() error {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	return rootCmd.Execute()
}
