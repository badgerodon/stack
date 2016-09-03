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
	dup["host"] = "bitbucket.org"
	dup["query"] = ""
	path := []string{""}
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
	// /owner/repo/raw/branch/path
	path = append(path[:3], append([]string{"raw", branch}, path[3:]...)...)
	dup["path"] = strings.Join(path, "/")
	return dup
}

func (s bitbucket) Get(loc Location) (io.ReadCloser, error) {
	return HTTP.Get(s.Build(loc))
}

func (s bitbucket) Version(loc Location, previous string) (string, error) {
	return HTTP.Version(s.Build(loc), previous)
}
