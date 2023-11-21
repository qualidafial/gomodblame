package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <dependency> [dependency [dependency...]]\n", os.Args[0])
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("gomodblame: ")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	log.Println("Reading dependency graph...")
	graph, err := ReadDependencyGraph()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Read dependency graph with %d edges", graph.Size())

	log.Println("Filtering dependency graph...")
	graph = graph.SubgraphDependingOn(func(module string) bool {
		for _, arg := range flag.Args() {
			if strings.Contains(module, arg) {
				log.Printf("Including %q", module)
				return true
			}
		}
		return false
	})
	log.Printf("Subgraph contains %d edges", graph.Size())

	log.Println("Organizing graph nodes and edges...")

	// Mermaid renders better when nodes are presented in an ideal order.
	// Iteratively find root modules (those that no other modules depend on),
	// output their edges to other nodes, and remove them from the graph.
	// Repeat until the graph is empty.
	var nodes []string

	// Organize edges to be output immediately after the node declaration.
	// This ensures Mermaid encounters the node before its edges, and biases
	// the layout to flow in one direction.
	edgesByDestinationNode := Multimap[string, string]{}

	for graph.Size() > 0 {
		module, ok := graph.FindRootModule()
		if !ok {
			log.Printf("No root modules left, choosing one at random. %d edges remaining", graph.Size())
			module, ok = graph.FindModule(func(module string) bool {
				return true
			})
		} else {
			nodes = append(nodes, module)
		}

		for _, dependency := range graph.Dependencies(module) {
			edgesByDestinationNode.Add(dependency, module)
			graph.Remove(module, dependency)
			if !graph.HasDependants(dependency) && !graph.HasDependencies(dependency) {
				// leaf node, output it immediately
				nodes = append(nodes, dependency)
			}
		}
	}

	nodeNames := map[string]string{}
	nodeName := func(module string) string {
		if name, ok := nodeNames[module]; ok {
			return name
		}
		name := fmt.Sprintf("n%d", len(nodeNames))
		nodeNames[module] = name

		return name
	}

	fmt.Println("graph LR;")
	for _, node := range nodes {
		fmt.Printf("    %s[\"%s\"];\n", nodeName(node), node)
		for from := range edgesByDestinationNode[node] {
			fmt.Printf("    %s --> %s;\n", nodeName(from), nodeName(node))
		}
	}
}

type DependencyGraph struct {
	dependencies Multimap[string, string]
	dependants   Multimap[string, string]
}

func (g *DependencyGraph) Size() int {
	var count int
	for _, dependencies := range g.dependencies {
		count += len(dependencies)
	}
	return count
}

func (g *DependencyGraph) Add(module, dependency string) {
	if g.dependencies == nil {
		g.dependencies = Multimap[string, string]{}
	}
	if g.dependants == nil {
		g.dependants = Multimap[string, string]{}
	}
	g.dependencies.Add(module, dependency)
	g.dependants.Add(dependency, module)
}

func (g *DependencyGraph) Remove(module, dependency string) {
	if g.dependencies == nil {
		g.dependencies = Multimap[string, string]{}
	}
	if g.dependants == nil {
		g.dependants = Multimap[string, string]{}
	}
	g.dependencies.Remove(module, dependency)
	g.dependants.Remove(dependency, module)
}

func (g *DependencyGraph) Contains(module, dependency string) bool {
	return g.dependencies.Contains(module, dependency)
}

func (g *DependencyGraph) DependsOn(module, dependency string) bool {
	// Direct dependency
	if g.Contains(module, dependency) {
		return true
	}

	// Indirect dependency
	moduleDependencies := g.dependencies[module]
	for moduleDependency := range moduleDependencies {
		if g.DependsOn(moduleDependency, dependency) {
			return true
		}
	}

	// No dependency
	return false
}

func (g *DependencyGraph) All(yield func(module, dependency string) bool) bool {
	for module, dependencies := range g.dependencies {
		for dependency := range dependencies {
			if !yield(module, dependency) {
				return false
			}
		}
	}
	return true
}

func (g *DependencyGraph) HasDependencies(module string) bool {
	return g.dependencies.ContainsKey(module)
}

func (g *DependencyGraph) HasDependants(module string) bool {
	return g.dependants.ContainsKey(module)
}

func (g *DependencyGraph) Dependencies(module string) []string {
	return g.dependencies[module].Slice()
}

func (g *DependencyGraph) Dependants(module string) []string {
	return g.dependants[module].Slice()
}

func (g *DependencyGraph) SubgraphDependingOn(f func(module string) bool) *DependencyGraph {
	subgraph := &DependencyGraph{}

	visited := Set[string]{}

	for module := range g.dependants {
		if f(module) {
			addDependants(subgraph, g, module, visited)
		}
	}

	return subgraph
}

func addDependants(dst, src *DependencyGraph, dependency string, visited Set[string]) {
	if visited.Contains(dependency) {
		return
	}
	visited.Add(dependency)

	for module := range src.dependants[dependency] {
		dst.Add(module, dependency)
		addDependants(dst, src, module, visited)
	}
}

func (g *DependencyGraph) FindRootModule() (string, bool) {
	return g.FindModule(func(module string) bool {
		return !g.HasDependants(module)
	})
}

func (g *DependencyGraph) FindModule(f func(module string) bool) (string, bool) {
	for module := range g.dependencies {
		if f(module) {
			return module, true
		}
	}
	return "", false
}

type Multimap[K, V comparable] map[K]Set[V]

func (m Multimap[K, V]) Add(key K, value V) {
	set, ok := m[key]
	if !ok {
		set = make(Set[V])
		m[key] = set
	}

	set.Add(value)
}

func (m Multimap[K, V]) Remove(key K, value V) {
	set, ok := m[key]
	if !ok {
		return
	}

	set.Remove(value)
	if len(set) == 0 {
		delete(m, key)
	}
}

func (m Multimap[K, V]) Contains(key K, value V) bool {
	_, ok := m[key][value]
	return ok
}

func (m Multimap[K, V]) ContainsKey(key K) bool {
	return len(m[key]) > 0
}

func (m Multimap[K, V]) Inverse() Multimap[V, K] {
	inv := Multimap[V, K]{}
	for key, values := range m {
		for value := range values {
			inv.Add(value, key)
		}
	}
	return inv
}

func (m Multimap[K, V]) Clone() Multimap[K, V] {
	clone := Multimap[K, V]{}
	for key, values := range m {
		for value := range values {
			clone.Add(key, value)
		}
	}
	return clone
}

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

func (s Set[T]) Remove(value T) {
	delete(s, value)
}

func (s Set[T]) Contains(value T) bool {
	_, ok := s[value]
	return ok
}

func (s Set[T]) Slice() []T {
	out := make([]T, 0, len(s))
	for value := range s {
		out = append(out, value)
	}
	return out
}

func ReadDependencyGraph() (*DependencyGraph, error) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("looking for go: %v", err)
	}

	cmd := exec.Command(goPath, "mod", "graph")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("starting go mod graph: %v", err)
	}

	graph := &DependencyGraph{}

	s := bufio.NewScanner(out)
	for s.Scan() {
		line := s.Text()
		module, dependency, ok := strings.Cut(line, " ")
		if !ok {
			return nil, fmt.Errorf("parsing line %q", line)
		}
		graph.Add(module, dependency)
	}
	if s.Err() != nil {
		return nil, fmt.Errorf("scanning go mod graph: %v", s.Err())
	}

	return graph, cmd.Wait()
}
