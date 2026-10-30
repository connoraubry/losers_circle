package webgraph

import (
	"encoding/json"
	"net/http"
)

// edgeJSON is one "beat" edge in the graph API response.
type edgeJSON struct {
	Winner string `json:"winner"`
	Loser  string `json:"loser"`
}

// cycleJSON is one distinct cycle in the graph API response.
type cycleJSON struct {
	ID    int      `json:"id"`
	Color string   `json:"color"`
	Teams []string `json:"teams"`
}

// graphJSON is the full response for GET /api/graph.
type graphJSON struct {
	Teams  []string    `json:"teams"`
	Edges  []edgeJSON  `json:"edges"`
	Cycles []cycleJSON `json:"cycles"`
}

// Handler returns an http.Handler serving the dataset list, graph data,
// and static frontend backed by c. Every handler only reads from c, which
// is built once up front, so requests never trigger a cycle computation.
func Handler(c *Cache) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/datasets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, c.List())
	})
	mux.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("dataset")
		entry, ok := c.Get(key)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown dataset: " + key})
			return
		}
		writeJSON(w, http.StatusOK, toGraphJSON(entry))
	})
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	return mux
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
