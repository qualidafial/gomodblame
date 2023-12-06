package graph

import (
	"github.com/qualidafial/gomodblame/internal/histogram"
	"github.com/qualidafial/gomodblame/internal/multimap"
	"github.com/qualidafial/gomodblame/internal/set"
)

type Graph[T comparable] struct {
	from, to histogram.Histogram[T]
	edges    multimap.Multimap[T, T]
}

func New[T comparable]() *Graph[T] {
	return &Graph[T]{
		from:  histogram.New[T](),
		to:    histogram.New[T](),
		edges: multimap.New[T, T](),
	}
}

func (g *Graph[T]) Size() int {
	return g.edges.Size()
}

func (g *Graph[T]) Add(from, to T) {
	if !g.Contains(from, to) {
		g.from.Add(from)
		g.to.Add(to)
		g.edges.Add(from, to)
	}
}

func (g *Graph[T]) Remove(from, to T) {
	if g.Contains(from, to) {
		g.from.Remove(from)
		g.to.Remove(to)
		g.edges.Remove(from, to)
	}
}

func (g *Graph[T]) Contains(from, to T) bool {
	return g.edges.Contains(from, to)
}

func (g *Graph[T]) DependsOn(from, to T) bool {
	// Direct reference
	if g.Contains(from, to) {
		return true
	}

	if !g.from.Contains(from) || !g.to.Contains(to) {
		return false
	}

	// Indirect reference
	for node := range g.edges[from] {
		if g.DependsOn(node, to) {
			return true
		}
	}

	// No reference
	return false
}

func (g *Graph[T]) All(yield func(from, to T) bool) bool {
	return g.edges.All(yield)
}

func (g *Graph[T]) HasEdgesFrom(from T) bool {
	return g.from[from] > 0
}

func (g *Graph[T]) HasEdgesTo(to T) bool {
	return g.to[to] > 0
}

func (g *Graph[T]) EdgesFrom(from T) []T {
	return g.edges[from].Slice()
}

func (g *Graph[T]) Dependants(to T) []T {
	return g.edges.Inverse()[to].Slice()
}

func (g *Graph[T]) SubgraphFrom(f func(from T) bool) *Graph[T] {
	subgraph := New[T]()

	visited := set.Set[T]{}

	for from := range g.edges {
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

	for to := range src.edges[from] {
		dst.Add(from, to)
		addSubgraphFrom(dst, src, to, visited)
	}
}

func (g *Graph[T]) SubgraphTo(f func(to T) bool) *Graph[T] {
	return g.Inverse().SubgraphFrom(f).Inverse()
}

func (g *Graph[T]) FindFromRoot() (T, bool) {
	return g.FindFrom(func(from T) bool {
		return !g.HasEdgesTo(from)
	})
}

func (g *Graph[T]) FindFrom(f func(module T) bool) (T, bool) {
	for from := range g.edges {
		if f(from) {
			return from, true
		}
	}
	var zero T
	return zero, false
}

func (g *Graph[T]) Inverse() *Graph[T] {
	inverse := New[T]()
	g.All(func(from, to T) bool {
		inverse.Add(to, from)
		return true
	})
	return inverse
}
