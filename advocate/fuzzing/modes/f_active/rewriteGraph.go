// Copyright (c) 2024 Erik Kassubek
//
// File: waitGroup.go
// Brief: Rewrite for negative wait group counter
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_active

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/fuzzing/f_constraints"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/results/bugs"
	"advocate/utils/types"
	"fmt"
)

type Edge struct {
	to  int
	rev int
	cap int
}

type Dinic struct {
	graph [][]Edge
	level []int
	iter  []int
}

func NewDinic(n int) *Dinic {
	return &Dinic{
		graph: make([][]Edge, n),
		level: make([]int, n),
		iter:  make([]int, n),
	}
}

func (d *Dinic) AddEdge(from, to, cap int) {
	forward := Edge{
		to:  to,
		rev: len(d.graph[to]),
		cap: cap,
	}

	reverse := Edge{
		to:  from,
		rev: len(d.graph[from]),
		cap: 0,
	}

	d.graph[from] = append(d.graph[from], forward)
	d.graph[to] = append(d.graph[to], reverse)
}

func (d *Dinic) bfs(s, t int) bool {
	for i := range d.level {
		d.level[i] = -1
	}

	queue := []int{s}
	d.level[s] = 0

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		for _, e := range d.graph[v] {
			if e.cap > 0 && d.level[e.to] < 0 {
				d.level[e.to] = d.level[v] + 1
				queue = append(queue, e.to)
			}
		}
	}

	return d.level[t] >= 0
}

func (d *Dinic) dfs(v, t, flow int) int {
	if v == t {
		return flow
	}

	for d.iter[v] < len(d.graph[v]) {
		i := d.iter[v]
		e := &d.graph[v][i]

		if e.cap > 0 && d.level[e.to] == d.level[v]+1 {
			f := flow
			if e.cap < f {
				f = e.cap
			}

			pushed := d.dfs(e.to, t, f)

			if pushed > 0 {
				to := e.to
				rev := e.rev

				d.graph[v][i].cap -= pushed
				d.graph[to][rev].cap += pushed

				return pushed
			}
		}

		d.iter[v]++
	}

	return 0
}

func (d *Dinic) MaxFlow(s, t int) int {
	total := 0

	for d.bfs(s, t) {
		for i := range d.iter {
			d.iter[i] = 0
		}

		for {
			pushed := d.dfs(s, t, int(^uint(0)>>1))
			if pushed == 0 {
				break
			}
			total += pushed
		}
	}

	return total
}

// Reachable returns the vertices reachable from s
// through edges with positive residual capacity.
func (d *Dinic) Reachable(s int) []bool {
	seen := make([]bool, len(d.graph))
	queue := []int{s}

	seen[s] = true

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		for _, e := range d.graph[v] {
			if e.cap > 0 && !seen[e.to] {
				seen[e.to] = true
				queue = append(queue, e.to)
			}
		}
	}

	return seen
}

