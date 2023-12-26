package tools

import (
	"html/template"
	"time"
)

type Whole struct {
	Title          string
	Body           Inner
	MatchupSection MatchupSection
	Graph          HTMLGraph
}

type HTMLGraph struct {
	GraphString template.HTML
}

type MatchupSection struct {
	Controls MatchupControls
	Matchups []Matchup
}

type MatchupControls struct {
	ValidWeeks []int
	Week       int
	Year       int
}

type Matchup struct {
	Team1 Team
	Team2 Team
}
type Team struct {
	Name string
}
type Inner struct {
	Title string
	Body  string
}

type Week struct {
	Year  int
	Week  int
	Games []Game
}

type Game struct {
	Home string
	Away string

	HomeScore int
	AwayScore int

	Date     time.Time
	Complete bool
}
