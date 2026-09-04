// Package nfldata loads scraped NFL season data (teams and game results).
package nfldata

// Game is a single NFL game result.
type Game struct {
	Season    int    `json:"season"`
	Week      int    `json:"week"`
	GameType  string `json:"game_type"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeScore *int   `json:"home_score"`
	AwayScore *int   `json:"away_score"`
	Status    string `json:"status"`
	// Winner is the winning team's abbreviation, or "" for a tie or a
	// game that hasn't been played yet.
	Winner string `json:"winner"`
}

// Played reports whether the game has a final result. The scraper's
// "status" field is not reliable on its own (some unplayed future games
// are marked "final" with null scores), so this also requires both
// scores to be present.
func (g Game) Played() bool {
	return g.Status == "final" && g.HomeScore != nil && g.AwayScore != nil
}

// Tie reports whether a played game ended in a tie.
func (g Game) Tie() bool {
	return g.Played() && g.Winner == ""
}

// Loser returns the losing team's abbreviation and true, or "", false if
// the game hasn't been played or ended in a tie.
func (g Game) Loser() (string, bool) {
	switch {
	case !g.Played() || g.Winner == "":
		return "", false
	case g.Winner == g.HomeTeam:
		return g.AwayTeam, true
	default:
		return g.HomeTeam, true
	}
}

// Season is one year's worth of scraped schedule/results.
type Season struct {
	SeasonYear int      `json:"season"`
	Teams      []string `json:"teams"`
	Games      []Game   `json:"games"`
}
