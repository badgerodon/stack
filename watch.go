package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/badgerodon/stack/sync"
)

func Watch(endpoint string) {
	serviceManager.Start()

	syncer, err := sync.Parse(endpoint)
	if err != nil {
		panic(err)
	}

	watcher := syncer.Watch()
	for range watcher.C {
		log.Printf("[watch] detected change\n")
		nm := filepath.Join(os.TempDir(), "badgerodon-stack-config.yaml")
		err := syncer.Sync(nm)
		if err != nil {
			log.Printf("[watch] error syncing: %v\n", err)
			time.Sleep(time.Minute)
			continue
		}
		f, err := os.Open(nm)
		if err != nil {
			log.Printf("[watch] error opening config: %v\n", err)
			time.Sleep(time.Minute)
			continue
		}
		cfg, err := ParseConfig(f)
		f.Close()
		if err != nil {
			log.Printf("[watch] error parsing config: %v\n", err)
			time.Sleep(time.Minute)
			continue
		}
		err = install(cfg)
		if err != nil {
			log.Printf("[watch] error installing: %v\n", err)
			continue
		}
	}
}
