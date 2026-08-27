package main

import (
	"flag"

	"github.com/connoraubry/losers_circle/src/scraper"
	"github.com/connoraubry/losers_circle/src/tools"
	log "github.com/sirupsen/logrus"
)

var (
	year  = flag.Int("year", tools.DefaultYear(), "Year to scrape")
	start = flag.Int("start", 1, "Start of range")
	end   = flag.Int("end", 18, "Last week to scrape")
	force = flag.Bool("force", false, "Force a re-scrape")
	all   = flag.Bool("all", false, "Scrape all years, even if not completed")
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	flag.Parse()

	//get old data
	data, err := tools.LoadFile2(*year)
	if err != nil {
		var err2 error
		data, err2 = tools.InitFile(*year)
		if err2 != nil {
			log.Fatal(err2)
		}
	}

	opts := scraper.Config{
		Year:  *year,
		Start: *start,
		End:   *end,
		Data:  data,
		Force: *force,
		All:   *all,
	}

	updateData := scraper.Scrape(opts)

	tools.SaveFile2(updateData)
}
