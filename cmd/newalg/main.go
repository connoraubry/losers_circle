package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/connoraubry/losers_circle/src/stems"
	"github.com/connoraubry/losers_circle/src/stems/graph"
	"github.com/connoraubry/losers_circle/src/tools"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	year     = flag.Int("year", time.Now().Year(), "year to generate")
	week     = flag.Int("week", 0, "week to parse through")
	logLevel = flag.String("logLevel", "DEBUG", "loglevel")
)

func parseLog() {
	switch *logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	flag.Parse()
	parseLog()

	if *year == 2024 {
		log.Fatal("No games for the current year!")
	}

	ParseYear(*year, *week)

}

func ParseYear(year, week int) {
	log.WithField("year", year).Info("Parsing Year")

	weeks := tools.LoadFile(year, 0)

	fmt.Println(weeks)
	// tools.GetLongestCycle(weeks)
	if week != 0 {
		weeks = weeks[:week]

	}

	g := graph.New()
	for _, week := range weeks {
		for _, game := range week.Games {
			if !game.Complete {
				continue
			}

			if game.HomeScore > game.AwayScore {
				g.AddConnection(game.Home, game.Away)
			} else if game.AwayScore > game.HomeScore {
				g.AddConnection(game.Away, game.Home)
			}
		}
	}

	for nodeIdx := range g.Nodes {
		fmt.Println(nodeIdx, g.Nodes[nodeIdx].Name)
	}

	stems := stems.New()
	stems.Init(g)

	// for second := 2; second < 17; second += 1 {
	// 	first := second
	// 	g.Stems.ProcessNextLevel(first, second)
	// 	first += 1
	// 	g.Stems.ProcessNextLevel(first, second)
	// }

	stems.ProcessNextLevel(2, 2)  //3
	stems.ProcessNextLevel(3, 2)  //4
	stems.ProcessNextLevel(3, 3)  //5
	stems.ProcessNextLevel(4, 3)  //6
	stems.ProcessNextLevel(5, 5)  //9
	stems.ProcessNextLevel(9, 6)  //14
	stems.ProcessNextLevel(14, 3) //16
	stems.ProcessNextLevel(9, 9)  //17
	stems.ProcessNextLevel(16, 17)

	// g.Stems.ProcessNextLevel(9, 9)
	// g.Stems.Levels[2].Print()

	countBits := 0

	for _, level := range stems.Levels {
		for _, start := range level.Starts {
			for _, end := range start.Ends {
				countBits += len(end.Bitmasks)
			}
		}
	}
	fmt.Printf("%v bitmasks found\n", countBits)
	logrus.WithField("Bits", countBits).Info("Total size")

	// g.PrintCycles()

	v := make(map[int]int)
	stemsPerLevel := make(map[int]int)

	for levelIdx, level := range stems.Levels {
		for _, start := range level.Starts {
			for _, end := range start.Ends {

				v[len(end.Bitmasks)] += 1
				stemsPerLevel[levelIdx] += len(end.Bitmasks)
			}
		}
	}
	fmt.Println(v)
	fmt.Println(stemsPerLevel)
}
