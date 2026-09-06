package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func ConnectDatabase() error {
	db, err := sql.Open("sqlite3", "./wordly.db")
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
