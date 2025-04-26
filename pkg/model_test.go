package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGraph(t *testing.T) {
	g, err := NewGraph("../testdata/input.txt")
	require.NoError(t, err)

	t.Run("edges", func(t *testing.T) {
		numEdges := g.NumEdges()
		require.Equal(t, 55, numEdges)
	})

	t.Run("vertices", func(t *testing.T) {
		numVertices := g.NumVertices()
		require.Equal(t, 29, numVertices)
	})

	t.Run("start_node", func(t *testing.T) {
		require.Equal(t, "A0", g.Source.Id)
	})

	t.Run("end_nodes", func(t *testing.T) {
		expected := []string{"A24", "A15", "A21", "A10", "A28"}
		actual := make([]string, len(g.Sinks))
		for i, sink := range g.Sinks {
			actual[i] = sink.Id
		}
		require.ElementsMatch(t, expected, actual)
	})
}

func TestMaxFlow(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		g, err := NewGraph("../testdata/basic.txt")
		require.NoError(t, err)

		maxFlow, err := g.MaxFlowToMultipleSinks()
		require.NoError(t, err)
		require.Equal(t, 15, maxFlow)
	})

	t.Run("complex", func(t *testing.T) {
		g, err := NewGraph("../testdata/complex.txt")
		require.NoError(t, err)

		maxFlow, err := g.MaxFlowToMultipleSinks()
		require.NoError(t, err)
		require.Equal(t, 15, maxFlow)
	})

	t.Run("cycle", func(t *testing.T) {
		g, err := NewGraph("../testdata/cycle.txt")
		require.NoError(t, err)

		maxFlow, err := g.MaxFlowToMultipleSinks()
		require.NoError(t, err)
		require.Equal(t, 5, maxFlow)
	})

	t.Run("duplicate_edge_not_allowed", func(t *testing.T) {
		_, err := NewGraph("../testdata/dup_edge.txt")
		require.ErrorIs(t, err, ErrEdgeExists)

	})

}
