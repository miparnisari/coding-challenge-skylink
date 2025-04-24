package pkg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/dominikbraun/graph"
)

var (
	EdgeRegex      = regexp.MustCompile(`TRANSMISSION: NODE (?P<s>.*) RELAYS (?P<e>.*) UNDER QUOTA (?P<q>.*)`)
	StartNodeRegex = regexp.MustCompile(`ALERT: PRIMARY NODE IS (?P<s>.*)`)
	EndNodesRegex  = regexp.MustCompile(`FINAL ARRIVAL POINTS ARE (?P<e>.*)`)
)

type CityGraph struct {
	graph.Graph[string, string]
	start             string
	ends              []string
	capacitiesOfEdges map[string]int
	flowsOfEdges      map[string]int
}

func NewEdgeId(a, b string) string {
	return a + "-" + b
}

func ReverseEdgeId(id string) string {
	split := strings.Split(id, "-")
	if len(split) != 2 {
		panic("invalid id format")
	}
	return split[1] + "-" + split[0]
}

func NewGraph(filename string) (*CityGraph, error) {
	cityGraph := &CityGraph{
		ends:              make([]string, 0),
		capacitiesOfEdges: make(map[string]int),
		flowsOfEdges:      make(map[string]int),
	}
	gr := graph.New(graph.StringHash, graph.Directed(), graph.Weighted())
	cityGraph.Graph = gr

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open graph file: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("could not close graph file: %v", err)
		}
	}(f)

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat graph file: %w", err)
	}

	size := fi.Size()
	if size <= 0 || size != int64(int(size)) {
		return nil, fmt.Errorf("invalid graph file size: %d", size)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("could not call mmap: %w", err)
	}

	defer func() {
		if err := syscall.Munmap(data); err != nil {
			log.Printf("could not call munmap: %v", err)
		}
	}()

	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		err := parseLine(line, cityGraph)
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("could not scan read: %w", err)
	}

	return cityGraph, nil
}

type pathh struct {
	edgeId           string
	residualCapacity int
}

func (cg *CityGraph) GetPath(start, end string, path *[]pathh) (*[]pathh, error) {
	if start == end {
		return path, nil
	}

	adjMap, err := cg.AdjacencyMap()
	if err != nil {
		return nil, err
	}
	edgesMap, ok := adjMap[start]
	if !ok {
		return nil, fmt.Errorf("could not find outgoing edges for node %s", start)
	}

	for _, edge := range edgesMap {
		edgeId := NewEdgeId(edge.Source, edge.Target)
		residualCapacity := cg.capacitiesOfEdges[edgeId] - cg.flowsOfEdges[edgeId]
		neww := pathh{edgeId: edgeId, residualCapacity: residualCapacity}
		if residualCapacity > 0 && !slices.Contains(*path, neww) {
			*path = append(*path, neww)
			result, err := cg.GetPath(edge.Target, end, path)
			if err != nil {
				return nil, err
			}
			if result != nil {
				return result, nil
			}
		}
	}
	return nil, nil
}

