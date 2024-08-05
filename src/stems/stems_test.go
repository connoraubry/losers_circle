package stems

import (
	"math"
	"testing"

	"github.com/connoraubry/losers_circle/src/stems/graph"
)

type DummyConnections [][2]string

func makeDummyS(connections DummyConnections) *Stems {
	g := graph.New()

	for _, cnx := range connections {
		g.AddConnection(cnx[0], cnx[1])
	}
	s := New()
	s.Init(g)
	return s
}

func makeDummyConnections() DummyConnections {
	d := DummyConnections{
		{"1", "2"},
		{"1", "3"},
		{"2", "4"},
		{"3", "5"},
		{"4", "6"},
		{"5", "6"},
	}
	return d
}

func TestInit(t *testing.T) {

	connections := makeDummyConnections()
	s := makeDummyS(connections)
	count := 0
	expected := len(connections)
	for _, entry := range s.Levels[2].Starts {
		for _, end := range entry.Ends {
			count += len(end.Bitmasks)
		}
	}

	if count != expected {
		t.Fatalf("Error counting length 2 connections: Got %v expected %v", count, expected)
	}
}

func TestDummyLevel3(t *testing.T) {
	connections := makeDummyConnections()
	s := makeDummyS(connections)
	s.ProcessNextLevel(2, 2)
	count := 0
	expected := 4

	var foundBitmasks []BitmaskStruct

	for _, entry := range s.Levels[3].Starts {
		for _, end := range entry.Ends {
			count += len(end.Bitmasks)
			foundBitmasks = append(foundBitmasks, end.Bitmasks...)
		}
	}

	if count != expected {
		t.Fatalf("Error counting length 3 connections: Got %v expected %v. Found: %v", count, expected, foundBitmasks)
	}
}

func TestDummyLevel4(t *testing.T) {
	connections := makeDummyConnections()
	s := makeDummyS(connections)
	s.ProcessNextLevel(2, 2)
	s.ProcessNextLevel(3, 2)
	count := 0
	expected := 2

	var foundBitmasks []BitmaskStruct

	for _, entry := range s.Levels[4].Starts {
		for _, end := range entry.Ends {
			count += len(end.Bitmasks)
			foundBitmasks = append(foundBitmasks, end.Bitmasks...)
		}
	}

	if count != expected {
		t.Fatalf("Error counting length 3 connections: Got %v expected %v. Found: %v", count, expected, foundBitmasks)
	}
}

// func TestEndAdd(t *testing.T) {
//
// 	e := NewStemEnd()
// 	bs := BitmaskStruct{
// 		Bitmask: 32,
// 	}
//
// 	e.Add(bs)
//
// 	if len(e.Bitmasks) != 1 {
// 		t.Errorf("Bitmasks list not increasing after added test")
// 	}
//
// 	bs2 := BitmaskStruct{
// 		Bitmask: 123,
// 	}
//
// 	e.Add(bs2)
//
// 	if len(e.Bitmasks) != 2 {
// 		t.Errorf("Second bitmask add didn't work")
// 	}
//
// 	e.Add(bs2)
// 	if len(e.Bitmasks) != 2 {
// 		t.Errorf("Third bitmask add not ignored")
// 	}
// }

func TestIdToBitmask(t *testing.T) {
	for i := -10; i < 50; i++ {
		var expected uint32
		if i < 32 {
			expected = uint32(math.Pow(2, float64(i)))
		} else {
			expected = 0
		}
		output := IdToBitmask(i)
		if expected != output {
			t.Errorf("IdToBitmask(%v) == %v. Expected %v", i, output, expected)
		}
	}
}
