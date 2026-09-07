package game

import (
	"crypto/sha256"
	"time"

	"github.com/nikhil25803/wordly/internal/db"
)

func GetTodaysPuzzle() (string, error) {

	today := time.Now().Format("2006-01-02")

	// If today's puzzle already exists in the database, return it
	todaysPuzzle, err := db.GetPuzzleWordByDate(today)
	if err == nil {
		return todaysPuzzle, nil
	}

	totalWords, err := db.GetWordCount()
	if err != nil {
		return "Not able to get word count", err
	}

	if totalWords == 0 {
		return "No words in the database", nil
	}

	hash := sha256.Sum256([]byte(today))

	wordIndex := int(hash[0]) % totalWords

	var word string
	err = db.DB.QueryRow(
		"SELECT word FROM words LIMIT 1 OFFSET ?",
		wordIndex,
	).Scan(&word)

	if err != nil {
		return "Not able to get today's puzzle", err
	}

	// Add todays puzzle to the database
	_, err = db.DB.Exec(
		"INSERT INTO puzzles (word_id, puzzle_date) VALUES ((SELECT id FROM words WHERE word = ?), ?)",
		word,
		today,
	)

	if err != nil {
		return "Not able to insert today's puzzle", err
	}

	return word, nil
}
