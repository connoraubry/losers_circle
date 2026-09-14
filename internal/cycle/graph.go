// Package cycle finds cycles in the head-to-head win graph of a season:
// a cycle is a sequence of teams where each beat the next, and the last
// beat the first.
package cycle

import (
	"fmt"
	"io"
	"math/bits"
	"sort"
	"strings"
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

// Teams returns the season's team abbreviations.
func (g *Graph) Teams() []string {
	return append([]string(nil), g.teams...)
}

// Edges returns every "beat" edge in the graph as (winner, loser) pairs.
func (g *Graph) Edges() [][2]string {
	var edges [][2]string
	for i, nbs := range g.adj {
		for _, j := range nbs {
			edges = append(edges, [2]string{g.teams[i], g.teams[j]})
		}
	}
	return edges
}

// InCycle reports whether the edge winner->loser lies on some existing
// cycle. An edge inside a strongly connected component of size >= 2 always
// closes a cycle, since the component guarantees a path back from loser to
// winner.
func (g *Graph) InCycle(winner, loser string) bool {
	wi, ok := g.index[winner]
	if !ok {
		return false
	}
	li, ok := g.index[loser]
	if !ok {
		return false
	}
	scc := g.sccID[wi]
	return scc == g.sccID[li] && g.sccSize[scc] >= 2
}

// NodeInCycle reports whether team lies on some existing cycle, i.e.
// shares a strongly connected component of size >= 2 with another team.
func (g *Graph) NodeInCycle(team string) bool {
	i, ok := g.index[team]
	if !ok {
		return false
	}
	return g.sccSize[g.sccID[i]] >= 2
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

// maxSweepGames caps how many games Sweep will exhaustively enumerate
// (2^n combinations); a real NFL week never comes close to this.
const maxSweepGames = 20

// SweepCause is one game result that contributes an edge to a
// SweepGrouping's example cycle.
type SweepCause struct {
	Game   nfldata.Game
	Winner string
	Loser  string
}

// SweepGrouping is a distinct set of teams (identified by which teams
// appear in Cycle) that could become mutually entangled in a new cycle
// depending on how a batch of unplayed games turns out, but aren't already
// entangled in g. Cycle is one example loop through those teams (the last
// beats the first, matching Longest's convention). Causes lists the subset
// of the swept games whose results form the edges of that example cycle;
// edges already present in g are omitted since they aren't caused by the
// sweep.
type SweepGrouping struct {
	Cycle  []string
	Causes []SweepCause
}

// Sweep enumerates every combination of outcomes for the given unplayed
// games (2^n combinations, where n is the number of games with recognized
// teams) and reports each distinct new set of teams that could become
// mutually entangled in a cycle, which g alone doesn't already show, along
// with the number of combinations actually enumerated. It returns an error
// if games is larger than can be swept exhaustively.
func (g *Graph) Sweep(games []nfldata.Game) ([]SweepGrouping, int, error) {
	var edges []sweepEdge
	for _, gm := range games {
		hi, ok := g.index[gm.HomeTeam]
		if !ok {
			continue
		}
		ai, ok := g.index[gm.AwayTeam]
		if !ok {
			continue
		}
		edges = append(edges, sweepEdge{hi, ai, gm})
	}
	if len(edges) == 0 {
		return nil, 0, nil
	}
	if len(edges) > maxSweepGames {
		return nil, 0, fmt.Errorf("too many games to sweep exhaustively (%d games, max %d)", len(edges), maxSweepGames)
	}

	baseline := map[string]bool{}
	for _, members := range groupByComponent(g.teams, g.sccID) {
		if len(members) >= 2 {
			baseline[groupKey(g.teams, members)] = true
		}
	}

	found := make(map[string]SweepGrouping)
	total := 1 << len(edges)
	for mask := 0; mask < total; mask++ {
		adj := make([][]int, len(g.teams))
		copy(adj, g.adj)
		for i, e := range edges {
			wi, li := e.homeIdx, e.awayIdx
			if mask&(1<<i) != 0 {
				wi, li = e.awayIdx, e.homeIdx
			}
			adj[wi] = append(append([]int(nil), adj[wi]...), li)
		}

		sccID := tarjanSCC(adj)
		sizes := sccSizes(sccID)
		for _, members := range groupByComponent(g.teams, sccID) {
			if sizes[sccID[members[0]]] < 2 {
				continue
			}
			key := groupKey(g.teams, members)
			if baseline[key] {
				continue
			}
			if _, ok := found[key]; ok {
				continue
			}
			cycle := findCycle(adj, sccID, sccID[members[0]], members[0])
			if cycle == nil {
				continue
			}
			teams := make([]string, len(cycle))
			for i, idx := range cycle {
				teams[i] = g.teams[idx]
			}
			found[key] = SweepGrouping{Cycle: teams, Causes: sweepCauses(g, edges, mask, cycle)}
		}
	}

	result := make([]SweepGrouping, 0, len(found))
	for _, grouping := range found {
		result = append(result, grouping)
	}
	sort.Slice(result, func(i, j int) bool {
		if len(result[i].Cycle) != len(result[j].Cycle) {
			return len(result[i].Cycle) > len(result[j].Cycle)
		}
		return strings.Join(result[i].Cycle, ",") < strings.Join(result[j].Cycle, ",")
	})
	return result, total, nil
}

// groupByComponent buckets team indices by their component id (from comp,
// as returned by tarjanSCC), keyed by component id.
func groupByComponent(teams []string, comp []int) map[int][]int {
	groups := make(map[int][]int)
	for i := range teams {
		groups[comp[i]] = append(groups[comp[i]], i)
	}
	return groups
}

// groupKey returns a stable key identifying a set of team indices,
// independent of order.
func groupKey(teams []string, members []int) string {
	names := make([]string, len(members))
	for i, idx := range members {
		names[i] = teams[idx]
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}

// findCycle returns one directed cycle through comp starting and ending at
// start, using only edges within comp, or nil if none is found (shouldn't
// happen for a component of size >= 2, since it's strongly connected).
func findCycle(adj [][]int, sccID []int, comp, start int) []int {
	visited := map[int]bool{start: true}
	path := []int{start}

	var dfs func(cur int) bool
	dfs = func(cur int) bool {
		for _, nb := range adj[cur] {
			if sccID[nb] != comp {
				continue
			}
			if nb == start {
				return true
			}
			if visited[nb] {
				continue
			}
			visited[nb] = true
			path = append(path, nb)
			if dfs(nb) {
				return true
			}
			path = path[:len(path)-1]
			visited[nb] = false
		}
		return false
	}
	if dfs(start) {
		return path
	}
	return nil
}

// sweepEdge is one unplayed game under consideration by Sweep, with its
// teams pre-resolved to indices.
type sweepEdge struct {
	homeIdx, awayIdx int
	game             nfldata.Game
}

// sweepCauses returns the subset of edges, as resolved by mask, whose
// result forms an edge of cycle that isn't already present in g (i.e. the
// swept results that actually cause this cycle to close).
func sweepCauses(g *Graph, edges []sweepEdge, mask int, cycle []int) []SweepCause {
	var causes []SweepCause
	for i := range cycle {
		from, to := cycle[i], cycle[(i+1)%len(cycle)]
		if containsInt(g.adj[from], to) {
			continue
		}
		for j, e := range edges {
			wi, li := e.homeIdx, e.awayIdx
			if mask&(1<<j) != 0 {
				wi, li = e.awayIdx, e.homeIdx
			}
			if wi == from && li == to {
				causes = append(causes, SweepCause{Game: e.game, Winner: g.teams[wi], Loser: g.teams[li]})
				break
			}
		}
	}
	return causes
}

// containsInt reports whether v appears in s.
func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
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
