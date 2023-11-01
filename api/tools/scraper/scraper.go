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

type Config struct {
	UseDB bool
	DBOpt db.Options
	Week  int
}

type Scraper struct {
	cfg       Config
	Collector *colly.Collector
	DB        *gorm.DB

	Games []Game
}

func New(cfg Config) *Scraper {
	s := &Scraper{}

	s.cfg = cfg

	c := colly.NewCollector()

	if s.cfg.UseDB {
		s.DB = db.NewDB(s.cfg.DBOpt)
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
		game, err := ProcessGame(e)
		if err != nil {
			return
		}
		s.Games = append(s.Games, game)
	})

	c.OnScraped(func(r *colly.Response) {
		log.WithField("url", r.Request.URL).Info("Finished")
	})

	s.Collector = c

	return s
}

func ProcessGame(e *colly.HTMLElement) (Game, error) {

	g := Game{}
	var err error
	var date time.Time

	e.ForEach("tr", func(idx int, elem *colly.HTMLElement) {

		switch idx {
		case 0:
			date, err = time.Parse("Jan _2, 2006", elem.Text)
			if err != nil {
				return
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

	if err != nil {
		return g, fmt.Errorf("Error parsing date: %v", err)
	}
	return g, nil
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

func (s *Scraper) GetTeamByName(name string) db.Team {

	var team db.Team

	homeDBmatch := db.Team{
		Name: name,
	}
	if s.cfg.UseDB {
		s.DB.Where(homeDBmatch).FirstOrCreate(&team)
	}

	return team
}

func (s *Scraper) ScrapeYear(year int) error {
	log.Debug("Entering Scrape Year Function")
	var yearDB db.Year
	if s.cfg.UseDB {
		s.DB.Where(&db.Year{Year: year}).FirstOrCreate(&yearDB)
	}

	for week := 1; week < 19; week++ {

		if s.cfg.Week != 0 && s.cfg.Week != week {
			continue
		}

		//Get Week database entry
		var weekDB db.Week
		if s.cfg.UseDB {
			match := db.Week{Week: week, YearID: int(yearDB.ID)}
			s.DB.Where(match).FirstOrCreate(&weekDB)
		}

		url := BuildURL(year, week)
		log.WithField("url", url).Info("Attempting to connect to url")

		s.Collector.Visit(url)

		for _, game := range s.Games {
			log.WithFields(log.Fields{"home": game.Home, "away": game.Away}).Debug("Entering game")

			homeDB := s.GetTeamByName(game.Home)
			awayDB := s.GetTeamByName(game.Away)

			log.WithFields(
				log.Fields{"homeid": homeDB.ID, "awayid": awayDB.ID},
			).Debug("Got home and away teams")

			log.Debug("Creating gameDB")

			var gameDB db.Game
			if s.cfg.UseDB {
				gameDBmatch := db.Game{
					HomeID: int(homeDB.ID),
					AwayID: int(awayDB.ID),
					WeekID: int(weekDB.ID),
					YearID: int(yearDB.ID),
					Date:   game.Date,
				}
				res := s.DB.Where(gameDBmatch).FirstOrCreate(&gameDB)
				if res.Error != nil {
					log.Fatal("Fatal error with gamedb: ", res.Error)
				}
			}

			log.Debug("Got Game DB")
			gameDB.HomePoints = game.HomeScore
			gameDB.AwayPoints = game.AwayScore
			gameDB.YearID = int(yearDB.ID)

			gameDB.Tie = game.AwayScore == game.HomeScore

			if time.Now().After(game.Date) {
				gameDB.Completed = true
			}
			s.DB.Save(&gameDB)
		}
		time.Sleep(3 * time.Second)
	}

	return nil
}
