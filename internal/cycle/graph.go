// Package cycle finds cycles in the head-to-head win graph of a season:
// a cycle is a sequence of teams where each beat the next, and the last
// beat the first.
package cycle

import (
	"fmt"
	"io"
	"math/bits"
	"sync/atomic"
	"time"

	"losers_circle/internal/nfldata"
)

// Graph is a directed graph of teams where an edge from A to B means A
// beat B in a head-to-head game.
type Graph struct {
	teams []string
	index map[string]int
	adj   [][]int // adj[i] = distinct indices j such that i beat j

	sccID   []int // sccID[i] = strongly connected component id of team i
	sccSize []int // sccSize[c] = number of teams in component c
}

// Build constructs a Graph from a season's played, non-tie games. A split
// season series (each team beat the other once) produces edges in both
// directions, since every game contributes its own edge.
func Build(s nfldata.Season) *Graph {
	g := &Graph{
		teams: append([]string(nil), s.Teams...),
		index: make(map[string]int, len(s.Teams)),
	}
	for i, t := range g.teams {
		g.index[t] = i
	}
	g.adj = make([][]int, len(g.teams))

	seen := make([]map[int]bool, len(g.teams))
	for i := range seen {
		seen[i] = map[int]bool{}
	}

	for _, game := range s.Games {
		loser, ok := game.Loser()
		if !ok {
			continue
		}
		wi, ok := g.index[game.Winner]
		if !ok {
			continue
		}
		li, ok := g.index[loser]
		if !ok {
			continue
		}
		if seen[wi][li] {
			continue
		}
		seen[wi][li] = true
		g.adj[wi] = append(g.adj[wi], li)
	}

	g.sccID = tarjanSCC(g.adj)
	g.sccSize = sccSizes(g.sccID)

	return g
}

// Longest returns the largest cycle containing team, as an ordered list of
// team abbreviations (team beat the next, ..., the last beat team). It
// returns nil if team is in no cycle.
func (g *Graph) Longest(team string) []string {
	return g.longest(team, nil)
}

// longest is Longest's implementation. If visits is non-nil, it's
// incremented once per node the search visits, for progress reporting; the
// extra nil check costs nothing measurable when visits is nil.
func (g *Graph) longest(team string, visits *int64) []string {
	start, ok := g.index[team]
	if !ok {
		return nil
	}

	// A team can only be in a cycle if it shares a strongly connected
	// component with at least one other team; a singleton component means
	// no path leads back to it.
	scc := g.sccID[start]
	n := g.sccSize[scc]
	if n < 2 {
		return nil
	}

	var best []int

	visited := uint32(1) << start
	path := make([]int, 0, n)
	path = append(path, start)

	var dfs func(cur int, visited uint32, path []int)
	dfs = func(cur int, visited uint32, path []int) {
		if visits != nil {
			atomic.AddInt64(visits, 1)
		}
		// Even using every unvisited team in this component, we
		// couldn't beat the best cycle found so far.
		if len(path)+(n-bits.OnesCount32(visited)) <= len(best) {
			return
		}
		for _, nb := range g.adj[cur] {
			// Cross-component edges can never close back into a
			// cycle with start.
			if g.sccID[nb] != scc {
				continue
			}
			if nb == start {
				if len(path) > len(best) {
					best = append([]int(nil), path...)
				}
				continue
			}
			bit := uint32(1) << nb
			if visited&bit != 0 {
				continue
			}
			dfs(nb, visited|bit, append(path, nb))
		}
	}
	dfs(start, visited, path)

	if best == nil {
		return nil
	}
	result := make([]string, len(best))
	for i, idx := range best {
		result[i] = g.teams[idx]
	}
	return result
}

// ShortestPath returns the shortest directed path from `from` to `to`
// (inclusive of both endpoints), or nil if no path exists.
func (g *Graph) ShortestPath(from, to string) []string {
	fi, ok := g.index[from]
	if !ok {
		return nil
	}
	ti, ok := g.index[to]
	if !ok {
		return nil
	}

	prev := make([]int, len(g.teams))
	for i := range prev {
		prev[i] = -1
	}
	visited := make([]bool, len(g.teams))
	visited[fi] = true
	queue := []int{fi}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == ti {
			break
		}
		for _, nb := range g.adj[cur] {
			if visited[nb] {
				continue
			}
			visited[nb] = true
			prev[nb] = cur
			queue = append(queue, nb)
		}
	}
	if !visited[ti] {
		return nil
	}

	var path []int
	for at := ti; at != -1; at = prev[at] {
		path = append(path, at)
		if at == fi {
			break
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	result := make([]string, len(path))
	for i, idx := range path {
		result[i] = g.teams[idx]
	}
	return result
}

// PotentialCycle describes an unplayed game where Winner beating Loser
// would create a new cycle in the win graph. Cycle is the resulting loop
// (Loser beat the next, ..., the last beat Loser), matching the ordering
// convention of Longest.
type PotentialCycle struct {
	Game   nfldata.Game
	Winner string
	Loser  string
	Cycle  []string
}

// PotentialCycles checks each unplayed game in games against g and returns
// one PotentialCycle per possible outcome (home win, away win) that would
// close a new cycle: a result W-beats-L creates a cycle exactly when a path
// L -> ... -> W already exists in g, since the new edge W->L closes it.
// Already-played games are skipped.
func (g *Graph) PotentialCycles(games []nfldata.Game) []PotentialCycle {
	var result []PotentialCycle
	for _, game := range games {
		if game.Played() {
			continue
		}
		for _, outcome := range [2][2]string{
			{game.HomeTeam, game.AwayTeam},
			{game.AwayTeam, game.HomeTeam},
		} {
			winner, loser := outcome[0], outcome[1]
			path := g.ShortestPath(loser, winner)
			if path == nil {
				continue
			}
			result = append(result, PotentialCycle{Game: game, Winner: winner, Loser: loser, Cycle: path})
		}
	}
	return result
}

// LongestByTeam returns each team's longest cycle (as returned by Longest),
// keyed by team abbreviation. Teams with no cycle are omitted.
func (g *Graph) LongestByTeam() map[string][]string {
	result := make(map[string][]string)
	for _, t := range g.teams {
		if c := g.Longest(t); c != nil {
			result[t] = c
		}
	}
	return result
}

// LongestByTeamProgress behaves like LongestByTeam, but periodically writes
// elapsed time and the number of search nodes visited so far to w (useful
// for long-running searches on larger graphs). A final summary line is
// always written when the search completes.
func (g *Graph) LongestByTeamProgress(w io.Writer, interval time.Duration) map[string][]string {
	var visits int64
	start := time.Now()

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(w, "  ...%.1fs elapsed, %d nodes analyzed\n", time.Since(start).Seconds(), atomic.LoadInt64(&visits))
			case <-done:
				return
			}
		}
	}()

	result := make(map[string][]string)
	for _, t := range g.teams {
		if c := g.longest(t, &visits); c != nil {
			result[t] = c
		}
	}
	close(done)

	fmt.Fprintf(w, "  done in %.1fs, %d nodes analyzed\n", time.Since(start).Seconds(), atomic.LoadInt64(&visits))
	return result
}
