// Command stats prints summary statistics about the loaded NFL season data.
package main

import (
	"flag"
	"fmt"
	"os"

	"losers_circle/internal/nfldata"
)

func main() {
	dir := flag.String("dir", "data", "directory containing season JSON files")
	flag.Parse()

	seasons, err := nfldata.LoadDir(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(seasons) == 0 {
		fmt.Fprintf(os.Stderr, "no season files found in %s\n", *dir)
		os.Exit(1)
	}

	for _, s := range seasons {
		printSeasonStats(s)
	}
}

func printSeasonStats(s nfldata.Season) {
	var played, unplayed, ties int
	wins := map[string]int{}
	losses := map[string]int{}

	for _, g := range s.Games {
		if !g.Played() {
			unplayed++
			continue
		}
		played++
		if g.Tie() {
			ties++
			continue
		}
		if loser, ok := g.Loser(); ok {
			losses[loser]++
			wins[g.Winner]++
		}
	}

	statusFinal := 0
	for _, g := range s.Games {
		if g.Status == "final" {
			statusFinal++
		}
	}

	fmt.Printf("season %d\n", s.SeasonYear)
	fmt.Printf("  teams:    %d\n", len(s.Teams))
	fmt.Printf("  games:    %d (played: %d, unplayed: %d, ties: %d)\n", len(s.Games), played, unplayed, ties)
	if statusFinal != played {
		fmt.Printf("  note: %d games have status=final but missing scores (not counted as played)\n", statusFinal-played)
	}

	undefeated, winless := 0, 0
	for _, t := range s.Teams {
		if losses[t] == 0 && wins[t] > 0 {
			undefeated++
		}
		if wins[t] == 0 && losses[t] > 0 {
			winless++
		}
	}
	fmt.Printf("  undefeated teams (with a win): %d\n", undefeated)
	fmt.Printf("  winless teams (with a loss):   %d\n", winless)
	fmt.Println()
}
