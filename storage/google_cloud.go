package storage

import (
	"fmt"
	"io"
)

// gs://bucket[/prefix]

type (
	GoogleCloudStorageProvider byte
)

func (gcsp GoogleCloudStorageProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (gcsp GoogleCloudStorageProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (gcsp GoogleCloudStorageProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
