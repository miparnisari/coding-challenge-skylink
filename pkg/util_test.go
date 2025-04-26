package pkg

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveIndices(t *testing.T) {
	tests := []struct {
		name            string
		input           []int
		indicesToRemove []int
		expectedResult  []int
	}{
		{
			name:            "remove middle element",
			input:           []int{1, 2, 3, 4, 5},
			indicesToRemove: []int{2},
			expectedResult:  []int{1, 2, 4, 5},
		},
		{
			name:            "remove first and last element",
			input:           []int{1, 2, 3, 4, 5},
			indicesToRemove: []int{0, 4},
			expectedResult:  []int{2, 3, 4},
		},
		{
			name:            "remove all elements",
			input:           []int{1, 2, 3},
			indicesToRemove: []int{0, 1, 2},
			expectedResult:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveIndices(tt.input, tt.indicesToRemove)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}
