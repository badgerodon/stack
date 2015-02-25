package sync

import (
	"log"
	"sync"
	"time"

	"github.com/badgerodon/stack/storage"
)

type Watcher struct {
	C       <-chan struct{}
	done    chan struct{}
	stopped bool
	mu      sync.Mutex
}

func Watch(loc storage.Location) (*Watcher, error) {
	provider, err := storage.GetProvider(loc)
	if err != nil {
		return nil, err
	}
	return newWatcher(func(done <-chan struct{}, change chan<- struct{}) {
		previous := ""
		changed := true
		previous, _ = provider.Version(loc, previous)
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()
		for {
			if changed {
				select {
				case change <- struct{}{}:
					changed = false
				case <-done:
					return
				}
			} else {
				select {
				case <-ticker.C:
					next, err := provider.Version(loc, previous)
					if err != nil {
						log.Println("[watcher] error getting version:", err)
						time.Sleep(time.Minute)
						continue
					}
					if previous != next {
						changed = true
						previous = next
					}
				case <-done:
					return
				}
			}
		}
	}), nil
}

func newWatcher(f func(done <-chan struct{}, change chan<- struct{})) *Watcher {
	done := make(chan struct{})
	change := make(chan struct{})
	go f(done, change)
	return &Watcher{
		C:       change,
		done:    done,
		stopped: false,
	}
}

func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.stopped {
		w.stopped = true
		close(w.done)
	}
}
