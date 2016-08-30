package storage

import (
	"io"
	"strings"
)

type githubStorage struct{}

// GitHub stores data on GitHub
var GitHub githubStorage

func (s githubStorage) build(loc Location) Location {
	dup := make(Location)
	for k, v := range loc {
		dup[k] = v
	}
	dup["scheme"] = "https"
	dup["host"] = "api.github.com"
	path := []string{"", "repos"}
	if loc.Host() != "" {
		path = append(path, loc.Host())
	}
	if loc.Path() != "" {
		p := loc.Path()
		if strings.HasPrefix(p, "/") {
			p = p[1:]
		}
		path = append(path, strings.Split(p, "/")...)
	}
	// inject 'contents' into path
	// /repos/owner/repo/contents/path
	path = append(path[:4], append([]string{"contents"}, path[4:]...)...)
	dup["path"] = strings.Join(path, "/")
	dup["Headers.Accept"] = "application/vnd.github.v3.raw"
	return dup
}

func (s githubStorage) Get(loc Location) (io.ReadCloser, error) {
	return HTTP.Get(s.build(loc))
}

func (s githubStorage) Version(loc Location, previous string) (string, error) {
	return HTTP.Version(s.build(loc), previous)
}
