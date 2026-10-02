package internal

import (
	"path/filepath"
	"testing"
)

func TestCommandFlags(t *testing.T) {
	if rootCmd.Flags().Lookup("today") != nil {
		t.Fatal("spoiler flag --today still exists")
	}
	if rootCmd.Flags().Lookup("words") == nil || rootCmd.Flags().Lookup("reset") == nil || rootCmd.Flags().Lookup("stats") == nil {
		t.Fatal("expected --words, --reset, and --stats flags")
	}

	t.Setenv("WORDLY_DB", filepath.Join(t.TempDir(), "wordly.db"))
	rootCmd.SetArgs([]string{"--words", "--stats"})
	t.Cleanup(func() {
		words = false
		reset = false
		stats = false
		rootCmd.Flags().Lookup("words").Changed = false
		rootCmd.Flags().Lookup("reset").Changed = false
		rootCmd.Flags().Lookup("stats").Changed = false
		rootCmd.SetArgs(nil)
	})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("--words and --stats were accepted together")
	}
}
