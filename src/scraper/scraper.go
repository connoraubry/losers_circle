package scraper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("vim-go")
}

type Config struct {
	Week int
}

type Game struct {
	Home string
	Away string

	HomeScore int
	AwayScore int

	Date     time.Time
	Complete bool
}

type Scraper struct {
	cfg       Config
	Collector *colly.Collector

	Games []Game
}

func New(cfg Config) *Scraper {
	s := &Scraper{}

	s.cfg = cfg

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		log.WithField("url", r.URL).Info("Scraper: Visiting")
		s.Games = make([]Game, 0)
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "pro-football-reference.com/*",
		Delay:       3 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Error("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.WithField("url", r.Request.URL).Info("Finished")
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

			g.Complete = time.Now().After(date)

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

type Week struct {
	Year  int
	Week  int
	Games []Game
}

func (s *Scraper) ScrapeYear(year int) []Week {
	log.Debug("Entering Scrape Year Function")

	var weeks []Week
	for week := 1; week < 19; week++ {

		if s.cfg.Week != 0 && s.cfg.Week != week {
			continue
		}

		url := BuildURL(year, week)
		log.WithField("url", url).Info("Attempting to connect to URL")

		s.Collector.Visit(url)

		W := Week{Year: year, Week: week}
		for _, game := range s.Games {
			logFields := log.Fields{"home": game.Home, "away": game.Away}
			log.WithFields(logFields).Debug("Entering game")
			W.Games = append(W.Games, game)
		}
		weeks = append(weeks, W)
		time.Sleep(3 * time.Second)
	}
	return weeks
}
