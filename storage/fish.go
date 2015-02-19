package storage

import (
	"fmt"
	"io"
)

// fish://user[:password]@other.host[:port]/[relative|/absolute]_path

type (
	FishProvider byte
)

func (fp FishProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fp FishProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (fp FishProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
