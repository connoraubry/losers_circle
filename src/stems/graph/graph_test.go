package graph

import (
	"fmt"
	"slices"
	"testing"
)

var connectionTestCases = []struct {
	first  string
	second string
}{
	{"1", "2"},
	{"1", "3"},
	{"2", "4"},
	{"3", "5"},
	{"4", "6"},
	{"5", "6"},
}

var nodeTestCases = []string{
	"test1",
	"test2",
	"test3",
	"test4",
	"test5",
	"test6",
}

func TestAddNode(t *testing.T) {
	g := New()

	for idx, name := range nodeTestCases {
		entry_idx, ok := g.AddNode(name)

		if entry_idx != idx {
			t.Errorf("g.AddNode(%v) == %v. Expected %v", name, entry_idx, idx)
		}

		if len(g.Nodes) != idx+1 {
			t.Errorf("len(g.Nodes) == %v. Expected %v", len(g.Nodes), idx+1)
		}

		if !ok {
			t.Errorf("g.AddNode(%v) should not return ok", name)
		}
	}

	for idx, name := range nodeTestCases {
		saved_id := g.NameToId[name]
		if saved_id != idx {
			t.Errorf("g.NameToId[%v] == %v. Expected %v", name, saved_id, idx)
		}
	}

	//try re-adding nodes
	for idx, name := range nodeTestCases {
		entry_idx, ok := g.AddNode(name)
		if entry_idx != idx {
			t.Errorf("g.AddNode(%v) == %v. Expected %v", name, entry_idx, idx)
		}
		if ok {
			t.Errorf("g.AddNode(%v) returned OK. should return false", name)
		}
	}
}

func TestAddNodeMax(t *testing.T) {
	g := New()

	for idx := 0; idx < 34; idx++ {
		name := fmt.Sprintf("%v", idx)

		entry_idx, ok := g.AddNode(name)

		if idx < 32 {
			if entry_idx != idx {
				t.Errorf("g.AddNode(%v) == %v. Expected %v", name, entry_idx, idx)
			}

			if !ok {
				t.Errorf("g.AddNode(%v) should return ok", name)
			}
		} else {
			if entry_idx != -1 {
				t.Errorf("g.AddNode(%v) == %v. Expected %v", name, entry_idx, -1)
			}

			if ok {
				t.Errorf("g.AddNode(%v) should not return ok", name)
			}

		}
	}
}

func TestAddNodes(t *testing.T) {
	g := New()
	g.AddNodes(nodeTestCases)
	for idx, name := range nodeTestCases {
		entry_idx := g.NameToId[name]
		if entry_idx != idx {
			t.Errorf("g.AddNode(%v) == %v. Expected %v", name, entry_idx, idx)
		}
	}
	if len(g.Nodes) != len(nodeTestCases) {
		t.Errorf("len(g.Nodes) == %v. Expected %v", len(g.Nodes), len(nodeTestCases))
	}
}

func TestAddConnections(t *testing.T) {
	g := New()

	expectedNodes := make(map[string]int)

	for _, cnx := range connectionTestCases {
		g.AddConnection(cnx.first, cnx.second)

		if _, ok := expectedNodes[cnx.first]; !ok {
			expectedNodes[cnx.first] = len(expectedNodes)
		}
		if _, ok := expectedNodes[cnx.second]; !ok {
			expectedNodes[cnx.second] = len(expectedNodes)
		}
	}

	for node, idx := range expectedNodes {
		if g.NameToId[node] != idx {
			t.Errorf("g.NameToId[%v] == %v. Expected %v", node, g.NameToId[node], idx)
		}
	}

	for _, cnx := range connectionTestCases {
		firstIdx := g.NameToId[cnx.first]
		secondIdx := g.NameToId[cnx.second]

		firstNode := g.Nodes[firstIdx]
		if !slices.Contains[[]int, int](firstNode.Outgoing, secondIdx) {
			t.Errorf("Node %v is not present in %v outgoing. %v", secondIdx, firstIdx, firstNode.Outgoing)
		}

		secondNode := g.Nodes[secondIdx]
		if !slices.Contains[[]int, int](secondNode.Incoming, firstIdx) {
			t.Errorf("Node %v is not present in %v outgoing. %v", firstIdx, secondIdx, secondNode.Incoming)
		}
	}
}

