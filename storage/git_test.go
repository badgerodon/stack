package storage

import "testing"

func TestGit(t *testing.T) {
	loc, _ := ParseLocation("git://example.com/owner/repository/path/to/file.txt")
	actual, _ := Git.info(loc)
	expect := gitInfo{
		url:    "https://git@example.com/owner/repository.git",
		branch: "refs/heads/master",
		path:   "path/to/file.txt",
	}

	if actual != expect {
		t.Errorf("expected `%v` but got `%v`", expect, actual)
	}
}
