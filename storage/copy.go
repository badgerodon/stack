package storage

import (
	"fmt"
	"io"
)

// copy://user[:password]@copy.com/some_dir

type (
	CopyCloudStorageProvider byte
)

func (ccsp CopyCloudStorageProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ccsp CopyCloudStorageProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (ccsp CopyCloudStorageProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
