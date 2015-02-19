package storage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

type (
	Provider interface {
		Delete(rawurl string) error
		Get(rawurl string) (io.ReadCloser, error)
		Put(rawurl string, rdr io.Reader) error
		List(rawurl string) ([]string, error)
		Version(rawurl string) (string, error)
	}
	Sizer interface {
		Size() (int64, error)
	}
	DeleteOnClose struct {
		*os.File
	}
)

func (doc DeleteOnClose) Close() error {
	doc.File.Close()
	return os.Remove(doc.File.Name())
}

func getSize(rdr io.Reader) (int64, error) {
	if szr, ok := rdr.(Sizer); ok {
		return szr.Size()
	}
	return 0, fmt.Errorf("could not find size implementation")
}

func GetProvider(rawurl string) (Provider, error) {
	if !strings.Contains(rawurl, "://") {
		return Local, nil
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "mega":
		return Mega, nil
	case "file":
		return Local, nil
	default:
		return nil, fmt.Errorf("unknown storage provider: %v", u.Scheme)
	}
}

func Delete(rawurl string) error {
	p, err := GetProvider(rawurl)
	if err != nil {
		return err
	}
	return p.Delete(rawurl)
}

func Get(rawurl string) (io.ReadCloser, error) {
	p, err := GetProvider(rawurl)
	if err != nil {
		return nil, err
	}
	return p.Get(rawurl)
}

func List(rawurl string) ([]string, error) {
	p, err := GetProvider(rawurl)
	if err != nil {
		return nil, err
	}
	return p.List(rawurl)
}

func Put(rawurl string, rdr io.Reader) error {
	p, err := GetProvider(rawurl)
	if err != nil {
		return err
	}
	return p.Put(rawurl, rdr)
}
