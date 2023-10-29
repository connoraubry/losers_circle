package main

import (
	"github.com/connoraubry/losers_circle/backend/api"
)

func main() {
	s := api.New()
	s.Serve()
}
