package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	weeks := s.ScrapeYear(*year)

	fmt.Println(weeks)
	bytes, err := json.Marshal(weeks)
	if err != nil {
		log.Error("Error marshaling weeks:", err)
	}

	filename := GenFilename(*year, *week)
	EnsureDir(filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()

	f.Write(bytes)

	fmt.Println(s.Games)
}

func EnsureDir(path string) {
	newpath := filepath.Dir(path)
	os.MkdirAll(newpath, 0o755)
}

func GenFilename(year, week int) string {
	if week != 0 {
		return fmt.Sprintf("data/nfl/fragment/%d/week_%02d.json", year, week)
	}

	return fmt.Sprintf("data/nfl/full/%d.json", year)
}
