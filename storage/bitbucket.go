package storage

import (
	"io"
	"strings"
)

type bitbucket struct {
	useHTTP bool
}

// BitBucket stores data on BitBucket
var BitBucket = bitbucket{useHTTP: false}

func (s bitbucket) BuildHTTP(loc Location) Location {
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

func (s bitbucket) BuildGit(loc Location) Location {
	// bitbucket://owner/repo/path
	// =>
	// git://bitbucket.org/owner/repo/path
	dup := make(Location)
	for k, v := range loc {
		dup[k] = v
	}
	dup["scheme"] = "git"
	dup["host"] = "bitbucket.org"
	dup["path"] = "/" + loc.Host() + loc.Path()
	return dup
}

func (s bitbucket) Get(loc Location) (io.ReadCloser, error) {
	if s.useHTTP {
		return HTTP.Get(s.BuildHTTP(loc))
	}
	return Git.Get(s.BuildGit(loc))
}

func (s bitbucket) Version(loc Location, previous string) (string, error) {
	if s.useHTTP {
		return HTTP.Version(s.BuildHTTP(loc), previous)
	}
	return Git.Version(s.BuildGit(loc), previous)
}
