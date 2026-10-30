package cycle

import (
	"sort"
	"strings"
)

// palette is cycled through for each distinct cycle found, longest cycle
// first, so colors stay stable and distinguishable across a typical
// season's handful of cycles.
var palette = []string{
	"#e6194b", "#3cb44b", "#4363d8", "#f58231", "#911eb4",
	"#42d4f4", "#f032e6", "#bfef45", "#fabed4", "#469990",
	"#dcbeff", "#9a6324", "#800000", "#aaffc3", "#808000",
}

// NeutralColor is used for edges/nodes not on any distinct cycle.
const NeutralColor = "#999999"

// DistinctCycle is one deduplicated cycle, with its canonical
// (rotation-normalized) team order and assigned color.
type DistinctCycle struct {
	Teams []string
	Color string
}

// DedupCycles collapses byTeam's per-team longest cycles into a distinct
// set, keyed by canonical rotation, and assigns each a palette color,
// longest cycle first (ties broken alphabetically for determinism).
func DedupCycles(byTeam map[string][]string) []DistinctCycle {
	seen := map[string][]string{}
	for _, c := range byTeam {
		if len(c) == 0 {
			continue
		}
		key, canon := canonicalize(c)
		if _, ok := seen[key]; !ok {
			seen[key] = canon
		}
	}

	cycles := make([]DistinctCycle, 0, len(seen))
	for _, canon := range seen {
		cycles = append(cycles, DistinctCycle{Teams: canon})
	}
	sort.Slice(cycles, func(i, j int) bool {
		if len(cycles[i].Teams) != len(cycles[j].Teams) {
			return len(cycles[i].Teams) > len(cycles[j].Teams)
		}
		return strings.Join(cycles[i].Teams, ",") < strings.Join(cycles[j].Teams, ",")
	})
	for i := range cycles {
		cycles[i].Color = palette[i%len(palette)]
	}
	return cycles
}

// canonicalize rotates cycle to start at its lexicographically smallest
// team, so equivalent loops (same sequence, different start point) produce
// the same key.
func canonicalize(c []string) (key string, canon []string) {
	minIdx := 0
	for i, t := range c {
		if t < c[minIdx] {
			minIdx = i
		}
	}
	canon = make([]string, len(c))
	for i := range c {
		canon[i] = c[(minIdx+i)%len(c)]
	}
	return strings.Join(canon, ","), canon
}

// CyclesWithEdge returns the indexes into cycles of every distinct cycle
// containing the consecutive pair winner->loser, in cycles' order. An edge
// on no cycle (nil result) is neutral; an edge on several gets one colored
// line per returned index.
func CyclesWithEdge(cycles []DistinctCycle, winner, loser string) []int {
	var idxs []int
	for ci, dc := range cycles {
		for i, t := range dc.Teams {
			if t == winner && dc.Teams[(i+1)%len(dc.Teams)] == loser {
				idxs = append(idxs, ci)
				break
			}
		}
	}
	return idxs
}
