package game

import (
	"crypto/sha256"
	"encoding/binary"
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
	if _, err := db.GetOrCreatePuzzleWord(existingDate, "which"); err != nil {
		t.Fatal(err)
	}
	word, err := getPuzzleForDate(existingDate)
	if err != nil {
		t.Fatal(err)
	}
	if word != "which" {
		t.Fatalf("got existing puzzle %q, want %q", word, "which")
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
	dateString := date.Format("2006-01-02")
	word, err = getPuzzleForDate(dateString)
	if err != nil {
		t.Fatal(err)
	}
	if word != want {
		t.Fatalf("puzzle for %s = %q, want index %d (%q)", dateString, word, index, want)
	}

	again, err := getPuzzleForDate(dateString)
	if err != nil {
		t.Fatal(err)
	}
	if again != word {
		t.Fatalf("puzzle changed from %q to %q", word, again)
	}
}
