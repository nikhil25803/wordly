package internal

import (
	"bytes"
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

func TestVersionFlag(t *testing.T) {
	var output bytes.Buffer
	rootCmd.SetOut(&output)
	rootCmd.SetArgs([]string{"--version"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
		if flag := rootCmd.Flags().Lookup("version"); flag != nil {
			_ = flag.Value.Set("false")
			flag.Changed = false
		}
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "wordly version dev\n"; got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}
