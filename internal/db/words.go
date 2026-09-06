package db

func GetWordCount() (int, error) {
	var count int

	err := DB.QueryRow(
		"SELECT COUNT(*) FROM words",
	).Scan(&count)

	return count, err
}
