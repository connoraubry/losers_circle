package webgraph

import (
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"time"
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
}

// Generate writes c's datasets as a static site under outDir: one
// self-contained HTML page per dataset (its graph data inlined, no
// fetch required), a shared app.js/style.css, and an index.html mirroring
// the first (newest) dataset so the site has a sensible landing page.
// Every page is a real, separately-addressable URL - no client-side
// routing - so browser back/forward works across datasets for free.
func Generate(c *Cache, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	generatedAt := time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")

	datasets := c.List()
	for _, d := range datasets {
		entry, _ := c.Get(d.Key)
		if err := writePage(outDir, d.Key+".html", d, datasets, toGraphJSON(entry), generatedAt); err != nil {
			return fmt.Errorf("page %s: %w", d.Key, err)
		}
	}
	if len(datasets) > 0 {
		entry, _ := c.Get(datasets[0].Key)
		if err := writePage(outDir, "index.html", datasets[0], datasets, toGraphJSON(entry), generatedAt); err != nil {
			return fmt.Errorf("index page: %w", err)
		}
	}

	return copyStaticAssets(outDir)
}

func writePage(outDir, filename string, current Dataset, datasets []Dataset, graph graphJSON, generatedAt string) error {
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
	})
}

// toGraphJSON serializes entry's already-built graph and already-
// precomputed cycles - no graph construction or cycle search happens here.
func toGraphJSON(entry Entry) graphJSON {
	edges := make([]edgeJSON, 0, len(entry.Graph.Edges()))
	for _, e := range entry.Graph.Edges() {
		edges = append(edges, edgeJSON{Winner: e[0], Loser: e[1]})
	}
	cycles := make([]cycleJSON, len(entry.Cycles))
	for i, dc := range entry.Cycles {
		cycles[i] = cycleJSON{ID: i, Color: dc.Color, Teams: dc.Teams}
	}
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
	return graphJSON{Teams: teams, TeamColors: colors, Edges: edges, Cycles: cycles}
}

// copyStaticAssets copies every file in the embedded frontend (staticFS,
// now just app.js and style.css - index.html is generated per dataset
// instead) into outDir, preserving its directory structure.
func copyStaticAssets(outDir string) error {
	return fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(staticFS, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(outDir, path)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}
