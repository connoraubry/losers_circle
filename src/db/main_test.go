package db

import (
	"os"
	"path"
	"testing"
)

func TestCreateDB(t *testing.T) {
	dname, err := os.MkdirTemp("", "sampledir")
	if err != nil {
		t.Fatalf("Error: Cannot make temp dir")
	}
	defer os.RemoveAll(dname)

	file := path.Join(dname, "tmp.sql")

	_, err = New(file)

	if err != nil {
		t.Fatalf("Error creating new database: %v", err)
	}

}

func TestAddTeam(t *testing.T) {
	dname, err := os.MkdirTemp("", "sampledir")
	if err != nil {
		t.Fatalf("Error: Cannot make temp dir")
	}
	defer os.RemoveAll(dname)
	file := path.Join(dname, "tmp.sql")

	s, err := New(file)
	if err != nil {
		t.Fatalf("Error creating new databae: %v", err)
	}

	teamName := "Test team"

	err = s.AddTeam(teamName)
	if err != nil {
		t.Fatalf("Error adding team to test: %v", teamName)
	}

	allTeams, err := s.GetTeams()
	if err != nil {
		t.Fatalf("Error fetching teams")
	}

	if len(allTeams) != 1 {
		t.Fatalf("Error: len(allTeams) = %v. Expected 1", len(allTeams))
	}

	if allTeams[0].Name != teamName {
		t.Fatalf("Error: allTeams[0].Name == %v. Expected %v", allTeams[0].Name, teamName)
	}
}

func TestAddStem(t *testing.T) {

	dname, err := os.MkdirTemp("", "sampledir")
	if err != nil {
		t.Fatalf("Error: Cannot make temp dir")
	}
	defer os.RemoveAll(dname)
	file := path.Join(dname, "tmp.sql")

	s, err := New(file)
	if err != nil {
		t.Fatalf("Error creating new databae: %v", err)
	}

	var teams = []string{"A", "B", "C", "D", "E", "F"}
	err = s.AddTeams(teams)
	if err != nil {
		t.Fatalf("Error adding teams %v: %v", teams, err)
	}

	exampleStem := Stem{
		Level:   2,
		Mask:    2,
		StartID: 0,
		EndID:   1,
	}

	err = s.AddStem(exampleStem)
	if err != nil {
		t.Fatalf("Error adding stem baseline %v: %v", exampleStem, err)
	}

	stems, err := s.GetStemsFromLevel(2)
	if err != nil {
		t.Fatalf("s.GetStemsFromLevel(2) failed: %v", err)
	}

	if len(stems) != 1 {
		t.Fatalf("Wrong number of stems returned: %v", len(stems))
	}
}

var StemTestCase = []struct {
	Name string
	Stem Stem
}{
	{Name: "TestCase1", Stem: Stem{Level: 2, Mask: 2, StartID: 0, EndID: 1}},
	{Name: "TestCase2", Stem: Stem{Level: 2, Mask: 4, StartID: 0, EndID: 2}},
	{Name: "TestCase3", Stem: Stem{Level: 2, Mask: 2, StartID: 1, EndID: 3}},
	{Name: "TestCase4", Stem: Stem{Level: 3, Mask: 6, StartID: 0, EndID: 3, StartStem: 2, EndStem: 3}},
}

func TestAddStems(t *testing.T) {

	dname, err := os.MkdirTemp("", "sampledir")
	if err != nil {
		t.Fatalf("Error: Cannot make temp dir")
	}
	defer os.RemoveAll(dname)
	file := path.Join(dname, "tmp.sql")

	s, err := New(file)
	if err != nil {
		t.Fatalf("Error creating new databae: %v", err)
	}

	var teams = []string{"A", "B", "C"}
	err = s.AddTeams(teams)
	if err != nil {
		t.Fatalf("Error adding teams %v: %v", teams, err)
	}

	for _, testCase := range StemTestCase {
		name := testCase.Name
		stem := testCase.Stem
		err = s.AddStem(stem)
		if err != nil {
			t.Errorf("Error adding stem %v %v: %v", name, stem, err)
		}

	}

	stems, err := s.GetStemsFromLevel(2)
	if err != nil {
		t.Fatalf("s.GetStemsFromLevel(2) failed: %v", err)
	}

	if len(stems) != 3 {
		t.Fatalf("Wrong number of stems returned: %v", len(stems))
	}

	stems, err = s.GetStemsFromLevel(3)
	if err != nil {
		t.Fatalf("s.GetStemsFromLevel(2) failed: %v", err)
	}

	if len(stems) != 1 {
		t.Fatalf("Wrong number of stems returned: %v", len(stems))
	}
}
