package db

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	gorm.Model

	Name     string
	Nickname string
}

type Game struct {
	gorm.Model

	HomeID     int
	Home       Team
	HomePoints int
	AwayID     int
	Away       Team
	AwayPoints int

	Completed bool
	Tie       bool

	Date time.Time

	WeekID int
	Week   Week
	YearID int
	Year   Year
}

type Week struct {
	gorm.Model

	Week int

	YearID int
	Year   Year
}

type Year struct {
	gorm.Model
	Year int `gorm:"index:unqiue,"`
}
