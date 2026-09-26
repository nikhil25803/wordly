package db

import (
	"database/sql"
	"strings"
)

func GetWordCount() (int, error) {
	var count int

	err := DB.QueryRow(
		"SELECT COUNT(*) FROM words",
	).Scan(&count)

	return count, err
}

func GetWordByIndex(index int) (string, error) {
	if index < 0 {
		return "", sql.ErrNoRows
	}

	var word string
	err := DB.QueryRow(
		"SELECT word FROM words ORDER BY id LIMIT 1 OFFSET ?",
		index,
	).Scan(&word)

	return word, err
}

func WordExists(word string) (bool, error) {
	word = strings.ToLower(strings.TrimSpace(word))
	if len(word) != 5 {
		return false, nil
	}

	var exists bool
	err := DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM words WHERE word = ?)",
		word,
	).Scan(&exists)

	return exists, err
}

func GetPuzzleWordByDate(date string) (string, error) {
	var word string
	err := DB.QueryRow(
		"SELECT w.word FROM puzzles p JOIN words w ON p.word_id = w.id WHERE p.puzzle_date = ?",
		date,
	).Scan(&word)

	if err != nil {
		return "", err
	}

	return word, nil
}

func GetOrCreatePuzzleWord(date, word string) (string, error) {
	_, err := DB.Exec(
		`INSERT INTO puzzles (word_id, puzzle_date)
		 SELECT id, ? FROM words WHERE word = ?
		 ON CONFLICT(puzzle_date) DO NOTHING`,
		date, strings.ToLower(strings.TrimSpace(word)),
	)
	if err != nil {
		return "", err
	}

	return GetPuzzleWordByDate(date)
}
