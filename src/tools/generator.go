package tools

import (
	"bytes"
	"fmt"
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

func GenerateMain(week int) []byte {
	log.WithField("week", week).Info("Generating main file")

	var b bytes.Buffer
	t := template.Must(template.ParseGlob("static/templates/*.html"))

	weeks := LoadFile(2023, 0)
	w := weeks[week-1]

	page := Whole{
		Body: Inner{
			Title: "Test Inner Title",
			Body:  "Test Inner Body",
		},
		Title:          "Circle of Suck",
		MatchupSection: matchupSelection(w),
		Graph:          dummyGraph(),
	}
	fmt.Printf("%+v", page)
	t.ExecuteTemplate(&b, "base", page)

	fmt.Println(w)

	return b.Bytes()
}
