package db

import "net/url"

func IsUrl(path string) bool {
	url, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	return url.Scheme != ""
}

func IsValidSqldUrl(path string) (bool, string) {
	url, err := url.ParseRequestURI(path)
	if err != nil {
		return false, ""
	}
	return url.Scheme == "libsql" || url.Scheme == "wss" || url.Scheme == "ws" || url.Scheme == "http" || url.Scheme == "https", url.Scheme
}
