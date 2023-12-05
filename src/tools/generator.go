package tools

import (
	"bytes"
	"html/template"
)

func GenerateMatchups() []byte {
	var b bytes.Buffer
	matchups := dummyMatchups()
	t := template.Must(template.ParseFiles("static/templates/matchups.html"))
	t.ExecuteTemplate(&b, "matchups", matchups)

	return b.Bytes()
}

func GenerateMain() []byte {
	var b bytes.Buffer
	t := template.Must(template.ParseGlob("static/templates/*.html"))

	page := Whole{
		Body: Inner{
			Title: "Test Inner Title",
			Body:  "Test Inner Body"},
		Title:    "Circle of Suck",
		Matchups: dummyMatchups(),
	}
	t.ExecuteTemplate(&b, "base", page)
	return b.Bytes()
}
