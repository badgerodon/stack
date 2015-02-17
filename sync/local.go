package sync

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/fsnotify.v1"
)

type (
	LocalSource struct {
		src string
	}
)

func (ls LocalSource) Watch() *Watcher {
	fw, err := fsnotify.NewWatcher()
	// TODO: fallback to polling
	if err != nil {
		log.Printf("[LocalSource] error creating file system watcher: %v\n", err)
		return nil
	}

	err = fw.Add(ls.src)
	if err != nil {
		log.Printf("[LocalSource] error creating file system watcher: %v\n", err)
		return nil
	}
	return NewWatcher(func(done <-chan struct{}, change chan<- struct{}) {
		defer fw.Close()

		changed := false

		for {
			if changed {
				select {
				case <-done:
					return
				case change <- struct{}{}:
					changed = false
				}
			} else {
				select {
				case <-done:
					return
				case evt := <-fw.Events:
					log.Println(evt)
					changed = true
				}
			}
		}
	})
}

func (ls LocalSource) Sync(dst string) error {
	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()

	sf, err := os.Open(ls.src)
	if err != nil {
		return err
	}
	defer sf.Close()

	_, err = io.Copy(df, sf)
	if err != nil {
		return err
	}
	return nil
}

func (ls LocalSource) Ext() string {
	return filepath.Ext(ls.src)
}
