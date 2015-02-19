package storage

import (
	"fmt"
	"io"
)

// (gdocs|gdrive)://user[:password]@other.host/some_dir

type (
	GoogleDocsProvider byte
)

func (gdp GoogleDocsProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (gdp GoogleDocsProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (gdp GoogleDocsProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
