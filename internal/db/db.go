package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

//go:embed wordly.db
var initialDatabase []byte

func ConnectDatabase() error {
	path, err := databasePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(path, initialDatabase, 0o600); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	DB = db

	return createTables()
}

func databasePath() (string, error) {
	if path := os.Getenv("WORDLY_DB"); path != "" {
		return path, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wordly", "wordly.db"), nil
}
