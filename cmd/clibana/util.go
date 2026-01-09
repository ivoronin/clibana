package main

import (
	"fmt"
	"strings"
)

func mustAssertType[T any](obj interface{}) T { //nolint:ireturn
	var zero T

	if val, ok := obj.(T); ok {
		return val
	}

	panic(fmt.Sprintf("failed to assert type: %v to %T", obj, zero))
}

func getNestedField(source map[string]interface{}, field string) (string, bool) {
	var value interface{} = source

	fieldParts := strings.Split(field, ".")
	for _, part := range fieldParts {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			value = nestedMap[part]
		} else {
			return "", false
		}
	}

	if strValue, ok := value.(string); ok {
		return strValue, true
	}

	return "", false
}
