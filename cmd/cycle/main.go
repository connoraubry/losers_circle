// Command cycle runs the losers circle cycle for a single NFL season.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"losers_circle/internal/nfldata"
)

func main() {
	dir := flag.String("dir", "data", "directory containing season JSON files")
	season := flag.Int("season", defaultSeason(time.Now()), "NFL season year to load (alias: -year)")
	flag.IntVar(season, "year", *season, "NFL season year to load (alias: -season)")
	maxWeek := flag.Int("week", 0, "only include games through this week (0 = all weeks)")
	flag.Parse()

	path := filepath.Join(*dir, fmt.Sprintf("%d.json", *season))
	s, err := nfldata.LoadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	s = nfldata.FilterMaxWeek(s, *maxWeek)

	nfldata.PrintSeasonStats(os.Stdout, s)
}

// defaultSeason returns the current NFL season year for the given time: the
// current year, unless it's January or February, in which case the season
// that started the previous year is still ongoing (playoffs/offseason).
func defaultSeason(t time.Time) int {
	if t.Month() == time.January || t.Month() == time.February {
		return t.Year() - 1
	}
	return t.Year()
}
