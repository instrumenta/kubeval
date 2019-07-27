package kubeval

import (
	"bytes"
	"fmt"
	"runtime"
)

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
