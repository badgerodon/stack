package storage

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

// [file://][relative|/absolute]/local/path

type (
	LocalProvider byte
	localref      struct {
		path string
	}
)

var Local LocalProvider

func init() {
	Register("file", Local)
	Register("local", Local)
}

func (lp LocalProvider) Delete(location Location) error {
	return os.Remove(location.Path())
}

func (lp LocalProvider) Get(location Location) (io.ReadCloser, error) {
	return os.Open(location.Path())
}

func (lp LocalProvider) Put(location Location, rdr io.Reader) error {
	f, err := os.Create(location.Path())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, rdr)
	if err != nil {
		return err
	}
	return nil
}

func (lp LocalProvider) List(loc Location) ([]string, error) {
	root := loc.Path()
	if root == "" {
		root = "./"
	}
	files := []string{}
	err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if p == root {
			return nil
		}
		files = append(files, p)
		if fi.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	return files, err
}

func (lp LocalProvider) Version(loc Location, previous string) (string, error) {
	fi, err := os.Stat(loc.Path())
	if err != nil {
		return "", err
	}
	return fi.ModTime().Format(time.RFC3339Nano), nil
}
