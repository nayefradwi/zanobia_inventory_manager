package common

import "time"

func HasAllValues[T any, P string | int](values []T, allValues []T, getKey func(value T) P) bool {
	mapValues := make(map[P]bool)
	for _, value := range allValues {
		key := getKey(value)
		mapValues[key] = true
	}
	for _, value := range values {
		key := getKey(value)
		if !mapValues[key] {
			return false
		}
	}
	return true
}

func GetUtcDateOnlyString() string {
	y, m, d := time.Now().UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

func GetUtcDateOnlyStringFromTime(t time.Time) string {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

func GetValues[K comparable, V any](input map[K]V) []V {
	values := make([]V, 0)
	for _, value := range input {
		values = append(values, value)
	}
	return values
}
