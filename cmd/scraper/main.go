package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/connoraubry/losers_circle/src/scraper"
	log "github.com/sirupsen/logrus"
)

var (
	year = flag.Int("year", time.Now().Year(), "Year to scrape")
	week = flag.Int("week", 0, "Week to scrape. (0 to scrape all weeks)")
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	flag.Parse()

	opts := scraper.Config{
		Week: *week,
	}

	s := scraper.New(opts)
	s.ScrapeYear(*year)

	fmt.Println(s.Games)
}
