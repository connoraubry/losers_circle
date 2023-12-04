package main

import (
	"github.com/connoraubry/losers_circle/server/src/api"
)

func main() {
	s := api.New()
	s.Serve()
}
