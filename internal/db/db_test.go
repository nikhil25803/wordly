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

func TestGetOrCreatePuzzleWordIsIdempotent(t *testing.T) {
	connectTestDatabase(t)

	const date = "2099-01-01"
	word, err := GetOrCreatePuzzleWord(date, "which")
	if err != nil {
		t.Fatal(err)
	}
	if word != "which" {
		t.Fatalf("got puzzle word %q, want %q", word, "which")
	}

	word, err = GetOrCreatePuzzleWord(date, "there")
	if err != nil {
		t.Fatal(err)
	}
	if word != "which" {
		t.Fatalf("existing puzzle changed to %q", word)
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
			word, err := GetOrCreatePuzzleWord("2099-01-03", candidate)
			results <- result{word, err}
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

	if _, err := GetOrCreatePuzzleWord("2099-01-02", "zzzzz"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown word error = %v, want sql.ErrNoRows", err)
	}
}
