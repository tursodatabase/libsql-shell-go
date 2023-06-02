package db

import "net/url"

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
	return url.Scheme == "libsql" || url.Scheme == "wss" || url.Scheme == "ws" || url.Scheme == "http" || url.Scheme == "https", url.Scheme
}
