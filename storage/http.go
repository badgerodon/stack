package storage

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type httpStorage struct{}

// HTTP stores data over http
var HTTP httpStorage

func init() {
	Register("http", HTTP)
	Register("https", HTTP)
}

func (s httpStorage) req(loc Location) *http.Request {
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
	for n, v := range loc {
		if strings.HasPrefix(n, "Headers.") {
			req.Header.Set(n[len("Headers."):], v)
		}
	}
	if user, ok := loc["user"]; ok {
		pw := loc["password"]
		req.SetBasicAuth(user, pw)
	}
	return req
}

func (s httpStorage) Get(loc Location) (io.ReadCloser, error) {
	req := s.req(loc)
	req.Method = "GET"
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode/100 == 2 {
		return res.Body, nil
	}

	res.Body.Close()
	return nil, fmt.Errorf("bad status (%v): %v", req, res.Status)
}

func (s httpStorage) Version(loc Location, previous string) (string, error) {
	req := s.req(loc)
	req.Method = "HEAD"
	req.Header.Set("If-None-Match", previous)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode == 302 || res.StatusCode == 304 {
		return previous, nil
	}

	if res.StatusCode/100 == 2 {
		return res.Header.Get("ETag"), nil
	}

	return "", fmt.Errorf("bad status (%v): %v", req, res.Status)
}
