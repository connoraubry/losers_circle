package cycle

import (
	"reflect"
	"testing"

	"losers_circle/internal/nfldata"
)

func game(winner, loser string) nfldata.Game {
	w, l := 21, 14
	return nfldata.Game{
		HomeTeam:  winner,
		AwayTeam:  loser,
		HomeScore: &w,
		AwayScore: &l,
		Status:    "final",
		Winner:    winner,
	}
}

func TestLongestThreeCycle(t *testing.T) {
	s := nfldata.Season{
		Teams: []string{"A", "B", "C"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "C"),
			game("C", "A"),
		},
	}
	g := Build(s)

	got := g.Longest("A")
	if len(got) != 3 {
		t.Fatalf("Longest(A) = %v, want a 3-team cycle", got)
	}
}

func TestLongestNoCycle(t *testing.T) {
	s := nfldata.Season{
		Teams: []string{"A", "B", "C"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "C"),
		},
	}
	g := Build(s)

	if got := g.Longest("A"); got != nil {
		t.Fatalf("Longest(A) = %v, want nil (no cycle)", got)
	}
}

func TestLongestSplitSeriesTwoCycle(t *testing.T) {
	s := nfldata.Season{
		Teams: []string{"A", "B"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "A"),
		},
	}
	g := Build(s)

	got := g.Longest("A")
	want := []string{"A", "B"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Longest(A) = %v, want %v", got, want)
	}
}
