package tools

import (
	"fmt"
	"html/template"
	"math/rand"

	"github.com/connoraubry/losers_circle/src/graph"
	log "github.com/sirupsen/logrus"
)

func matchupSelection(week Week) MatchupSection {
	ms := MatchupSection{}
	ms.Controls.ValidWeeks = getValidWeeks()
	ms.Controls.Week = week.Week
	ms.Controls.Year = week.Year

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
	ms.Controls.ValidWeeks = getValidWeeks()
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

func GetLongestCycle(season []Week) []string {
	log.Info("Entering GetLongestCycle")

	gg := graph.New()

	for _, week := range season {
		for _, game := range week.Games {

			var cnx graph.Connection

			if game.HomeScore > game.AwayScore {
				cnx = graph.NewCnx(game.Home, game.Away)
			} else if game.AwayScore > game.HomeScore {
				cnx = graph.NewCnx(game.Away, game.Home)
			}

			gg.AddConnection(cnx)
		}
	}
	log.Info("Evalulating Graph")
	gg.EvaluateCycles()

	log.Info("Finding Longest Cycle")
	var longestCycle []string
	for _, cycle := range gg.NodeToCycle {
		if len(cycle) > len(longestCycle) {
			longestCycle = cycle
		}
	}
	return longestCycle
}

func GetGraph(year, week int, season []Week) HTMLGraph {
	log.Info("Entering GetGraph")

	longestCycle := LoadLongestCycle2(year, week)

	// longestCycle = GetLongestCycle(season[:week])
	return CycleToHTMLGraph(longestCycle)
}
func CycleToHTMLGraph(longestCycle []string) HTMLGraph {
	log.Debug("Entering CycleToHTMLGraph")
	teamToAbbr := GetTeamToAbbr()
	teamString := ""
	stFmt := "<span><span class='%s'>%s</span> ></span> "

	if len(longestCycle) > 0 {
		for _, name := range longestCycle {
			teamString = teamString + fmt.Sprintf(stFmt, teamToAbbr[name], name)
		}
		teamString = teamString + fmt.Sprintf("<span class='%s'>%s</span>", teamToAbbr[longestCycle[0]], longestCycle[0])
	} else {
		teamString = "No cycles found!"
	}

	return HTMLGraph{GraphString: template.HTML(teamString)}
}

// func dummyGraph() HTMLGraph {
// 	g := HTMLGraph{}

// 	teams := GetAllTeams()
// 	rand.Shuffle(len(teams), func(i, j int) { teams[i], teams[j] = teams[j], teams[i] })
// 	teamString := ""
// 	for i := 0; i < len(teams); i++ {
// 		teamString = teamString + fmt.Sprintf("%s > ", teams[i].Name)
// 	}
// 	teamString = teamString + teams[0].Name
// 	g.GraphString = template.HTML(teamString)
// 	return g
// }

func getValidWeeks() []int {
	var vs []int

	for i := 1; i <= 18; i++ {
		vs = append(vs, i)
	}
	return vs
}
