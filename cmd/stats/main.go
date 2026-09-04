// Command stats prints summary statistics about the loaded NFL season data.
package main

import (
	"flag"
	"fmt"
	"os"

	"losers_circle/internal/nfldata"
)

func main() {
	dir := flag.String("dir", "data", "directory containing season JSON files")
	flag.Parse()

	seasons, err := nfldata.LoadDir(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(seasons) == 0 {
		fmt.Fprintf(os.Stderr, "no season files found in %s\n", *dir)
		os.Exit(1)
	}

	for _, s := range seasons {
		nfldata.PrintSeasonStats(os.Stdout, s)
	}
}
