package storage

import (
	"fmt"
	"io"
)

// cf+http://container_name

type (
	RackSpaceProvider byte
)

func (rsp RackSpaceProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (rsp RackSpaceProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (rsp RackSpaceProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
