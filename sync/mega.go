package sync

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/t3rm1n4l/go-mega"
)

type MegaSource struct {
	email, password, path string
}

var MegaPollingInterval time.Duration

func init() {
	interval, _ := strconv.Atoi(os.Getenv("POLLING_INTERVAL"))
	if interval <= 0 {
		interval = 15
	}
	MegaPollingInterval = time.Second * time.Duration(interval)
}

func NewMegaSourceFromString(str string) *MegaSource {
	email := os.Getenv("MEGA_EMAIL")
	password := os.Getenv("MEGA_PASSWORD")
	path := str
	prefixSz := len("mega://")
	if len(path) > prefixSz {
		path = path[prefixSz:]
	}
	if strings.Contains(path, "@") {
		idx := strings.LastIndex(path, "@")
		email = path[:idx]
		path = path[idx+1:]
		if strings.Contains(email, ":") {
			idx = strings.LastIndex(email, ":")
			password = email[idx+1:]
			email = email[:idx]
		}
	}
	if strings.HasPrefix(path, "mega.co.nz/") {
		path = path[10:]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return &MegaSource{email, password, path}
}

func (ms *MegaSource) Watch() *Watcher {
	return NewWatcher(func(done <-chan struct{}, change chan<- struct{}) {
		ticker := time.NewTicker(MegaPollingInterval)
		defer ticker.Stop()

		changed := false
		var lastTimeStamp time.Time

		for {
			// if we've changed, signal the watcher
			if changed {
				select {
				case <-done:
					return
				// this gives us a chance to cleanup if no one ever listens
				//   after 15 seconds we'll just loop back around and contine waiting
				case <-ticker.C:
				case change <- struct{}{}:
					changed = false
				}
			} else {
				select {
				case <-done:
					return
				case <-ticker.C:
					t, err := ms.timestamp()
					if err == nil {
						if t != lastTimeStamp {
							lastTimeStamp = t
							changed = true
						}
					} else {
						log.Printf("[mega] error getting timestamp: %v", err)
					}
				}
			}
		}
	})
}

func (ms *MegaSource) timestamp() (time.Time, error) {
	log.Println("CHECKING")
	client := mega.New()
	err := client.Login(ms.email, ms.password)
	if err != nil {
		return time.Time{}, fmt.Errorf("error logging in: %v", err)
	}
	nodes, err := client.FS.PathLookup(client.FS.GetRoot(), strings.Split(ms.path, "/"))
	if err != nil {
		return time.Time{}, err
	}
	if len(nodes) == 0 {
		return time.Time{}, fmt.Errorf("File not found")
	}
	return nodes[len(nodes)-1].GetTimeStamp(), nil
}

func (ms *MegaSource) Sync(dst string) error {
	client := mega.New()
	err := client.Login(ms.email, ms.password)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	nodes, err := client.FS.PathLookup(client.FS.GetRoot(), strings.Split(ms.path, "/"))
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return fmt.Errorf("File not found")
	}
	err = client.DownloadFile(nodes[len(nodes)-1], dst, nil)
	if err != nil {
		return fmt.Errorf("error downloading file: %v", err)
	}
	return nil
}

func (ms MegaSource) Ext() string {
	return path.Ext(ms.path)
}
