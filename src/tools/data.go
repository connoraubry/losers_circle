package tools

import (
	"math/rand"
)

func dummyMatchups() []Matchup {

	teams := GetAllTeams()
	rand.Shuffle(len(teams), func(i, j int) { teams[i], teams[j] = teams[j], teams[i] })

	var matchups []Matchup
	for i := 0; i < len(teams); i += 2 {
		newMatch := Matchup{
			Team1: Team{Name: teams[i].Abbr},
			Team2: Team{Name: teams[i+1].Abbr},
		}
		matchups = append(matchups, newMatch)
	}
	return matchups
}
