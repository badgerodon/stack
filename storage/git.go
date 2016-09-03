package storage

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/minio/blake2b-simd"
	"github.com/pkg/errors"

	"gopkg.in/src-d/go-git.v3"
)

type gitStorage struct{}
type gitInfo struct {
	url    string
	branch string
	path   string
}

func (info gitInfo) String() string {
	return fmt.Sprintf("gitInfo(url=%s branch=%s path=%s)",
		info.url, info.branch, info.path)
}

// Git retrieves a file via `git archive`
var Git gitStorage

func (s gitStorage) info(loc Location) (gitInfo, error) {
	// git://user@server/owner/repo/path
	user := loc["user"]
	if user == "" {
		user = "git"
	}

	host := loc.Host()
	parts := strings.SplitN(loc.Path(), "/", 4)
	if len(parts) < 4 {
		return gitInfo{}, errors.New("invalid git path, expected /owner/repo/path")
	}
	owner := parts[1]
	repo := parts[2]
	path := parts[3]

	branch := loc.Query().Get("branch")
	if branch == "" {
		branch = "master"
	}
	branch = "refs/heads/" + branch

	return gitInfo{
		url:    fmt.Sprintf("https://%s@%s/%s/%s.git", user, host, owner, repo),
		branch: branch,
		path:   path,
	}, nil
}

func (s gitStorage) Get(loc Location) (io.ReadCloser, error) {
	info, err := s.info(loc)
	if err != nil {
		return nil, errors.Wrap(err, "invalid git location")
	}

	r, err := git.NewRepository(info.url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get git repository for: %v", info)
	}

	err = r.Pull("origin", info.branch)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pull from git (%v)", info)
	}

	h, err := r.Head("origin")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hash from git")
	}

	c, err := r.Commit(h)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get commit `%v` from git", h)
	}

	f, err := c.File(info.path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get file `%s` from git", info.path)
	}

	content, err := f.Contents()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get file contents from git")
	}

	return nopCloser{strings.NewReader(content)}, nil
}

func (s gitStorage) Version(loc Location, previous string) (string, error) {
	// just hash the file contents
	contents, err := s.Get(loc)
	if err != nil {
		return "", err
	}
	defer contents.Close()

	h := blake2b.New512()
	io.Copy(h, contents)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs), nil
}
