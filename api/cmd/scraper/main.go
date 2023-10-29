package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/connoraubry/losers_circle/backend/tools/db"
	"github.com/connoraubry/losers_circle/backend/tools/scraper"
	log "github.com/sirupsen/logrus"
)

var (
	logLevel = flag.String("logLevel", "info", "Level for log outputs [debug, info, warning, error]")
	year     = flag.Int("year", time.Now().Year(), "Year to scrape")
	month    = flag.Int("month", 0, "Month to scrape (0 to scrape all months)")

	dbSkip = flag.Bool("skipDB", false, "Don't use database")
	dbHost = flag.String("db-host", "localhost", "Host of db")
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {

	flag.Parse()

	opts := scraper.Config{
		UseDB: !*dbSkip,
		DBOpt: db.Options{
			Host: *dbHost,
		},
	}

	s := scraper.New(opts)
	s.ScrapeYear(*year)

	fmt.Println(s.Games)
}
