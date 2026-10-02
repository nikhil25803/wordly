package db

import (
	"errors"
	"time"
)

var ErrHistoryFinished = errors.New("game is already finished")

func GetOrCreateHistory(userID, puzzleID int) (History, error) {
	_, err := DB.Exec(
		`INSERT INTO history (user_id, puzzle_id, guessed, attempts, played_at)
		 VALUES (?, ?, FALSE, 0, ?)
		 ON CONFLICT(user_id, puzzle_id) DO NOTHING`,
		userID, puzzleID, time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return History{}, err
	}

	var history History
	err = DB.QueryRow(
		`SELECT id, user_id, puzzle_id, guessed, attempts, CAST(played_at AS TEXT)
		 FROM history WHERE user_id = ? AND puzzle_id = ?`,
		userID, puzzleID,
	).Scan(
		&history.ID, &history.UserID, &history.PuzzleID,
		&history.Guessed, &history.Attempts, &history.PlayedAt,
	)
	return history, err
}

func GetGuesses(historyID int) ([]Guess, error) {
	rows, err := DB.Query(
		`SELECT id, history_id, attempt_number, word
		 FROM guesses WHERE history_id = ? ORDER BY attempt_number`,
		historyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guesses []Guess
	for rows.Next() {
		var guess Guess
		if err := rows.Scan(&guess.ID, &guess.HistoryID, &guess.AttemptNumber, &guess.Word); err != nil {
			return nil, err
		}
		guesses = append(guesses, guess)
	}
	return guesses, rows.Err()
}

func SaveGuess(historyID int, word string, won bool) (History, error) {
	tx, err := DB.Begin()
	if err != nil {
		return History{}, err
	}
	defer tx.Rollback()

	var history History
	err = tx.QueryRow(
		`SELECT id, user_id, puzzle_id, guessed, attempts, CAST(played_at AS TEXT)
		 FROM history WHERE id = ?`,
		historyID,
	).Scan(
		&history.ID, &history.UserID, &history.PuzzleID,
		&history.Guessed, &history.Attempts, &history.PlayedAt,
	)
	if err != nil {
		return History{}, err
	}
	if history.Guessed || history.Attempts >= 6 {
		return History{}, ErrHistoryFinished
	}

	history.Attempts++
	if _, err := tx.Exec(
		`INSERT INTO guesses (history_id, attempt_number, word) VALUES (?, ?, ?)`,
		history.ID, history.Attempts, word,
	); err != nil {
		return History{}, err
	}
	if _, err := tx.Exec(
		`UPDATE history SET attempts = ?, guessed = ? WHERE id = ?`,
		history.Attempts, won, history.ID,
	); err != nil {
		return History{}, err
	}
	history.Guessed = won

	if err := tx.Commit(); err != nil {
		return History{}, err
	}
	return history, nil
}

func GetResults(userID int) ([]Result, error) {
	rows, err := DB.Query(
		`SELECT CAST(p.puzzle_date AS TEXT), h.guessed, h.attempts
		 FROM history h JOIN puzzles p ON p.id = h.puzzle_id
		 WHERE h.user_id = ? AND (h.guessed = TRUE OR h.attempts >= 6)
		 ORDER BY p.puzzle_date`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var result Result
		if err := rows.Scan(&result.PuzzleDate, &result.Won, &result.Attempts); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func ResetUserGames(userID int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`DELETE FROM guesses WHERE history_id IN (SELECT id FROM history WHERE user_id = ?)`,
		userID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM history WHERE user_id = ?", userID); err != nil {
		return err
	}
	return tx.Commit()
}
