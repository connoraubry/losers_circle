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

// TestLongestBridgeNode covers a team that has both a win and a loss (so
// naive in/out-degree pruning wouldn't remove it) but still isn't part of
// any cycle, because it only bridges between two separate cycles: A<->B
// and C<->D, joined by A->X->C. Only SCC-based pruning catches this.
func TestLongestBridgeNode(t *testing.T) {
	s := nfldata.Season{
		Teams: []string{"A", "B", "C", "D", "X"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "A"),
			game("C", "D"),
			game("D", "C"),
			game("A", "X"),
			game("X", "C"),
		},
	}
	g := Build(s)

	if got := g.Longest("X"); got != nil {
		t.Fatalf("Longest(X) = %v, want nil (X only bridges two cycles)", got)
	}
	if got := g.Longest("A"); len(got) != 2 {
		t.Fatalf("Longest(A) = %v, want a 2-team cycle", got)
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
