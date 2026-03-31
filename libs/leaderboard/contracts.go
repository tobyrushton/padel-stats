package leaderboard

// Entry is an API-safe leaderboard row.
type Entry struct {
	Rank            int    `json:"rank"`
	PlayerID        int64  `json:"playerId" format:"int64"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Username        string `json:"username"`
	ScoreDifference int    `json:"scoreDifference"`
	Wins            int    `json:"wins"`
	Losses          int    `json:"losses"`
	GamesPlayed     int    `json:"gamesPlayed"`
}

// EntryRecord is a persistence-layer leaderboard row.
type EntryRecord struct {
	PlayerID        int64
	FirstName       string
	LastName        string
	Username        string
	ScoreDifference int
	Wins            int
	Losses          int
	GamesPlayed     int
}
