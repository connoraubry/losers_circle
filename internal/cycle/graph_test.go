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

func unplayed(home, away string, week int) nfldata.Game {
	return nfldata.Game{HomeTeam: home, AwayTeam: away, Week: week, Status: "scheduled"}
}

func TestShortestPath(t *testing.T) {
	s := nfldata.Season{
		Teams: []string{"A", "B", "C", "D"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "C"),
		},
	}
	g := Build(s)

	if got, want := g.ShortestPath("A", "C"), []string{"A", "B", "C"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ShortestPath(A, C) = %v, want %v", got, want)
	}
	if got := g.ShortestPath("C", "A"); got != nil {
		t.Fatalf("ShortestPath(C, A) = %v, want nil", got)
	}
	if got := g.ShortestPath("A", "D"); got != nil {
		t.Fatalf("ShortestPath(A, D) = %v, want nil (D isolated)", got)
	}
}

func TestPotentialCyclesClosesCycle(t *testing.T) {
	// A beat B beat C; an unplayed C-vs-A game closes a cycle iff C beats A.
	s := nfldata.Season{
		Teams: []string{"A", "B", "C"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "C"),
		},
	}
	g := Build(s)

	pcs := g.PotentialCycles([]nfldata.Game{unplayed("C", "A", 3)})
	if len(pcs) != 1 {
		t.Fatalf("PotentialCycles = %v, want exactly one closing outcome", pcs)
	}
	pc := pcs[0]
	if pc.Winner != "C" || pc.Loser != "A" {
		t.Fatalf("PotentialCycles = %+v, want C beating A to close the cycle", pc)
	}
	if want := []string{"A", "B", "C"}; !reflect.DeepEqual(pc.Cycle, want) {
		t.Fatalf("Cycle = %v, want %v", pc.Cycle, want)
	}
}

func TestPotentialCyclesSkipsPlayedGames(t *testing.T) {
	s := nfldata.Season{
		Teams: []string{"A", "B", "C"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "C"),
		},
	}
	g := Build(s)

	pcs := g.PotentialCycles([]nfldata.Game{game("C", "A")})
	if pcs != nil {
		t.Fatalf("PotentialCycles = %v, want nil for an already-played game", pcs)
	}
}

func TestPotentialCyclesNoPath(t *testing.T) {
	s := nfldata.Season{
		Teams: []string{"A", "B", "C"},
		Games: []nfldata.Game{
			game("A", "B"),
		},
	}
	g := Build(s)

	pcs := g.PotentialCycles([]nfldata.Game{unplayed("B", "C", 3)})
	if pcs != nil {
		t.Fatalf("PotentialCycles = %v, want nil (no existing path could close a cycle)", pcs)
	}
}

func TestSweepFindsCycleAcrossCombination(t *testing.T) {
	// A beat D already. A 4-cycle only closes if all three upcoming games
	// go the "wrong" way (B beats A, C beats B, D beats C), out of 8
	// possible combinations.
	s := nfldata.Season{
		Teams: []string{"A", "B", "C", "D"},
		Games: []nfldata.Game{
			game("A", "D"),
		},
	}
	g := Build(s)

	groupings, total, err := g.Sweep([]nfldata.Game{
		unplayed("B", "A", 2),
		unplayed("C", "B", 2),
		unplayed("D", "C", 2),
	})
	if err != nil {
		t.Fatalf("Sweep returned error: %v", err)
	}
	if total != 8 {
		t.Fatalf("total = %d, want 8", total)
	}
	if len(groupings) != 1 {
		t.Fatalf("groupings = %+v, want exactly one", groupings)
	}
	if want := []string{"A", "D", "C", "B"}; !reflect.DeepEqual(groupings[0].Cycle, want) {
		t.Fatalf("Cycle = %v, want %v", groupings[0].Cycle, want)
	}
	if len(groupings[0].Causes) != 3 {
		t.Fatalf("Causes = %+v, want 3 (the A->D edge is pre-existing, not a cause)", groupings[0].Causes)
	}
}

func TestSweepExcludesExistingCycle(t *testing.T) {
	// A and B already form a cycle; an unplayed C-D game can't be a "new"
	// grouping involving A/B since it's disjoint from them.
	s := nfldata.Season{
		Teams: []string{"A", "B", "C", "D"},
		Games: []nfldata.Game{
			game("A", "B"),
			game("B", "A"),
		},
	}
	g := Build(s)

	groupings, _, err := g.Sweep([]nfldata.Game{unplayed("C", "D", 2)})
	if err != nil {
		t.Fatalf("Sweep returned error: %v", err)
	}
	if len(groupings) != 0 {
		t.Fatalf("groupings = %v, want none (C-D alone can't form a cycle, and A-B isn't new)", groupings)
	}
}

func TestSweepTooManyGames(t *testing.T) {
	s := nfldata.Season{Teams: []string{"A", "B"}, Games: nil}
	g := Build(s)

	var games []nfldata.Game
	for i := 0; i < maxSweepGames+1; i++ {
		games = append(games, unplayed("A", "B", 2))
	}
	if _, _, err := g.Sweep(games); err == nil {
		t.Fatalf("Sweep with %d games: want error, got nil", len(games))
	}
}
