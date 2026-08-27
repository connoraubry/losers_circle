package tools

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type TeamInfo struct {
	Name string `json:"name"`
	Abbr string `json:"abbr"`
}
type LeagueInfo struct {
	League string     `json:"league"`
	Teams  []TeamInfo `json:"teams"`
}

func GetTeamsFromWeek(w Week) []string {
	var teams []string

	for _, game := range w.Games {
		teams = append(teams, game.Away, game.Home)
	}

	return teams
}

func GetAllTeams() []TeamInfo {
	var leagueInfo LeagueInfo

	yamlFile, err := os.Open("./data/nfl.yaml")
	if err != nil {
		panic(err)
	}
	defer yamlFile.Close()

	bytes, _ := io.ReadAll(yamlFile)
	err = yaml.Unmarshal(bytes, &leagueInfo)
	if err != nil {
		fmt.Println(err)
	}
	return leagueInfo.Teams
}

func GetTeamToAbbr() map[string]string {
	var leagueInfo LeagueInfo

	yamlFile, err := os.Open("./data/nfl.yaml")
	if err != nil {
		panic(err)
	}
	defer yamlFile.Close()

	bytes, _ := io.ReadAll(yamlFile)
	err = yaml.Unmarshal(bytes, &leagueInfo)
	if err != nil {
		fmt.Println(err)
	}

	ttabbr := make(map[string]string)
	for _, team := range leagueInfo.Teams {
		ttabbr[team.Name] = team.Abbr
	}
	return ttabbr
}
