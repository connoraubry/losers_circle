package tools

import "html/template"

type Whole struct {
	Title    string
	Body     Inner
	Matchups []Matchup
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

type Page struct {
	Title string
	Body  template.HTML
}
