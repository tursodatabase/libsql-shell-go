package db

import "net/url"

func IsUrl(path string) bool {
	_, err := url.ParseRequestURI(path)
	return err == nil
}

func IsValidTursoUrl(path string) bool {
	url, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	return url.Scheme == "http" || url.Scheme == "https" || url.Scheme == "libsql"
}