// Create a new trace for a negative wait group counter (done before add) or unlock before lock
//
// Parameter:
//   - bug Bug: The bug to create a trace for
func rewriteGraph(bug bugs.Bug) (*f_constraints.Constraint, error) {
	switch bug.Type {
	case helper.PNegWG:
		log.Info("Start rewrite for negative wait group counter...")
	case helper.PUnlockBeforeLock:
		log.Info("Start rewrite for unlock before lock...")
	}

	// unlock before lock: arg1 - unlock, arg2 - lock
	// negative wait group: arg1 - done, arg2 - add

	elements := types.MergeLists(bug.TraceElement1, bug.TraceElement2)
	dec := types.SliceAsMap(bug.TraceElement1)
	inc := types.SliceAsMap(bug.TraceElement2)
	n := len(bug.TraceElement1) + len(bug.TraceElement2)

	if n == 0 {
		return nil, fmt.Errorf("No element in bug")
	}

	// Map each element to its vertex index.
	index := make(map[trace.Element]int, n)

	for i, x := range elements {
		if _, exists := index[x]; exists {
			// Duplicate elements are not supported.
			return nil, fmt.Errorf("Duplicate element in bug")
		}
		index[x] = i
	}

	// Build the precedence graph.
	// An edge u -> v means u must appear before v.
	type Pair struct {
		from int
		to   int
	}

	var precedence []Pair
	adj := make([][]int, n)

	addPrecedence := func(u, v int) {
		precedence = append(precedence, Pair{u, v})
		adj[u] = append(adj[u], v)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			switch a_clock.GetHappensBefore(elements[i].GetVC(a_clock.Strong), elements[j].GetVC(a_clock.Strong)) {
			case a_hb.Before:
				addPrecedence(i, j)
			case a_hb.After:
				addPrecedence(j, i)
			}
		}
	}

	if _, ok := topoSubset(adj, nil); !ok {
		return nil, fmt.Errorf("Cyclic dependency graph")
	}

	// Construct the maximum-weight closure network.
	//
	// Profit = -counter contribution:
	//   dec: +1
	//   inc: -1
	//
	// Selecting a vertex means including it in the prefix
	// that should make the counter negative.

	source := n
	sink := n + 1

	dinic := NewDinic(n + 2)

	totalPositive := 0
	profit := make([]int, n)

	for i, x := range elements {
		switch {
		case dec[x] && inc[x]:
			return nil, fmt.Errorf("Non disjoint bug sets")

		case dec[x]:
			profit[i] = 1
			dinic.AddEdge(source, i, 1)
			totalPositive++

		case inc[x]:
			profit[i] = -1
			dinic.AddEdge(i, sink, 1)

		default:
			// An element in neither set contributes zero.
			profit[i] = 0
		}
	}

	// If b must follow a, selecting b requires selecting a.
	//
	// Therefore, for every precedence edge a -> b,
	// add a high-capacity edge b -> a.
	//
	// Cutting such an edge must cost more than any possible
	// benefit from selecting additional vertices.
	inf := totalPositive + 1

	for _, e := range precedence {
		dinic.AddEdge(e.to, e.from, inf)
	}

	dinic.MaxFlow(source, sink)

	// The source side of the minimum cut gives a
	// maximum-profit predecessor-closed subset.
	reachable := dinic.Reachable(source)

	selected := make([]bool, n)
	totalProfit := 0

	for i := 0; i < n; i++ {
		if reachable[i] {
			selected[i] = true
			totalProfit += profit[i]
		}
	}

	// Positive profit means the counter will be negative
	// after processing this subset.
	if totalProfit <= 0 {
		return nil, fmt.Errorf("No possible reordering")
	}

	// Topologically sort the selected prefix.
	first, ok := topoSubset(adj, selected)
	if !ok {
		return nil, fmt.Errorf("No possible topological reordering of prefix")
	}

	// Then topologically sort the remaining elements.
	remaining := make([]bool, n)

	for i := 0; i < n; i++ {
		remaining[i] = !selected[i]
	}

	second, ok := topoSubset(adj, remaining)
	if !ok {
		return nil, fmt.Errorf("No possible topological reordering")
	}

	result := f_constraints.NewConstraint()

	for _, i := range first {
		result.Add(elements[i])
	}

	for _, i := range second {
		result.Add(elements[i])
	}

	// Defensive verification of the counter requirement.
	counter := 0
	wentNegative := false

	for _, x := range result.Elems {
		if inc[x] {
			counter++
		}
		if dec[x] {
			counter--
		}

		if counter < 0 {
			wentNegative = true
			break
		}
	}

	if !wentNegative {
		return nil, fmt.Errorf("Counter never went negative")
	}

	return &result, nil
}

// topoSubset returns a topological ordering of the vertices
// included in subset.
//
// If subset == nil, all vertices are included.
//
// Returns ok=false if the induced graph contains a cycle.
func topoSubset(adj [][]int, subset []bool) ([]int, bool) {
	n := len(adj)
	indegree := make([]int, n)
	count := 0

	included := func(v int) bool {
		return subset == nil || subset[v]
	}

	for v := 0; v < n; v++ {
		if !included(v) {
			continue
		}

		count++

		for _, to := range adj[v] {
			if included(to) {
				indegree[to]++
			}
		}
	}

	queue := make([]int, 0, count)

	for v := 0; v < n; v++ {
		if included(v) && indegree[v] == 0 {
			queue = append(queue, v)
		}
	}

	order := make([]int, 0, count)

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		order = append(order, v)

		for _, to := range adj[v] {
			if !included(to) {
				continue
			}

			indegree[to]--

			if indegree[to] == 0 {
				queue = append(queue, to)
			}
		}
	}

	return order, len(order) == count
}
