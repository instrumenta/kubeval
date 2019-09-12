package kubeval

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

func getObject(body map[string]interface{}, key string) (map[string]interface{}, error) {
	value, found := body[key]
	if !found {
		return nil, fmt.Errorf("Missing '%s' key", key)
	}
	if value == nil {
		return nil, fmt.Errorf("Missing '%s' value", key)
	}
	typedValue, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected object value for key '%s'", key)
	}
	return typedValue, nil
}

func getStringAt(body map[string]interface{}, path []string) (string, error) {
	obj := body
	visited := []string{}
	var last interface{} = body
	for _, key := range path {
		visited = append(visited, key)

		typed, ok := last.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("Expected object at key '%s'", strings.Join(visited, "."))
		}
		obj = typed

		value, found := obj[key]
		if !found {
			return "", fmt.Errorf("Missing '%s' key", strings.Join(visited, "."))
		}
		last = value
	}
	typed, ok := last.(string)
	if !ok {
		return "", fmt.Errorf("Expected string value for key '%s'", strings.Join(visited, "."))
	}
	return typed, nil
}

func getString(body map[string]interface{}, key string) (string, error) {
	value, found := body[key]
	if !found {
		return "", fmt.Errorf("Missing '%s' key", key)
	}
	if value == nil {
		return "", fmt.Errorf("Missing '%s' value", key)
	}
	typedValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Expected string value for key '%s'", key)
	}
	return typedValue, nil
}

// detectLineBreak returns the relevant platform specific line ending
func detectLineBreak(haystack []byte) string {
	windowsLineEnding := bytes.Contains(haystack, []byte("\r\n"))
	if windowsLineEnding && runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// in is a method which tests whether the `key` is in the set
func in(set []string, key string) bool {
	for _, k := range set {
		if k == key {
			return true
		}
	}
	return false
}
