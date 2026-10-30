// Command buildsite generates a static, interactive viewer of the losers
// circle win/loss graph: HTML/JS/CSS plus one JSON data file per dataset,
// written to -out. The result needs no server-side logic to host - any
// static file server will do. Cycle detection, the expensive part, is
// precomputed once per dataset and persisted under -cache-dir, so
// re-running this command doesn't redo that search unless the underlying
// season data changed.
package main

import (
	"flag"
	"log"
	"time"

	"losers_circle/internal/webgraph"
)

func main() {
	dir := flag.String("dir", "data", "directory containing season JSON files")
	cacheDir := flag.String("cache-dir", "cache", "directory to persist precomputed cycles in")
	out := flag.String("out", "site", "directory to write the static site to")
	flag.Parse()

	start := time.Now()
	cache, err := webgraph.Build(*dir, *cacheDir)
	if err != nil {
		log.Fatalf("building dataset cache: %v", err)
	}
	log.Printf("precomputed %d datasets in %s", len(cache.List()), time.Since(start).Round(time.Millisecond))

	if err := webgraph.Generate(cache, *out); err != nil {
		log.Fatalf("generating site: %v", err)
	}
	log.Printf("wrote static site to %s", *out)
}
