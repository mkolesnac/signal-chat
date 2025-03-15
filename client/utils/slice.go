package utils

type Predicate[T any] func(T) bool

func Filter[T any](slice []T, predicate Predicate[T]) []T {
	var filtered []T
	for _, value := range slice {
		if predicate(value) {
			filtered = append(filtered, value)
		}
	}

	return filtered
}
