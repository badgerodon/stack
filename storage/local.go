package storage

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// [file://][relative|/absolute]/local/path

type (
	LocalProvider byte
	localref      struct {
		path string
	}
)

var Local LocalProvider

func (lp LocalProvider) parse(rawurl string) localref {
	if strings.HasPrefix(rawurl, "file://") {
		rawurl = rawurl[7:]
	}
	return localref{rawurl}
}

func (lp LocalProvider) Delete(rawurl string) error {
	ref := lp.parse(rawurl)
	return os.Remove(ref.path)
}

func (lp LocalProvider) Get(rawurl string) (io.ReadCloser, error) {
	ref := lp.parse(rawurl)
	return os.Open(ref.path)
}

func (lp LocalProvider) Put(rawurl string, rdr io.Reader) error {
	ref := lp.parse(rawurl)
	f, err := os.Create(ref.path)
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

func (lp LocalProvider) List(rawurl string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (lp LocalProvider) Version(rawurl string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
