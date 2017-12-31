// Copyright ©2014 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multi

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/internal/uid"
)

var (
	wug *WeightedUndirectedGraph

	_ graph.Graph                        = wug
	_ graph.Undirected                   = wug
	_ graph.WeightedUndirected           = wug
	_ graph.Multigraph                   = wug
	_ graph.UndirectedMultigraph         = wug
	_ graph.WeightedUndirectedMultigraph = wug
)

// WeightedUndirectedGraph implements a generalized undirected graph.
type WeightedUndirectedGraph struct {
	EdgeWeightFunc func([]graph.WeightedLine) float64

	nodes map[int64]graph.Node
	lines map[int64]map[int64]map[int64]graph.WeightedLine

	nodeIDs uid.Set
	lineIDs uid.Set
}

// NewWeightedUndirectedGraph returns an WeightedUndirectedGraph with the specified self and absent
// edge weight values.
func NewWeightedUndirectedGraph() *WeightedUndirectedGraph {
	return &WeightedUndirectedGraph{
		nodes: make(map[int64]graph.Node),
		lines: make(map[int64]map[int64]map[int64]graph.WeightedLine),

		nodeIDs: uid.NewSet(),
		lineIDs: uid.NewSet(),
	}
}

// NewNode returns a new unique Node to be added to g. The Node's ID does
// not become valid in g until the Node is added to g.
func (g *WeightedUndirectedGraph) NewNode() graph.Node {
	if len(g.nodes) == 0 {
		return Node(0)
	}
	if int64(len(g.nodes)) == uid.Max {
		panic("simple: cannot allocate node: no slot")
	}
	return Node(g.nodeIDs.NewID())
}

// AddNode adds n to the graph. It panics if the added node ID matches an existing node ID.
func (g *WeightedUndirectedGraph) AddNode(n graph.Node) {
	if _, exists := g.nodes[n.ID()]; exists {
		panic(fmt.Sprintf("simple: node ID collision: %d", n.ID()))
	}
	g.nodes[n.ID()] = n
	g.lines[n.ID()] = make(map[int64]map[int64]graph.WeightedLine)
	g.nodeIDs.Use(n.ID())
}

// RemoveNode removes n from the graph, as well as any edges attached to it. If the node
// is not in the graph it is a no-op.
func (g *WeightedUndirectedGraph) RemoveNode(n graph.Node) {
	if _, ok := g.nodes[n.ID()]; !ok {
		return
	}
	delete(g.nodes, n.ID())

	for from := range g.lines[n.ID()] {
		delete(g.lines[from], n.ID())
	}
	delete(g.lines, n.ID())

	g.nodeIDs.Release(n.ID())
}

// NewLine returns a new WeightedLine from the source to the destination node.
// The returned WeightedLine will have a graph-unique ID.
// The Line's ID does not become valid in g until the Line is added to g.
func (g *WeightedUndirectedGraph) NewLine(from, to graph.Node) graph.WeightedLine {
	return &WeightedLine{F: from, T: to, UID: g.lineIDs.NewID()}
}

// SetWeighted adds l, a line from one node to another. If the nodes do not exist, they are added.
func (g *WeightedUndirectedGraph) SetWeighted(l graph.WeightedLine) {
	var (
		from = l.From()
		fid  = from.ID()
		to   = l.To()
		tid  = to.ID()
		lid  = l.ID()
	)

	if !g.Has(from) {
		g.AddNode(from)
	}
	if g.lines[fid][tid] == nil {
		g.lines[fid][tid] = make(map[int64]graph.WeightedLine)
	}
	if !g.Has(to) {
		g.AddNode(to)
	}
	if g.lines[tid][fid] == nil {
		g.lines[tid][fid] = make(map[int64]graph.WeightedLine)
	}

	g.lines[fid][tid][lid] = l
	g.lines[tid][fid][lid] = l
	g.lineIDs.Use(lid)
}

// RemoveLine removes l from the graph, leaving the terminal nodes. If the line does not exist
// it is a no-op.
func (g *WeightedUndirectedGraph) RemoveLine(l graph.WeightedLine) {
	from, to := l.From(), l.To()
	if _, ok := g.nodes[from.ID()]; !ok {
		return
	}
	if _, ok := g.nodes[to.ID()]; !ok {
		return
	}

	delete(g.lines[from.ID()], to.ID())
	delete(g.lines[to.ID()], from.ID())
	if len(g.lines[to.ID()][from.ID()]) == 0 {
		delete(g.lines[to.ID()], from.ID())
	}
	g.lineIDs.Release(l.ID())
}

// Node returns the node in the graph with the given ID.
func (g *WeightedUndirectedGraph) Node(id int64) graph.Node {
	return g.nodes[id]
}

// Has returns whether the node exists within the graph.
func (g *WeightedUndirectedGraph) Has(n graph.Node) bool {
	_, ok := g.nodes[n.ID()]
	return ok
}

