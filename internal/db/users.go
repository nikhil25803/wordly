package db

import (
	"errors"
	"os/user"
	"strings"
	"time"
)

func GetCurrentUser() (User, error) {
	osUser, err := user.Current()
	if err != nil {
		return User{}, err
	}

	username := strings.TrimSpace(osUser.Username)
	if username == "" {
		return User{}, errors.New("current username is empty")
	}

	_, err = DB.Exec(
		`INSERT INTO users (username, registered_at) VALUES (?, ?)
		 ON CONFLICT(username) DO NOTHING`,
		username, time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return User{}, err
	}

	var currentUser User
	err = DB.QueryRow(
		`SELECT id, username, CAST(registered_at AS TEXT) FROM users WHERE username = ?`,
		username,
	).Scan(&currentUser.ID, &currentUser.Username, &currentUser.RegisteredAt)

	return currentUser, err
}
