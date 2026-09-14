// Package cyclesvg renders a cycle.Graph as an SVG image: teams as nodes,
// "beat" relationships as directed edges. Edges belonging to a distinct
// cycle are colored per cycle, with a legend describing each one.
package cyclesvg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"

	"losers_circle/internal/cycle"
)

// palette is cycled through for each distinct cycle found, longest cycle
// first, so colors stay stable and distinguishable across a typical
// season's handful of cycles.
var palette = []string{
	"#e6194b", "#3cb44b", "#4363d8", "#f58231", "#911eb4",
	"#42d4f4", "#f032e6", "#bfef45", "#fabed4", "#469990",
	"#dcbeff", "#9a6324", "#800000", "#aaffc3", "#808000",
}

// neutralColor is used for edges/nodes not on any cycle.
const neutralColor = "#999999"

// fallbackCycleColor is used for edges that lie on some cycle (per
// cycle.Graph.InCycle) but aren't covered by any distinct cycle derived
// from byTeam, since only the longest cycle per team is known.
const fallbackCycleColor = "#555555"

// Options configures Render.
type Options struct {
	// OnlyCycles restricts the rendered graph to teams/edges that lie on an
	// existing cycle, instead of every team/edge in g.
	OnlyCycles bool
}

// Render writes an SVG of g to path. byTeam is g.LongestByTeam()'s result,
// used to derive the set of distinct cycles for coloring and the legend.
func Render(path string, g *cycle.Graph, byTeam map[string][]string, opts Options) error {
	cycles := dedupCycles(byTeam)

	ctx := context.Background()
	gv, err := graphviz.New(ctx)
	if err != nil {
		return fmt.Errorf("start graphviz: %w", err)
	}
	defer gv.Close()

	graph, err := gv.Graph(graphviz.WithDirectedType(graphviz.Directed))
	if err != nil {
		return fmt.Errorf("create graph: %w", err)
	}
	defer graph.Close()
	graph.SetRankDir(cgraph.LRRank)

	nodes := map[string]*cgraph.Node{}
	getNode := func(name string) (*cgraph.Node, error) {
		if n, ok := nodes[name]; ok {
			return n, nil
		}
		n, err := graph.CreateNodeByName(name)
		if err != nil {
			return nil, err
		}
		n.SetShape(cgraph.EllipseShape)
		nodes[name] = n
		return n, nil
	}

	usedTeams := map[string]bool{}
	for _, e := range g.Edges() {
		winner, loser := e[0], e[1]
		inCycle := g.InCycle(winner, loser)
		if opts.OnlyCycles && !inCycle {
			continue
		}

		wn, err := getNode(winner)
		if err != nil {
			return fmt.Errorf("create node %s: %w", winner, err)
		}
		ln, err := getNode(loser)
		if err != nil {
			return fmt.Errorf("create node %s: %w", loser, err)
		}
		usedTeams[winner] = true
		usedTeams[loser] = true

		edge, err := graph.CreateEdgeByName(winner+"->"+loser, wn, ln)
		if err != nil {
			return fmt.Errorf("create edge %s->%s: %w", winner, loser, err)
		}
		edge.SetColor(edgeColor(cycles, winner, loser, inCycle))
	}

	for name, n := range nodes {
		if !usedTeams[name] {
			continue
		}
		if !inAnyCycle(cycles, name) && !g.NodeInCycle(name) {
			n.SetColor(neutralColor)
		}
	}

	if err := addLegend(graph, cycles); err != nil {
		return fmt.Errorf("add legend: %w", err)
	}

	var buf bytes.Buffer
	if err := gv.Render(ctx, graph, graphviz.SVG, &buf); err != nil {
		return fmt.Errorf("render svg: %w", err)
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// addLegend adds a cluster subgraph listing each distinct cycle's color and
// team sequence, so the colors in the main graph can be identified.
func addLegend(graph *cgraph.Graph, cycles []distinctCycle) error {
	if len(cycles) == 0 {
		return nil
	}

	legend, err := graph.CreateSubGraphByName("cluster_legend")
	if err != nil {
		return err
	}
	legend.SetLabel("Cycles")

	for i, dc := range cycles {
		n, err := legend.CreateNodeByName(fmt.Sprintf("legend_%d", i))
		if err != nil {
			return err
		}
		n.SetShape(cgraph.BoxShape)
		n.SetStyle(cgraph.FilledNodeStyle)
		n.SetColor(dc.color)
		n.SetFontColor("#ffffff")
		n.SetLabel(strings.Join(dc.teams, " -> "))
	}
	return nil
}

// distinctCycle is one deduplicated cycle used for coloring, with its
// canonical (rotation-normalized) team order and assigned color.
type distinctCycle struct {
	teams []string
	color string
}

// dedupCycles collapses byTeam's per-team longest cycles into a distinct
// set, keyed by canonical rotation, and assigns each a palette color,
// longest cycle first (ties broken alphabetically for determinism).
func dedupCycles(byTeam map[string][]string) []distinctCycle {
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

	cycles := make([]distinctCycle, 0, len(seen))
	for _, canon := range seen {
		cycles = append(cycles, distinctCycle{teams: canon})
	}
	sort.Slice(cycles, func(i, j int) bool {
		if len(cycles[i].teams) != len(cycles[j].teams) {
			return len(cycles[i].teams) > len(cycles[j].teams)
		}
		return strings.Join(cycles[i].teams, ",") < strings.Join(cycles[j].teams, ",")
	})
	for i := range cycles {
		cycles[i].color = palette[i%len(palette)]
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

// edgeColor returns the color for edge winner->loser: the color of the
// longest distinct cycle containing that consecutive pair (cycles is
// sorted longest-first), or fallbackCycleColor/neutralColor depending on
// whether the edge lies on some cycle at all.
func edgeColor(cycles []distinctCycle, winner, loser string, inCycle bool) string {
	for _, dc := range cycles {
		for i, t := range dc.teams {
			if t == winner && dc.teams[(i+1)%len(dc.teams)] == loser {
				return dc.color
			}
		}
	}
	if inCycle {
		return fallbackCycleColor
	}
	return neutralColor
}

// inAnyCycle reports whether team appears in any distinct cycle.
func inAnyCycle(cycles []distinctCycle, team string) bool {
	for _, dc := range cycles {
		for _, t := range dc.teams {
			if t == team {
				return true
			}
		}
	}
	return false
}
