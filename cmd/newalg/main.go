package main

import (
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/connoraubry/losers_circle/src/db"
	"github.com/connoraubry/losers_circle/src/stems"
	"github.com/connoraubry/losers_circle/src/stems/graph"
	"github.com/connoraubry/losers_circle/src/tools"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	year     = flag.Int("year", time.Now().Year(), "year to generate")
	week     = flag.Int("week", 0, "week to parse through")
	logLevel = flag.String("logLevel", "DEBUG", "loglevel")

	firstLevel = flag.Int("start", 2, "first stem level")
	endLevel   = flag.Int("end", 2, "end stem level")

	numGames = flag.Int("games", 80, "num games to process")
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

	// ParseYear(*year, *week)
	// LoadInitSQL()
	// DoLevel(*firstLevel, *endLevel)
	// DoLevel2(*firstLevel, *endLevel)
	// for i := 2; i < 10; i++ {
	// 	fmt.Printf("Level %d:\n", i)
	// 	fmt.Printf("  ")
	// 	Validate(i)
	// 	fmt.Printf("  ")
	// 	GetNumCycles(i)
	// }
	LoadOneByOne(*numGames)
}

func GetNumCycles(level int) {
	s, err := db.New("./NFL_2023.db")
	if err != nil {
		log.Fatalf("Error opening db")
	}

	stems, err := s.GetStemsFromLevel(level)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for _, stem := range stems {

		if stem.EndID == stem.StartID {
			count += 1
		}

	}

	fmt.Printf("%v/%v are cycles\n", count, len(stems))
}

func Validate(level int) {
	s, err := db.New("./NFL_2023.db")
	if err != nil {
		log.Fatalf("Error opening db")
	}

	stems, err := s.GetStemsFromLevel(level)
	if err != nil {
		log.Fatal(err)
	}
	badCount := 0
	for _, stem := range stems {

		if GetNumOnes(stem.Mask) != level-1 {
			badCount += 1
		}

	}

	fmt.Printf("%v/%v are bad masks\n", badCount, len(stems))
}

func GetNumOnes(mask uint32) int {
	count := 0
	for mask > 0 {
		if mask&1 == 1 {
			count += 1
		}
		mask = mask >> 1
	}
	return count
}
func DoLevel2(startLevel, endLevel int) {

	s, err := db.New("./NFL_2023.db")
	if err != nil {
		log.Fatalf("Error creating new sql db: %v", err)
	}

	slv, err := s.GetStemsFromLevel(startLevel)
	if err != nil {
		log.Fatalf("Error getting stems from level %v: %v", startLevel, err)
	}

	slog.Info("Starting main loop")

	start := time.Now()
	bar := progressbar.Default(int64(len(slv)))

	count := 0
	for _, startStem := range slv {
		endStems, err := s.GetValidRows(endLevel, startStem.EndID, startStem.Mask)
		if err != nil {
			log.Fatal(err)
		}
		for _, endStem := range endStems {

			if startStem.Mask&endStem.Mask > 0 {
				continue
			}

			newStem := db.Stem{

				Level:     startLevel + endLevel - 1,
				Mask:      startStem.Mask | endStem.Mask,
				StartID:   startStem.StartID,
				EndID:     endStem.EndID,
				StartStem: startStem.Id,
				EndStem:   endStem.Id,
			}
			s.StemChan <- newStem
			count += 1
		}
		bar.Add(1)
	}

	fmt.Println("Execution time: ", time.Since(start))
	fmt.Println(count)

}

func DoLevel(startLevel, endLevel int) {

	s, err := db.New("./NFL_2023.db")
	if err != nil {
		log.Fatalf("Error creating new sql db: %v", err)
	}

	slv, err := s.GetStemsFromLevel(startLevel)
	if err != nil {
		log.Fatalf("Error getting stems from level %v: %v", startLevel, err)
	}
	var elv []db.Stem
	if startLevel == endLevel {
		elv = slv
	} else {
		elv, err = s.GetStemsFromLevel(endLevel)
		if err != nil {
			log.Fatalf("Error getting stems from level %v: %v", endLevel, err)
		}
	}

	slog.Info("Creating end map")

	endMap := make(map[int][]db.Stem)
	for _, entry := range elv {
		endMap[entry.StartID] = append(endMap[entry.StartID], entry)
	}

	startMap := make(map[int][]db.Stem)
	for _, entry := range slv {
		startMap[entry.StartID] = append(startMap[entry.StartID], entry)
	}

	slog.Info("Starting main loop")

	start := time.Now()
	bar := progressbar.Default(int64(len(slv)))
	count := 0

	for _, startStem := range slv {
		for _, endStem := range endMap[startStem.EndID] {

			if startStem.Mask&endStem.Mask > 0 {
				continue
			}

			newStem := db.Stem{

				Level:     startLevel + endLevel - 1,
				Mask:      startStem.Mask | endStem.Mask,
				StartID:   startStem.StartID,
				EndID:     endStem.EndID,
				StartStem: startStem.Id,
				EndStem:   endStem.Id,
			}
			s.StemChan <- newStem
			count += 1
		}
		bar.Add(1)
	}

	fmt.Println("Execution time: ", time.Since(start))
	fmt.Println(count)

}

func LoadInitSQL() {
	weeks := tools.LoadFile(2023, 0)

	g := graph.New()
	g.LoadFromWeeks(weeks)
	for nodeIdx := range g.Nodes {
		fmt.Println(nodeIdx, g.Nodes[nodeIdx].Name)
	}

	s, err := db.New("./NFL_2023.db")
	if err != nil {
		log.Fatalf("Error creating new sql db: %v", err)
	}

	for _, node := range g.Nodes {
		s.AddTeam(node.Name)
	}

	newTeams, err := s.GetTeams()
	if err != nil {
		log.Fatalf("could not get teams")
	}

	nameToId := make(map[string]int)
	for _, team := range newTeams {
		nameToId[team.Name] = team.Id
	}

	for _, node := range g.Nodes {
		for _, next := range node.Outgoing {
			nextName := g.Nodes[next].Name

			startID := nameToId[node.Name]
			endID := nameToId[nextName]
			mask := stems.IdToBitmask(endID)

			stem := db.Stem{
				Level:   2,
				Mask:    mask,
				StartID: startID,
				EndID:   endID,
			}

			s.StemChan <- stem
		}
	}

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

	// stems.ProcessNextLevel(2, 2)  //3
	// stems.ProcessNextLevel(3, 2)  //4
	// stems.ProcessNextLevel(3, 3)  //5
	// stems.ProcessNextLevel(4, 3)  //6
	// stems.ProcessNextLevel(5, 5)  //9
	// stems.ProcessNextLevel(9, 6)  //14
	// stems.ProcessNextLevel(14, 3) //16
	// stems.ProcessNextLevel(9, 9)  //17
	// stems.ProcessNextLevel(16, 17)

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

func LoadOneByOne(maxint int) {
	s, err := db.New("./NFL_2023.db")
	if err != nil {
		log.Fatalf("Error creating new sql db: %v", err)
	}

	weeks := tools.LoadFile(2023, 0)

	g := graph.New()
	g.LoadFromWeeks(weeks)
	for _, node := range g.Nodes {
		s.AddTeam(node.Name)
	}

	count := 0
	// pb := progressbar.Default(int64(maxint))
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

			start := time.Now()
			err = s.AddConnection(first, second)
			if err != nil {
				fmt.Println(err)
			}
			t := time.Since(start)

			fmt.Printf("Execution time for entry %v->%v %v: %v\n",
				first, second, count, t)

			count += 1

			if count > maxint {
				return
			}
		}
	}
}
