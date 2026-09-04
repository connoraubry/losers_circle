package nfldata

import (
	"fmt"
	"io"
)

// FilterMaxWeek returns a copy of s containing only games with Week <=
// maxWeek. A maxWeek of 0 or less returns s unchanged.
func FilterMaxWeek(s Season, maxWeek int) Season {
	if maxWeek <= 0 {
		return s
	}
	filtered := s
	filtered.Games = nil
	for _, g := range s.Games {
		if g.Week <= maxWeek {
			filtered.Games = append(filtered.Games, g)
		}
	}
	return filtered
}

// PrintSeasonStats writes summary statistics about a season to w.
func PrintSeasonStats(w io.Writer, s Season) {
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

	fmt.Fprintf(w, "season %d\n", s.SeasonYear)
	fmt.Fprintf(w, "  teams:    %d\n", len(s.Teams))
	fmt.Fprintf(w, "  games:    %d (played: %d, unplayed: %d, ties: %d)\n", len(s.Games), played, unplayed, ties)
	if statusFinal != played {
		fmt.Fprintf(w, "  note: %d games have status=final but missing scores (not counted as played)\n", statusFinal-played)
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
	fmt.Fprintf(w, "  undefeated teams (with a win): %d\n", undefeated)
	fmt.Fprintf(w, "  winless teams (with a loss):   %d\n", winless)
	fmt.Fprintln(w)
}
