package main

import (
	"flag"
	"fmt"
	"slices"

	"github.com/connoraubry/losers_circle/src/algorithms"
	"github.com/connoraubry/losers_circle/src/tools"
	log "github.com/sirupsen/logrus"
)

var (
	// genType = flag.String("type", "games", "Type of generation [html, games]")
	year = flag.Int("year", tools.DefaultYear(), "Year to generate")
	algo = flag.String("alg", "greedy", "Algorithm to run [greedy,]")
)

func main() {
	flag.Parse()

	switch *algo {
	case "greedy":
		RunGreedyAlgorithm(*year)
	default:
		fmt.Println("bad!")
	}

}

func RunGreedyAlgorithm(year int) {
	log.WithField("year", year).Info("Parsing Year")

	data, err := tools.LoadFile2(year)
	if err != nil {
		log.Fatal(err)
	}

	weekToCycle := make(map[string][]string)

	overWriteFlag := false
	var maxLongest []string
	lastLongest := 0

	last_week := 18
	for _, week := range data.Weeks {
		last_week = week.Week
		if !week.Completed {
			break
		}
		// last_week = week.Week
	}

	var lastMissing []string

	for i := 1; i <= last_week; i++ {
		log.WithFields(
			log.Fields{"week": i, "lastLongest": lastLongest}).Info("Evaluating week.")
		if overWriteFlag {
			weekToCycle[fmt.Sprintf("%02d", i+1)] = maxLongest
			continue
		}

		cycle := algorithms.GetLongestCycleDFS(data.Weeks[:i])
		weekToCycle[fmt.Sprintf("%02d", i+1)] = cycle

		if len(cycle) == 32 {
			overWriteFlag = true
			maxLongest = cycle
		}

		lastLongest = len(cycle)
		// fmt.Printf("%+v\n", cycle)
		fmt.Printf("Longest cycle (length %d)\n", len(cycle))
		for _, team := range cycle {
			fmt.Println(team)
		}

		fmt.Printf("\nTeams not in cycle:\n")
		lastMissing = nil
		for _, team := range tools.GetTeamsFromWeek(data.Weeks[0]) {
			if !slices.Contains(cycle, team) {
				lastMissing = append(lastMissing, team)
				fmt.Println(team)
			}
		}
		fmt.Println(lastMissing)

	}

	fmt.Println("Finding next potential games to increase circle:")
	algorithms.EvalutateNextMinimumGames(data.Weeks, 7, lastMissing, lastLongest)
}
