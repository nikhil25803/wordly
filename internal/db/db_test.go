package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func connectTestDatabase(t *testing.T) {
	t.Helper()
	t.Setenv("WORDLY_DB", filepath.Join(t.TempDir(), "wordly.db"))
	if err := ConnectDatabase(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { DB.Close() })
}

func TestEmbeddedDatabase(t *testing.T) {
	connectTestDatabase(t)

	count, err := GetWordCount()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("embedded database contains no words")
	}
}

func TestGetCurrentUserIsIdempotent(t *testing.T) {
	connectTestDatabase(t)

	first, err := GetCurrentUser()
	if err != nil {
		t.Fatal(err)
	}
	second, err := GetCurrentUser()
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("current user changed: first=%+v second=%+v", first, second)
	}

	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d users, want 1", count)
	}
}

func TestWordHandlers(t *testing.T) {
	connectTestDatabase(t)

	word, err := GetWordByIndex(0)
	if err != nil {
		t.Fatal(err)
	}
	if word != "which" {
		t.Fatalf("got first word %q, want %q", word, "which")
	}
	if _, err := GetWordByIndex(-1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("negative index error = %v, want sql.ErrNoRows", err)
	}
	count, err := GetWordCount()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := GetWordByIndex(count); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("out-of-range index error = %v, want sql.ErrNoRows", err)
	}

	for _, test := range []struct {
		word string
		want bool
	}{
		{" WHICH ", true},
		{"zzzzz", false},
		{"bad", false},
	} {
		exists, err := WordExists(test.word)
		if err != nil {
			t.Fatal(err)
		}
		if exists != test.want {
			t.Errorf("WordExists(%q) = %v, want %v", test.word, exists, test.want)
		}
	}
}

func TestGetOrCreatePuzzleIsIdempotent(t *testing.T) {
	connectTestDatabase(t)

	const date = "2099-01-01"
	puzzle, err := GetOrCreatePuzzle(date, "which")
	if err != nil {
		t.Fatal(err)
	}
	if puzzle.Word != "which" {
		t.Fatalf("got puzzle word %q, want %q", puzzle.Word, "which")
	}

	puzzle, err = GetOrCreatePuzzle(date, "there")
	if err != nil {
		t.Fatal(err)
	}
	if puzzle.Word != "which" {
		t.Fatalf("existing puzzle changed to %q", puzzle.Word)
	}

	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM puzzles WHERE puzzle_date = ?", date).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d puzzles for date, want 1", count)
	}

	type result struct {
		word string
		err  error
	}
	results := make(chan result, 2)
	for _, candidate := range []string{"which", "there"} {
		go func() {
			puzzle, err := GetOrCreatePuzzle("2099-01-03", candidate)
			results <- result{puzzle.Word, err}
		}()
	}
	first := <-results
	second := <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("concurrent puzzle creation errors: %v, %v", first.err, second.err)
	}
	if first.word != second.word {
		t.Fatalf("concurrent puzzle words differ: %q, %q", first.word, second.word)
	}

	if _, err := GetOrCreatePuzzle("2099-01-02", "zzzzz"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown word error = %v, want sql.ErrNoRows", err)
	}
}

func TestGamePersistenceAndReset(t *testing.T) {
	connectTestDatabase(t)

	currentUser, err := GetCurrentUser()
	if err != nil {
		t.Fatal(err)
	}
	puzzle, err := GetOrCreatePuzzle("2099-05-01", "which")
	if err != nil {
		t.Fatal(err)
	}
	history, err := GetOrCreateHistory(currentUser.ID, puzzle.ID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := GetOrCreateHistory(currentUser.ID, puzzle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != history.ID {
		t.Fatalf("history changed from %d to %d", history.ID, again.ID)
	}

	if _, err := SaveGuess(history.ID, "there", false); err != nil {
		t.Fatal(err)
	}
	finished, err := SaveGuess(history.ID, "which", true)
	if err != nil {
		t.Fatal(err)
	}
	if !finished.Guessed || finished.Attempts != 2 {
		t.Fatalf("finished history = %+v", finished)
	}
	if _, err := SaveGuess(history.ID, "their", false); !errors.Is(err, ErrHistoryFinished) {
		t.Fatalf("finished history accepted guess: %v", err)
	}
	guesses, err := GetGuesses(history.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(guesses) != 2 || guesses[0].Word != "there" || guesses[1].Word != "which" {
		t.Fatalf("saved guesses = %+v", guesses)
	}
	results, err := GetResults(currentUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].Won || results[0].Attempts != 2 {
		t.Fatalf("saved results = %+v", results)
	}

	result, err := DB.Exec(
		"INSERT INTO users (username, registered_at) VALUES (?, ?)",
		"another-user", "2099-05-01T00:00:00Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	otherUserID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := GetOrCreateHistory(int(otherUserID), puzzle.ID); err != nil {
		t.Fatal(err)
	}

	if err := ResetUserGames(currentUser.ID); err != nil {
		t.Fatal(err)
	}
	var currentCount, otherCount, userCount int
	if err := DB.QueryRow("SELECT COUNT(*) FROM history WHERE user_id = ?", currentUser.ID).Scan(&currentCount); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow("SELECT COUNT(*) FROM history WHERE user_id = ?", otherUserID).Scan(&otherCount); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", currentUser.ID).Scan(&userCount); err != nil {
		t.Fatal(err)
	}
	if currentCount != 0 || otherCount != 1 || userCount != 1 {
		t.Fatalf("reset counts: current=%d other=%d user=%d", currentCount, otherCount, userCount)
	}
}
