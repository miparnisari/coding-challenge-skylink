package pkg

import (
	"container/list"
	"fmt"
	"math"
	"slices"
	"strings"
)

var (
	ErrEdgeNotFound = fmt.Errorf("edge not found")
	ErrEdgeExists   = fmt.Errorf("edge already exists")
)

type Vertex struct {
	Id string
}

type Edge struct {
	Id    string
	Start Vertex
	End   Vertex
	Cap   int
	Flow  int
}

func NewEdge(from, to Vertex, cap int) *Edge {
	return &Edge{
		Id:    fmt.Sprintf("%s-%s", from.Id, to.Id),
		Start: from,
		End:   to,
		Cap:   cap,
	}
}

type Network struct {
	Source *Vertex
	Sinks  []*Vertex

	Vertices  map[string]Vertex
	EdgesFrom map[string][]*Edge
}

func (cg *Network) NumVertices() int {
	return len(cg.Vertices)
}

func (cg *Network) NumEdges() int {
	total := 0
	for _, edges := range cg.EdgesFrom {
		total += len(edges)
	}
	return total
}

func (cg *Network) GetOrAddVertex(id string) Vertex {
	vertex, ok := cg.Vertices[id]
	if ok {
		return vertex
	}
	vertex = Vertex{Id: id}
	cg.Vertices[id] = vertex
	return vertex
}

func (cg *Network) RemoveVertex(s string) {
	delete(cg.Vertices, s)
	delete(cg.EdgesFrom, s)

	for from, edges := range cg.EdgesFrom {
		indicesToRemove := make([]int, 0)
		for i, edge := range edges {
			if edge.End.Id == s {
				indicesToRemove = append(indicesToRemove, i)
			}
		}
		cg.EdgesFrom[from] = RemoveIndices(edges, indicesToRemove)
	}

	return
}

func (cg *Network) Edge(start, end string) (*Edge, error) {
	edgesMap, ok := cg.EdgesFrom[start]
	if ok {
		for _, edge := range edgesMap {
			if edge.End.Id == end {
				return edge, nil
			}
		}
	}

	return nil, ErrEdgeNotFound
}

func (cg *Network) AddEdge(start, end string, weight int) error {
	_, ok := cg.EdgesFrom[start]
	if !ok {
		cg.EdgesFrom[start] = make([]*Edge, 0)
	}
	for _, edge := range cg.EdgesFrom[start] {
		if edge.End.Id == end {
			return ErrEdgeExists
		}
	}

	e := NewEdge(cg.Vertices[start], cg.Vertices[end], weight)
	cg.EdgesFrom[start] = append(cg.EdgesFrom[start], e)
	return nil
}

// BFS returns true if we arrived to the end. It updates "parent".
func (cg *Network) BFS(start, end Vertex, parent map[Vertex]Vertex) bool {
	visited := make(map[string]bool)
	queue := list.New()
	queue.PushBack(start)
	visited[start.Id] = true

	for queue.Len() > 0 {
		cur := queue.Remove(queue.Front())
		curVertex := cur.(Vertex)

		for _, edge := range cg.EdgesFrom[curVertex.Id] {
			_, vis := visited[edge.End.Id]
			if !vis && edge.Cap > 0 {
				visited[edge.End.Id] = true
				queue.PushBack(edge.End)
				if parent != nil {
					parent[edge.End] = curVertex
				}
			}
		}
	}

	return visited[end.Id]
}

func (cg *Network) MaxFlowToMultipleSinks() (int, error) {
	if cg.Source == nil {
		return 0, fmt.Errorf("no start node specified")
	}
	if len(cg.Sinks) == 0 {
		return 0, fmt.Errorf("no end node(s) specified")
	}

	// build artificial sink and then clean it up
	sink := cg.GetOrAddVertex("artificial")
	defer cg.RemoveVertex("artificial")
	for _, sink := range cg.Sinks {
		cg.AddEdge(sink.Id, "artificial", math.MaxInt)
	}

	summ := 0
	parent := make(map[Vertex]Vertex)

	for cg.BFS(*cg.Source, sink, parent) {
		pathFlow := math.MaxInt
		s := sink
		for s != *cg.Source {
			edge, err := cg.Edge(parent[s].Id, s.Id)
			if err == nil {
				pathFlow = min(pathFlow, edge.Cap)
			} else {
				pathFlow = min(pathFlow, 0)
			}
			s = parent[s]
		}
		summ += pathFlow
		v := sink
		for v != *cg.Source {
			u := parent[v]
			edge, err := cg.Edge(u.Id, v.Id)
			if err == nil {
				edge.Cap -= pathFlow
			}
			reverse, err := cg.Edge(v.Id, u.Id)
			if err == nil {
				reverse.Cap += pathFlow
			}
			v = parent[v]
		}

		path := make([]string, 0)
		v = sink
		for v != *cg.Source {
			path = append(path, v.Id)
			v = parent[v]
		}
		path = append(path, cg.Source.Id)
		slices.Reverse(path)
		fmt.Println("path", strings.Join(path, "->"), "flow:", pathFlow)
	}
	return summ, nil
}