// Nodes returns all the nodes in the graph.
func (g *WeightedUndirectedGraph) Nodes() []graph.Node {
	if len(g.nodes) == 0 {
		return nil
	}
	nodes := make([]graph.Node, len(g.nodes))
	i := 0
	for _, n := range g.nodes {
		nodes[i] = n
		i++
	}
	return nodes
}

// Edges returns all the edges in the graph. Each edge in the returned slice
// is a multi.Edge.
func (g *WeightedUndirectedGraph) Edges() []graph.Edge {
	if len(g.lines) == 0 {
		return nil
	}
	var edges []graph.Edge
	seen := make(map[int64]struct{})
	for _, u := range g.lines {
		for _, e := range u {
			var lines Edge
			for _, l := range e {
				lid := l.ID()
				if _, ok := seen[lid]; ok {
					continue
				}
				seen[lid] = struct{}{}
				lines = append(lines, l)
			}
			if len(lines) != 0 {
				edges = append(edges, lines)
			}
		}
	}
	return edges
}

// From returns all nodes in g that can be reached directly from n.
func (g *WeightedUndirectedGraph) From(n graph.Node) []graph.Node {
	if !g.Has(n) {
		return nil
	}

	nodes := make([]graph.Node, len(g.lines[n.ID()]))
	i := 0
	for from := range g.lines[n.ID()] {
		nodes[i] = g.nodes[from]
		i++
	}
	return nodes
}

// HasEdgeBetween returns whether an edge exists between nodes x and y.
func (g *WeightedUndirectedGraph) HasEdgeBetween(x, y graph.Node) bool {
	_, ok := g.lines[x.ID()][y.ID()]
	return ok
}

// Lines returns the lines from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *WeightedUndirectedGraph) Lines(u, v graph.Node) []graph.Line {
	return g.LinesBetween(u, v)
}

// LinesBetween returns the lines between nodes x and y.
func (g *WeightedUndirectedGraph) LinesBetween(x, y graph.Node) []graph.Line {
	edge := g.lines[x.ID()][y.ID()]
	if len(edge) == 0 {
		return nil
	}
	var lines []graph.Line
	seen := make(map[int64]struct{})
	for _, l := range edge {
		lid := l.ID()
		if _, ok := seen[lid]; ok {
			continue
		}
		seen[lid] = struct{}{}
		lines = append(lines, l)
	}
	return lines
}

// Edge returns the edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
// The returned graph.Edge is a multi.WeightedEdge if an edge exists.
func (g *WeightedUndirectedGraph) Edge(u, v graph.Node) graph.Edge {
	return g.WeightedEdge(u, v)
}

// EdgeBetween returns the edge between nodes x and y.
func (g *WeightedUndirectedGraph) EdgeBetween(x, y graph.Node) graph.Edge {
	return g.WeightedEdge(x, y)
}

// WeightedEdge returns the weighted edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
// The returned graph.WeightedEdge is a multi.WeightedEdge if an edge exists.
func (g *WeightedUndirectedGraph) WeightedEdge(u, v graph.Node) graph.WeightedEdge {
	lines := g.WeightedLines(u, v)
	if len(lines) == 0 {
		return nil
	}
	return WeightedEdge{Lines: lines, WeightFunc: g.EdgeWeightFunc}
}

// WeightedEdgeBetween returns the weighted edge between nodes x and y.
func (g *WeightedUndirectedGraph) WeightedEdgeBetween(x, y graph.Node) graph.WeightedEdge {
	return g.WeightedEdge(x, y)
}

// WeightedLines returns the lines from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *WeightedUndirectedGraph) WeightedLines(u, v graph.Node) []graph.WeightedLine {
	return g.WeightedLinesBetween(u, v)
}

// WeightedLinesBetween returns the lines between nodes x and y.
func (g *WeightedUndirectedGraph) WeightedLinesBetween(x, y graph.Node) []graph.WeightedLine {
	edge := g.lines[x.ID()][y.ID()]
	if len(edge) == 0 {
		return nil
	}
	var lines []graph.WeightedLine
	seen := make(map[int64]struct{})
	for _, l := range edge {
		lid := l.ID()
		if _, ok := seen[lid]; ok {
			continue
		}
		seen[lid] = struct{}{}
		lines = append(lines, l)
	}
	return lines
}

// Weight returns the weight for the lines between x and y summarised by the receiver's
// EdgeWeightFunc. Weight returns true if an edge exists between x and y, false otherwise.
func (g *WeightedUndirectedGraph) Weight(x, y graph.Node) (w float64, ok bool) {
	lines := g.WeightedLines(x, y)
	return WeightedEdge{Lines: lines, WeightFunc: g.EdgeWeightFunc}.Weight(), len(lines) != 0
}

// Degree returns the degree of n in g.
func (g *WeightedUndirectedGraph) Degree(n graph.Node) int {
	if _, ok := g.nodes[n.ID()]; !ok {
		return 0
	}
	var deg int
	for _, e := range g.lines[n.ID()] {
		deg += len(e)
	}
	return deg
}
