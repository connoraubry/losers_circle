package scraper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/connoraubry/losers_circle/backend/tools/db"
)

type ScraperConfig struct {
}

type Scraper struct {
	cfg       ScraperConfig
	Collector *colly.Collector
	DB        *gorm.DB

	Games []Game

	UseDB bool
}

func New(useDB bool) *Scraper {
	s := &Scraper{}

	s.UseDB = useDB
	c := colly.NewCollector()

	if s.UseDB {
		s.DB = db.NewDB(db.Options{Host: "localhost"})
	}

	c.OnRequest(func(r *colly.Request) {
		log.WithField("url", r.URL).Info("Scraper: Visitng")
		s.Games = make([]Game, 0)
	})
	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: "pro-football-reference.com/*",
		// Set a delay between requests to these domains
		Delay: 3 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})
	c.OnError(func(_ *colly.Response, err error) {
		log.Error("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.WithField("url", r.Request.URL).Info("Visited")
	})

	c.OnHTML("table.teams tbody", func(e *colly.HTMLElement) {
		game := ProcessGame(e)
		s.Games = append(s.Games, game)
	})

	c.OnScraped(func(r *colly.Response) {
		log.WithField("url", r.Request.URL).Info("Finished")
	})

	s.Collector = c

	return s
}

func ProcessGame(e *colly.HTMLElement) Game {

	g := Game{}

	e.ForEach("tr", func(idx int, elem *colly.HTMLElement) {

		switch idx {
		case 0:
			date, err := time.Parse("Jan _2, 2006", elem.Text)
			if err != nil {
				log.Fatalf("Error parsing date: %v %v", elem.Text, err)
			}
			g.Date = date
		case 1:
			row := ProcessTeamRow(elem)
			g.Away = row.Team
			g.AwayScore = row.Score
		case 2:
			row := ProcessTeamRow(elem)
			g.Home = row.Team
			g.HomeScore = row.Score
		}

	})
	return g
}

type Game struct {
	Home string
	Away string

	HomeScore int
	AwayScore int

	Date     time.Time
	Complete bool
}

type Row struct {
	Team     string
	Score    int
	IsWinner bool
}

func ProcessTeamRow(e *colly.HTMLElement) Row {
	r := Row{Score: -1}

	if e.Attr("class") == "winner" {
		r.IsWinner = true
	}
	e.ForEach("td", func(idx int, elem *colly.HTMLElement) {
		switch idx {
		case 0:
			r.Team = elem.Text
		case 1:
			score, err := strconv.Atoi(elem.Text)
			if err != nil {
				score = -1
			}
			r.Score = score
		}
	})

	return r
}

func BuildURL(year, week int) string {
	return fmt.Sprintf("https://www.pro-football-reference.com/years/%v/week_%v.htm", year, week)
}

func (s *Scraper) ScrapeYear(year int) error {
	log.Debug("Entering Scrape Year Function")
	var yearDB db.Year
	if s.UseDB {
		s.DB.Where(&db.Year{Year: year}).FirstOrCreate(&yearDB)
	}

	fmt.Println(yearDB)

	for week := 1; week < 19; week++ {
		var weekDB db.Week
		if s.UseDB {
			match := db.Week{Week: week, YearID: int(yearDB.ID)}
			s.DB.Where(match).FirstOrCreate(&weekDB)
		}

		url := BuildURL(year, week)
		log.WithField("url", url).Info("Attempting to connect to url")

		s.Collector.Visit(url)

		for _, game := range s.Games {
			log.WithFields(log.Fields{"home": game.Home, "away": game.Away}).Debug("Entering game")

			var homeDB, awayDB db.Team

			homeDBmatch := db.Team{
				Name: game.Home,
			}
			awayDBmatch := db.Team{
				Name: game.Away,
			}
			if s.UseDB {
				s.DB.Where(homeDBmatch).FirstOrCreate(&homeDB)
			}
			if s.UseDB {
				s.DB.Where(awayDBmatch).FirstOrCreate(&awayDB)
			}

			log.WithFields(log.Fields{"homeid": homeDB.ID, "awayid": awayDB.ID})
			log.Debug("Got home and away teams")

			var winnerID int
			if game.HomeScore > game.AwayScore {
				winnerID = int(homeDB.ID)
			} else if game.AwayScore > game.HomeScore {
				winnerID = int(awayDB.ID)
			}

			log.Debug("Creating gameDB")

			var gameDB db.Game
			if s.UseDB {
				gameDBmatch := db.Game{
					HomeID: int(homeDB.ID),
					AwayID: int(awayDB.ID),
					WeekID: int(weekDB.ID),
					Date:   game.Date,
				}
				res := s.DB.Where(gameDBmatch).FirstOrCreate(&gameDB)
				if res.Error != nil {
					log.Fatal("Fatal error with gamedb:", res.Error)
				}
			}

			log.Debug("Got Game DB")

			gameDB.HomePoints = game.HomeScore
			gameDB.AwayPoints = game.AwayScore
			gameDB.WinnerID = winnerID

			gameDB.Tie = game.AwayScore == game.HomeScore
			s.DB.Save(&gameDB)
		}
		time.Sleep(3 * time.Second)
	}

	return nil
}
