// Command cycle runs the losers circle cycle for a single NFL season.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"losers_circle/internal/cycle"
	"losers_circle/internal/nfldata"
)

func main() {
	dir := flag.String("dir", "data", "directory containing season JSON files")
	season := flag.Int("season", defaultSeason(time.Now()), "NFL season year to load (alias: -year)")
	flag.IntVar(season, "year", *season, "NFL season year to load (alias: -season)")
	maxWeek := flag.Int("week", 0, "only include games through this week (0 = all weeks)")
	progress := flag.Bool("progress", false, "print elapsed time and nodes analyzed while solving")
	flag.Parse()

	path := filepath.Join(*dir, fmt.Sprintf("%d.json", *season))
	s, err := nfldata.LoadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	s = nfldata.FilterMaxWeek(s, *maxWeek)

	nfldata.PrintSeasonStats(os.Stdout, s)
	printLongestCycles(os.Stdout, s, *progress)
}

// printLongestCycles prints each team's longest cycle (a loop of teams
// where each beat the next, and the last beat the first), longest first.
func printLongestCycles(w *os.File, s nfldata.Season, progress bool) {
	g := cycle.Build(s)

	var byTeam map[string][]string
	if progress {
		byTeam = g.LongestByTeamProgress(os.Stderr, time.Second)
	} else {
		byTeam = g.LongestByTeam()
	}

	teams := append([]string(nil), s.Teams...)
	sort.Slice(teams, func(i, j int) bool {
		li, lj := len(byTeam[teams[i]]), len(byTeam[teams[j]])
		if li != lj {
			return li > lj
		}
		return teams[i] < teams[j]
	})

	fmt.Fprintln(w, "longest cycles:")
	for _, t := range teams {
		c := byTeam[t]
		if c == nil {
			fmt.Fprintf(w, "  %-4s (0): none\n", t)
			continue
		}
		fmt.Fprintf(w, "  %-4s (%d): %s\n", t, len(c), strings.Join(c, " -> "))
	}
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
