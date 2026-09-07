package db

func GetWordCount() (int, error) {
	var count int

	err := DB.QueryRow(
		"SELECT COUNT(*) FROM words",
	).Scan(&count)

	return count, err
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

func AddPuzzle(word string, date string) error {
	var wordID int
	err := DB.QueryRow(
		"SELECT id FROM words WHERE word = ?",
		word,
	).Scan(&wordID)

	if err != nil {
		return err
	}

	_, err = DB.Exec(
		"INSERT INTO puzzles (word_id, puzzle_date) VALUES (?, ?)",
		wordID, date,
	)

	return err
}
