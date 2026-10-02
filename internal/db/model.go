package db

type Word struct {
	ID   int
	Word string
}

type Puzzle struct {
	ID         int
	WordID     int
	PuzzleDate string
	Word       string
}

type User struct {
	ID           int
	Username     string
	RegisteredAt string
}

type History struct {
	ID       int
	UserID   int
	PuzzleID int
	Guessed  bool
	Attempts int
	PlayedAt string
}

type Guess struct {
	ID            int
	HistoryID     int
	AttemptNumber int
	Word          string
}

type Result struct {
	PuzzleDate string
	Won        bool
	Attempts   int
}
