package main

import "fmt"

func install(cfg *Config) error {
	pl := NewPortLock(49001)
	pl.Lock()
	defer pl.Unlock()

	prevCfg := ReadConfig()
	defer SaveConfig(cfg)

	err := ProcessSources(prevCfg, cfg)
	if err != nil {
		return fmt.Errorf("error processing sources: %v", err)
	}

	err = ProcessApplications(prevCfg, cfg)
	if err != nil {
		return fmt.Errorf("error processing applications: %v", err)
	}

	return nil
}
