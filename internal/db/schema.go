package db

func createTables() error {
	if err := createWordsTable(); err != nil {
		return err
	}

	if err := createPuzzlesTable(); err != nil {
		return err
	}

	if err := createUsersTable(); err != nil {
		return err
	}

	if err := createHistoryTable(); err != nil {
		return err
	}

	if err := createGuessesTable(); err != nil {
		return err
	}
	return nil
}

/*
	Table words {
	  id integer [primary key, not null]
	  word varchar [not null, unique]
	}
*/
func createWordsTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS words (
			id INTEGER PRIMARY KEY,
			word TEXT NOT NULL UNIQUE
		);
	`)

	return err
}

/*
	Table puzzles {
	id integer [primary key, not null]
	word_id integer [not null, ref: > words.id]
	puzzle_date date [not null, unique]
	}
*/

func createPuzzlesTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS puzzles (
			id INTEGER PRIMARY KEY,
			word_id INTEGER NOT NULL REFERENCES words(id),
			puzzle_date DATE NOT NULL UNIQUE
		);
	`)

	return err
}

/*
	Table users {
	id integer [primary key, not null]
	username varchar [not null, unique]
	registered_at timestamp [not null]
	}
*/

func createUsersTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			registered_at TIMESTAMP NOT NULL
		);
	`)

	return err
}

/*
	Table history {
	id integer [primary key, not null]
	user_id integer [not null, ref: > users.id]
	puzzle_id integer [not null, ref: > puzzles.id]
	guessed bool [not null]
	attempts integer [not null]
	played_at timestamp [not null]
	}
*/

func createHistoryTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			puzzle_id INTEGER NOT NULL REFERENCES puzzles(id),
			guessed BOOLEAN NOT NULL,
			attempts INTEGER NOT NULL,
			played_at TIMESTAMP NOT NULL
		);
	`)

	return err
}

/*
Table guesses {
id integer [primary key, not null]
history_id integer [not null, ref: > history.id]
attempt_number integer [not null]
word varchar [not null]
}
*/
func createGuessesTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS guesses (
			id INTEGER PRIMARY KEY,
			history_id INTEGER NOT NULL REFERENCES history(id),
			attempt_number INTEGER NOT NULL,
			word TEXT NOT NULL
		);
	`)

	return err
}
