package algorithms

import (
	"fmt"

	"github.com/connoraubry/losers_circle/src/graph"
	"github.com/connoraubry/losers_circle/src/tools"
)

func makeGraphFromGames(games []tools.Game) *graph.Graph {
	g := graph.New()
	for _, game := range games {
		if !game.Complete {
			continue
		}
		first := ""
		second := ""

		if game.HomeScore > game.AwayScore {
			first = game.Home
			second = game.Away
		} else if game.AwayScore > game.HomeScore {
			first = game.Away
			second = game.Home
		} else {
			continue
		}

		g.AddConnection(graph.NewCnx(first, second))
	}
	return g
}

func makeGraph(weeks []tools.Week) *graph.Graph {
	g := graph.New()
	for _, wk := range weeks {
		for _, game := range wk.Games {
			if !game.Complete {
				continue
			}
			first := ""
			second := ""

			if game.HomeScore > game.AwayScore {
				first = game.Home
				second = game.Away
			} else if game.AwayScore > game.HomeScore {
				first = game.Away
				second = game.Home
			} else {
				continue
			}

			g.AddConnection(graph.NewCnx(first, second))
		}
	}
	return g
}

func GetLongestCycleDFS(weeks []tools.Week) []string {
	gg := makeGraph(weeks)

	gg.EvaluateCycles()
	var longestCycle []string
	for _, cycle := range gg.NodeToCycle {
		if len(cycle) > len(longestCycle) {
			longestCycle = cycle
		}
	}
	return longestCycle
}

func EvalutateNextMinimumGames(weeks []tools.Week, lastFullWeek int, targetTeams []string, longest int) {

	//add future relevant games
	var futureGames []tools.Game
	for _, week := range weeks[lastFullWeek:] {
		for _, game := range week.Games {
			for _, team := range targetTeams {
				if game.Away == team || game.Home == team {
					futureGames = append(futureGames, game)
				}
			}
		}
	}

	type Res struct {
		Win  string
		Loss string
	}

	maxGames := 10

	for _, game := range futureGames {
		var results = []Res{
			{Win: game.Home, Loss: game.Away},
			{Win: game.Away, Loss: game.Home},
		}

		for _, result := range results {

			gg := makeGraph(weeks[:lastFullWeek])
			gg.AddConnection(graph.NewCnx(result.Win, result.Loss))
			gg.EvaluateCycles()
			var longestCycle []string
			for _, cycle := range gg.NodeToCycle {
				if len(cycle) > len(longestCycle) {
					longestCycle = cycle
				}
			}
			if len(longestCycle) > longest {
				fmt.Println("longest cycle",
					len(longestCycle),
					"Winner", result.Win,
					"Loser", result.Loss,
				)
				maxGames -= 1
				if maxGames < 0 {
					return
				}
			}
		}

	}
	// fmt.Println(futureGames)
}

// type Res struct {
// 	Win  string
// 	Loss string
// }
//
// func FastestTo32Recursive(games []tools.Game, futureGames []tools.Game) (bool, []Res) {
//
// 	game := futureGames[0]
//
// 	var results = []Res{
// 		{Win: game.Home, Loss: game.Away},
// 		{Win: game.Away, Loss: game.Home},
// 	}
//
// 	for _, res := range results {
// 		gg := makeGraphFromGames(games)
// 		gg.AddConnection(graph.NewCnx(res.Win, res.Loss))
// 		gg.EvaluateCycles()
// 		var longestCycle []string
// 		for _, cycle := range gg.NodeToCycle {
// 			if len(cycle) > len(longestCycle) {
// 				longestCycle = cycle
// 			}
// 		}
//
// 		if len(longestCycle) == 32 {
// 			return true, []Res{res}
// 		} else {
// 			ok, recRes := FastestTo32Recursive(append(games, game), futureGames[1:])
// 			if ok == true {
// 				return true, append([]Res{res}, recRes...)
// 			}
// 		}
//
// 	}
//
// 	return false, nil
//
// }
//
// func FastestTo32(weeks []tools.Week, lastFullWeek int, targetTeams []string) {
//
// 	var completedGames []tools.Game
//
// 	for _, week := range weeks {
// 		completedGames = append(completedGames, week.Games...)
// 		if !week.Completed {
// 			break
// 		}
// 	}
//
// 	var futureGames []tools.Game
// 	for _, week := range weeks[lastFullWeek:] {
// 		for _, game := range week.Games {
// 			for _, team := range targetTeams {
// 				if game.Away == team || game.Home == team {
// 					futureGames = append(futureGames, game)
// 				}
// 			}
// 		}
// 	}
//
//
//
// }
