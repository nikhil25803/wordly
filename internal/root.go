package internal

import (
	"fmt"
	"os"

	"github.com/nikhil25803/wordly/internal/db"
	"github.com/nikhil25803/wordly/internal/game"
	"github.com/spf13/cobra"
)

var (
	words bool
	today bool
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

		if today {
			word, err := game.GetTodaysPuzzle()
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), word)
			return nil
		}

		return nil
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
		&today,
		"today",
		false,
		"Show today's puzzle word",
	)
}

func Execute() error {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	return rootCmd.Execute()
}
