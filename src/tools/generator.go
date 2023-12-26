package tools

import (
	"bytes"
	"html/template"

	log "github.com/sirupsen/logrus"
)

func GenerateMatchups() []byte {
	var b bytes.Buffer
	matchups := dummyMatchupSection()
	t := template.Must(template.ParseFiles("static/templates/matchups.html"))
	t.ExecuteTemplate(&b, "matchups", matchups)
	return b.Bytes()
}

func GenerateMain(year, week int) []byte {
	log.WithField("week", week).Info("Generating main file")

	var b bytes.Buffer
	t := template.Must(template.ParseGlob("static/templates/*.html"))

	weeks := LoadFile(year, 0)
	w := weeks[week-1]

	log.WithField("weeksLen", len(weeks)).Info("Generating main page")

	page := Whole{
		Body: Inner{
			Title: "Test Inner Title",
			Body:  "Test Inner Body",
		},
		Title:          "Circle of Suck",
		MatchupSection: matchupSelection(w),
		Graph:          GetGraph(year, week, weeks),
	}
	t.ExecuteTemplate(&b, "base", page)

	return b.Bytes()
}

func GenerateWeek(week int) []byte {
	log.WithField("week", week).Info("Generating week")

	var b bytes.Buffer
	t := template.Must(template.ParseFiles("static/templates/matchups.html"))

	weeks := LoadFile(2023, 0)
	w := weeks[week-1]

	matchup := matchupSelection(w)

	t.ExecuteTemplate(&b, "matchups", matchup)

	return b.Bytes()
}
