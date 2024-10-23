package device_manager

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateIndexForPartitionedDevice(t *testing.T) {
	var tests []struct {
		description      string
		originalIndex    int
		partitionIndex   int
		partitionsLength int
		expected         int
	}

	// first element: partitionIndex, second element: finalIndex
	boardMatrix := [][]int{
		{0, 1, 2, 3, 4, 5, 6, 7},
		{8, 9, 10, 11, 12, 13, 14, 15},
		{16, 17, 18, 19, 20, 21, 22, 23},
		{24, 25, 26, 27, 28, 29, 30, 31},
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			tests = append(tests, struct {
				description      string
				originalIndex    int
				partitionIndex   int
				partitionsLength int
				expected         int
			}{
				description:      fmt.Sprintf("Original Board %d, Partition %d", i, j),
				originalIndex:    i,
				partitionIndex:   j,
				partitionsLength: 8,
				expected:         boardMatrix[i][j],
			})
		}
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			generatedFinalIndex := generateIndexForPartitionedDevice(tc.originalIndex, tc.partitionIndex, tc.partitionsLength)
			assert.Equal(t, tc.expected, generatedFinalIndex)
		})
	}
}
