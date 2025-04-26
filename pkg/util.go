package pkg

func RemoveIndices[T any](input []T, indicesToRemove []int) []T {
	if len(indicesToRemove) == 0 {
		return append([]T(nil), input...)
	}

	indexMap := make(map[int]struct{})
	for _, idx := range indicesToRemove {
		if idx >= 0 && idx < len(input) {
			indexMap[idx] = struct{}{}
		}
	}

	result := make([]T, 0, len(input)-len(indexMap))
	for i, val := range input {
		if _, found := indexMap[i]; !found {
			result = append(result, val)
		}
	}
	return result
}
