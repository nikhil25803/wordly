package db

import (
	"path/filepath"
	"testing"
)

func TestEmbeddedDatabase(t *testing.T) {
	t.Setenv("WORDLY_DB", filepath.Join(t.TempDir(), "wordly.db"))
	if err := ConnectDatabase(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { DB.Close() })

	count, err := GetWordCount()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("embedded database contains no words")
	}
}
