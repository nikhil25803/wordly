package db

import (
	"os/user"
)

func GetCurrentUser() (string, error) {
	osUser, err := user.Current()
	if err != nil {
		return "", err
	}

	return osUser.Username, nil
}
