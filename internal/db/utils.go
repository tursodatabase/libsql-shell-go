package db

import "net/url"

func IsUrl(path string) bool {
	url, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	return url.Scheme != ""
}

func IsValidTursoUrl(path string) bool {
	url, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	return url.Scheme == "libsql" || url.Scheme == "wss" || url.Scheme == "ws"
}
