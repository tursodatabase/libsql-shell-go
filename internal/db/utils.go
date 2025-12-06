package db

import (
	"net/url"
	"strings"
	"unicode"
)

func IsUrl(uri string) bool {
	url, err := url.ParseRequestURI(uri)
	if err != nil {
		return false
	}
	return url.Scheme != ""
}

func IsValidSqldUrl(uri string) (bool, string) {
	url, err := url.ParseRequestURI(uri)
	if err != nil {
		return false, ""
	}
	// go-libsql supports libsql://, http://, https:// schemes (no websocket)
	return url.Scheme == "libsql" || url.Scheme == "http" || url.Scheme == "https", url.Scheme
}

func EscapeSingleQuotes(value string) string {
	return strings.Replace(value, "'", "''", -1)
}

func startsWithNumber(name string) bool {
	firstChar := rune(name[0])
	return unicode.IsNumber(firstChar)
}

func NeedsEscaping(name string) bool {
	if len(name) == 0 {
		return true
	}
	if startsWithNumber(name) {
		return true
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) && char != rune('_') {
			return true
		}
	}
	return false
}
