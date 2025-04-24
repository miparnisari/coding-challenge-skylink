package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGraph(t *testing.T) {
	g, err := NewGraph("../testdata/input.txt")
	require.NoError(t, err)

	t.Run("traits", func(t *testing.T) {
		require.True(t, g.Traits().IsWeighted)
		require.True(t, g.Traits().IsDirected)
		require.False(t, g.Traits().PreventCycles)
	})

	t.Run("edges", func(t *testing.T) {
		numEdges, err := g.Size()
		require.NoError(t, err)
		require.Equal(t, 55*2, numEdges)
	})

	t.Run("vertices", func(t *testing.T) {
		numVertices, err := g.Order()
		require.NoError(t, err)
		require.Equal(t, 29, numVertices)
	})

	t.Run("start_node", func(t *testing.T) {
		require.Equal(t, "A0", g.start)
	})

	t.Run("end_nodes", func(t *testing.T) {
		expected := []string{"A24", "A15", "A21", "A10", "A28"}
		require.ElementsMatch(t, expected, g.ends)
	})

	t.Run("capacities", func(t *testing.T) {
		edgeId := NewEdgeId("A22", "A13")
		require.Equal(t, 19, g.capacitiesOfEdges[edgeId])
	})

	t.Run("capacities", func(t *testing.T) {
		edgeId := NewEdgeId("A22", "A13")
		require.Equal(t, 0, g.flowsOfEdges[edgeId])
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

}
