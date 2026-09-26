package game

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"time"

	"github.com/nikhil25803/wordly/internal/db"
)

func GetTodaysPuzzle() (string, error) {
	return getPuzzleForDate(time.Now().UTC().Format("2006-01-02"))
}

func getPuzzleForDate(date string) (string, error) {
	puzzle, err := db.GetPuzzleWordByDate(date)
	if err == nil {
		return puzzle, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	totalWords, err := db.GetWordCount()
	if err != nil {
		return "", err
	}
	if totalWords == 0 {
		return "", errors.New("no words in the database")
	}

	hash := sha256.Sum256([]byte(date))
	wordIndex := int(binary.BigEndian.Uint64(hash[:8]) % uint64(totalWords))
	word, err := db.GetWordByIndex(wordIndex)
	if err != nil {
		return "", err
	}

	return db.GetOrCreatePuzzleWord(date, word)
}

func StartGame() (string, error) {
	user, err := db.GetCurrentUser()
	if err != nil {
		return "", err
	}

	return "Game started for user: " + user.Username, nil
}
