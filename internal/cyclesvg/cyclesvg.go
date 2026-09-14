// Package cyclesvg renders a cycle.Graph as an SVG image: teams as nodes,
// "beat" relationships as directed edges. Edges belonging to a distinct
// cycle are colored per cycle, with a legend describing each one.
package cyclesvg

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"os"
	"regexp"
	"sort"
	"strconv"
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

// neutralColor is used for edges/nodes not on any distinct cycle.
const neutralColor = "#999999"

// Options configures Render.
type Options struct {
	// OnlyCycles renders only the distinct cycles themselves - their
	// member teams and the edges that form each loop - instead of the
	// full win/loss graph.
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
	if opts.OnlyCycles {
		// A pure cycle draws as an actual, compact ring under circo,
		// instead of the long line dot/LR-rank produces for a simple loop.
		gv.SetLayout(graphviz.CIRCO)
	} else {
		graph.SetRankDir(cgraph.LRRank)
	}

	nodes := map[string]*cgraph.Node{}
	nodeColor := map[string]string{}
	getNode := func(name, color string) (*cgraph.Node, error) {
		if n, ok := nodes[name]; ok {
			if nodeColor[name] == neutralColor && color != neutralColor {
				nodeColor[name] = color
			}
			return n, nil
		}
		n, err := graph.CreateNodeByName(name)
		if err != nil {
			return nil, err
		}
		n.SetShape(cgraph.EllipseShape)
		nodes[name] = n
		nodeColor[name] = color
		return n, nil
	}
	addEdge := func(winner, loser, color string) error {
		wn, err := getNode(winner, color)
		if err != nil {
			return fmt.Errorf("create node %s: %w", winner, err)
		}
		ln, err := getNode(loser, color)
		if err != nil {
			return fmt.Errorf("create node %s: %w", loser, err)
		}
		edge, err := graph.CreateEdgeByName(winner+"->"+loser, wn, ln)
		if err != nil {
			return fmt.Errorf("create edge %s->%s: %w", winner, loser, err)
		}
		edge.SetColor(color)
		return nil
	}

	if opts.OnlyCycles {
		// Only the distinct cycles' own edges - nothing else in the graph.
		for _, dc := range cycles {
			for i, t := range dc.teams {
				next := dc.teams[(i+1)%len(dc.teams)]
				if err := addEdge(t, next, dc.color); err != nil {
					return err
				}
			}
		}
	} else {
		for _, e := range g.Edges() {
			winner, loser := e[0], e[1]
			if err := addEdge(winner, loser, edgeColor(cycles, winner, loser)); err != nil {
				return err
			}
		}
	}

	for name, n := range nodes {
		if nodeColor[name] == neutralColor {
			continue
		}
		n.SetColor(nodeColor[name])
		n.SetStyle(cgraph.FilledNodeStyle)
		n.SetFontColor("#ffffff")
	}

	var buf bytes.Buffer
	if err := gv.Render(ctx, graph, graphviz.SVG, &buf); err != nil {
		return fmt.Errorf("render svg: %w", err)
	}

	out, err := addLegend(buf.Bytes(), cycles)
	if err != nil {
		return fmt.Errorf("add legend: %w", err)
	}
	return os.WriteFile(path, out, 0o644)
}

// svgOpenTag matches the opening <svg width="..pt" height="..pt" of
// graphviz's SVG output, to recover its rendered size.
var svgOpenTag = regexp.MustCompile(`<svg\s+width="([\d.]+)pt"\s+height="([\d.]+)pt"`)

