package nfldata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// LoadFile loads a single season JSON file.
func LoadFile(path string) (Season, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Season{}, fmt.Errorf("read %s: %w", path, err)
	}
	var s Season
	if err := json.Unmarshal(b, &s); err != nil {
		return Season{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return s, nil
}

// LoadDir loads every *.json file in dir as a Season, sorted by season year.
func LoadDir(dir string) ([]Season, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	seasons := make([]Season, 0, len(matches))
	for _, path := range matches {
		s, err := LoadFile(path)
		if err != nil {
			return nil, err
		}
		seasons = append(seasons, s)
	}
	sort.Slice(seasons, func(i, j int) bool { return seasons[i].SeasonYear < seasons[j].SeasonYear })
	return seasons, nil
}
