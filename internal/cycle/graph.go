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
	start, ok := g.index[team]
	if !ok {
		return nil
	}
	best := g.longestIdx(start, nil, nil)
	if best == nil {
		return nil
	}
	return g.namesOf(best)
}

// namesOf converts team indices to their abbreviations.
func (g *Graph) namesOf(idx []int) []string {
	result := make([]string, len(idx))
	for i, v := range idx {
		result[i] = g.teams[v]
	}
	return result
}

// longestIdx returns the largest cycle containing team index start, as an
// ordered list of team indices, or nil if start is in no cycle. If seed is
// non-nil, it must already be a valid cycle through start (typically
// another team's cached longest cycle, rotated to begin at start); it's
// used as the search's initial lower bound, both to prune faster and so
// that, when nothing longer exists, the result converges exactly onto seed
// rather than an arbitrary same-length alternative. If visits is non-nil,
// it's incremented once per node the search visits, for progress
// reporting; the extra nil check costs nothing measurable when visits is
// nil.
func (g *Graph) longestIdx(start int, seed []int, visits *int64) []int {
	// A team can only be in a cycle if it shares a strongly connected
	// component with at least one other team; a singleton component means
	// no path leads back to it.
	scc := g.sccID[start]
	n := g.sccSize[scc]
	if n < 2 {
		return nil
	}

	best := append([]int(nil), seed...)

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

	return best
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
// with the number of combinations actually enumerated and how many of those
// combinations produce at least one new cycle. It returns an error if games
// is larger than can be swept exhaustively.
func (g *Graph) Sweep(games []nfldata.Game) (groupings []SweepGrouping, total, newCycleCombos int, err error) {
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
		return nil, 0, 0, nil
	}
	if len(edges) > maxSweepGames {
		return nil, 0, 0, fmt.Errorf("too many games to sweep exhaustively (%d games, max %d)", len(edges), maxSweepGames)
	}

	baseline := map[string]bool{}
	for _, members := range groupByComponent(g.teams, g.sccID) {
		if len(members) >= 2 {
			baseline[groupKey(g.teams, members)] = true
		}
	}

	found := make(map[string]SweepGrouping)
	total = 1 << len(edges)
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
		maskHasNewCycle := false
		for _, members := range groupByComponent(g.teams, sccID) {
			if sizes[sccID[members[0]]] < 2 {
				continue
			}
			if baseline[groupKey(g.teams, members)] {
				continue
			}
			cycle := findCycle(adj, sccID, sccID[members[0]], members[0])
			if cycle == nil {
				continue
			}
			maskHasNewCycle = true
			// Key on the teams actually shown in Cycle, not the full SCC
			// membership: the component can include peripheral teams that
			// findCycle's simple loop doesn't pass through, and different
			// masks can pull in different peripherals around the same
			// minimal cycle, which would otherwise register as distinct
			// groupings despite printing identically.
			key := groupKey(g.teams, cycle)
			if _, ok := found[key]; ok {
				continue
			}
			teams := make([]string, len(cycle))
			for i, idx := range cycle {
				teams[i] = g.teams[idx]
			}
			found[key] = SweepGrouping{Cycle: teams, Causes: sweepCauses(g, edges, mask, cycle)}
		}
		if maskHasNewCycle {
			newCycleCombos++
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
	return result, total, newCycleCombos, nil
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
//
// Teams sharing a strongly connected component reuse and warm-start each
// other's searches (see longestIdx): once a cycle is found that spans an
// entire component, that's provably the longest possible cycle in it (a
// simple cycle can't exceed its component's size), so every team on it is
// resolved without any further search. Teams on a shorter shared cycle
// still warm-start from it, pruning faster and converging on that same
// cycle when nothing longer exists for them, rather than an arbitrary
// same-length alternative.
func (g *Graph) LongestByTeam() map[string][]string {
	return g.longestByTeam(nil)
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

	result := g.longestByTeam(&visits)
	close(done)

	fmt.Fprintf(w, "  done in %.1fs, %d nodes analyzed\n", time.Since(start).Seconds(), atomic.LoadInt64(&visits))
	return result
}

// sccCycleCache tracks, for one strongly connected component, the longest
// cycle found so far among its teams, and whether that cycle has been
// proven maximal for the whole component (its length equals the
// component's size).
type sccCycleCache struct {
	best  []int
	final bool
}

// longestByTeam is LongestByTeam's implementation; see its docs for the
// per-component reuse strategy. visits is optional, as in longestIdx.
func (g *Graph) longestByTeam(visits *int64) map[string][]string {
	result := make(map[string][]string)
	caches := make(map[int]*sccCycleCache)

	for _, t := range g.teams {
		start, ok := g.index[t]
		if !ok {
			continue
		}
		scc := g.sccID[start]
		n := g.sccSize[scc]
		if n < 2 {
			continue
		}

		cache := caches[scc]
		if cache == nil {
			cache = &sccCycleCache{}
			caches[scc] = cache
		}

		if cache.final && containsInt(cache.best, start) {
			result[t] = g.namesOf(rotateTo(cache.best, start))
			continue
		}

		var seed []int
		if containsInt(cache.best, start) {
			seed = rotateTo(cache.best, start)
		}

		best := g.longestIdx(start, seed, visits)
		if best == nil {
			continue
		}
		result[t] = g.namesOf(best)

		if len(best) > len(cache.best) {
			cache.best = best
			if len(best) == n {
				cache.final = true
			}
		}
	}
	return result
}

// rotateTo rotates cycle, a closed loop of team indices, so it begins at
// start, which must already be a member.
func rotateTo(cycle []int, start int) []int {
	offset := 0
	for i, v := range cycle {
		if v == start {
			offset = i
			break
		}
	}
	rotated := make([]int, len(cycle))
	for i := range cycle {
		rotated[i] = cycle[(offset+i)%len(cycle)]
	}
	return rotated
}
