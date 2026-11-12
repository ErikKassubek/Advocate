// Copyright (c) 2024 
//
// File: analysisGraph.go
// Brief: Functions to use a graph for analysis. Used fop negative wait group
//   counter and unlock before lock
//
//
// License: BSD-3-Clause

package a_scenarios

import (
	"gosc/analysis/a_hb"
	"gosc/analysis/hb/a_clock"
	"gosc/trace"
	"fmt"
	"math"
)

var (
	source = trace.EmptyWait(-1)
	drain  = trace.EmptyWait(-2)
)

// Build a st graph for a wait group.
// The graph has the following structure:
// - a start node s
// - a end node t
// - edges from s to all done operations
// - edges from all add operations to t
// - edges from done to add if the add happens before the done
//
// Parameter:
//   - increases map[int][]TraceElement: Operations that increase the "counter" (adds and locks)
//   - decreases map[int][]TraceElement: Operations that decrease the "counter" (dones and unlocks)
//
// Returns:
//   - []Edge: The graph
func buildResidualGraph(increases []trace.Element, decreases []trace.Element) map[trace.Element][]trace.Element {
	graph := make(map[trace.Element][]trace.Element, 0)
	graph[&source] = []trace.Element{}
	graph[&drain] = []trace.Element{}

	// add edges from s to all done operations
	for _, elem := range decreases {
		// graph[elem] = []trace.Element{}
		graph[&source] = append(graph[&source], elem)
	}

	// add edges from all add operations to t
	for _, elem := range increases {
		graph[elem] = []trace.Element{&drain}
	}

	// add edge from done to add if the add happens before the done
	for _, elemDecrease := range decreases {
		for _, elemIncrease := range increases {
			if a_clock.GetHappensBefore(elemIncrease.VC(a_clock.Strong), elemDecrease.VC(a_clock.Strong)) == a_hb.Before {
				graph[elemDecrease] = append(graph[elemDecrease], elemIncrease)
			}
		}
	}

	return graph
}

// Calculate the maximum flow of a graph using the ford fulkerson algorithm
//
// Parameter:
//   - graph map[TraceElement][]TraceElement: The graph
//
// Returns:
//   - int: The maximum flow
//   - map[TraceElement][]TraceElement: The graph with max flow
//   - error
func calculateMaxFlow(graph map[trace.Element][]trace.Element) (int, map[trace.Element][]trace.Element, error) {
	maxFlow := 0
	maxNumberRounds := 0
	for _, val := range graph {
		maxNumberRounds += len(val)
	}
	maxNumberRounds = 1e5 * int(math.Pow(float64(maxNumberRounds), 3.))

	for i := 0; i < int(maxNumberRounds); i++ { // max number rounds to prevent infinite loop
		path, flow := findPath(graph)
		if flow == 0 {
			return maxFlow, graph, nil
		}

		maxFlow += flow
		for i := 0; i < len(path)-1; i++ {
			graph[path[i]] = append(graph[path[i]], path[i+1])
			graph[path[i+1]] = remove(graph[path[i+1]], path[i])
		}
	}

	return maxFlow, graph, fmt.Errorf("Too many rounds")
}

// Find a path in a graph using a breadth-fifoirst search
//
// Parameter:
//   - map[TraceElement][]TraceElement
//
// Returns:
//   - []TraceElement: The path
//   - int: The flow
func findPath(graph map[trace.Element][]trace.Element) ([]trace.Element, int) {
	visited := make(map[trace.Element]bool, 0)

	queue := []trace.Element{&source}
	visited[&source] = true
	parents := make(map[trace.Element]trace.Element, 0)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node.IsEqual(&drain) {
			path := []trace.Element{}
			for !node.IsEqual(&source) {
				path = append(path, node)
				node = parents[node]
			}
			path = append(path, &source)

			return path, 1
		}

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
				parents[neighbor] = node
			}
		}
	}
	return []trace.Element{}, 0
}

// Remove an element from a list
//
// Parameter:
//   - list []TraceElement: The list
//   - element TraceElement: The element to remove
//
// Returns:
//   - []string: The list without the element
func remove(list []trace.Element, element trace.Element) []trace.Element {
	for i, e := range list {
		if element.IsEqual(e) {
			list = append(list[:i], list[i+1:]...)
			return list
		}
	}
	return list
}