func (cg *CityGraph) MaxFlowToMultipleSinks() (int, error) {
	if cg.start == "" {
		return 0, fmt.Errorf("no start node found")
	}
	if len(cg.ends) == 0 {
		return 0, fmt.Errorf("no end node(s) found")
	}

	// build artificial sink and then clean it up
	_ = cg.AddVertex("artificial")
	for _, end := range cg.ends {
		_ = cg.AddEdge(end, "artificial")
		edgeId := NewEdgeId(end, "artificial")
		cg.capacitiesOfEdges[edgeId] = math.MaxInt
		defer func() {
			cg.capacitiesOfEdges[edgeId] = 0
		}()
	}
	defer func() {
		_ = cg.RemoveVertex("artificial")
	}()

	emptyPath := make([]pathh, 0)
	path, err := cg.GetPath(cg.start, "artificial", &emptyPath)
	if err != nil {
		return -1, err
	}
	for path != nil {
		flow := math.MaxInt
		for _, pathElement := range *path {
			flow = min(flow, pathElement.residualCapacity)
		}
		for _, pathElement := range *path {
			cg.flowsOfEdges[pathElement.edgeId] += flow
			cg.flowsOfEdges[ReverseEdgeId(pathElement.edgeId)] -= flow
		}
		emptyPath = make([]pathh, 0)
		path, err = cg.GetPath(cg.start, "artificial", &emptyPath)
		if err != nil {
			return -1, err
		}
	}
	//  return sum(edge.flow for edge in self.network[source.name])
	adjMap, err := cg.AdjacencyMap()
	if err != nil {
		return -1, err
	}
	edgesMap, ok := adjMap[cg.start]
	if !ok {
		return -1, fmt.Errorf("could not find outgoing edges for node %s", cg.start)
	}

	summ := 0
	for _, edge := range edgesMap {
		edgeId := NewEdgeId(edge.Source, edge.Target)
		summ += cg.flowsOfEdges[edgeId]
	}

	return summ, nil
}

func parseLine(line string, cityGraph *CityGraph) error {
	edgeMatch := EdgeRegex.FindStringSubmatch(line)
	if edgeMatch != nil {
		start := edgeMatch[EdgeRegex.SubexpIndex("s")]
		end := edgeMatch[EdgeRegex.SubexpIndex("e")]
		weightStr := edgeMatch[EdgeRegex.SubexpIndex("q")]
		weightInt, err := strconv.Atoi(weightStr)
		if err != nil {
			return fmt.Errorf("could not convert weight to int: %w", err)
		}

		_, err = cityGraph.Vertex(start)
		if err != nil && errors.Is(err, graph.ErrVertexNotFound) {
			_ = cityGraph.AddVertex(start)
		}
		_, err = cityGraph.Vertex(end)
		if err != nil && errors.Is(err, graph.ErrVertexNotFound) {
			_ = cityGraph.AddVertex(end)
		}

		_, err = cityGraph.Edge(start, end)
		if err != nil && !errors.Is(err, graph.ErrEdgeAlreadyExists) {
			_ = cityGraph.AddEdge(start, end, graph.EdgeWeight(weightInt))
			edgeId := NewEdgeId(start, end)
			cityGraph.capacitiesOfEdges[edgeId] = weightInt
			// return edge
			_ = cityGraph.AddEdge(end, start, graph.EdgeWeight(0))
			edgeId = NewEdgeId(end, start)
			cityGraph.capacitiesOfEdges[edgeId] = 0
		}
	} else {
		startMatch := StartNodeRegex.FindStringSubmatch(line)
		if startMatch != nil {
			start := startMatch[StartNodeRegex.SubexpIndex("s")]
			_, err := cityGraph.Vertex(start)
			if err != nil && errors.Is(err, graph.ErrVertexNotFound) {
				_ = cityGraph.AddVertex(start)
			}
			cityGraph.start = start
		} else {
			endsMatch := EndNodesRegex.FindStringSubmatch(line)
			if endsMatch != nil {
				ends := endsMatch[EndNodesRegex.SubexpIndex("e")]
				endsSplit := strings.Split(ends, ",")
				for _, end := range endsSplit {
					cleanEnd := strings.TrimSpace(end)
					_, err := cityGraph.Vertex(cleanEnd)
					if err != nil && errors.Is(err, graph.ErrVertexNotFound) {
						_ = cityGraph.AddVertex(cleanEnd)
					}
					// check that the end is reachable.
					adjMap, err := cityGraph.PredecessorMap()
					if err != nil {
						return err
					}
					adjacents, ok := adjMap[cleanEnd]
					if !ok || len(adjacents) == 0 {
						return fmt.Errorf("end %v is not reachable", cleanEnd)
					}
					cityGraph.ends = append(cityGraph.ends, cleanEnd)
				}
			} else {
				return fmt.Errorf("could not parse line: %s", line)
			}
		}
	}
	return nil
}
