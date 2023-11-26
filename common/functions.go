package common

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"
)

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

func StructToMap(input interface{}) (map[string]interface{}, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(encoded, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func Base64Decode(input string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func GetBoolFromContext(ctx context.Context, key any) bool {
	value := ctx.Value(key)
	if value == nil {
		return false
	}
	if val, ok := value.(bool); ok {
		return val
	}
	return false
}

func SetBoolToContext(ctx context.Context, key any, value string) context.Context {
	if value == "true" {
		return context.WithValue(ctx, key, true)
	}
	return context.WithValue(ctx, key, false)
}

func GetIntFromContext(ctx context.Context, key any) int {
	value := ctx.Value(key)
	if value == nil {
		return 0
	}
	if val, ok := value.(int); ok {
		return val
	}
	return 0
}
