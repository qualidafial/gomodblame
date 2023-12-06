package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/qualidafial/gomodblame/internal/graph"
	"github.com/qualidafial/gomodblame/internal/multimap"
)

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [-from <module>] [-to <module>]\n", os.Args[0])
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("gomodblame: ")

	var from, to, outFile string

	flag.StringVar(&from, "from", "", "only include modules depended on by this module")
	flag.StringVar(&to, "to", "", "only include modules that depend on this module")
	flag.StringVar(&outFile, "o", "", "write output to this file instead of stdout")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 0 {
		usage()
		os.Exit(1)
	}

	log.Println("Reading dependency graph...")
	graph, err := ReadDependencyGraph()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Dependency graph contains %d edges", graph.Size())

	if from != "" {
		log.Printf("Filtering to modules depended on by %q", from)
		graph = graph.SubgraphFrom(func(module string) bool {
			return strings.Contains(module, from)
		})
		log.Printf("Subgraph contains %d edges", graph.Size())
	}
	if to != "" {
		log.Printf("Filtering to modules that depend on %q", to)
		graph = graph.SubgraphTo(func(module string) bool {
			return strings.Contains(module, to)
		})
		log.Printf("Subgraph contains %d edges", graph.Size())
	}

	log.Println("Organizing graph nodes and edges...")

	// Mermaid renders better when nodes are presented in an ideal order.
	// Iteratively find root modules (those that no other modules depend on),
	// output their edges to other nodes, and remove them from the graph.
	// Repeat until the graph is empty.
	var nodes []string

	// Organize edges to be output immediately after the node declaration.
	// This ensures Mermaid encounters the node before its edges, and biases
	// the layout to flow in one direction.
	edgesByTo := multimap.Multimap[string, string]{}

	for graph.Size() > 0 {
		module, ok := graph.FindFromRoot()
		if !ok {
			log.Printf("No root modules left, choosing one at random. %d edges remaining", graph.Size())
			module, ok = graph.FindFrom(func(module string) bool {
				return true
			})
		} else {
			nodes = append(nodes, module)
		}

		for _, dependency := range graph.EdgesFrom(module) {
			edgesByTo.Add(dependency, module)
			graph.Remove(module, dependency)
			if !graph.HasEdgesTo(dependency) && !graph.HasEdgesFrom(dependency) {
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

	var out io.Writer = os.Stdout
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		out = f
	}

	_, err = fmt.Fprintln(out, "graph LR;")
	if err != nil {
		log.Fatal(err)
	}
	for _, node := range nodes {
		_, err = fmt.Fprintf(out, "    %s[\"%s\"];\n", nodeName(node), node)
		if err != nil {
			log.Fatal(err)
		}
		for from := range edgesByTo[node] {
			_, err = fmt.Fprintf(out, "    %s --> %s;\n", nodeName(from), nodeName(node))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func ReadDependencyGraph() (*graph.Graph[string], error) {
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

	graph := graph.New[string]()

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
