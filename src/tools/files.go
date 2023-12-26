package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func EnsureDir(path string) {
	newpath := filepath.Dir(path)
	os.MkdirAll(newpath, 0o755)
}

func GenFilename(year, week int) string {
	if week != 0 {
		return fmt.Sprintf("data/nfl/fragment/%d/week_%02d.json", year, week)
	}

	return fmt.Sprintf("data/nfl/full/%d.json", year)
}

func SaveFile(weeks []Week, year, week int) {
	bytes, err := json.MarshalIndent(weeks, "", "  ")
	if err != nil {
		log.Error("Error marshaling weeks:", err)
	}

	filename := GenFilename(year, week)
	EnsureDir(filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()

	f.Write(bytes)
}

func LoadFile(year, week int) []Week {
	filename := GenFilename(year, week)

	f, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()

	var w []Week

	bytes, _ := io.ReadAll(f)
	json.Unmarshal(bytes, &w)

	return w
}

func GenCycleFilename(year int) string {
	return fmt.Sprintf("data/nfl/compiled_circle/%d.json", year)
}

func SaveLongestCycles(year int, weekToCycle map[string][]string) {
	log.Debug("Entering SaveLongestCycles")
	filename := GenCycleFilename(year)
	EnsureDir(filename)

	bytes, err := json.MarshalIndent(weekToCycle, "", "  ")
	if err != nil {
		log.Error("Error marshalling longest cycle", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()

	f.Write(bytes)
}
func LoadLongestCycle(year int) map[int][]string {
	filename := GenCycleFilename(year)

	f, err := os.Open(filename)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()

	x := make(map[int][]string)
	bytes, _ := io.ReadAll(f)
	json.Unmarshal(bytes, &x)

	return x
}

func LoadLongestCycle2(year, week int) []string {
	log.Info("Loading longest cycle from file")
	m := LoadLongestCycle(year)
	cyc := m[week]
	return cyc
}
