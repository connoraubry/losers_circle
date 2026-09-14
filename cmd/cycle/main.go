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
	upcoming := flag.Bool("upcoming", false, "check whether results in the current and next week's remaining games would create a cycle")
	flag.Parse()

	if *upcoming && *maxWeek > 0 {
		fmt.Fprintln(os.Stderr, "error: -upcoming cannot be combined with -week")
		os.Exit(1)
	}

	path := filepath.Join(*dir, fmt.Sprintf("%d.json", *season))
	s, err := nfldata.LoadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	upcomingGames := upcomingGames(s.Games)
	s = nfldata.FilterMaxWeek(s, *maxWeek)

	nfldata.PrintSeasonStats(os.Stdout, s)
	g := cycle.Build(s)
	printLongestCycles(os.Stdout, g, s.Teams, *progress)

	if *upcoming {
		printPotentialCycles(os.Stdout, g, upcomingGames)
	}
}

// printLongestCycles prints each team's longest cycle (a loop of teams
// where each beat the next, and the last beat the first), longest first.
func printLongestCycles(w *os.File, g *cycle.Graph, teamNames []string, progress bool) {
	var byTeam map[string][]string
	if progress {
		byTeam = g.LongestByTeamProgress(os.Stderr, time.Second)
	} else {
		byTeam = g.LongestByTeam()
	}

	teams := append([]string(nil), teamNames...)
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

// upcomingGames returns the unplayed games in the earliest week with any
// unplayed game, plus the following week. It returns nil if every game has
// been played.
func upcomingGames(games []nfldata.Game) []nfldata.Game {
	minWeek := -1
	for _, gm := range games {
		if gm.Played() {
			continue
		}
		if minWeek == -1 || gm.Week < minWeek {
			minWeek = gm.Week
		}
	}
	if minWeek == -1 {
		return nil
	}

	var result []nfldata.Game
	for _, gm := range games {
		if gm.Played() {
			continue
		}
		if gm.Week == minWeek || gm.Week == minWeek+1 {
			result = append(result, gm)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Week != result[j].Week {
			return result[i].Week < result[j].Week
		}
		if result[i].Date != result[j].Date {
			return result[i].Date < result[j].Date
		}
		return result[i].Time < result[j].Time
	})
	return result
}

// printPotentialCycles reports unplayed games in the current and next week
// whose result would close a new cycle in the win graph.
func printPotentialCycles(w *os.File, g *cycle.Graph, games []nfldata.Game) {
	if len(games) == 0 {
		fmt.Fprintln(w, "\nupcoming games that could create a cycle: season complete, no upcoming games")
		return
	}

	minWeek, maxWeek := games[0].Week, games[0].Week
	for _, gm := range games {
		if gm.Week < minWeek {
			minWeek = gm.Week
		}
		if gm.Week > maxWeek {
			maxWeek = gm.Week
		}
	}

	pcs := g.PotentialCycles(games)

	fmt.Fprintf(w, "\nupcoming games that could create a cycle (weeks %d-%d):\n", minWeek, maxWeek)
	if len(pcs) == 0 {
		fmt.Fprintln(w, "  none")
		return
	}
	for _, pc := range pcs {
		fmt.Fprintf(w, "  week %d, %s @ %s (%s): if %s wins, closes %s\n",
			pc.Game.Week, pc.Game.AwayTeam, pc.Game.HomeTeam, pc.Game.Date,
			pc.Winner, strings.Join(pc.Cycle, " -> "))
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
