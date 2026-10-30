// Command server serves an interactive, browser-based viewer of the
// losers circle win/loss graph. It precomputes every dataset's cycles at
// startup (persisting results under -cache-dir so a restart doesn't redo
// the expensive search unless the underlying season data changed), then
// serves them over HTTP alongside a small static frontend.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"losers_circle/internal/webgraph"
)

func main() {
	dir := flag.String("dir", "data", "directory containing season JSON files")
	cacheDir := flag.String("cache-dir", "cache", "directory to persist precomputed cycles in")
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	start := time.Now()
	cache, err := webgraph.Build(*dir, *cacheDir)
	if err != nil {
		log.Fatalf("building dataset cache: %v", err)
	}
	log.Printf("precomputed %d datasets in %s", len(cache.List()), time.Since(start).Round(time.Millisecond))

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, webgraph.Handler(cache)); err != nil {
		log.Fatal(err)
	}
}
