package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/connoraubry/losers_circle/backend/tools/db"
	"github.com/connoraubry/losers_circle/backend/tools/graph"
	log "github.com/sirupsen/logrus"
)

func ping(w http.ResponseWriter, r *http.Request) {
	log.Info("Calling handler ping")
	fmt.Fprintf(w, "pong")
}

func (s *Server) Data(w http.ResponseWriter, r *http.Request) {
	log.Info("Calling handler data")
	s.Val = s.Val + 1
	fmt.Fprintf(w, "%v", s.Val)
}

// GET /teams
func (s *Server) teams(w http.ResponseWriter, r *http.Request) {
	log.Info("Calling handler teams")

	teams := db.GetTeamsInASeason(s.DB, 2023)

	var names []string
	for _, t := range teams {
		names = append(names, t.Name)
	}
	bytes, err := json.Marshal(names)
	if err != nil {
		log.Error("Error marshalling names", err)
	}
	w.Write(bytes)
}

func (s *Server) Test(w http.ResponseWriter, r *http.Request) {
	log.Info("Calling handler Test")

	fmt.Fprintf(w, "test successful")
}

func (s *Server) GetLargestCircle(w http.ResponseWriter, r *http.Request) {
	log.Info("Calling handler GetLargestCycle")

	games := db.GetGames(s.DB, 2023)
	log.WithField("gameLen", len(games)).Info("Games acquired")

	g := graph.New()

	for _, game := range games {
		var winner, loser string
		if game.AwayPoints > game.HomePoints {
			winner = game.Away.Name
			loser = game.Home.Name
		} else if game.HomePoints > game.AwayPoints {
			winner = game.Home.Name
			loser = game.Away.Name
		}

		cnx := graph.NewCnx(winner, loser)
		g.AddConnection(cnx)

	}
	log.Info("Added connections. Evalutating graphs")
	g.EvaluateCycles()
	log.Info("Evaluated cycles")

	var longestCycle []string
	for _, cycle := range g.NodeToCycle {
		if len(cycle) > len(longestCycle) {
			longestCycle = cycle
		}
		// fmt.Printf("%v: %v %v\n", team, len(cycle), cycle)
	}

	bytes, err := json.Marshal(longestCycle)
	if err != nil {
		log.Error("Error marhsalling cycle", err)
	}
	w.Write(bytes)
}
