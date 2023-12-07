package graph

import (
	"github.com/qualidafial/gomodblame/internal/multimap"
	"github.com/qualidafial/gomodblame/internal/set"
)

type Graph[T comparable] struct {
	nodes   set.Set[T]
	fromTos multimap.Multimap[T, T]
	toFroms multimap.Multimap[T, T]
}

func New[T comparable]() *Graph[T] {
	return &Graph[T]{
		nodes:   set.New[T](),
		fromTos: multimap.New[T, T](),
		toFroms: multimap.New[T, T](),
	}
}

func (g *Graph[T]) NodeCount() int {
	return len(g.nodes)
}

func (g *Graph[T]) EdgeCount() int {
	return g.fromTos.Size()
}

func (g *Graph[T]) AnyNode() (T, bool) {
	for node := range g.nodes {
		return node, true
	}
	var zero T
	return zero, false
}

func (g *Graph[T]) Add(from, to T) {
	if !g.ContainsEdge(from, to) {
		g.nodes.Add(from)
		g.fromTos.Add(from, to)
		g.toFroms.Add(to, from)
	}
}

func (g *Graph[T]) Remove(from, to T) {
	if g.ContainsEdge(from, to) {
		g.fromTos.Remove(from, to)
		g.toFroms.Remove(to, from)
		if !g.HasEdgesFrom(from) && !g.HasEdgesTo(from) {
			g.nodes.Remove(from)
		}
		if !g.HasEdgesFrom(to) && !g.HasEdgesTo(to) {
			g.nodes.Remove(to)
		}
	}
}

func (g *Graph[T]) ContainsNode(node T) bool {
	return g.nodes.Contains(node)
}

func (g *Graph[T]) ContainsEdge(from, to T) bool {
	return g.fromTos.Contains(from, to)
}

func (g *Graph[T]) All(yield func(from, to T) bool) bool {
	return g.fromTos.All(yield)
}

func (g *Graph[T]) HasEdgesFrom(from T) bool {
	return len(g.fromTos[from]) > 0
}

func (g *Graph[T]) HasEdgesTo(to T) bool {
	return len(g.toFroms[to]) > 0
}

func (g *Graph[T]) EdgesFrom(from T) []T {
	return g.fromTos[from].Slice()
}

func (g *Graph[T]) EdgesTo(to T) []T {
	return g.toFroms[to].Slice()
}

func (g *Graph[T]) Map(fn func(T) T) *Graph[T] {
	mapped := New[T]()

	for from, tos := range g.fromTos {
		newFrom := fn(from)
		for to := range tos {
			newTo := fn(to)
			mapped.Add(newFrom, newTo)
		}
	}

	return mapped
}

func (g *Graph[T]) SubgraphFrom(f func(from T) bool) *Graph[T] {
	subgraph := New[T]()

	visited := set.Set[T]{}

	for from := range g.fromTos {
		if f(from) {
			addSubgraphFrom(subgraph, g, from, visited)
		}
	}

	return subgraph
}

func addSubgraphFrom[T comparable](dst, src *Graph[T], from T, visited set.Set[T]) {
	if visited.Contains(from) {
		return
	}
	visited.Add(from)

	for to := range src.fromTos[from] {
		dst.Add(from, to)
		addSubgraphFrom(dst, src, to, visited)
	}
}

func (g *Graph[T]) SubgraphUntil(f func(to T) bool) *Graph[T] {
	subgraph := New[T]()

	visited := set.Set[T]{}

	for _, node := range g.RootNodes() {
		_ = addSubgraphFromUntil(subgraph, g, node, f, visited)
	}

	return subgraph
}

func addSubgraphFromUntil[T comparable](dst, src *Graph[T], from T, until func(to T) bool, visited set.Set[T]) bool {
	if dst.ContainsNode(from) {
		return true
	}

	if visited.Contains(from) {
		return false
	}
	visited.Add(from)

	var found bool

	for to := range src.fromTos[from] {
		if until(to) || addSubgraphFromUntil(dst, src, to, until, visited) {
			dst.Add(from, to)
			found = true
		}
	}

	return found
}

func (g *Graph[T]) SubgraphTo(f func(to T) bool) *Graph[T] {
	return g.Inverse().SubgraphFrom(f).Inverse()
}

func (g *Graph[T]) FindRootNode() (T, bool) {
	return g.FindFrom(func(from T) bool {
		return !g.HasEdgesTo(from)
	})
}

func (g *Graph[T]) RootNodes() []T {
	var roots []T

	for node := range g.nodes {
		if !g.HasEdgesTo(node) {
			roots = append(roots, node)
		}
	}

	return roots
}

func (g *Graph[T]) FindLeafNode() (T, bool) {
	return g.FindTo(func(to T) bool {
		return !g.HasEdgesFrom(to)
	})
}

func (g *Graph[T]) LeafNodes() []T {
	var leaves []T

	for node := range g.nodes {
		if !g.HasEdgesFrom(node) {
			leaves = append(leaves, node)
		}
	}

	return leaves
}

func (g *Graph[T]) FindFrom(f func(module T) bool) (T, bool) {
	for from := range g.fromTos {
		if f(from) {
			return from, true
		}
	}
	var zero T
	return zero, false
}

func (g *Graph[T]) FindTo(f func(module T) bool) (T, bool) {
	for to := range g.toFroms {
		if f(to) {
			return to, true
		}
	}
	var zero T
	return zero, false
}

func (g *Graph[T]) Inverse() *Graph[T] {
	return &Graph[T]{
		nodes:   g.nodes,
		fromTos: g.toFroms,
		toFroms: g.fromTos,
	}
}
