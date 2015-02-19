package storage

import (
	"fmt"
	"io"
)

// dpbx:///some_dir
// dropbox:///some_dir

type (
	DropBoxProvider byte
)

func (dbp DropBoxProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (dbp DropBoxProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (dbp DropBoxProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
