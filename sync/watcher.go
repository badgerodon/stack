package sync

import "sync"

type Watcher struct {
	C       <-chan struct{}
	done    chan struct{}
	stopped bool
	mu      sync.Mutex
}

func NewWatcher(f func(done <-chan struct{}, change chan<- struct{})) *Watcher {
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
