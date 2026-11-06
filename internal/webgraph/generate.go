package webgraph

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"losers_circle/internal/cycle"
)

// edgeJSON is one "beat" edge embedded in a dataset's page.
type edgeJSON struct {
	Winner string `json:"winner"`
	Loser  string `json:"loser"`
}

// cycleJSON is one distinct cycle embedded in a dataset's page.
type cycleJSON struct {
	ID    int      `json:"id"`
	Color string   `json:"color"`
	Teams []string `json:"teams"`
}

// teamColorJSON is one team's node fill/border colors, embedded in a
// dataset's page. Teams with no known colors are simply absent from
// graphJSON.TeamColors; the viewer falls back to a neutral color for
// those.
type teamColorJSON struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

// graphJSON is a dataset's full graph data, embedded directly into its
// page (see pageTemplate) rather than fetched separately.
type graphJSON struct {
	Teams      []string                 `json:"teams"`
	TeamColors map[string]teamColorJSON `json:"teamColors"`
	Edges      []edgeJSON               `json:"edges"`
	Cycles     []cycleJSON              `json:"cycles"`
	// Weeks holds one cumulative-through-week snapshot per played NFL
	// regular-season week, for the week slider. Empty for non-NFL
	// datasets. TeamColors is shared with the full graph above rather
	// than repeated per week.
	Weeks []weekJSON `json:"weeks,omitempty"`
}

// weekJSON is one week's cumulative graph, embedded alongside the full
// season's graphJSON.
type weekJSON struct {
	Week   int         `json:"week"`
	Edges  []edgeJSON  `json:"edges"`
	Cycles []cycleJSON `json:"cycles"`
}

//go:embed templates/page.html.tmpl
var pageTemplateSrc string

var pageTemplate = template.Must(template.New("page").Parse(pageTemplateSrc))

// pageData is pageTemplate's input. Datasets and Graph are plain Go
// values (not pre-marshaled strings): html/template JSON-encodes them
// automatically since they're used inside a <script> block.
type pageData struct {
	Key          string
	CurrentLabel string
	Datasets     []Dataset
	Graph        graphJSON
	GeneratedAt  string
	AppJSName    string
	StyleName    string
}

// Generate writes c's datasets as a static site under outDir: one
// self-contained HTML page per dataset (its graph data inlined, no
// fetch required), a shared app-<hash>.js/style-<hash>.css, and an
// index.html mirroring the first (newest) dataset so the site has a
// sensible landing page. Every page is a real, separately-addressable
// URL - no client-side routing - so browser back/forward works across
// datasets for free.
//
// outDir is wiped and rewritten from scratch each run (rather than
// overwriting in place), so a season/conference that's no longer
// produced - or a previous build's differently-hashed asset filenames -
// doesn't linger as an orphaned file.
func Generate(c *Cache, outDir string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clear %s: %w", outDir, err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	generatedAt := time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")
	appJSName, err := writeHashedAsset(outDir, "app.js")
	if err != nil {
		return err
	}
	styleName, err := writeHashedAsset(outDir, "style.css")
	if err != nil {
		return err
	}

	datasets := c.List()
	for _, d := range datasets {
		entry, _ := c.Get(d.Key)
		if err := writePage(outDir, d.Key+".html", d, datasets, toGraphJSON(entry), generatedAt, appJSName, styleName); err != nil {
			return fmt.Errorf("page %s: %w", d.Key, err)
		}
	}
	if len(datasets) > 0 {
		entry, _ := c.Get(datasets[0].Key)
		if err := writePage(outDir, "index.html", datasets[0], datasets, toGraphJSON(entry), generatedAt, appJSName, styleName); err != nil {
			return fmt.Errorf("index page: %w", err)
		}
	}

	return nil
}

// writeHashedAsset reads the embedded static asset at path (relative to
// staticFS, e.g. "app.js"), writes it into outDir under a filename that
// embeds a short content hash (e.g. "app-fc17b43e.js"), and returns that
// filename. Content-addressing the filename itself - rather than e.g. a
// "?v=" query param - means a CDN's default static-asset caching (which
// ignores how often the underlying file content changes) never has a
// stale file to serve under a URL that's still in use: a changed file
// gets a new URL outright, and the old one simply isn't written on the
// next build (outDir having just been wiped by Generate).
func writeHashedAsset(outDir, path string) (string, error) {
	data, err := fs.ReadFile(staticFS, path)
	if err != nil {
		return "", fmt.Errorf("read asset %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])[:8]
	ext := filepath.Ext(path)
	name := fmt.Sprintf("%s-%s%s", strings.TrimSuffix(path, ext), hash, ext)
	if err := os.WriteFile(filepath.Join(outDir, name), data, 0o644); err != nil {
		return "", fmt.Errorf("write asset %s: %w", name, err)
	}
	return name, nil
}

func writePage(outDir, filename string, current Dataset, datasets []Dataset, graph graphJSON, generatedAt, appJSName, styleName string) error {
	f, err := os.Create(filepath.Join(outDir, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	return pageTemplate.Execute(f, pageData{
		Key:          current.Key,
		CurrentLabel: current.Label,
		Datasets:     datasets,
		Graph:        graph,
		GeneratedAt:  generatedAt,
		AppJSName:    appJSName,
		StyleName:    styleName,
	})
}

// toGraphJSON serializes entry's already-built graph and already-
// precomputed cycles (full season, plus one snapshot per week for the NFL
// week slider) - no graph construction or cycle search happens here.
func toGraphJSON(entry Entry) graphJSON {
	teams := entry.Graph.Teams()
	byTeam := nflTeamColors
	if entry.Dataset.League == "cfb" || entry.Dataset.League == "cbb" {
		// CBB reuses cfbTeamColors directly: ESPN uses the same team
		// abbreviations for the same school across both sports.
		byTeam = cfbTeamColors
	}
	colors := make(map[string]teamColorJSON, len(teams))
	for _, team := range teams {
		if c, ok := byTeam[team]; ok {
			colors[team] = teamColorJSON{Primary: c.Primary, Secondary: c.Secondary}
		}
	}

	weeks := make([]weekJSON, len(entry.Weeks))
	for i, w := range entry.Weeks {
		weeks[i] = weekJSON{Week: w.Week, Edges: edgesJSON(w.Graph), Cycles: cyclesJSON(w.Cycles)}
	}

	return graphJSON{
		Teams:      teams,
		TeamColors: colors,
		Edges:      edgesJSON(entry.Graph),
		Cycles:     cyclesJSON(entry.Cycles),
		Weeks:      weeks,
	}
}

func edgesJSON(g *cycle.Graph) []edgeJSON {
	edges := make([]edgeJSON, 0, len(g.Edges()))
	for _, e := range g.Edges() {
		edges = append(edges, edgeJSON{Winner: e[0], Loser: e[1]})
	}
	return edges
}

func cyclesJSON(dcs []cycle.DistinctCycle) []cycleJSON {
	cycles := make([]cycleJSON, len(dcs))
	for i, dc := range dcs {
		cycles[i] = cycleJSON{ID: i, Color: dc.Color, Teams: dc.Teams}
	}
	return cycles
}