// addLegend prepends a compact, wrapping legend row (one swatch + label per
// distinct cycle) above svg, extending the canvas down by only as much
// height as the legend actually needs rather than leaving it wherever
// Graphviz would otherwise place a cluster. svg is left untouched if there
// are no cycles to show.
func addLegend(svg []byte, cycles []distinctCycle) ([]byte, error) {
	if len(cycles) == 0 {
		return svg, nil
	}

	loc := svgOpenTag.FindSubmatchIndex(svg)
	if loc == nil {
		return nil, fmt.Errorf("unrecognized graphviz svg output")
	}
	graphWidth, err := strconv.ParseFloat(string(svg[loc[2]:loc[3]]), 64)
	if err != nil {
		return nil, fmt.Errorf("parse graph width: %w", err)
	}
	graphHeight, err := strconv.ParseFloat(string(svg[loc[4]:loc[5]]), 64)
	if err != nil {
		return nil, fmt.Errorf("parse graph height: %w", err)
	}

	bodyStart := bytes.IndexByte(svg[loc[1]:], '>')
	if bodyStart < 0 {
		return nil, fmt.Errorf("unrecognized graphviz svg output")
	}
	bodyStart += loc[1] + 1
	bodyEnd := bytes.LastIndex(svg, []byte("</svg>"))
	if bodyEnd < 0 {
		return nil, fmt.Errorf("unrecognized graphviz svg output")
	}
	body := svg[bodyStart:bodyEnd]

	const (
		padding    = 10.0
		rowHeight  = 22.0
		swatchSize = 14.0
		labelGapX  = 6.0
		itemGapX   = 24.0
		fontSize   = 11.0
		charWidth  = 6.2 // rough average glyph width at fontSize, for wrapping
	)
	rowMaxWidth := graphWidth
	if rowMaxWidth < 300 {
		rowMaxWidth = 300
	}

	type placedItem struct {
		color, label string
		x, y         float64
	}
	var items []placedItem
	x, y := padding, padding
	rowContentWidth := 0.0
	for _, dc := range cycles {
		label := legendLabel(dc.teams)
		width := swatchSize + labelGapX + float64(len(label))*charWidth
		if x > padding && x+width > rowMaxWidth {
			x = padding
			y += rowHeight
		}
		items = append(items, placedItem{color: dc.color, label: label, x: x, y: y})
		x += width + itemGapX
		if x-itemGapX > rowContentWidth {
			rowContentWidth = x - itemGapX
		}
	}
	legendWidth := rowContentWidth + padding
	legendHeight := y + rowHeight + padding

	finalWidth := graphWidth
	if legendWidth > finalWidth {
		finalWidth = legendWidth
	}
	finalHeight := legendHeight + graphHeight
	graphXOffset := (finalWidth - graphWidth) / 2
	legendXOffset := (finalWidth - legendWidth) / 2

	var out bytes.Buffer
	fmt.Fprint(&out, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>`+"\n")
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%.2fpt" height="%.2fpt" viewBox="0.00 0.00 %.2f %.2f">`+"\n",
		finalWidth, finalHeight, finalWidth, finalHeight)

	fmt.Fprintf(&out, `<g transform="translate(%.2f,0)">`+"\n", legendXOffset)
	for _, it := range items {
		fmt.Fprintf(&out, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" stroke="none"/>`+"\n",
			it.x, it.y, swatchSize, swatchSize, it.color)
		fmt.Fprintf(&out, `<text x="%.2f" y="%.2f" font-family="Helvetica,Arial,sans-serif" font-size="%.0f" fill="#000000">%s</text>`+"\n",
			it.x+swatchSize+labelGapX, it.y+swatchSize-3, fontSize, html.EscapeString(it.label))
	}
	out.WriteString("</g>\n")

	fmt.Fprintf(&out, `<g transform="translate(%.2f,%.2f)">`+"\n", graphXOffset, legendHeight)
	out.Write(body)
	out.WriteString("</g>\n</svg>\n")

	return out.Bytes(), nil
}

// maxLegendTeams caps how many teams a legend label spells out before
// collapsing the middle, so a large cycle (e.g. spanning most of a league)
// doesn't blow up the legend's width.
const maxLegendTeams = 6

// legendLabel returns cycle's team sequence for display in the legend,
// collapsing the middle of long cycles to their first and last few teams
// plus a total count.
func legendLabel(teams []string) string {
	if len(teams) <= maxLegendTeams {
		return strings.Join(teams, " -> ")
	}
	half := maxLegendTeams / 2
	head := strings.Join(teams[:half], " -> ")
	tail := strings.Join(teams[len(teams)-half:], " -> ")
	return fmt.Sprintf("%s -> ... -> %s (%d teams)", head, tail, len(teams))
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

// edgeColor returns the color of the longest distinct cycle containing the
// consecutive pair winner->loser (cycles is sorted longest-first), or
// neutralColor if no distinct cycle contains that edge.
func edgeColor(cycles []distinctCycle, winner, loser string) string {
	for _, dc := range cycles {
		for i, t := range dc.teams {
			if t == winner && dc.teams[(i+1)%len(dc.teams)] == loser {
				return dc.color
			}
		}
	}
	return neutralColor
}
