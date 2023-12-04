package graph

import (
	"fmt"

	"github.com/connoraubry/losers_circle/server/src/graph/set"
)

type Graph struct {
	Nodes       map[string]*Node
	NodeToCycle map[string][]string
	Valid       set.Set
	Verbose     bool
}

type Node struct {
	Name     string
	Idx      int
	Outgoing []*Node
	Incoming []*Node
}

type Connection struct {
	From string
	To   string
}

func NewCnx(from, to string) Connection {
	return Connection{
		From: from,
		To:   to,
	}
}

func New() *Graph {
	g := &Graph{}
	g.Nodes = make(map[string]*Node)
	g.NodeToCycle = make(map[string][]string)
	g.Valid = *set.New()

	return g
}

func (g *Graph) Print() {
	fmt.Printf("Graph\n")
	fmt.Printf("Num nodes: %v\n", len(g.Nodes))

	for nodeName := range g.Nodes {
		fmt.Printf("Nodes: %v: %+v\n", nodeName, g.NodeToCycle[nodeName])
	}
}

func (g *Graph) AddNodes(nodeList []string) {
	for _, name := range nodeList {
		g.AddNode(name)
	}
}

func (g *Graph) AddNode(name string) *Node {
	g.Nodes[name] = &Node{Name: name, Idx: len(g.Nodes)}
	g.Valid.Add(name)
	g.NodeToCycle[name] = []string{}
	return g.Nodes[name]
}

func (g *Graph) AddConnection(con Connection) {
	nodeA, ok := g.Nodes[con.From]
	if !ok {
		nodeA = g.AddNode(con.From)
	}
	nodeB, ok := g.Nodes[con.To]
	if !ok {
		nodeB = g.AddNode(con.To)
	}
	nodeA.Outgoing = append(nodeA.Outgoing, nodeB)
	nodeB.Incoming = append(nodeB.Incoming, nodeA)
}

func (g *Graph) RemoveConnection(con Connection) {
	nodeA := g.Nodes[con.From]
	nodeB := g.Nodes[con.To]

	removeIdx := -1
	for idx, node := range nodeA.Outgoing {
		if node == nodeB {
			removeIdx = idx
			break
		}
	}
	if removeIdx != -1 {
		nodeA.Outgoing = append(nodeA.Outgoing[:removeIdx], nodeA.Outgoing[removeIdx+1:]...)
	}

	removeIdx = -1
	for idx, node := range nodeB.Incoming {
		if node == nodeA {
			removeIdx = idx
			break
		}
	}
	if removeIdx != -1 {
		nodeB.Incoming = append(nodeB.Incoming[:removeIdx], nodeB.Incoming[removeIdx+1:]...)
	}
}

func (g *Graph) ResetEval() {
	g.ResetValidity()
	g.NodeToCycle = make(map[string][]string)
}

func (g *Graph) ResetValidity() {
	for _, node := range g.Nodes {
		g.Valid.Add(node.Name)
	}
}

func (g *Graph) EvaluateValidity() int {
	for _, node := range g.Nodes {
		g.SetValidityOutgoing(node)
		g.SetValidityIncoming(node)
	}
	valid_count := 0
	for _, node := range g.Nodes {
		if g.Valid.IsIn(node.Name) {
			valid_count += 1
		}
	}

	if g.Verbose {
		for _, node := range g.Nodes {
			fmt.Println(node.Name, g.Valid.IsIn(node.Name), len(node.Incoming), len(node.Outgoing))
		}
	}
	return valid_count
}

func (g *Graph) SetValidityIncoming(node *Node) {

	if !g.Valid.IsIn(node.Name) {
		return
	}

	invalid := true
	for _, n := range node.Incoming {
		//if there are any valid parents, break
		if g.Valid.IsIn(n.Name) {
			invalid = false
			break
		}
	}

	if invalid {
		g.Valid.Remove(node.Name)
		for _, n := range node.Outgoing {
			g.SetValidityIncoming(n)
		}
	}
}
func (g *Graph) SetValidityOutgoing(node *Node) {
	if !g.Valid.IsIn(node.Name) {
		return
	}

	invalid := true
	for _, n := range node.Outgoing {
		//if there are any valid parents, break
		if g.Valid.IsIn(n.Name) {
			invalid = false
			break
		}
	}

	if invalid {
		g.Valid.Remove(node.Name)
		for _, n := range node.Incoming {
			g.SetValidityOutgoing(n)
		}
	}
}

func (g *Graph) EvaluateCycles() {

	g.EvaluateValidity()

	count := 0
	for nodeName := range g.Nodes {
		//fmt.Println("Evaluating", nodeName)
		if !g.Valid.IsIn(nodeName) {
			continue
		}

		if cycle, ok := g.NodeToCycle[nodeName]; ok {
			if len(cycle) > (len(g.Nodes) / 2) {
				continue
			}
		}

		cycle, innerCount := g.FindLongestCycleWithTeam(nodeName)
		for _, entry := range cycle {
			g.NodeToCycle[entry] = cycle
		}
		g.NodeToCycle[nodeName] = cycle
		count += innerCount
	}
}
func (g *Graph) FindLongestCycleWithTeam(node string) ([]string, int) {
	visited := make(map[string]int)
	var res []string
	count := 0

	maxTarget := g.Valid.Len()
	//fmt.Println("Max target", maxTarget)

	res, innerCount := g.FindLongestCycleRecursive(node, visited, maxTarget)
	count += innerCount

	return res, count
}

func (g *Graph) FindLongestCycleRecursive(node string, visited map[string]int, targetLen int) ([]string, int) {
	var result []string
	var count int = 0
	n := len(visited)
	if level, ok := visited[node]; ok {
		if level == 0 {
			result = make([]string, n)
			for node, level := range visited {
				result[level] = node
			}
		}
		return result, 1
	} else if n >= targetLen {
		return result, 1
	}

	visited[node] = n

	for _, nextNode := range g.Nodes[node].Outgoing {
		if !g.Valid.IsIn(nextNode.Name) {
			continue
		}

		maxCycle, innerCount := g.FindLongestCycleRecursive(nextNode.Name, visited, targetLen)
		count += innerCount
		if len(maxCycle) > len(result) {
			result = maxCycle
		}

		if len(maxCycle) == targetLen {
			return result, count
		}
	}

	delete(visited, node)
	return result, count
}
