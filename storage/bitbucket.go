package storage

import (
	"io"
	"strings"
)

type bitbucket struct{}

// BitBucket stores data on BitBucket
var BitBucket bitbucket

func (s bitbucket) Build(loc Location) Location {
	dup := make(Location)
	for k, v := range loc {
		dup[k] = v
	}
	dup["scheme"] = "https"
	dup["host"] = "api.bitbucket.org"
	dup["query"] = ""
	path := []string{"", "1.0", "repositories"}
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
	branch := loc.Query().Get("branch")
	if branch == "" {
		branch = "master"
	}
	// inject 'raw/branch'
	// https://api.bitbucket.org/1.0/repositories/{accountname}/{repo_slug}/raw/{revision}/{path}
	path = append(path[:5], append([]string{"raw", branch}, path[5:]...)...)
	dup["path"] = strings.Join(path, "/")
	return dup
}

func (s bitbucket) Get(loc Location) (io.ReadCloser, error) {
	return HTTP.Get(s.Build(loc))
}

func (s bitbucket) Version(loc Location, previous string) (string, error) {
	return HTTP.Version(s.Build(loc), previous)
}
