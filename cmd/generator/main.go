package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/connoraubry/losers_circle/src/tools"
	log "github.com/sirupsen/logrus"
)

var (
	genType = flag.String("type", "", "Type of generation [html, games]")
	year    = flag.Int("year", time.Now().Year(), "year to generate")
)

func main() {
	flag.Parse()

	switch *genType {
	case "games":
		fmt.Println("parsing games")
		ParseYear(*year)
	case "html":
		fmt.Println("parsing html")
	default:
		fmt.Println("Not declared!")
	}

	// output_dir := "./public"
	// GenerateAll(output_dir)

	test := tools.LoadLongestCycle(*year)
	fmt.Println(test)
}

func ParseYear(year int) {
	log.WithField("year", year).Info("Parsing Year")

	weeks := tools.LoadFile(year, 0)

	weekToCycle := make(map[string][]string)

	overWriteFlag := false
	var maxLongest []string

	for i := 1; i < len(weeks); i++ {

		if overWriteFlag {
			weekToCycle[fmt.Sprintf("%02d", i+1)] = maxLongest
			continue
		}

		cycle := tools.GetLongestCycle(weeks[:i+1])
		weekToCycle[fmt.Sprintf("%02d", i+1)] = cycle

		if len(cycle) == 32 {
			overWriteFlag = true
			maxLongest = cycle
		}
	}

	tools.SaveLongestCycles(year, weekToCycle)
}
