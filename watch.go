package main

import (
	"log"
	"time"

	"github.com/badgerodon/stack/storage"
	"github.com/badgerodon/stack/sync"
	"github.com/cenkalti/backoff"
)

func watch(src string) error {
	loc, err := storage.ParseLocation(src)
	if err != nil {
		return err
	}

	watcher, err := sync.Watch(loc)
	if err != nil {
		return err
	}
	defer watcher.Stop()

	// TODO: a better approach here would be to use a channel to retry on,
	//       then if you jacked up the config, it would pick up the change
	//       in the middle of all the retries. As it stands now it would take a
	//       minute to fix itself.
	eb := backoff.NewExponentialBackOff()
	eb.MaxElapsedTime = time.Minute

	for range watcher.C {
		log.Println("[watch] new version")
		backoff.Retry(func() error {
			err := apply(src)
			if err != nil {
				log.Printf("[watch] error installing: %v\n", err)
			}
			return err
		}, eb)
	}

	return nil
}
