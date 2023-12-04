package main

import (
	"flag"

	"github.com/connoraubry/losers_circle/src/web"
	"github.com/sirupsen/logrus"
)

var (
	useDB = flag.Bool("useDB", false, "Use DB connection")
)

func main() {
	logrus.Info("Starting server")
	flag.Parse()
	server := web.New(*useDB)
	server.Run()
}
