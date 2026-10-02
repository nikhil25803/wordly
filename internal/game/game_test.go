package game

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikhil25803/wordly/internal/db"
)

func setupTestDatabase(t *testing.T) {
	t.Helper()
	t.Setenv("WORDLY_DB", filepath.Join(t.TempDir(), "wordly.db"))
	if err := db.ConnectDatabase(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.DB.Close() })
}

func TestGetPuzzleForDate(t *testing.T) {
	setupTestDatabase(t)

	const existingDate = "2099-01-01"
	if _, err := db.GetOrCreatePuzzle(existingDate, "which"); err != nil {
		t.Fatal(err)
	}
	puzzle, err := getPuzzleForDate(existingDate)
	if err != nil {
		t.Fatal(err)
	}
	if puzzle.Word != "which" {
		t.Fatalf("got existing puzzle %q, want %q", puzzle.Word, "which")
	}

	count, err := db.GetWordCount()
	if err != nil {
		t.Fatal(err)
	}
	if count <= 256 {
		t.Fatalf("test dictionary has %d words, need more than 256", count)
	}
	date := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	var index int
	for attempts := 0; attempts < 1000 && index <= 255; attempts++ {
		date = date.AddDate(0, 0, 1)
		hash := sha256.Sum256([]byte(date.Format("2006-01-02")))
		index = int(binary.BigEndian.Uint64(hash[:8]) % uint64(count))
	}
	if index <= 255 {
		t.Fatal("could not find a test date selecting beyond the first 256 words")
	}

	want, err := db.GetWordByIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	puzzle, err = getPuzzleForDate(date.Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	if puzzle.Word != want {
		t.Fatalf("puzzle = %q, want index %d (%q)", puzzle.Word, index, want)
	}
}

func TestEvaluateGuessHandlesDuplicateLetters(t *testing.T) {
	guess := evaluateGuess("apple", "alley")
	want := [...]LetterState{Correct, Present, Absent, Present, Absent}
	for i, state := range want {
		if guess.Tiles[i].State != state {
			t.Errorf("tile %d state = %v, want %v", i, guess.Tiles[i].State, state)
		}
	}
}

func TestGameValidationResumeAndWin(t *testing.T) {
	setupTestDatabase(t)
	const date = "2099-02-01"
	if _, err := db.GetOrCreatePuzzle(date, "which"); err != nil {
		t.Fatal(err)
	}

	current, err := startGameForDate(date)
	if err != nil {
		t.Fatal(err)
	}
	if current.Answer() != "" {
		t.Fatal("answer was revealed before completion")
	}
	if err := current.SubmitGuess("bad"); !errors.Is(err, ErrGuessLength) {
		t.Fatalf("short guess error = %v, want ErrGuessLength", err)
	}
	if err := current.SubmitGuess("ZZZZZ"); !errors.Is(err, ErrWordNotFound) {
		t.Fatalf("unknown guess error = %v, want ErrWordNotFound", err)
	}
	if len(current.Guesses) != 0 {
		t.Fatalf("invalid guesses consumed %d attempts", len(current.Guesses))
	}
	if err := current.SubmitGuess(" THERE "); err != nil {
		t.Fatal(err)
	}

	resumed, err := startGameForDate(date)
	if err != nil {
		t.Fatal(err)
	}
	if len(resumed.Guesses) != 1 || resumed.Guesses[0].Word != "there" {
		t.Fatalf("resumed guesses = %+v, want [there]", resumed.Guesses)
	}
	if err := resumed.SubmitGuess("which"); err != nil {
		t.Fatal(err)
	}
	if !resumed.Done || !resumed.Won || resumed.Answer() != "which" {
		t.Fatalf("completed game = done:%v won:%v answer:%q", resumed.Done, resumed.Won, resumed.Answer())
	}
	if resumed.Stats.Played != 1 || resumed.Stats.Wins != 1 || resumed.Stats.Distribution[1] != 1 {
		t.Fatalf("unexpected stats: %+v", resumed.Stats)
	}
	if err := resumed.SubmitGuess("their"); !errors.Is(err, ErrGameFinished) {
		t.Fatalf("post-game error = %v, want ErrGameFinished", err)
	}

	reopened, err := startGameForDate(date)
	if err != nil {
		t.Fatal(err)
	}
	if !reopened.Done || len(reopened.Guesses) != 2 {
		t.Fatalf("reopened game = done:%v guesses:%d", reopened.Done, len(reopened.Guesses))
	}
}

func TestGameLossAfterSixGuesses(t *testing.T) {
	setupTestDatabase(t)
	const date = "2099-03-01"
	if _, err := db.GetOrCreatePuzzle(date, "which"); err != nil {
		t.Fatal(err)
	}
	current, err := startGameForDate(date)
	if err != nil {
		t.Fatal(err)
	}
	for _, word := range []string{"there", "their", "about", "would", "these", "other"} {
		if err := current.SubmitGuess(word); err != nil {
			t.Fatal(err)
		}
	}
	if !current.Done || current.Won || current.Answer() != "which" {
		t.Fatalf("lost game = done:%v won:%v answer:%q", current.Done, current.Won, current.Answer())
	}
	if current.Stats.Played != 1 || current.Stats.Wins != 0 {
		t.Fatalf("unexpected loss stats: %+v", current.Stats)
	}
}

func TestCalculateStats(t *testing.T) {
	results := []db.Result{
		{PuzzleDate: "2099-04-01", Won: true, Attempts: 2},
		{PuzzleDate: "2099-04-02", Won: true, Attempts: 3},
		{PuzzleDate: "2099-04-04", Won: true, Attempts: 1},
		{PuzzleDate: "2099-04-05", Won: false, Attempts: 6},
		{PuzzleDate: "2099-04-06", Won: true, Attempts: 4},
	}
	stats := calculateStats(results, "2099-04-06")
	if stats.Played != 5 || stats.Wins != 4 || stats.WinPercentage != 80 {
		t.Fatalf("unexpected totals: %+v", stats)
	}
	if stats.CurrentStreak != 1 || stats.MaxStreak != 2 {
		t.Fatalf("unexpected streaks: %+v", stats)
	}
	if stats.Distribution != [MaxAttempts]int{1, 1, 1, 1, 0, 0} {
		t.Fatalf("unexpected distribution: %v", stats.Distribution)
	}
	if stale := calculateStats(results, "2099-04-07"); stale.CurrentStreak != 0 {
		t.Fatalf("missed day retained current streak: %+v", stale)
	}
}

func TestGetCurrentUserStatsDoesNotStartGame(t *testing.T) {
	setupTestDatabase(t)

	stats, err := GetCurrentUserStats()
	if err != nil {
		t.Fatal(err)
	}
	if stats != (Stats{}) {
		t.Fatalf("empty stats = %+v, want zero values", stats)
	}

	for _, table := range []string{"puzzles", "history"} {
		var count int
		if err := db.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("--stats behavior created %d rows in %s", count, table)
		}
	}
}
