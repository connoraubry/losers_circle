package cycle

// tarjanSCC partitions the graph's nodes into strongly connected
// components using Tarjan's algorithm, returning a component id per node
// index. A team can only be part of a cycle if its component has 2 or
// more teams in it (no self-loops exist, since a team never plays itself).
func tarjanSCC(adj [][]int) []int {
	n := len(adj)
	index := make([]int, n)
	low := make([]int, n)
	onStack := make([]bool, n)
	comp := make([]int, n)
	for i := range index {
		index[i] = -1
	}

	var stack []int
	nextIndex := 0
	nextComp := 0

	var strongconnect func(v int)
	strongconnect = func(v int) {
		index[v] = nextIndex
		low[v] = nextIndex
		nextIndex++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range adj[v] {
			switch {
			case index[w] == -1:
				strongconnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			case onStack[w]:
				if index[w] < low[v] {
					low[v] = index[w]
				}
			}
		}

		if low[v] == index[v] {
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp[w] = nextComp
				if w == v {
					break
				}
			}
			nextComp++
		}
	}

	for v := 0; v < n; v++ {
		if index[v] == -1 {
			strongconnect(v)
		}
	}
	return comp
}

// sccSizes counts how many nodes fall in each component id from comp.
func sccSizes(comp []int) []int {
	sizes := make([]int, len(comp))
	for _, c := range comp {
		sizes[c]++
	}
	return sizes
}
