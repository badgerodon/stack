package storage

import (
	"fmt"
	"io"
)

// azure://container_name

type (
	AzureProvider byte
)

func (az AzureProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (az AzureProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (az AzureProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
