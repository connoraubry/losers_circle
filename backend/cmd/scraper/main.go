package main

import (
	"fmt"

	"github.com/connoraubry/losers_circle/backend/tools/scraper"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)

}

func main() {
	fmt.Println("vim-go")
	s := scraper.New(true)
	s.ScrapeYear(2023)

	fmt.Println(s.Games)
}
