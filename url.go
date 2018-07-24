package main

import (
	"net/url"
	"regexp"
)

var SubUrlPattern = regexp.MustCompile(`\A(.+)://([^\/]+)(.*)\z`)

func urlParse(s string) (*url.URL, error) {
	r, err := url.Parse(s)
	if err == nil {
		return r, nil
	}
	parts := SubUrlPattern.FindStringSubmatch(s)
	if len(parts) < 4 {
		return nil, err
	}
	return &url.URL{
		Scheme: parts[1],
		Host:   parts[2],
		Path:   parts[3],
	}, nil
}