func TestAddTooManyConnections(t *testing.T) {
	g := New()
	for i := 0; i < 18; i++ {
		first := fmt.Sprintf("%v", i)
		second := fmt.Sprintf("%v", i+32)

		success := g.AddConnection(first, second)
		if i < 16 && !success {
			t.Errorf("Expected success adding index %v", i)
		}
		if i >= 16 && success {
			t.Errorf("Expected failure adding index %v", i)
		}
	}
	success := g.AddConnection("3", "1204124")
	if success {
		t.Errorf("Expected failure adding test case")
	}
}

func TestRemoveConnectionEmptyGraph(t *testing.T) {
	g := New()

	err := g.RemoveConnection("test", "test2")
	if err == nil {
		t.Errorf("Expected error removing connection on empty graph")
	}

	g.AddNode("test")

	err = g.RemoveConnection("test", "test2")
	if err == nil {
		t.Errorf("Expected error removing connection on empty graph")
	}

	g.AddNode("test2")

	err = g.RemoveConnection("test", "test2")
	if err != nil {
		t.Errorf("Did not expect error. Nodes are initialized")
	}
}

func TestRemoveConnection(t *testing.T) {
	g := New()

	for _, cnx := range connectionTestCases {
		g.AddConnection(cnx.first, cnx.second)
	}

	for _, cnx := range connectionTestCases {
		lenOut := len(g.Nodes[g.NameToId[cnx.first]].Outgoing)
		lenIn := len(g.Nodes[g.NameToId[cnx.second]].Incoming)
		g.RemoveConnection(cnx.first, cnx.second)

		newLenOut := len(g.Nodes[g.NameToId[cnx.first]].Outgoing)
		newLenIn := len(g.Nodes[g.NameToId[cnx.second]].Incoming)

		if newLenOut != lenOut-1 {
			t.Errorf("len(first.outgoing) == %v. Expected %v", newLenOut, lenOut)
		}
		if newLenIn != lenIn-1 {
			t.Errorf("len(second.incoming) == %v. Expected %v", newLenIn, lenIn)
		}

	}

	err := g.RemoveConnection("test", "test2")
	if err == nil {
		t.Errorf("Expected error removing connection on empty graph")
	}

	g.AddNode("test")

	err = g.RemoveConnection("test", "test2")
	if err == nil {
		t.Errorf("Expected error removing connection on empty graph")
	}

	g.AddNode("test2")

	err = g.RemoveConnection("test", "test2")
	if err != nil {
		t.Errorf("Did not expect error. Nodes are initialized")
	}
}

func TestNodeToString(t *testing.T) {
	g := New()

	idx, _ := g.AddNode("test1")

	expectedString := fmt.Sprintf("test1:\n  Wins: []\n  Loss: []\n")
	if expectedString != g.NodeToString(idx) {
		t.Errorf("g.NodeToString(%v) gave wrong output: %v %v", idx, expectedString, g.NodeToString(idx))
	}

	var outgoing = []int{1, 2, 3}
	g.AddNode("test2")
	g.AddNode("test3")
	g.AddNode("test 4")
	g.Nodes[idx].Outgoing = outgoing

	expectedString = fmt.Sprintf("test1:\n  Wins: [test2 test3 test 4]\n  Loss: []\n")
	if expectedString != g.NodeToString(idx) {
		t.Errorf("g.NodeToString(%v) gave wrong output: %v %v", idx, expectedString, g.NodeToString(idx))
	}

	var incoming = []int{3, 2}
	g.Nodes[idx].Incoming = incoming

	expectedString = fmt.Sprintf("test1:\n  Wins: [test2 test3 test 4]\n  Loss: [test 4 test3]\n")
	if expectedString != g.NodeToString(idx) {
		t.Errorf("g.NodeToString(%v) gave wrong output: %v %v", idx, expectedString, g.NodeToString(idx))
	}
}
