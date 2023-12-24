package main

import (
	"os"
	"path/filepath"

	"github.com/connoraubry/losers_circle/src/tools"
	log "github.com/sirupsen/logrus"
	"github.com/yosssi/gohtml"
)

func main() {

	output_dir := "./public"
	GenerateAll(output_dir)
}

func GenerateAll(output_dir string) {

	main := tools.GenerateMain(10)
	main = gohtml.FormatBytes(main)
	output_path := filepath.Join(output_dir, "index.html")
	err := os.WriteFile(output_path, main, 0644)

	if err != nil {
		log.Error(err)
	}

}
