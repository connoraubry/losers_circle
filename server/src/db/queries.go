package db

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetAllTeams(db *gorm.DB, year int) int {
	log.Info("Calling handler GetAllTeams")

	weeks := GetWeeks(db, year)

	fmt.Println(weeks)
	return 3
}

func GetYear(db *gorm.DB, year int) Year {
	var y Year
	db.Where(&Year{Year: year}).First(&y)
	return y
}

func GetWeeks(db *gorm.DB, year int) []Week {
	y := GetYear(db, year)
	var weeks []Week
	db.Where(&Week{Year: y}).Find(&weeks)
	return weeks
}

func GetGames(db *gorm.DB, year int) []Game {
	var games []Game
	db.Preload("Away").Preload("Home").Where(&Game{YearID: int(GetYear(db, year).ID)}).Find(&games)
	return games
}

func GetTeamsInASeason(db *gorm.DB, year int) []Team {
	var teams []Team
	games := GetGames(db, year)
	uniqueTeamIDs := make(map[int]bool)
	for _, game := range games {
		uniqueTeamIDs[game.AwayID] = true
		uniqueTeamIDs[game.HomeID] = true
	}
	for id := range uniqueTeamIDs {
		var team Team
		db.First(&team, id)
		teams = append(teams, team)
	}
	return teams
}

func GetName(db *gorm.DB, teamID int) string {
	var team Team
	db.First(&team, teamID)
	return team.Name
}
