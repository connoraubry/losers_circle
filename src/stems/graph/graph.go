package graph

import (
	"fmt"
	"slices"
)

type Graph struct {
	Nodes    []*Node
	NameToId map[string]int
}

type Node struct {
	ID       int
	Name     string
	Outgoing []int
	Incoming []int
}

func (g *Graph) NodeToString(idx int) string {
	n := g.Nodes[idx]

	var outgoingNames []string
	var incomingNames []string

	for _, oIdx := range n.Outgoing {
		newName := g.Nodes[oIdx].Name
		outgoingNames = append(outgoingNames, newName)
	}
	for _, nIdx := range n.Incoming {
		newName := g.Nodes[nIdx].Name
		incomingNames = append(incomingNames, newName)
	}

	s := fmt.Sprintf("%v:\n", n.Name)
	s = s + fmt.Sprintf("  Wins: %v\n", outgoingNames)
	s = s + fmt.Sprintf("  Loss: %v\n", incomingNames)

	return s
}

func New() *Graph {
	g := &Graph{}
	g.NameToId = make(map[string]int)
	return g
}

func (g *Graph) AddNode(name string) (int, bool) {

	if id, ok := g.NameToId[name]; ok {
		return id, false
	}

	if len(g.Nodes) == 32 {
		return -1, false
	}

	id := len(g.Nodes)
	newNode := &Node{
		ID:   id,
		Name: name,
	}

	g.Nodes = append(g.Nodes, newNode)
	g.NameToId[name] = id

	return id, true
}

func (g *Graph) AddNodes(names []string) {
	for _, name := range names {
		g.AddNode(name)
	}
}

func (g *Graph) AddConnection(start, end string) bool {
	nodeA, ok := g.NameToId[start]
	if !ok {
		nodeA, ok = g.AddNode(start)
		if !ok {
			return false
		}
	}

	nodeB, ok := g.NameToId[end]
	if !ok {
		nodeB, ok = g.AddNode(end)
		if !ok {
			return false
		}
	}

	if !slices.Contains(g.Nodes[nodeA].Outgoing, nodeB) {
		g.Nodes[nodeA].Outgoing = append(g.Nodes[nodeA].Outgoing, nodeB)
	}
	if !slices.Contains(g.Nodes[nodeB].Incoming, nodeA) {
		g.Nodes[nodeB].Incoming = append(g.Nodes[nodeB].Incoming, nodeA)
	}

	return true
}

func (g *Graph) RemoveConnection(start, end string) error {

	startIdx, ok := g.NameToId[start]
	if !ok {
		return fmt.Errorf("Node %v not present in graph", start)
	}
	endIdx, ok := g.NameToId[end]
	if !ok {
		return fmt.Errorf("Node %v not present in graph", end)
	}

	startNode := g.Nodes[startIdx]
	endNode := g.Nodes[endIdx]

	startNode.Outgoing = slices.DeleteFunc(startNode.Outgoing, func(n int) bool {
		return n == endIdx
	})

	endNode.Incoming = slices.DeleteFunc(endNode.Incoming, func(n int) bool {
		return n == startIdx
	})

	return nil
}

// func (g *Graph) PrintCycles() {
// 	for s_idx, loop := range g.Stems.Loops {
// 		first := g.Nodes[s_idx].Name
//
// 		fmt.Printf("Longest loop containing %v (len: %v):\n  ", first, len(loop)+1)
// 		for _, elem := range loop {
// 			fmt.Printf("%v, ", g.Nodes[elem].Name)
// 		}
// 		fmt.Printf("\n")
// 	}
// }
