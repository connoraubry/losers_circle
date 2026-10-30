package webgraph

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
)

// edgeJSON is one "beat" edge in a dataset's graph data file.
type edgeJSON struct {
	Winner string `json:"winner"`
	Loser  string `json:"loser"`
}

// cycleJSON is one distinct cycle in a dataset's graph data file.
type cycleJSON struct {
	ID    int      `json:"id"`
	Color string   `json:"color"`
	Teams []string `json:"teams"`
}

// graphJSON is a dataset's full graph data file, written to
// data/<key>.json.
type graphJSON struct {
	Teams  []string    `json:"teams"`
	Edges  []edgeJSON  `json:"edges"`
	Cycles []cycleJSON `json:"cycles"`
}

// Generate writes c's datasets as a static site under outDir: the
// embedded frontend (staticFS) plus one JSON data file per dataset and a
// datasets.json index, ready to be served by any static file host.
func Generate(c *Cache, outDir string) error {
	dataDir := filepath.Join(outDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}

	if err := writeJSONFile(filepath.Join(dataDir, "datasets.json"), c.List()); err != nil {
		return err
	}
	for _, d := range c.List() {
		entry, _ := c.Get(d.Key)
		if err := writeJSONFile(filepath.Join(dataDir, d.Key+".json"), toGraphJSON(entry)); err != nil {
			return err
		}
	}

	return copyStaticAssets(outDir)
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
	return graphJSON{Teams: entry.Graph.Teams(), Edges: edges, Cycles: cycles}
}

func writeJSONFile(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// copyStaticAssets copies every file in the embedded frontend (staticFS)
// into outDir, preserving its directory structure.
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
