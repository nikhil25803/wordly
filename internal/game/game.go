package game

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"strings"
	"time"

	"github.com/nikhil25803/wordly/internal/db"
)

const (
	WordLength  = 5
	MaxAttempts = 6
)

var (
	ErrGuessLength  = errors.New("guess must be five letters")
	ErrWordNotFound = errors.New("word is not in the dictionary")
	ErrGameFinished = errors.New("game is already finished")
)

type LetterState uint8

const (
	Absent LetterState = iota
	Present
	Correct
)

type Tile struct {
	Letter byte
	State  LetterState
}

type EvaluatedGuess struct {
	Word  string
	Tiles [WordLength]Tile
}

type Stats struct {
	Played        int
	Wins          int
	WinPercentage int
	CurrentStreak int
	MaxStreak     int
	Distribution  [MaxAttempts]int
}

type Game struct {
	User       db.User
	PuzzleDate string
	Guesses    []EvaluatedGuess
	Done       bool
	Won        bool
	Stats      Stats

	answer    string
	historyID int
}

func getPuzzleForDate(date string) (db.Puzzle, error) {
	puzzle, err := db.GetPuzzleByDate(date)
	if err == nil {
		return puzzle, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.Puzzle{}, err
	}

	totalWords, err := db.GetWordCount()
	if err != nil {
		return db.Puzzle{}, err
	}
	if totalWords == 0 {
		return db.Puzzle{}, errors.New("no words in the database")
	}

	hash := sha256.Sum256([]byte(date))
	wordIndex := int(binary.BigEndian.Uint64(hash[:8]) % uint64(totalWords))
	word, err := db.GetWordByIndex(wordIndex)
	if err != nil {
		return db.Puzzle{}, err
	}

	return db.GetOrCreatePuzzle(date, word)
}

func StartGame() (*Game, error) {
	return startGameForDate(time.Now().UTC().Format("2006-01-02"))
}

func startGameForDate(date string) (*Game, error) {
	user, err := db.GetCurrentUser()
	if err != nil {
		return nil, err
	}
	puzzle, err := getPuzzleForDate(date)
	if err != nil {
		return nil, err
	}
	history, err := db.GetOrCreateHistory(user.ID, puzzle.ID)
	if err != nil {
		return nil, err
	}
	savedGuesses, err := db.GetGuesses(history.ID)
	if err != nil {
		return nil, err
	}

	game := &Game{
		User:       user,
		PuzzleDate: date,
		Done:       history.Guessed || history.Attempts >= MaxAttempts,
		Won:        history.Guessed,
		answer:     puzzle.Word,
		historyID:  history.ID,
	}
	for _, guess := range savedGuesses {
		game.Guesses = append(game.Guesses, evaluateGuess(puzzle.Word, guess.Word))
	}
	if game.Done {
		if err := game.loadStats(); err != nil {
			return nil, err
		}
	}
	return game, nil
}

func (g *Game) SubmitGuess(word string) error {
	if g.Done {
		return ErrGameFinished
	}

	word = strings.ToLower(strings.TrimSpace(word))
	if !isFiveLetters(word) {
		return ErrGuessLength
	}
	exists, err := db.WordExists(word)
	if err != nil {
		return err
	}
	if !exists {
		return ErrWordNotFound
	}

	guess := evaluateGuess(g.answer, word)
	history, err := db.SaveGuess(g.historyID, word, word == g.answer)
	if err != nil {
		if errors.Is(err, db.ErrHistoryFinished) {
			return ErrGameFinished
		}
		return err
	}
	g.Guesses = append(g.Guesses, guess)
	g.Won = history.Guessed
	g.Done = g.Won || history.Attempts >= MaxAttempts
	if g.Done {
		return g.loadStats()
	}
	return nil
}

func (g *Game) Answer() string {
	if !g.Done {
		return ""
	}
	return g.answer
}

func ResetCurrentUser() (db.User, error) {
	user, err := db.GetCurrentUser()
	if err != nil {
		return db.User{}, err
	}
	return user, db.ResetUserGames(user.ID)
}

func GetCurrentUserStats() (Stats, error) {
	user, err := db.GetCurrentUser()
	if err != nil {
		return Stats{}, err
	}
	results, err := db.GetResults(user.ID)
	if err != nil {
		return Stats{}, err
	}
	return calculateStats(results, time.Now().UTC().Format("2006-01-02")), nil
}

func (g *Game) loadStats() error {
	results, err := db.GetResults(g.User.ID)
	if err != nil {
		return err
	}
	g.Stats = calculateStats(results, g.PuzzleDate)
	return nil
}

func calculateStats(results []db.Result, today string) Stats {
	var stats Stats
	var streak int
	var previous time.Time

	for _, result := range results {
		date, err := time.Parse("2006-01-02", result.PuzzleDate)
		if err != nil {
			continue
		}
		stats.Played++
		if !result.Won {
			streak = 0
			previous = date
			continue
		}

		stats.Wins++
		if result.Attempts >= 1 && result.Attempts <= MaxAttempts {
			stats.Distribution[result.Attempts-1]++
		}
		if streak > 0 && date.Sub(previous) == 24*time.Hour {
			streak++
		} else {
			streak = 1
		}
		if streak > stats.MaxStreak {
			stats.MaxStreak = streak
		}
		previous = date
	}
	if stats.Played > 0 {
		stats.WinPercentage = stats.Wins * 100 / stats.Played
	}
	if len(results) > 0 && results[len(results)-1].Won && results[len(results)-1].PuzzleDate == today {
		stats.CurrentStreak = streak
	}
	return stats
}

func isFiveLetters(word string) bool {
	if len(word) != WordLength {
		return false
	}
	for i := range word {
		if word[i] < 'a' || word[i] > 'z' {
			return false
		}
	}
	return true
}

func evaluateGuess(answer, guess string) EvaluatedGuess {
	result := EvaluatedGuess{Word: guess}
	remaining := [26]int{}

	for i := 0; i < WordLength; i++ {
		result.Tiles[i].Letter = guess[i]
		if guess[i] == answer[i] {
			result.Tiles[i].State = Correct
		} else {
			remaining[answer[i]-'a']++
		}
	}
	for i := 0; i < WordLength; i++ {
		if result.Tiles[i].State == Correct {
			continue
		}
		letter := guess[i] - 'a'
		if remaining[letter] > 0 {
			result.Tiles[i].State = Present
			remaining[letter]--
		}
	}
	return result
}
