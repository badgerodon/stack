package storage

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type (
	HTTPProvider struct{}
)

var HTTP = &HTTPProvider{}

func init() {
	Register("http", HTTP)
	Register("https", HTTP)
}

func (hp *HTTPProvider) req(loc Location) *http.Request {
	u := url.URL{
		Scheme: "http",
	}
	if s, ok := loc["scheme"]; ok {
		u.Scheme = s
	}
	if h, ok := loc["host"]; ok {
		u.Host = h
	}
	if p, ok := loc["path"]; ok {
		u.Path = p
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	if user, ok := loc["user"]; ok {
		pw := loc["password"]
		req.SetBasicAuth(user, pw)
	}
	return req
}

func (hp *HTTPProvider) Delete(loc Location) error {
	panic("not implemented")
}

func (hp *HTTPProvider) Get(loc Location) (io.ReadCloser, error) {
	req := hp.req(loc)
	req.Method = "GET"
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 == 2 {
		return res.Body, nil
	} else {
		res.Body.Close()
		return nil, fmt.Errorf("bad status: %v", res.Status)
	}
}

func (hp *HTTPProvider) Put(loc Location, rdr io.Reader) error {
	panic("not implemented")
}

func (hp *HTTPProvider) List(loc Location) ([]string, error) {
	panic("not implemented")
}

func (hp *HTTPProvider) Version(loc Location, previous string) (string, error) {
	req := hp.req(loc)
	req.Method = "HEAD"
	req.Header.Set("If-None-Match", previous)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode == 302 {
		return previous, nil
	}
	if res.StatusCode/100 == 2 {
		return res.Header.Get("ETag"), nil
	}
	return "", fmt.Errorf("bad status: %v", res.Status)
}
