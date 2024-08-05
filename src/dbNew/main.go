package main

import (
	"database/sql"
	"fmt"
	"log"
	"slices"

	_ "github.com/mattn/go-sqlite3"
)

const SQL_PATH = "./nfl.db"

type State struct {
	DB *sql.DB
}

type Team struct {
	Id   int
	Name string
}

type Stem struct {
	Id        int
	Level     int
	Mask      uint32
	StartID   int
	EndID     int
	StartStem int
	EndStem   int
}

func main() {
	s, err := New(SQL_PATH)

	fmt.Println(err)
	err = s.InsertDummyData()
	fmt.Println(err)
	err = s.AddTeam("beep")
	fmt.Printf("Error adding beep: %v\n", err)
	err = s.AddTeams([]string{"beep", "Baltimore", "Bal", "test", "boop"})
	fmt.Printf("Error adding teams: %v\n", err)

	fmt.Println("Added teams")
	teams, err := s.GetTeams()
	fmt.Println(teams, err)

	exStem := Stem{
		Level:   3,
		Mask:    2,
		StartID: 0,
		EndID:   1,
	}
	err = s.AddStem(exStem)
	fmt.Println(err)
	nexStem := Stem{
		Level:     3,
		Mask:      51,
		StartID:   0,
		EndID:     3,
		StartStem: 1,
		EndStem:   41,
	}
	err = s.AddStem(nexStem)
	stems, err := s.GetStemsFromLevel(3)
	fmt.Println(stems)
}

func New(filepath string) (*State, error) {
	s := &State{}

	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("Error opening db: %v", err)
	}

	s.DB = db

	sqlStmt := `
CREATE TABLE IF NOT EXISTS team(
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS stem(
    id          INTEGER PRIMARY KEY,
	level		INTEGER NOT NULL,
    mask        INTEGER NOT NULL,
    start_id    INTEGER NOT NULL,
    end_id      INTEGER NOT NULL,
    start_stem  INTEGER,
    end_stem    INTEGER,

    FOREIGN KEY(start_id) REFERENCES team(id)
    FOREIGN KEY(end_id) REFERENCES team(id)
    FOREIGN KEY(start_stem) REFERENCES stem(id)
    FOREIGN KEY(end_stem) REFERENCES stem(id)

	UNIQUE(mask, start_id, end_id)
); `
	_, err = s.DB.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("Error executing sql create table statement: %v", err)
	}

	return s, s.DB.Ping()
}

func (s *State) Close() error {
	return s.DB.Close()
}

func (s *State) AddTeam(team string) error {
	query := "INSERT OR IGNORE INTO team(name) values (?)"

	_, err := s.DB.Exec(query, team)
	if err != nil {
		return fmt.Errorf("Error adding team %v: %v", team, err)
	}

	return err
}

func (s *State) AddTeams(teams []string) error {

	db_teams, err := s.GetTeams()
	if err != nil {
		return err
	}

	var teams_to_add []string
	for _, team := range teams {
		if !slices.ContainsFunc(db_teams, func(n Team) bool {
			return n.Name == team
		}) {
			teams_to_add = append(teams_to_add, team)
		}
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("Error with db.Begin(): %v", err)
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO team(name) VALUES (?)")
	if err != nil {
		return fmt.Errorf("Error with tx.Prepare(): %v", err)
	}
	defer stmt.Close()

	for _, team := range teams_to_add {
		_, err = stmt.Exec(team)
		if err != nil {
			return fmt.Errorf("Error running statment with team %v: %v", team, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Error committing tx: %v", err)
	}

	return nil
}

func (s *State) InsertDummyData() error {

	var teams = []string{
		"Bal", "KC", "Detroit", "Bears", "49ers", "Minnesota Vikings",
	}

	return s.AddTeams(teams)

}

func (s *State) GetTeams() ([]Team, error) {

	var teams []Team

	queryString := "select id, name from team"
	rows, err := s.DB.Query(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error reading db query %v: %v", queryString, err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, fmt.Errorf("Error scanning row: %v", err)
		}
		teams = append(teams, Team{Id: id, Name: name})
	}

	err = rows.Err()
	if err != nil {
		log.Fatal("Error with rows: ", err)
		return nil, fmt.Errorf("Error with rows: %v", err)
	}
	return teams, nil
}

func (s *State) addStemBaseline(stem Stem) error {
	query := "INSERT OR IGNORE INTO stem(mask, level, start_id, end_id) values (?, ?, ?, ?)"

	_, err := s.DB.Exec(query, stem.Mask, stem.Level, stem.StartID, stem.EndID)
	if err != nil {
		return fmt.Errorf("Error adding team %v: %v", stem, err)
	}

	return err
}

func (s *State) AddStem(stem Stem) error {

	if stem.Level == 2 {
		return s.addStemBaseline(stem)
	}

	query := "INSERT OR IGNORE INTO stem(mask, level, start_id, end_id, start_stem, end_stem) values (?, ?, ?, ?, ?, ?)"

	_, err := s.DB.Exec(query, stem.Mask, stem.Level, stem.StartID, stem.EndID, stem.StartStem, stem.EndStem)
	if err != nil {
		return fmt.Errorf("Error adding team %v: %v", stem, err)
	}

	return err
}

func (s *State) GetStemsFromLevel(level int) ([]Stem, error) {

	var stems []Stem

	queryString := "select id,mask,start_id,end_id,start_stem,end_stem from stem where level = ?"
	rows, err := s.DB.Query(queryString, level)
	if err != nil {
		return nil, fmt.Errorf("Error reading db query %v: %v", queryString, err)
	}
	defer rows.Close()
	for rows.Next() {
		var s Stem

		var startStem sql.NullInt64
		var endStem sql.NullInt64
		err = rows.Scan(&s.Id, &s.Mask, &s.StartID, &s.EndID, &startStem, &endStem)
		if err != nil {
			return nil, fmt.Errorf("Error scanning row: %v", err)
		}

		// start, end stem values can be null. Get them safely here
		if startStem.Valid {
			s.StartStem = int(startStem.Int64)
		}
		if endStem.Valid {
			s.EndStem = int(endStem.Int64)
		}

		stems = append(stems, s)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal("Error with rows: ", err)
		return nil, fmt.Errorf("Error with rows: %v", err)
	}
	return stems, nil
}
