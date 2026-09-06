package db

func createTables() error {
	if err := createWordsTable(); err != nil {
		return err
	}
	return nil
}

func createWordsTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS words (
			id INTEGER PRIMARY KEY,
			word TEXT NOT NULL UNIQUE
		);
	`)

	return err
}
