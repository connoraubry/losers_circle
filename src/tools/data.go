package tools

import (
	"fmt"
	"math/rand"
)

func matchupSelection(week Week) MatchupSection {
	ms := MatchupSection{}
	ms.Controls.ValidYears = getValidWeeks()

	teamToAbbr := GetTeamToAbbr()

	for _, game := range week.Games {

		team1 := Team{Name: teamToAbbr[game.Away]}
		team2 := Team{Name: teamToAbbr[game.Home]}

		m := Matchup{
			Team1: team1,
			Team2: team2,
		}
		ms.Matchups = append(ms.Matchups, m)
	}
	return ms
}

func dummyMatchupSection() MatchupSection {
	ms := MatchupSection{}
	ms.Controls.ValidYears = getValidWeeks()
	ms.Matchups = dummyMatchups()
	return ms
}
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
func dummyGraph() Graph {
	g := Graph{}

	teams := GetAllTeams()
	rand.Shuffle(len(teams), func(i, j int) { teams[i], teams[j] = teams[j], teams[i] })
	teamString := ""
	for i := 0; i < len(teams); i++ {
		teamString = teamString + fmt.Sprintf("%s > ", teams[i].Name)
	}
	teamString = teamString + teams[0].Name
	g.GraphString = teamString
	return g
}
func getValidWeeks() []string {
	var vs []string

	for i := 1; i <= 18; i++ {
		vs = append(vs, fmt.Sprintf("%d", i))
	}
	return vs
}
