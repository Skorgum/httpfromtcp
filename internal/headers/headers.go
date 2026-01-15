package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Override(key, value string) {
	h[key] = value
}

func NewHeaders() Headers {
	return make(map[string]string)
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func isTokenChar(c byte) bool {
	if c >= 'A' && c <= 'Z' ||
		c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' {
		return true
	}

	for _, ch := range tokenChars {
		if ch == c {
			return true
		}
	}
	return false
}

func validTokens(data []byte) bool {
	for _, c := range data {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	rawKey := string(parts[0])
	if rawKey != strings.TrimRight(rawKey, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", rawKey)
	}
	key := string(bytes.TrimSpace(parts[0]))
	if !validTokens([]byte(key)) {
		return 0, false, fmt.Errorf("invalid header token found: %s", key)
	}
	key = strings.ToLower(key)

	value := string(bytes.TrimSpace(parts[1]))

	h.Set(key, value)

	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if existing, ok := h[key]; ok {
		h[key] = existing + ", " + value
	} else {
		h[key] = value
	}
}

func (h Headers) Get(key string) string {
	key = strings.ToLower(key)
	if value, ok := h[key]; ok {
		return value
	}
	return ""
}
