package storage

import (
	"io"
	"net/http"
)

type (
	HTTPProvider struct{}
)

var HTTP = &HTTPProvider{}

func init() {
	//Register("http", HTTP)
	//Register("https", HTTP)
}

func (hp *HTTPProvider) Delete(rawurl string) error {
	panic("not implemented")
}

func (hp *HTTPProvider) Get(rawurl string) (io.ReadCloser, error) {
	panic("not implemented")
}

func (hp *HTTPProvider) Put(rawurl string, rdr io.Reader) error {
	panic("not implemented")
}

func (hp *HTTPProvider) List(rawurl string) ([]string, error) {
	panic("not implemented")
}

func (hp *HTTPProvider) Version(rawurl, previous string) (string, error) {
	req, err := http.NewRequest("HEAD", rawurl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("If-None-Match", previous)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode == 302 {
		return previous, nil
	}
	return res.Header.Get("ETag"), nil
}
