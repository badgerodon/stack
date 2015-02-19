package storage

import (
	"fmt"
	"io"
)

// ftp[s]://user[:password]@other.host[:port]/some_dir

type (
	FTPProvider byte
)

func (fp FTPProvider) Get(rawurl string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fp FTPProvider) Put(rawurl string, rdr io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (fp FTPProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
