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
	"losers_circle/internal/cyclesvg"
	"losers_circle/internal/nfldata"
)

func main() {
	dir := flag.String("dir", "data", "directory containing season JSON files")
	season := flag.Int("season", defaultSeason(time.Now()), "NFL season year to load (alias: -year)")
	flag.IntVar(season, "year", *season, "NFL season year to load (alias: -season)")
	maxWeek := flag.Int("week", 0, "only include games through this week (0 = all weeks)")
	progress := flag.Bool("progress", false, "print elapsed time and nodes analyzed while solving")
	upcoming := flag.Bool("upcoming", false, "check whether results in the current and next week's remaining games would create a cycle")
	sweep := flag.Bool("sweep", false, "exhaustively check every combination of results in the next full week for new cycles")
	graphSVG := flag.Bool("graph", false, "write the full win/loss graph to graph.svg")
	cycleGraphSVG := flag.Bool("cycle-graph", false, "write only the teams/edges on an existing cycle to cycles.svg")
	flag.Parse()

	if *upcoming && *maxWeek > 0 {
		fmt.Fprintln(os.Stderr, "error: -upcoming cannot be combined with -week")
		os.Exit(1)
	}
	if *sweep && *maxWeek > 0 {
		fmt.Fprintln(os.Stderr, "error: -sweep cannot be combined with -week")
		os.Exit(1)
	}

	path := filepath.Join(*dir, fmt.Sprintf("%d.json", *season))
	s, err := nfldata.LoadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	upcomingGames := upcomingGames(s.Games)
	nextWeekGames := nextFullWeekGames(s.Games)
	s = nfldata.FilterMaxWeek(s, *maxWeek)

	nfldata.PrintSeasonStats(os.Stdout, s)
	g := cycle.Build(s)
	printLongestCycles(os.Stdout, g, s.Teams, *progress)

	if *upcoming {
		printPotentialCycles(os.Stdout, g, upcomingGames)
	}
	if *sweep {
		printSweep(os.Stdout, g, nextWeekGames)
	}

	if *graphSVG || *cycleGraphSVG {
		byTeam := g.LongestByTeam()
		if *graphSVG {
			if err := cyclesvg.Render("graph.svg", g, byTeam, cyclesvg.Options{}); err != nil {
				fmt.Fprintln(os.Stderr, "error writing graph.svg:", err)
				os.Exit(1)
			}
			fmt.Fprintln(os.Stdout, "\nwrote graph.svg")
		}
		if *cycleGraphSVG {
			if err := cyclesvg.Render("cycles.svg", g, byTeam, cyclesvg.Options{OnlyCycles: true}); err != nil {
				fmt.Fprintln(os.Stderr, "error writing cycles.svg:", err)
				os.Exit(1)
			}
			fmt.Fprintln(os.Stdout, "wrote cycles.svg")
		}
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

// nextFullWeekGames returns the unplayed games in the next week that has no
// played games at all yet (as opposed to the current, possibly
// partially-played week). It returns nil if there's no such week in the
// data.
func nextFullWeekGames(games []nfldata.Game) []nfldata.Game {
	hasPlayed := map[int]bool{}
	hasUnplayed := map[int]bool{}
	for _, gm := range games {
		if gm.Played() {
			hasPlayed[gm.Week] = true
		} else {
			hasUnplayed[gm.Week] = true
		}
	}

	minWeek := -1
	for w := range hasUnplayed {
		if minWeek == -1 || w < minWeek {
			minWeek = w
		}
	}
	if minWeek == -1 {
		return nil
	}
	week := minWeek
	if hasPlayed[minWeek] {
		week++
	}

	var result []nfldata.Game
	for _, gm := range games {
		if gm.Week == week && !gm.Played() {
			result = append(result, gm)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Date != result[j].Date {
			return result[i].Date < result[j].Date
		}
		return result[i].Time < result[j].Time
	})
	return result
}

// maxSweepGroupingsShown caps how many distinct groupings printSweep lists,
// largest cycle first, to keep the report scannable in a highly-entangled
// week.
const maxSweepGroupingsShown = 15

// printSweep reports every distinct new cycle that could result from some
// combination of outcomes in games (expected to be a single week's slate).
func printSweep(w *os.File, g *cycle.Graph, games []nfldata.Game) {
	if len(games) == 0 {
		fmt.Fprintln(w, "\nweek sweep: season complete, no full upcoming week")
		return
	}
	week := games[0].Week

	groupings, total, newCycleCombos, err := g.Sweep(games)
	if err != nil {
		fmt.Fprintf(w, "\nweek %d sweep: %v\n", week, err)
		return
	}

	fmt.Fprintf(w, "\nweek %d sweep: %d distinct new cycles possible across %d result combinations (%d combinations produce a new cycle):\n",
		week, len(groupings), total, newCycleCombos)
	if len(groupings) == 0 {
		fmt.Fprintln(w, "  none")
		return
	}
	shown := groupings
	if len(shown) > maxSweepGroupingsShown {
		shown = shown[:maxSweepGroupingsShown]
	}
	for _, sg := range shown {
		causes := make([]string, len(sg.Causes))
		for i, c := range sg.Causes {
			causes[i] = fmt.Sprintf("%s beats %s", c.Winner, c.Loser)
		}
		fmt.Fprintf(w, "  %s (%d teams): if %s\n", strings.Join(sg.Cycle, " -> "), len(sg.Cycle), strings.Join(causes, ", "))
	}
	if len(groupings) > maxSweepGroupingsShown {
		fmt.Fprintf(w, "  ...and %d more\n", len(groupings)-maxSweepGroupingsShown)
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
