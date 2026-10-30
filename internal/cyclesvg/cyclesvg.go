// Package cyclesvg renders a cycle.Graph as an SVG image: teams as nodes,
// "beat" relationships as directed edges. Edges belonging to a distinct
// cycle are colored per cycle, with a legend describing each one. A pair of
// teams on several distinct cycles gets one parallel line per cycle.
package cyclesvg

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"

	"losers_circle/internal/cycle"
)

// neutralColor is used for edges/nodes not on any distinct cycle.
const neutralColor = cycle.NeutralColor

// Options configures Render.
type Options struct {
	// OnlyCycles renders only the distinct cycles themselves - their
	// member teams and the edges that form each loop - instead of the
	// full win/loss graph.
	OnlyCycles bool
	// TrimNonCycleEdges renders the full win/loss graph, but drops any
	// edge between two teams that are both on some cycle unless that edge
	// is itself part of a cycle. Edges touching a team on no cycle at all
	// are left untouched. Ignored if OnlyCycles is set.
	TrimNonCycleEdges bool
}

// Render writes an SVG of g to path, plus a PNG of the same graph (same
// name, .png extension) next to it. The PNG is graphviz's own rendering of
// the graph and doesn't carry the SVG's post-processed legend. byTeam is
// g.LongestByTeam()'s result, used to derive the set of distinct cycles for
// coloring and the legend.
func Render(path string, g *cycle.Graph, byTeam map[string][]string, opts Options) error {
	cycles := cycle.DedupCycles(byTeam)

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
	// addEdge creates one visual edge. edgeKey disambiguates parallel edges
	// between the same winner/loser: edges sharing a key are coalesced by
	// graphviz into one line, while distinct keys render as separate,
	// automatically-offset parallel lines - which is how a pair that's part
	// of several distinct cycles ends up with one line per cycle.
	addEdge := func(winner, loser, color, edgeKey string) error {
		wn, err := getNode(winner, color)
		if err != nil {
			return fmt.Errorf("create node %s: %w", winner, err)
		}
		ln, err := getNode(loser, color)
		if err != nil {
			return fmt.Errorf("create node %s: %w", loser, err)
		}
		name := winner + "->" + loser
		if edgeKey != "" {
			name += "#" + edgeKey
		}
		edge, err := graph.CreateEdgeByName(name, wn, ln)
		if err != nil {
			return fmt.Errorf("create edge %s->%s: %w", winner, loser, err)
		}
		edge.SetColor(color)
		return nil
	}

	if opts.OnlyCycles {
		// Only the distinct cycles' own edges - nothing else in the graph.
		// Each cycle gets its own edgeKey so a pair shared by multiple
		// cycles still draws one line per cycle.
		for ci, dc := range cycles {
			for i, t := range dc.Teams {
				next := dc.Teams[(i+1)%len(dc.Teams)]
				if err := addEdge(t, next, dc.Color, strconv.Itoa(ci)); err != nil {
					return err
				}
			}
		}
	} else {
		inCycle := map[string]bool{}
		if opts.TrimNonCycleEdges {
			for _, dc := range cycles {
				for _, t := range dc.Teams {
					inCycle[t] = true
				}
			}
		}
		for _, e := range g.Edges() {
			winner, loser := e[0], e[1]
			idxs := cycle.CyclesWithEdge(cycles, winner, loser)
			if len(idxs) == 0 {
				if opts.TrimNonCycleEdges && inCycle[winner] && inCycle[loser] {
					continue
				}
				if err := addEdge(winner, loser, neutralColor, ""); err != nil {
					return err
				}
				continue
			}
			for _, ci := range idxs {
				if err := addEdge(winner, loser, cycles[ci].Color, strconv.Itoa(ci)); err != nil {
					return err
				}
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
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return err
	}

	var pngBuf bytes.Buffer
	if err := gv.Render(ctx, graph, graphviz.PNG, &pngBuf); err != nil {
		return fmt.Errorf("render png: %w", err)
	}
	pngPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".png"
	return os.WriteFile(pngPath, pngBuf.Bytes(), 0o644)
}

// svgOpenTag matches the opening <svg width="..pt" height="..pt" of
// graphviz's SVG output, to recover its rendered size.
var svgOpenTag = regexp.MustCompile(`<svg\s+width="([\d.]+)pt"\s+height="([\d.]+)pt"`)

// addLegend prepends a compact, wrapping legend row (one swatch + label per
// distinct cycle) above svg, extending the canvas down by only as much
// height as the legend actually needs rather than leaving it wherever
// Graphviz would otherwise place a cluster. svg is left untouched if there
// are no cycles to show.
func addLegend(svg []byte, cycles []cycle.DistinctCycle) ([]byte, error) {
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
		label := legendLabel(dc.Teams)
		width := swatchSize + labelGapX + float64(len(label))*charWidth
		if x > padding && x+width > rowMaxWidth {
			x = padding
			y += rowHeight
		}
		items = append(items, placedItem{color: dc.Color, label: label, x: x, y: y})
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
