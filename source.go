package main

import (
	"fmt"
	"log"
	"os"
)

func ProcessSources(prevCfg, newCfg *Config) error {
	// remove
	for _, pa := range prevCfg.Applications {
		found := false
		for _, na := range newCfg.Applications {
			if pa.SourceHash() == na.SourceHash() {
				found = true
				break
			}
		}
		if !found {
			log.Println("[sources] remove", pa.DownloadPath())
			os.Remove(pa.DownloadPath())
		}
	}
	// add
	for _, na := range newCfg.Applications {
		found := false
		for _, pa := range prevCfg.Applications {
			if na.SourceHash() == pa.SourceHash() {
				found = true
				break
			}
		}
		if found {
			log.Println("[sources] skip", na.DownloadPath())
		} else {
			log.Println("[sources] add", na.DownloadPath())
			err := na.Source.Sync(na.DownloadPath())
			if err != nil {
				return fmt.Errorf("error syncing: %v", err)
			}
		}
	}
	return nil
}
