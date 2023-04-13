package db

import "net/url"

func IsHttpUrl(path string) bool {
	url, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	return url.Scheme == "http" || url.Scheme == "https"
}
