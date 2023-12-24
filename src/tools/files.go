package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func GenFilename(year, week int) string {
	if week != 0 {
		return fmt.Sprintf("data/nfl/fragment/%d/week_%02d.json", year, week)
	}

	return fmt.Sprintf("data/nfl/full/%d.json", year)
}

func SaveFile(weeks []Week, year, week int) {
	bytes, err := json.Marshal(weeks)
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

func EnsureDir(path string) {
	newpath := filepath.Dir(path)
	os.MkdirAll(newpath, 0o755)
}
