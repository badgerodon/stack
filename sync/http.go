package sync

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type (
	HTTPSource struct {
		url string
	}
)

func (hs HTTPSource) Watch() *Watcher {
	return NewWatcher(func(done <-chan struct{}, change chan<- struct{}) {
		return
	})
}

func (hs HTTPSource) Sync(dst string) error {
	u, err := url.Parse(hs.url)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	if u.User != nil {
		pw, _ := u.User.Password()
		req.SetBasicAuth(u.User.Username(), pw)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	f, err := os.Open(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}
	return nil
}

func (hs HTTPSource) Ext() string {
	return filepath.Ext(strings.Split(hs.url, "?")[0])
}
