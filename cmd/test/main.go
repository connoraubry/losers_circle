package main

import (
	"fmt"

	"github.com/connoraubry/losers_circle/src/graph"
	"github.com/connoraubry/losers_circle/src/tools"
)

func main() {
	fmt.Println("hi")

	// data, _ := tools.LoadFile2(2024)
	// numGames := getNumGames(data.Weeks)

	// for i := 0; i < numGames; i++ {
	// 	start := time.Now()
	//
	// 	gg := makeGraphByGames(data.Weeks, i)
	// 	gg.EvaluateCycles()
	//
	// 	duration := time.Since(start)
	//
	// 	fmt.Printf("Seconds elapsed for %d games: %v\n", i, duration)
	// 	if duration > 5*time.Second {
	// 		break
	// 	}
	//
	// }

	data2, _ := tools.LoadFile2(2025)
	fmt.Println(getNumGames(data2.Weeks))
}

func getNumGames(weeks []tools.Week) int {
	count := 0
	for _, week := range weeks {
		for _, game := range week.Games {
			if game.Complete {
				count += 1
			}
		}
	}
	return count
}

func makeGraphByGames(weeks []tools.Week, numGames int) *graph.Graph {

	g := graph.New()
	count := 0
outerloop:
	for _, wk := range weeks {
		for _, game := range wk.Games {

			if count >= numGames {
				break outerloop
			}
			count += 1

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
