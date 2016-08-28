package sync

import (
	"log"
	"sync"
	"time"

	"github.com/badgerodon/stack/storage"
)

// PollInterval is the time in between looking for new versions
var PollInterval = time.Second * 15

// A Watcher watches for changes
type Watcher struct {
	C       <-chan struct{}
	done    chan struct{}
	stopped bool
	mu      sync.Mutex
}

// Watch looks for changes at the given location
func Watch(loc storage.Location) (*Watcher, error) {
	provider, err := storage.GetProvider(loc)
	if err != nil {
		return nil, err
	}
	return newWatcher(func(done <-chan struct{}, change chan<- struct{}) {
		previous := ""
		changed := true
		previous, _ = provider.Version(loc, previous)
		ticker := time.NewTicker(PollInterval)
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

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.stopped {
		w.stopped = true
		close(w.done)
	}
}
