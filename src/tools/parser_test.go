package tools

import (
	"fmt"
	"testing"
)

func TestParserB(t *testing.T) {
	p := NewParser()
	val := byte('f')
	res := p.ByteMap[val]
	fmt.Println(val)
	fmt.Println(res)
	if len(res) != 2 {
		t.Fatalf(`Byte map result length != %v. Got %v`, 2, len(res))
	}
	res0 := res[0]
	if !res0.Selected {
		t.Fatalf(`Res0 selected = false`)
	}
	if !res0.HomeWon {
		t.Fatalf(`Res0 homewon = false`)
	}

	res1 := res[1]
	if !res1.Selected {
		t.Fatalf(`Res1 selected = falase`)
	}
	if !res1.HomeWon {
		t.Fatalf(`Res1 homewon = false`)
	}
}
func TestParserNine(t *testing.T) {
	p := NewParser()
	val := byte('9')
	res := p.ByteMap[val]
	fmt.Println(val)
	fmt.Println(res)
	if len(res) != 2 {
		t.Fatalf(`Byte map result length != %v. Got %v`, 2, len(res))
	}
	res0 := res[0]
	if !res0.Selected {
		t.Fatalf(`Res0 selected = false`)
	}
	if res0.HomeWon {
		t.Fatalf(`Res0 homewon = true`)
	}

	res1 := res[1]
	if res1.Selected {
		t.Fatalf(`Res1 selected = true`)
	}
	if !res1.HomeWon {
		t.Fatalf(`Res1 homewon = false`)
	}
}

func TestParserWeek(t *testing.T) {
	p := NewParser()
	week := []byte("c2e")
	expectedGameList := []GameResult{
		{true, true}, {false, false},
		{false, false}, {true, false},
		{true, true}, {true, false},
	}
	weekRes := p.ParseWeek(week)

	if len(weekRes) != len(expectedGameList) {
		t.Fatalf(`ParseWeek result length != expected. %v != %v`, len(weekRes), len(expectedGameList))
	}
	for idx := range weekRes {
		truth := expectedGameList[idx]
		result := weekRes[idx]

		if truth != result {
			t.Fatalf("Structs differ. %v != %v", truth, result)
		}
	}

}
